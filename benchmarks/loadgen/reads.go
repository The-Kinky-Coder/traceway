package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type reader struct {
	cfg      config
	client   *http.Client
	stats    map[string]*latencyTracker
	ingester *ingester
}

func newReader(cfg config, client *http.Client, stats map[string]*latencyTracker) *reader {
	return &reader{cfg: cfg, client: client, stats: stats}
}

// attachIngester gives the reader access to a sample session id so it can hit
// /api/sessions/:id/recording with an id that actually exists.
func (r *reader) attachIngester(i *ingester) { r.ingester = i }

type readSpec struct {
	name   string
	method string
	path   string
	body   func() any
}

func (r *reader) specs() []readSpec {
	last24h := func() (string, string) {
		now := time.Now().UTC()
		return now.Add(-24 * time.Hour).Format(time.RFC3339), now.Format(time.RFC3339)
	}
	last1h := func() (string, string) {
		now := time.Now().UTC()
		return now.Add(-1 * time.Hour).Format(time.RFC3339), now.Format(time.RFC3339)
	}
	return []readSpec{
		{
			name: "endpoints_grouped", method: "POST", path: "/api/endpoints/grouped",
			body: func() any {
				f, t := last24h()
				return map[string]any{
					"fromDate":      f,
					"toDate":        t,
					"orderBy":       "count",
					"sortDirection": "desc",
					"pagination":    map[string]int{"page": 1, "pageSize": 50},
					"search":        "",
				}
			},
		},
		{
			name: "exceptions_grouped", method: "POST", path: "/api/exception-stack-traces",
			body: func() any {
				f, t := last24h()
				return map[string]any{
					"fromDate":        f,
					"toDate":          t,
					"orderBy":         "count desc",
					"pagination":      map[string]int{"page": 1, "pageSize": 50},
					"search":          "",
					"searchType":      "all",
					"includeArchived": false,
				}
			},
		},
		{name: "dashboard_overview", method: "GET", path: "/api/dashboard/overview"},
		{name: "logs", method: "POST", path: "/api/logs", body: func() any {
			f, t := last1h()
			return map[string]any{
				"fromDate":   f,
				"toDate":     t,
				"orderBy":    "timestamp desc",
				"pagination": map[string]int{"page": 1, "pageSize": 50},
			}
		}},
		// session_recording handled separately because the URL depends on a
		// live session id surfaced by the ingester.
	}
}

func (r *reader) run(ctx context.Context) {
	var wg sync.WaitGroup
	for _, spec := range r.specs() {
		spec := spec
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.loopSpec(ctx, spec)
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		r.loopRecording(ctx)
	}()
	wg.Wait()
}

func (r *reader) loopSpec(ctx context.Context, spec readSpec) {
	interval := time.Duration(float64(time.Second) / r.cfg.readRPS)
	if interval < 10*time.Millisecond {
		interval = 10 * time.Millisecond
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			r.doOne(ctx, spec)
		}
	}
}

func (r *reader) loopRecording(ctx context.Context) {
	interval := time.Duration(float64(time.Second) / r.cfg.readRPS)
	if interval < 10*time.Millisecond {
		interval = 10 * time.Millisecond
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			sid := ""
			if r.ingester != nil {
				sid = r.ingester.SampleSessionId()
			}
			if sid == "" {
				// Nothing seeded yet — quietly skip; don't record as ok or err.
				continue
			}
			r.doOne(ctx, readSpec{
				name:   "session_recording",
				method: "GET",
				path:   "/api/sessions/" + sid + "/recording",
			})
		}
	}
}

func (r *reader) doOne(ctx context.Context, spec readSpec) {
	tracker, ok := r.stats[spec.name]
	if !ok {
		return
	}

	u, err := url.Parse(r.cfg.target + spec.path)
	if err != nil {
		tracker.Record(0, err)
		return
	}
	q := u.Query()
	q.Set("projectId", r.cfg.projectId)
	u.RawQuery = q.Encode()

	var body []byte
	if spec.body != nil {
		body, err = json.Marshal(spec.body())
		if err != nil {
			tracker.Record(0, err)
			return
		}
	}

	req, err := http.NewRequestWithContext(ctx, spec.method, u.String(), bytes.NewReader(body))
	if err != nil {
		tracker.Record(0, err)
		return
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if r.cfg.jwt != "" {
		req.Header.Set("Authorization", "Bearer "+r.cfg.jwt)
	}

	start := time.Now()
	resp, err := r.client.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		tracker.Record(elapsed.Seconds()*1000, err)
		return
	}
	defer resp.Body.Close()
	_, _ = readAndDiscard(resp.Body)
	if resp.StatusCode >= 400 {
		tracker.Record(elapsed.Seconds()*1000, fmt.Errorf("status %d", resp.StatusCode))
		return
	}
	tracker.Record(elapsed.Seconds()*1000, nil)
}
