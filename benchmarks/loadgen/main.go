package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type config struct {
	target             string
	projectToken       string
	jwt                string
	projectId          string
	signal             string
	scenario           string
	duration           time.Duration
	stepDuration       time.Duration
	phase1BatchSizes   []int
	phase2RequestRates []float64
	phase1FixedRate    float64
	phase2BatchCap     int
	ingestErrThreshold        float64
	softCliffRatio            float64
	stepDrainSeconds          time.Duration
	interPhaseCooldownSeconds time.Duration
	phase2BisectMaxSteps      int
	phase2BisectTolerance     float64
	fillLevels         []int64
	readThresholdMs    int
	settleSeconds      time.Duration
	fillBatchSize      int
	fillRequestRate    float64
	reportOut          string
	tier               string
	mode               string
}

func main() {
	var (
		cfg              config
		phase1BatchesStr string
		phase2RatesStr   string
		fillLevelsStr    string
	)

	flag.StringVar(&cfg.target, "target", "", "Base URL of the system under test (e.g. http://10.0.0.2 or http://localhost:8087)")
	flag.StringVar(&cfg.projectToken, "token", "", "Project bearer token for OTLP ingest endpoints")
	flag.StringVar(&cfg.jwt, "jwt", "", "JWT for read endpoints (required when --scenario=read-probe)")
	flag.StringVar(&cfg.projectId, "project-id", "", "Project UUID for read endpoints (required when --scenario=read-probe)")
	flag.StringVar(&cfg.signal, "signal", "", "Which signal to benchmark: spans | metrics | logs (required)")
	flag.StringVar(&cfg.scenario, "scenario", "throughput", "Scenario: throughput (default, two-phase ingest ramp) | read-probe (ingest to fill levels and probe a read)")
	flag.DurationVar(&cfg.duration, "duration", 30*time.Minute, "Total run duration cap")
	flag.DurationVar(&cfg.stepDuration, "step-duration", 2*time.Minute, "Per-step hold time (throughput scenario only)")
	flag.StringVar(&phase1BatchesStr, "phase1-batch-sizes", "256,1024,4096,8192,16384", "Comma-separated batch sizes for Phase 1 (throughput scenario)")
	flag.StringVar(&phase2RatesStr, "phase2-request-rates", "1,5,25,100,400", "Comma-separated request rates for Phase 2 (throughput scenario)")
	flag.Float64Var(&cfg.phase1FixedRate, "phase1-fixed-rate", 5, "Fixed request rate during Phase 1 (req/sec)")
	flag.IntVar(&cfg.phase2BatchCap, "phase2-batch-cap", 16384, "Cap on Phase 2 batch size; Phase 2 uses min(this, Phase 1 winner). Bumped from the OTel collector default of 8192 because pgch SUTs can usefully exceed it.")
	flag.Float64Var(&cfg.ingestErrThreshold, "ingest-err-threshold", 0.05, "Step fails if combined (HTTP error + OTLP rejected) item rate exceeds this")
	flag.Float64Var(&cfg.softCliffRatio, "soft-cliff-ratio", 0.70, "Step fails when achieved req-rate is below this fraction of target — catches saturated-but-not-erroring cliffs. 0 disables.")
	flag.DurationVar(&cfg.stepDrainSeconds, "step-drain-seconds", 10*time.Second, "After step duration expires, wait up to this long for in-flight HTTP requests to complete before hard-canceling (reduces error-count noise at boundaries)")
	flag.DurationVar(&cfg.interPhaseCooldownSeconds, "inter-phase-cooldown-seconds", 30*time.Second, "Pause between Phase 1 and Phase 2 to let the SUT drain queues and recover from Phase 1's final-step load. 0 disables.")
	flag.IntVar(&cfg.phase2BisectMaxSteps, "phase2-bisect-max-steps", 3, "After Phase 2 finds the cliff, run up to this many bisection steps between the last passing and first failing rate to narrow the cliff. 0 disables.")
	flag.Float64Var(&cfg.phase2BisectTolerance, "phase2-bisect-tolerance", 0.20, "Bisection stops when (firstFailRate-lastPassRate)/lastPassRate falls below this fraction (e.g. 0.20 = stop when the cliff is pinned to within 20% of the last passing rate).")
	flag.StringVar(&fillLevelsStr, "fill-levels", "100000,1000000,10000000,100000000", "Comma-separated row counts to fill before probing a read (read-probe scenario)")
	flag.IntVar(&cfg.readThresholdMs, "read-threshold-ms", 5000, "Read latency threshold in ms; step fails if a probe exceeds it (read-probe scenario)")
	flag.DurationVar(&cfg.settleSeconds, "settle-seconds", 10*time.Second, "Wait between finishing ingest and probing the read (read-probe scenario)")
	flag.IntVar(&cfg.fillBatchSize, "fill-batch-size", 8192, "OTLP batch size used during the fill phase (read-probe scenario)")
	flag.Float64Var(&cfg.fillRequestRate, "fill-request-rate", 100, "OTLP request rate (req/sec) during the fill phase (read-probe scenario)")
	flag.StringVar(&cfg.reportOut, "report-out", "", "Path to write JSON results (required)")
	flag.StringVar(&cfg.tier, "tier", "local", "Hardware tier label embedded in output (e.g. ccx13)")
	flag.StringVar(&cfg.mode, "mode", "unknown", "DB mode label embedded in output (sqlite | pgch)")
	flag.Parse()

	if cfg.target == "" || cfg.projectToken == "" || cfg.reportOut == "" || cfg.signal == "" {
		fmt.Fprintln(os.Stderr, "missing required flag: --target, --token, --signal, --report-out")
		flag.Usage()
		os.Exit(2)
	}

	batches, err := parseInts(phase1BatchesStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid --phase1-batch-sizes: %v\n", err)
		os.Exit(2)
	}
	rates, err := parseFloats(phase2RatesStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid --phase2-request-rates: %v\n", err)
		os.Exit(2)
	}
	cfg.phase1BatchSizes = batches
	cfg.phase2RequestRates = rates

	fillLevels, err := parseInt64s(fillLevelsStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid --fill-levels: %v\n", err)
		os.Exit(2)
	}
	cfg.fillLevels = fillLevels

	if _, err := pickSender(cfg.signal); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	switch cfg.scenario {
	case "throughput":
	case "read-probe":
		if cfg.jwt == "" || cfg.projectId == "" {
			fmt.Fprintln(os.Stderr, "--scenario=read-probe requires --jwt and --project-id")
			os.Exit(2)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown --scenario %q (expected throughput|read-probe)\n", cfg.scenario)
		os.Exit(2)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	deadline := time.Now().Add(cfg.duration)
	ctx, cancelDeadline := context.WithDeadline(ctx, deadline)
	defer cancelDeadline()

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        500,
			MaxIdleConnsPerHost: 200,
			IdleConnTimeout:     60 * time.Second,
		},
	}

	startedAt := time.Now().UTC()
	ingestStats := newLatencyTracker()
	ing, err := newIngester(cfg, httpClient, ingestStats)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	out := finalReport{
		Tier:      cfg.tier,
		Mode:      cfg.mode,
		Signal:    cfg.signal,
		Scenario:  cfg.scenario,
		StartedAt: startedAt.Format(time.RFC3339),
	}

	// Persist a partial report after every step so a mid-run crash, OOM,
	// or dropped SSH session still leaves usable data on disk. Atomic
	// write via tempfile + rename means concurrent readers (and any partial
	// kill of the writer process itself) never see a half-written file.
	writeCheckpoint := func() {
		out.EndedAt = time.Now().UTC().Format(time.RFC3339)
		out.computeHeadline()
		if err := writeReportAtomic(cfg.reportOut, &out); err != nil {
			fmt.Fprintf(os.Stderr, "checkpoint write failed: %v\n", err)
		}
	}

	switch cfg.scenario {
	case "throughput":
		phase1 := runBatchSizeRamp(ctx, cfg, ing, ingestStats, func(p phaseResult) {
			out.Phase1 = &p
			writeCheckpoint()
		})
		out.Phase1 = &phase1
		// Cool-down between phases: Phase 1's last step often runs the SUT at
		// 70-99% of capacity, leaving its internal queues (CH merges, PG WAL,
		// HTTP handler pool) saturated. Jumping straight into Phase 2 with the
		// SUT still digesting that wave produces "0 OK / 0 errors" garbage
		// because new requests sit on the SUT-side TCP backlog without ever
		// reaching a handler or timing out.
		if cfg.interPhaseCooldownSeconds > 0 {
			fmt.Fprintf(stderrPrefix(), "inter-phase cooldown: %v\n", cfg.interPhaseCooldownSeconds)
			select {
			case <-time.After(cfg.interPhaseCooldownSeconds):
			case <-ctx.Done():
			}
		}
		phase2 := runRequestRateRamp(ctx, cfg, ing, ingestStats, phase1, func(p phaseResult) {
			out.Phase2 = &p
			writeCheckpoint()
		})
		out.Phase2 = &phase2
	case "read-probe":
		probe := runReadProbe(ctx, cfg, ing, ingestStats, httpClient, func(p readProbeResult) {
			out.ReadProbe = &p
			writeCheckpoint()
		})
		out.ReadProbe = &probe
	}

	// Final write — even if everything above ran cleanly, do one last
	// atomic rewrite so the file reflects the EndedAt and headline.
	writeCheckpoint()

	switch cfg.scenario {
	case "throughput":
		fmt.Fprintf(os.Stderr, "wrote %s: signal=%s max sustainable %s/sec = %.0f\n",
			cfg.reportOut, cfg.signal, cfg.signal, out.MaxSustainableItemsPerSec)
	case "read-probe":
		fmt.Fprintf(os.Stderr, "wrote %s: signal=%s max fill level passed = %d rows\n",
			cfg.reportOut, cfg.signal, out.MaxFillLevelPassed)
	}
}

// writeReportAtomic encodes the report into a sibling .tmp file and renames
// it over the destination, so a kill or disk-full mid-write can't leave a
// half-encoded file. Called after every step.
func writeReportAtomic(path string, report *finalReport) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

func parseInt64s(s string) ([]int64, error) {
	parts := strings.Split(s, ",")
	out := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse %q: %w", p, err)
		}
		out = append(out, v)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty list")
	}
	return out, nil
}

func parseInts(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("parse %q: %w", p, err)
		}
		out = append(out, v)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty list")
	}
	return out, nil
}

func parseFloats(s string) ([]float64, error) {
	parts := strings.Split(s, ",")
	out := make([]float64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return nil, fmt.Errorf("parse %q: %w", p, err)
		}
		out = append(out, v)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty list")
	}
	return out, nil
}
