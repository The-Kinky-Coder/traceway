package main

import (
	"context"
	"fmt"
	"time"
)

type stepResult struct {
	Step                 int             `json:"step"`
	BatchSize            int             `json:"batchSize"`
	RequestRate          float64         `json:"requestRate"`
	AttemptedItemsPerSec float64         `json:"attemptedItemsPerSec"`
	ActualItemsPerSec    float64         `json:"actualItemsPerSec"`
	Rejected             int64           `json:"rejected"`
	Ingest               latencySnapshot `json:"ingest"`
	Passed               bool            `json:"passed"`
	FailReason           string          `json:"failReason,omitempty"`
}

type phaseResult struct {
	Kind             string       `json:"kind"`
	FixedRequestRate float64      `json:"fixedRequestRate,omitempty"`
	FixedBatchSize   int          `json:"fixedBatchSize,omitempty"`
	Steps            []stepResult `json:"steps"`
	MaxBatchSize     int          `json:"maxBatchSize,omitempty"`
	MaxRequestRate   float64      `json:"maxRequestRate,omitempty"`
}

type finalReport struct {
	Tier                      string           `json:"tier"`
	Mode                      string           `json:"mode"`
	Signal                    string           `json:"signal"`
	Scenario                  string           `json:"scenario"`
	StartedAt                 string           `json:"startedAt"`
	EndedAt                   string           `json:"endedAt"`
	Phase1                    *phaseResult     `json:"phase1,omitempty"`
	Phase2                    *phaseResult     `json:"phase2,omitempty"`
	ReadProbe                 *readProbeResult `json:"readProbe,omitempty"`
	MaxSustainableItemsPerSec float64          `json:"maxSustainableItemsPerSec,omitempty"`
	MaxFillLevelPassed        int64            `json:"maxFillLevelPassed,omitempty"`
}

func (r *finalReport) computeHeadline() {
	if r.Phase2 != nil {
		// Use the achieved throughput of the last passing step. The
		// FixedBatchSize × MaxRequestRate formula is unreliable: at the cliff,
		// workers saturate on HTTP latency and never deliver the target rate,
		// so the formula overstates capacity. ActualItemsPerSec is measured
		// from real OK responses and tells the truth.
		var best float64
		for _, s := range r.Phase2.Steps {
			if s.Passed && s.ActualItemsPerSec > best {
				best = s.ActualItemsPerSec
			}
		}
		r.MaxSustainableItemsPerSec = best
	}
	// Fall back to Phase 1 when Phase 2 produced no passing steps. This
	// happens when Phase 1's heavy final step poisons the SUT and Phase 2
	// can't recover within its drain window — without this fallback the
	// headline would be 0 even though Phase 1 has perfectly valid data.
	if r.MaxSustainableItemsPerSec == 0 && r.Phase1 != nil {
		for _, s := range r.Phase1.Steps {
			if s.Passed && s.ActualItemsPerSec > r.MaxSustainableItemsPerSec {
				r.MaxSustainableItemsPerSec = s.ActualItemsPerSec
			}
		}
	}
	if r.ReadProbe != nil {
		r.MaxFillLevelPassed = r.ReadProbe.MaxFillLevelPassed
	}
}

// phaseCheckpoint is invoked after every step in the throughput phases so the
// caller can persist a partial report — we never want to lose progress when
// the process dies mid-run.
type phaseCheckpoint func(phaseResult)

