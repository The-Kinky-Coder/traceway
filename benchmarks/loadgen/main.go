package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type config struct {
	target              string
	projectToken        string
	jwt                 string
	projectId           string
	duration            time.Duration
	stepDuration        time.Duration
	initialEPS          float64
	maxEPS              float64
	reportShare         float64
	readRPS             float64
	ingestErrThreshold  float64
	readP99ThresholdMs  float64
	reportOut           string
	tier                string
	mode                string
}

func main() {
	var cfg config

	flag.StringVar(&cfg.target, "target", "", "Base URL of the system under test (e.g. http://10.0.0.2 or http://localhost:8087)")
	flag.StringVar(&cfg.projectToken, "token", "", "Project bearer token for /api/report + OTLP")
	flag.StringVar(&cfg.jwt, "jwt", "", "JWT for dashboard read endpoints")
	flag.StringVar(&cfg.projectId, "project-id", "", "Project UUID (passed as ?projectId= on read endpoints)")
	flag.DurationVar(&cfg.duration, "duration", 30*time.Minute, "Total run duration cap")
	flag.DurationVar(&cfg.stepDuration, "step-duration", 2*time.Minute, "Per-step hold time")
	flag.Float64Var(&cfg.initialEPS, "initial-eps", 100, "Starting target traces/sec")
	flag.Float64Var(&cfg.maxEPS, "max-eps", 50000, "Hard cap on target traces/sec")
	flag.Float64Var(&cfg.reportShare, "report-share", 0.7, "Fraction of ingest sent via /api/report (rest via OTLP)")
	flag.Float64Var(&cfg.readRPS, "read-rps", 1.0, "Read RPS per endpoint")
	flag.Float64Var(&cfg.ingestErrThreshold, "ingest-err-threshold", 0.05, "Step fails if ingest error rate exceeds this")
	flag.Float64Var(&cfg.readP99ThresholdMs, "read-p99-threshold-ms", 3000, "Step fails if any read endpoint p99 (ms) exceeds this")
	flag.StringVar(&cfg.reportOut, "report-out", "", "Path to write JSON results (required)")
	flag.StringVar(&cfg.tier, "tier", "local", "Hardware tier label embedded in output (e.g. ccx13)")
	flag.StringVar(&cfg.mode, "mode", "unknown", "DB mode label embedded in output (sqlite | pgch)")
	flag.Parse()

	if cfg.target == "" || cfg.projectToken == "" || cfg.reportOut == "" {
		fmt.Fprintln(os.Stderr, "missing required flag: --target, --token, --report-out")
		flag.Usage()
		os.Exit(2)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	deadline := time.Now().Add(cfg.duration)
	ctx, cancelDeadline := context.WithDeadline(ctx, deadline)
	defer cancelDeadline()

	httpClient := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     60 * time.Second,
		},
	}

	ingestStats := newLatencyTracker()
	readStats := map[string]*latencyTracker{
		"endpoints_grouped":  newLatencyTracker(),
		"exceptions_grouped": newLatencyTracker(),
		"dashboard_overview": newLatencyTracker(),
		"session_recording":  newLatencyTracker(),
		"logs":               newLatencyTracker(),
	}

	ingester := newIngester(cfg, httpClient, ingestStats)
	reader := newReader(cfg, httpClient, readStats)
	reader.attachIngester(ingester)

	go ingester.run(ctx)
	go reader.run(ctx)

	results := runRamp(ctx, cfg, ingester, ingestStats, readStats)

	out := finalReport{
		Tier:      cfg.tier,
		Mode:      cfg.mode,
		StartedAt: time.Now().UTC().Add(-cfg.duration).Format(time.RFC3339),
		EndedAt:   time.Now().UTC().Format(time.RFC3339),
		Steps:     results,
	}
	out.computeBreakingPoint()

	f, err := os.Create(cfg.reportOut)
	if err != nil {
		log.Fatalf("create %s: %v", cfg.reportOut, err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(&out); err != nil {
		log.Fatalf("write %s: %v", cfg.reportOut, err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s: max sustainable EPS = %.0f (%d step(s))\n", cfg.reportOut, out.MaxSustainableEPS, len(out.Steps))
}
