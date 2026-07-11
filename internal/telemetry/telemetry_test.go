package telemetry

import (
	"context"
	"errors"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

func TestEmitMetricAggregates(t *testing.T) {
	ctx := context.Background()
	r := New()
	_ = r.EmitMetric(ctx, ports.MetricSample{Name: "tokens", Value: 10})
	_ = r.EmitMetric(ctx, ports.MetricSample{Name: "tokens", Value: 5})
	count, sum, ok := r.MetricSnapshot("tokens")
	if !ok || count != 2 || sum != 15 {
		t.Fatalf("snapshot = %d,%v,%v", count, sum, ok)
	}
	if _, _, ok := r.MetricSnapshot("missing"); ok {
		t.Error("missing metric should report ok=false")
	}
}

func TestSpanNestingAndEnd(t *testing.T) {
	ctx := context.Background()
	r := New()
	ctx1, s1 := r.StartSpan(ctx, "outer", nil)
	_, s2 := r.StartSpan(ctx1, "inner", ports.Attributes{"k": {}})
	s2.SetAttribute("added", ports.AttrValue{})
	s2.RecordError(errors.New("boom"))
	s2.End(ports.SpanError)
	s1.End(ports.SpanOK)

	if s2.(*span).depth != s1.(*span).depth+1 {
		t.Error("inner span should be one deeper than outer")
	}
	if s2.(*span).status != ports.SpanError {
		t.Error("inner span status should be error")
	}
	// End is idempotent-ish: a second End does not overwrite the recorded status.
	s2.End(ports.SpanOK)
	if s2.(*span).status != ports.SpanError {
		t.Error("second End must not overwrite the first status")
	}
}

func TestTelemetryNeverFails(t *testing.T) {
	r := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := r.EmitMetric(ctx, ports.MetricSample{Name: "x", Value: 1}); err != nil {
		t.Errorf("telemetry must not fail on cancelled context: %v", err)
	}
	if err := r.Flush(context.Background()); err != nil {
		t.Errorf("flush: %v", err)
	}
}