// runBatchSizeRamp holds requestRate fixed (phase1FixedRate) and grows batch
// size step by step. Stops at the first failing step. Returns a phaseResult
// whose MaxBatchSize is the largest batch that passed. Calls `checkpoint`
// after each step (after both pass and fail) so partial state is durable.
func runBatchSizeRamp(ctx context.Context, cfg config, ing *ingester, ingest *latencyTracker, checkpoint phaseCheckpoint) phaseResult {
	res := phaseResult{
		Kind:             "batch-size-ramp",
		FixedRequestRate: cfg.phase1FixedRate,
	}

	ing.SetRequestRate(cfg.phase1FixedRate)

	for idx, batch := range cfg.phase1BatchSizes {
		if ctx.Err() != nil {
			break
		}
		ing.SetBatchSize(batch)
		s := runOneStep(ctx, cfg, ing, ingest, idx+1, batch, cfg.phase1FixedRate)
		res.Steps = append(res.Steps, s)
		fmt.Fprintf(stderrPrefix(), "phase1 step %d: batch=%d rate=%.1f items/s=%.0f p99=%.0fms err=%.2f%% passed=%t %s\n",
			s.Step, s.BatchSize, s.RequestRate, s.ActualItemsPerSec, s.Ingest.P99, s.Ingest.ErrRate*100, s.Passed, s.FailReason)
		if s.Passed {
			res.MaxBatchSize = batch
		}
		if checkpoint != nil {
			checkpoint(res)
		}
		if !s.Passed {
			break
		}
	}

	return res
}

// runRequestRateRamp holds batchSize fixed at min(phase1.MaxBatchSize, cfg.phase2BatchCap)
// and grows request rate step by step. Calls `checkpoint` after each coarse
// step and after each bisection step so partial state is durable.
func runRequestRateRamp(ctx context.Context, cfg config, ing *ingester, ingest *latencyTracker, phase1 phaseResult, checkpoint phaseCheckpoint) phaseResult {
	batch := phase1.MaxBatchSize
	if batch <= 0 {
		batch = cfg.phase2BatchCap
	}
	if batch > cfg.phase2BatchCap {
		batch = cfg.phase2BatchCap
	}

	res := phaseResult{
		Kind:           "request-rate-ramp",
		FixedBatchSize: batch,
	}
	if batch <= 0 {
		return res
	}

	ing.SetBatchSize(batch)

	var lastPassRate, firstFailRate float64
	stepNo := 0

	for _, rate := range cfg.phase2RequestRates {
		if ctx.Err() != nil {
			break
		}
		stepNo++
		ing.SetRequestRate(rate)
		s := runOneStep(ctx, cfg, ing, ingest, stepNo, batch, rate)
		res.Steps = append(res.Steps, s)
		fmt.Fprintf(stderrPrefix(), "phase2 step %d: batch=%d rate=%.1f items/s=%.0f p99=%.0fms err=%.2f%% passed=%t %s\n",
			s.Step, s.BatchSize, s.RequestRate, s.ActualItemsPerSec, s.Ingest.P99, s.Ingest.ErrRate*100, s.Passed, s.FailReason)
		if s.Passed {
			res.MaxRequestRate = rate
			lastPassRate = rate
		}
		if checkpoint != nil {
			checkpoint(res)
		}
		if !s.Passed {
			firstFailRate = rate
			break
		}
	}

	// Bisection refinement. After the coarse ramp finds an adjacent
	// (passing, failing) pair, halve the gap up to phase2BisectMaxSteps
	// times to pin the real cliff. Skipped when nothing failed (no cliff
	// in the configured range) or nothing passed (cliff below the
	// smallest configured rate — bisecting below it isn't informative).
	if cfg.phase2BisectMaxSteps > 0 && lastPassRate > 0 && firstFailRate > lastPassRate {
		for b := 0; b < cfg.phase2BisectMaxSteps; b++ {
			if ctx.Err() != nil {
				break
			}
			gap := (firstFailRate - lastPassRate) / lastPassRate
			if gap <= cfg.phase2BisectTolerance {
				break
			}
			mid := (lastPassRate + firstFailRate) / 2
			stepNo++
			ing.SetRequestRate(mid)
			s := runOneStep(ctx, cfg, ing, ingest, stepNo, batch, mid)
			res.Steps = append(res.Steps, s)
			fmt.Fprintf(stderrPrefix(), "phase2 bisect %d: batch=%d rate=%.1f items/s=%.0f p99=%.0fms err=%.2f%% passed=%t %s\n",
				s.Step, s.BatchSize, s.RequestRate, s.ActualItemsPerSec, s.Ingest.P99, s.Ingest.ErrRate*100, s.Passed, s.FailReason)
			if s.Passed {
				lastPassRate = mid
				res.MaxRequestRate = mid
			} else {
				firstFailRate = mid
			}
			if checkpoint != nil {
				checkpoint(res)
			}
		}
	}

	return res
}

