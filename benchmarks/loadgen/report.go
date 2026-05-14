package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type rTrace struct {
	Id                 string            `json:"id"`
	Endpoint           string            `json:"endpoint"`
	Duration           int64             `json:"duration"` // nanoseconds
	RecordedAt         time.Time         `json:"recordedAt"`
	StatusCode         int               `json:"statusCode"`
	BodySize           int               `json:"bodySize"`
	ClientIP           string            `json:"clientIP"`
	Attributes         map[string]string `json:"attributes"`
	Spans              []rSpan           `json:"spans"`
	IsTask             bool              `json:"isTask"`
	DistributedTraceId string            `json:"distributedTraceId"`
}

type rSpan struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	StartTime time.Time `json:"startTime"`
	Duration  int64     `json:"duration"`
}

type rSession struct {
	Id                 string            `json:"id"`
	StartedAt          time.Time         `json:"startedAt"`
	ClientIP           string            `json:"clientIP"`
	Attributes         map[string]string `json:"attributes"`
	DistributedTraceId string            `json:"distributedTraceId,omitempty"`
}

type rRecording struct {
	ExceptionId  string          `json:"exceptionId"`
	SessionId    string          `json:"sessionId,omitempty"`
	SegmentIndex int32           `json:"segmentIndex"`
	Events       json.RawMessage `json:"events"`
	StartedAt    *time.Time      `json:"startedAt,omitempty"`
	EndedAt      *time.Time      `json:"endedAt,omitempty"`
}

type rException struct {
	IsTask     bool              `json:"isTask"`
	StackTrace string            `json:"stackTrace"`
	RecordedAt time.Time         `json:"recordedAt"`
	Attributes map[string]string `json:"attributes"`
	IsMessage  bool              `json:"isMessage"`
}

type rMetric struct {
	Name       string            `json:"name"`
	Value      float64           `json:"value"`
	RecordedAt time.Time         `json:"recordedAt"`
	Tags       map[string]string `json:"tags,omitempty"`
}

type rFrame struct {
	StackTraces       []rException `json:"stackTraces"`
	Metrics           []rMetric    `json:"metrics"`
	Traces            []rTrace     `json:"traces"`
	SessionRecordings []rRecording `json:"sessionRecordings"`
	Sessions          []rSession   `json:"sessions"`
}

type rReport struct {
	CollectionFrames []rFrame `json:"collectionFrames"`
	AppVersion       string   `json:"appVersion"`
	ServerName       string   `json:"serverName"`
}

var endpointPaths = []string{
	"GET /api/users", "POST /api/users", "GET /api/orders", "POST /api/orders",
	"GET /api/products/:id", "PUT /api/products/:id", "GET /api/search",
	"POST /api/checkout", "GET /api/cart", "DELETE /api/cart/:id",
	"GET /api/auth/me", "POST /api/auth/login", "GET /api/health",
}

// sendReportFrame builds one CollectionFrame, POSTs it gzipped to /api/report,
// records latency on the ingest tracker, and returns a sample session id (when
// one was included in the frame) so the reader can later query its recording.
func sendReportFrame(ctx context.Context, client *http.Client, cfg config, rng *rand.Rand, stats *latencyTracker) string {
	frame := buildFrame(rng)

	body, err := encodeFrame(rReport{
		CollectionFrames: []rFrame{frame},
		AppVersion:       "bench-1.0.0",
		ServerName:       fmt.Sprintf("bench-%s-%s", cfg.tier, cfg.mode),
	})
	if err != nil {
		stats.Record(0, err)
		return ""
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.target+"/api/report", bytes.NewReader(body))
	if err != nil {
		stats.Record(0, err)
		return ""
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Authorization", "Bearer "+cfg.projectToken)

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		stats.Record(elapsed.Seconds()*1000, err)
		return ""
	}
	defer resp.Body.Close()
	// Drain so the connection can be reused — keep-alive is critical at high
	// frame rates; not draining leaves connections in CLOSE_WAIT.
	_, _ = readAndDiscard(resp.Body)

	if resp.StatusCode >= 400 {
		stats.Record(elapsed.Seconds()*1000, fmt.Errorf("status %d", resp.StatusCode))
		return ""
	}
	stats.Record(elapsed.Seconds()*1000, nil)

	if len(frame.Sessions) > 0 {
		return frame.Sessions[0].Id
	}
	return ""
}

