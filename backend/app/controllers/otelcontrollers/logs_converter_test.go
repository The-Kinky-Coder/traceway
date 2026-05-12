package otelcontrollers

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
)

// Pins the protojson base64-of-hex round-trip — see logs_converter.go.
func TestToLogRecord_IDEncoding(t *testing.T) {
	const wireTraceHex = "7b873c7bbf35739e79e1f7b9736739f7"
	const wireSpanHex = "7dfd3877775ae1bd"

	// protojson decodes bytes-typed fields as base64; hex chars are valid base64.
	jsonTrace, err := base64.StdEncoding.DecodeString(wireTraceHex)
	if err != nil || len(jsonTrace) != 24 {
		t.Fatalf("seed: jsonTrace len=%d err=%v", len(jsonTrace), err)
	}
	jsonSpan, err := base64.StdEncoding.DecodeString(wireSpanHex)
	if err != nil || len(jsonSpan) != 12 {
		t.Fatalf("seed: jsonSpan len=%d err=%v", len(jsonSpan), err)
	}
	binTrace, _ := hex.DecodeString(wireTraceHex)
	binSpan, _ := hex.DecodeString(wireSpanHex)

	tests := []struct {
		name       string
		traceBytes []byte
		spanBytes  []byte
		wantTrace  string
		wantSpan   string
	}{
		{"binary OTLP", binTrace, binSpan, wireTraceHex, wireSpanHex},
		{"JSON OTLP (protojson base64 roundtrip)", jsonTrace, jsonSpan, wireTraceHex, wireSpanHex},
		{"missing trace context", nil, nil, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := toLogRecord(
				testProjectId,
				&logspb.LogRecord{TraceId: tt.traceBytes, SpanId: tt.spanBytes},
				"svc", "", nil, "", "scope", "", nil,
			)
			if rec.TraceId != tt.wantTrace {
				t.Errorf("TraceId = %q (len %d), want %q", rec.TraceId, len(rec.TraceId), tt.wantTrace)
			}
			if rec.SpanId != tt.wantSpan {
				t.Errorf("SpanId = %q (len %d), want %q", rec.SpanId, len(rec.SpanId), tt.wantSpan)
			}
		})
	}
}
