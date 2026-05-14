package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	mathrand "math/rand"
	"time"

	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"
)

const otlpSpansPerRequest = 20

// sendOTLPTraces marshals an ExportTraceServiceRequest with otlpSpansPerRequest
// spans into protobuf, POSTs to /api/otel/v1/traces, and records latency.
// Protobuf (not protojson) because that's what production exporters send — it
// exercises the backend's protobuf decode path which is non-trivial CPU work
// that would be hidden if we sent JSON.
func sendOTLPTraces(ctx context.Context, client *http.Client, cfg config, rng *mathrand.Rand, stats *latencyTracker) {
	now := time.Now().UTC()
	resourceAttrs := []*commonpb.KeyValue{
		strAttr("service.name", "bench-loadgen"),
		strAttr("service.version", "1.0.0"),
		strAttr("deployment.environment", "bench"),
	}

	spans := make([]*tracepb.Span, otlpSpansPerRequest)
	for i := range spans {
		traceId := make([]byte, 16)
		spanId := make([]byte, 8)
		_, _ = rand.Read(traceId)
		_, _ = rand.Read(spanId)

		startNs := now.Add(-time.Duration(rng.Intn(500)) * time.Millisecond).UnixNano()
		duration := time.Duration(10+rng.Intn(990)) * time.Millisecond
		endNs := startNs + int64(duration)

		spans[i] = &tracepb.Span{
			TraceId:           traceId,
			SpanId:            spanId,
			Name:              endpointPaths[rng.Intn(len(endpointPaths))],
			Kind:              tracepb.Span_SPAN_KIND_SERVER,
			StartTimeUnixNano: uint64(startNs),
			EndTimeUnixNano:   uint64(endNs),
			Attributes: []*commonpb.KeyValue{
				strAttr("http.method", "GET"),
				intAttr("http.status_code", int64(pickStatus(rng))),
				strAttr("http.route", endpointPaths[rng.Intn(len(endpointPaths))]),
			},
			Status: &tracepb.Status{Code: tracepb.Status_STATUS_CODE_OK},
		}
	}

	req := &coltracepb.ExportTraceServiceRequest{
		ResourceSpans: []*tracepb.ResourceSpans{{
			Resource: &resourcepb.Resource{Attributes: resourceAttrs},
			ScopeSpans: []*tracepb.ScopeSpans{{
				Scope: &commonpb.InstrumentationScope{Name: "bench-loadgen", Version: "1.0.0"},
				Spans: spans,
			}},
		}},
	}

	body, err := proto.Marshal(req)
	if err != nil {
		stats.Record(0, err)
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.target+"/api/otel/v1/traces", bytes.NewReader(body))
	if err != nil {
		stats.Record(0, err)
		return
	}
	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("Authorization", "Bearer "+cfg.projectToken)

	start := time.Now()
	resp, err := client.Do(httpReq)
	elapsed := time.Since(start)
	if err != nil {
		stats.Record(elapsed.Seconds()*1000, err)
		return
	}
	defer resp.Body.Close()
	_, _ = readAndDiscard(resp.Body)
	if resp.StatusCode >= 400 {
		stats.Record(elapsed.Seconds()*1000, fmt.Errorf("status %d", resp.StatusCode))
		return
	}
	stats.Record(elapsed.Seconds()*1000, nil)
}

func strAttr(k, v string) *commonpb.KeyValue {
	return &commonpb.KeyValue{Key: k, Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: v}}}
}

func intAttr(k string, v int64) *commonpb.KeyValue {
	return &commonpb.KeyValue{Key: k, Value: &commonpb.AnyValue{Value: &commonpb.AnyValue_IntValue{IntValue: v}}}
}