func buildFrame(rng *rand.Rand) rFrame {
	now := time.Now().UTC()
	frame := rFrame{
		Traces:      make([]rTrace, 0, tracesPerFrame),
		Metrics:     make([]rMetric, 0, 5),
		Sessions:    nil,
		StackTraces: nil,
	}

	for i := 0; i < tracesPerFrame; i++ {
		t := rTrace{
			Id:         uuid.NewString(),
			Endpoint:   endpointPaths[rng.Intn(len(endpointPaths))],
			Duration:   int64(time.Duration(10+rng.Intn(990)) * time.Millisecond), // 10-1000ms
			RecordedAt: now.Add(-time.Duration(rng.Intn(5000)) * time.Millisecond),
			StatusCode: pickStatus(rng),
			BodySize:   200 + rng.Intn(8000),
			ClientIP:   "192.0.2.1",
			Attributes: map[string]string{"region": "bench", "tier": "x"},
			Spans:      buildSpans(rng, now),
			IsTask:     rng.Float64() < 0.10,
		}
		frame.Traces = append(frame.Traces, t)
	}

	for i := 0; i < 5; i++ {
		frame.Metrics = append(frame.Metrics, rMetric{
			Name:       "bench.metric." + []string{"cpu", "mem", "qps", "lat", "errs"}[i],
			Value:      rng.Float64() * 100,
			RecordedAt: now,
		})
	}

	if rng.Float64() < recordingChance {
		sid := uuid.NewString()
		started := now.Add(-2 * time.Second)
		ended := now
		frame.Sessions = []rSession{{
			Id:         sid,
			StartedAt:  started,
			ClientIP:   "192.0.2.1",
			Attributes: map[string]string{"browser": "bench"},
		}}
		frame.SessionRecordings = []rRecording{{
			ExceptionId:  uuid.NewString(),
			SessionId:    sid,
			SegmentIndex: 0,
			Events:       fakeRRWeb(rng),
			StartedAt:    &started,
			EndedAt:      &ended,
		}}
	} else if rng.Float64() < 0.2 {
		// Even without a recording, ship occasional sessions so the read mix
		// always has at least one session id to query.
		sid := uuid.NewString()
		frame.Sessions = []rSession{{
			Id:         sid,
			StartedAt:  now,
			ClientIP:   "192.0.2.1",
			Attributes: map[string]string{"browser": "bench"},
		}}
	}

	if rng.Float64() < exceptionChance {
		frame.StackTraces = []rException{{
			IsTask:     false,
			StackTrace: "RuntimeError: bench-synthetic\n  at bench:1\n  at bench:2",
			RecordedAt: now,
			Attributes: map[string]string{"env": "bench"},
		}}
	}

	return frame
}

func buildSpans(rng *rand.Rand, now time.Time) []rSpan {
	n := 2 + rng.Intn(4) // 2-5 spans
	spans := make([]rSpan, n)
	for i := range spans {
		spans[i] = rSpan{
			Id:        uuid.NewString(),
			Name:      []string{"db.query", "http.call", "cache.get", "render", "json.encode"}[rng.Intn(5)],
			StartTime: now.Add(-time.Duration(rng.Intn(500)) * time.Millisecond),
			Duration:  int64(time.Duration(1+rng.Intn(200)) * time.Millisecond),
		}
	}
	return spans
}

func pickStatus(rng *rand.Rand) int {
	r := rng.Float64()
	switch {
	case r < 0.92:
		return 200
	case r < 0.95:
		return 201
	case r < 0.98:
		return 404
	default:
		return 500
	}
}

// fakeRRWeb produces a ~100KB JSON array shaped like rrweb events. The backend
// stores this opaquely in S3/disk without inspecting the structure, so a valid
// JSON array of arbitrary size is sufficient to exercise the storage path.
func fakeRRWeb(rng *rand.Rand) json.RawMessage {
	var buf bytes.Buffer
	buf.WriteString("[")
	const targetBytes = 100_000
	first := true
	for buf.Len() < targetBytes {
		if !first {
			buf.WriteString(",")
		}
		first = false
		fmt.Fprintf(&buf, `{"type":%d,"timestamp":%d,"data":{"href":"https://bench.local/page/%d","text":"%s"}}`,
			2+rng.Intn(4), time.Now().UnixMilli(), rng.Intn(10000), randomString(rng, 64))
	}
	buf.WriteString("]")
	return buf.Bytes()
}

func randomString(rng *rand.Rand, n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 "
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rng.Intn(len(charset))]
	}
	return string(b)
}

func encodeFrame(r rReport) ([]byte, error) {
	js, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	var gz bytes.Buffer
	w := gzip.NewWriter(&gz)
	if _, err := w.Write(js); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return gz.Bytes(), nil
}
