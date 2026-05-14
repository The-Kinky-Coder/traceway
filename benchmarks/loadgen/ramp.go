package main

import (
	"context"
	"fmt"
	"time"
)

type stepResult struct {
	Step       int                        `json:"step"`
	TargetEPS  float64                    `json:"targetEps"`
	ActualEPS  float64                    `json:"actualEps"`
	Ingest     latencySnapshot            `json:"ingest"`
	Reads      map[string]latencySnapshot `json:"reads"`
	Passed     bool                       `json:"passed"`
	FailReason string                     `json:"failReason,omitempty"`
}

type finalReport struct {
	Tier               string       `json:"tier"`
	Mode               string       `json:"mode"`
	StartedAt          string       `json:"startedAt"`
	EndedAt            string       `json:"endedAt"`
	MaxSustainableEPS  float64      `json:"maxSustainableEps"`
	Steps              []stepResult `json:"steps"`
}

func (r *finalReport) computeBreakingPoint() {
	for i := len(r.Steps) - 1; i >= 0; i-- {
		if r.Steps[i].Passed {
			r.MaxSustainableEPS = r.Steps[i].ActualEPS
			return
		}
	}
}

// runRamp drives the EPS ramp: hold a target rate for stepDuration, snapshot
// stats, evaluate the breaking-point criteria, and either advance or stop.
// Doubling cadence (100 -> 200 -> 400 -> ...) gets us to the cliff in <10 steps
// for any realistic SUT, keeping total runtime bounded.
func runRamp(
	ctx context.Context,
	cfg config,
	ingester *ingester,
	ingest *latencyTracker,
	reads map[string]*latencyTracker,
) []stepResult {
	results := []stepResult{}
	target := cfg.initialEPS

	for stepNo := 1; target <= cfg.maxEPS; stepNo++ {
		ingester.SetRate(target)

		// Reset trackers at start of step so we measure only the new rate's
		// steady state, not the ramp-up transient from the previous step.
		ingest.SnapshotAndReset()
		for _, t := range reads {
			t.SnapshotAndReset()
		}

		stepCtx, cancel := context.WithTimeout(ctx, cfg.stepDuration)
		start := time.Now()
		<-stepCtx.Done()
		cancel()
		elapsed := time.Since(start)

		ingestSnap := ingest.SnapshotAndReset()
		readSnaps := map[string]latencySnapshot{}
		for name, t := range reads {
			readSnaps[name] = t.SnapshotAndReset()
		}

		// Both target and actual must be in the same unit (traces/sec).
		// Rate limiter ticks per request (= per frame), so requests/sec × the
		// batch size gives the trace throughput on the wire. OTLP requests
		// carry spans rather than traces; we count them at the same rate
		// because the user-facing question is "events/sec the system swallowed"
		// regardless of which protocol carried them.
		actualEPS := 0.0
		if elapsed > 0 {
			actualEPS = float64(ingestSnap.OK+ingestSnap.Errors) / elapsed.Seconds() * tracesPerFrame
		}

		passed, reason := evaluate(cfg, ingestSnap, readSnaps)

		res := stepResult{
			Step:       stepNo,
			TargetEPS:  target,
			ActualEPS:  actualEPS,
			Ingest:     ingestSnap,
			Reads:      readSnaps,
			Passed:     passed,
			FailReason: reason,
		}
		results = append(results, res)
		fmt.Fprintf(stderrPrefix(), "step %d: target=%.0f actual=%.0f ingest_p99=%.0fms ingest_err=%.2f%% passed=%t %s\n",
			stepNo, target, actualEPS, ingestSnap.P99, ingestSnap.ErrRate*100, passed, reason)

		if !passed || ctx.Err() != nil {
			return results
		}

		target *= 2
	}
	return results
}

func evaluate(cfg config, ingest latencySnapshot, reads map[string]latencySnapshot) (bool, string) {
	if ingest.ErrRate > cfg.ingestErrThreshold {
		return false, fmt.Sprintf("ingest error rate %.2f%% > %.2f%% threshold", ingest.ErrRate*100, cfg.ingestErrThreshold*100)
	}
	for name, s := range reads {
		// If a read endpoint never returned a single OK response during this
		// step the SUT is effectively unavailable for reads; treat that as a
		// hard fail even though P99 reads as 0 from the empty histogram.
		if s.OK == 0 && s.Errors > 0 {
			return false, fmt.Sprintf("read endpoint %q served zero OK responses", name)
		}
		if s.P99 > cfg.readP99ThresholdMs {
			return false, fmt.Sprintf("read endpoint %q p99 = %.0fms > %.0fms threshold", name, s.P99, cfg.readP99ThresholdMs)
		}
	}
	return true, ""
}
