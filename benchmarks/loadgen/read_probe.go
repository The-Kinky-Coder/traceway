package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type readProbeStep struct {
	FillLevelTarget      int64   `json:"fillLevelTarget"`
	RowsIngested         int64   `json:"rowsIngested"`
	IngestSecondsElapsed float64 `json:"ingestSecondsElapsed"`
	ReadLatencyMs        float64 `json:"readLatencyMs"`
	ReadOk               bool    `json:"readOk"`
	Passed               bool    `json:"passed"`
	FailReason           string  `json:"failReason,omitempty"`
}

type readProbeResult struct {
	ReadPath           string          `json:"readPath"`
	ReadThresholdMs    int             `json:"readThresholdMs"`
	SettleSeconds      int             `json:"settleSeconds"`
	FillBatchSize      int             `json:"fillBatchSize"`
	FillRequestRate    float64         `json:"fillRequestRate"`
	Steps              []readProbeStep `json:"steps"`
	MaxFillLevelPassed int64           `json:"maxFillLevelPassed"`
}

// runReadProbe walks the configured fill levels. For each level it ingests
// rows until totalIngested >= target, settles for cfg.settleSeconds, then
// issues one read probe. If the probe times out, errors, or exceeds
// cfg.readThresholdMs, the step fails and the loop stops without ingesting
// further.
// readProbeCheckpoint is invoked after every fill-level step so the caller
// can persist a partial report — we never want to lose progress when the
// process dies mid-run.
type readProbeCheckpoint func(readProbeResult)

func runReadProbe(ctx context.Context, cfg config, ing *ingester, ingestStats *latencyTracker, client *http.Client, checkpoint readProbeCheckpoint) readProbeResult {
	res := readProbeResult{
		ReadPath:        readPathForSignal(cfg.signal),
		ReadThresholdMs: cfg.readThresholdMs,
		SettleSeconds:   int(cfg.settleSeconds.Seconds()),
		FillBatchSize:   cfg.fillBatchSize,
		FillRequestRate: cfg.fillRequestRate,
	}

	ing.SetBatchSize(cfg.fillBatchSize)
	ing.SetRequestRate(cfg.fillRequestRate)

	// Reset the ingester counters before starting so totalIngested reflects
	// only what this scenario sent.
	ing.SnapshotAndResetItems()
	ingestStats.SnapshotAndReset()

	var totalIngested int64

	for _, target := range cfg.fillLevels {
		if ctx.Err() != nil {
			break
		}
		step := readProbeStep{FillLevelTarget: target}

		if totalIngested < target {
			fillStart := time.Now()
			ing.Start(ctx)
			pollFillProgress(ctx, ing, &totalIngested, target)
			ing.Stop()
			step.IngestSecondsElapsed = time.Since(fillStart).Seconds()
		}
		step.RowsIngested = totalIngested

		// Drain whatever bumped between Stop and Snapshot.
		extraAttempted, _ := ing.SnapshotAndResetItems()
		totalIngested += extraAttempted
		step.RowsIngested = totalIngested

		fmt.Fprintf(stderrPrefix(), "read-probe fill=%d rows reached in %.1fs (signal=%s) — settling %ds\n",
			step.RowsIngested, step.IngestSecondsElapsed, cfg.signal, res.SettleSeconds)

		select {
		case <-time.After(cfg.settleSeconds):
		case <-ctx.Done():
		}
		if ctx.Err() != nil {
			res.Steps = append(res.Steps, step)
			if checkpoint != nil {
				checkpoint(res)
			}
			break
		}

		latencyMs, readErr := probeRead(ctx, client, cfg)
		step.ReadLatencyMs = latencyMs
		switch {
		case readErr != nil:
			step.ReadOk = false
			step.Passed = false
			step.FailReason = fmt.Sprintf("read error: %v (latency %.0fms)", readErr, latencyMs)
		case latencyMs > float64(cfg.readThresholdMs):
			step.ReadOk = true
			step.Passed = false
			step.FailReason = fmt.Sprintf("read latency %.0fms > %dms threshold", latencyMs, cfg.readThresholdMs)
		default:
			step.ReadOk = true
			step.Passed = true
			res.MaxFillLevelPassed = target
		}

		fmt.Fprintf(stderrPrefix(), "read-probe target=%d read=%.0fms ok=%t passed=%t %s\n",
			target, latencyMs, step.ReadOk, step.Passed, step.FailReason)
		res.Steps = append(res.Steps, step)
		if checkpoint != nil {
			checkpoint(res)
		}
		if !step.Passed {
			break
		}
	}

	return res
}