// runOneStep resizes the worker pool for the new rate, holds the step for
// stepDuration, drains in-flight HTTP for stepDrainSeconds, then snapshots
// latency + item counters. The drain window matters: without it, every step
// boundary cancels ~workerCount in-flight requests mid-flight, inflating the
// recorded error count and depressing the OK count.
func runOneStep(ctx context.Context, cfg config, ing *ingester, ingest *latencyTracker, stepNo, batchSize int, requestRate float64) stepResult {
	ingest.SnapshotAndReset()
	ing.SnapshotAndResetItems()

	ing.Start(ctx)

	stepCtx, cancel := context.WithTimeout(ctx, cfg.stepDuration)
	start := time.Now()
	<-stepCtx.Done()
	cancel()

	// Stop accepting new requests; let in-flight ones complete. elapsed
	// includes the drain window so the rate divisor reflects the full window
	// during which items could have arrived.
	ing.StopAccepting()
	ing.WaitForDrain(cfg.stepDrainSeconds)
	elapsed := time.Since(start)
	ing.Stop()

	snap := ingest.SnapshotAndReset()
	attempted, rejected := ing.SnapshotAndResetItems()

	var attemptedIps, actualIps float64
	if elapsed > 0 {
		attemptedIps = float64(attempted) / elapsed.Seconds()
		actualItems := attempted - rejected
		// Discount failed HTTP requests too — their items never made it in.
		if attempted > 0 {
			httpFailItems := int64(float64(snap.Errors) / float64(snap.OK+snap.Errors) * float64(attempted))
			actualItems -= httpFailItems
		}
		if actualItems < 0 {
			actualItems = 0
		}
		actualIps = float64(actualItems) / elapsed.Seconds()
	}

	passed, reason := evaluateStep(cfg, snap, attempted, rejected, requestRate, elapsed)

	return stepResult{
		Step:                 stepNo,
		BatchSize:            batchSize,
		RequestRate:          requestRate,
		AttemptedItemsPerSec: attemptedIps,
		ActualItemsPerSec:    actualIps,
		Rejected:             rejected,
		Ingest:               snap,
		Passed:               passed,
		FailReason:           reason,
	}
}

// evaluateStep combines three failure criteria:
//  1. HTTP-level error rate + OTLP partial-success rejections > threshold
//     (combined item-error budget).
//  2. Soft cliff: achieved request rate is far below target. This catches
//     the "workers saturated but SUT not yet erroring" state, where latency
//     has cliffed to multiple seconds and only error-rate-based detection
//     would let the ramp keep climbing one more step before noticing.
func evaluateStep(cfg config, snap latencySnapshot, attempted, rejected int64, targetRate float64, elapsed time.Duration) (bool, string) {
	totalReq := snap.OK + snap.Errors
	if totalReq == 0 {
		return false, "no requests completed"
	}
	httpErrRate := float64(snap.Errors) / float64(totalReq)
	var rejectRate float64
	if attempted > 0 {
		rejectRate = float64(rejected) / float64(attempted)
	}
	combined := httpErrRate + rejectRate
	if combined > cfg.ingestErrThreshold {
		return false, fmt.Sprintf("combined error rate %.2f%% (http %.2f%% + rejected %.2f%%) > %.2f%% threshold",
			combined*100, httpErrRate*100, rejectRate*100, cfg.ingestErrThreshold*100)
	}
	if targetRate > 0 && elapsed > 0 && cfg.softCliffRatio > 0 {
		achievedRate := float64(snap.OK) / elapsed.Seconds()
		if achievedRate < targetRate*cfg.softCliffRatio {
			return false, fmt.Sprintf("achieved %.2f req/sec is below %.0f%% of target %.2f req/sec (workers saturated, SUT past cliff)",
				achievedRate, cfg.softCliffRatio*100, targetRate)
		}
	}
	return true, ""
}
