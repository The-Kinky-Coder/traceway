package main

import (
	"context"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"

	"golang.org/x/time/rate"
)

const (
	// One frame carries 10 traces, so frame rate = targetEPS / tracesPerFrame.
	// Matches Go SDK behavior where frames rotate every ~5s with whatever has
	// accumulated; benchmark uses a fixed batch size for predictable accounting.
	tracesPerFrame  = 10
	ingestWorkers   = 256
	recordingChance = 0.05 // 5% of frames carry a session recording segment
	exceptionChance = 0.02 // 2% of frames carry an exception
)

type ingester struct {
	cfg     config
	client  *http.Client
	stats   *latencyTracker
	limiter *rate.Limiter
	// rng pool: each worker grabs one to avoid contending on the global source
	rngPool sync.Pool
	// sessionsCreated counts sessions sent via /api/report so the reader can
	// query a real session ID for the recording playback endpoint.
	sessionsCreated atomic.Int64
	sampleSession   atomic.Value // string — most recently seeded session id
}

func newIngester(cfg config, client *http.Client, stats *latencyTracker) *ingester {
	framesPerSec := cfg.initialEPS / tracesPerFrame
	if framesPerSec < 1 {
		framesPerSec = 1
	}
	return &ingester{
		cfg:     cfg,
		client:  client,
		stats:   stats,
		limiter: rate.NewLimiter(rate.Limit(framesPerSec), int(framesPerSec)+1),
		rngPool: sync.Pool{New: func() any { return rand.New(rand.NewSource(rand.Int63())) }},
	}
}

func (i *ingester) SetRate(eps float64) {
	fps := eps / tracesPerFrame
	if fps < 1 {
		fps = 1
	}
	i.limiter.SetLimit(rate.Limit(fps))
	i.limiter.SetBurst(int(fps) + 1)
}

func (i *ingester) SampleSessionId() string {
	v := i.sampleSession.Load()
	if v == nil {
		return ""
	}
	return v.(string)
}

func (i *ingester) run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(ingestWorkers)
	for w := 0; w < ingestWorkers; w++ {
		go func() {
			defer wg.Done()
			rng := i.rngPool.Get().(*rand.Rand)
			defer i.rngPool.Put(rng)
			for {
				if err := i.limiter.Wait(ctx); err != nil {
					return
				}
				i.sendOne(ctx, rng)
			}
		}()
	}
	wg.Wait()
}

func (i *ingester) sendOne(ctx context.Context, rng *rand.Rand) {
	if rng.Float64() < i.cfg.reportShare {
		sessionId := sendReportFrame(ctx, i.client, i.cfg, rng, i.stats)
		if sessionId != "" {
			i.sampleSession.Store(sessionId)
			i.sessionsCreated.Add(1)
		}
	} else {
		sendOTLPTraces(ctx, i.client, i.cfg, rng, i.stats)
	}
}