// pollFillProgress drains attempted-item counters from the ingester until the
// target is reached or the context cancels. Sub-second polling keeps overshoot
// small (at fill-batch=8192 × 100 req/s the loadgen sends ~800k items/sec, so
// 200ms granularity overshoots by ≤160k items — negligible at 100M fill).
func pollFillProgress(ctx context.Context, ing *ingester, totalIngested *int64, target int64) {
	tick := time.NewTicker(200 * time.Millisecond)
	defer tick.Stop()
	for *totalIngested < target {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			attempted, _ := ing.SnapshotAndResetItems()
			*totalIngested += attempted
		}
	}
}

func readPathForSignal(signal string) string {
	switch signal {
	case "spans":
		return "/api/endpoints/grouped"
	case "metrics":
		return "/api/metrics/application"
	case "logs":
		return "/api/logs"
	}
	return ""
}

// probeRead issues one HTTP read against the signal-specific endpoint and
// returns wall-clock latency in ms plus any error. The request is hard-capped
// at threshold + 1s so a hanging SUT can't deadlock the loop.
func probeRead(ctx context.Context, client *http.Client, cfg config) (float64, error) {
	now := time.Now().UTC()
	fromDate := now.Add(-24 * time.Hour).Format(time.RFC3339)
	toDate := now.Format(time.RFC3339)

	probeCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.readThresholdMs+1000)*time.Millisecond)
	defer cancel()

	var req *http.Request
	var err error
	switch cfg.signal {
	case "spans":
		body, _ := json.Marshal(map[string]any{
			"fromDate":      fromDate,
			"toDate":        toDate,
			"orderBy":       "count",
			"sortDirection": "desc",
			"pagination":    map[string]int{"page": 1, "pageSize": 50},
			"search":        "",
		})
		u, _ := url.Parse(cfg.target + "/api/endpoints/grouped")
		q := u.Query()
		q.Set("projectId", cfg.projectId)
		u.RawQuery = q.Encode()
		req, err = http.NewRequestWithContext(probeCtx, http.MethodPost, u.String(), bytes.NewReader(body))
		if err != nil {
			return 0, err
		}
		req.Header.Set("Content-Type", "application/json")
	case "metrics":
		u, _ := url.Parse(cfg.target + "/api/metrics/application")
		q := u.Query()
		q.Set("projectId", cfg.projectId)
		q.Set("fromDate", fromDate)
		q.Set("toDate", toDate)
		u.RawQuery = q.Encode()
		req, err = http.NewRequestWithContext(probeCtx, http.MethodGet, u.String(), nil)
		if err != nil {
			return 0, err
		}
	case "logs":
		body, _ := json.Marshal(map[string]any{
			"fromDate":   fromDate,
			"toDate":     toDate,
			"orderBy":    "timestamp desc",
			"pagination": map[string]int{"page": 1, "pageSize": 50},
		})
		u, _ := url.Parse(cfg.target + "/api/logs")
		q := u.Query()
		q.Set("projectId", cfg.projectId)
		u.RawQuery = q.Encode()
		req, err = http.NewRequestWithContext(probeCtx, http.MethodPost, u.String(), bytes.NewReader(body))
		if err != nil {
			return 0, err
		}
		req.Header.Set("Content-Type", "application/json")
	default:
		return 0, fmt.Errorf("no read endpoint configured for signal %q", cfg.signal)
	}
	req.Header.Set("Authorization", "Bearer "+cfg.jwt)

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start).Seconds() * 1000
	if err != nil {
		return elapsed, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return elapsed, fmt.Errorf("status %d", resp.StatusCode)
	}
	return elapsed, nil
}
