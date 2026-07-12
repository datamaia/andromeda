package telemetry

import (
	"context"
	"errors"
	"testing"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/datamaia/andromeda/internal/ports"
)

func newOTel(t *testing.T) (*OTel, *sdktrace.TracerProvider) {
	t.Helper()
	tp := sdktrace.NewTracerProvider()
	mp := sdkmetric.NewMeterProvider()
	o := NewOTel(tp, mp)
	t.Cleanup(func() { o.Shutdown(context.Background()) })
	return o, tp
}

func TestOTelEmitMetricAndSpan(t *testing.T) {
	ctx := context.Background()
	o, _ := newOTel(t)

	if err := o.EmitMetric(ctx, ports.MetricSample{Name: "tokens.total", Value: 12}); err != nil {
		t.Fatalf("emit: %v", err)
	}
	sv := "provider"
	ctx2, span := o.StartSpan(ctx, "chat", ports.Attributes{"component": {String: &sv}})
	if ctx2 == nil {
		t.Fatal("StartSpan should return a derived context")
	}
	span.SetAttribute("model", ports.AttrValue{String: strptr("claude")})
	span.RecordError(errors.New("boom"))
	span.End(ports.SpanError)

	if err := o.Flush(ctx); err != nil {
		t.Errorf("flush: %v", err)
	}
}

func TestOTelNeverFailsOnCancelledContext(t *testing.T) {
	o, _ := newOTel(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := o.EmitMetric(ctx, ports.MetricSample{Name: "x", Value: 1}); err != nil {
		t.Errorf("telemetry must not fail on cancelled context: %v", err)
	}
}

func strptr(s string) *string { return &s }
