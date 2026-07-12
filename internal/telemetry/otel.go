package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/datamaia/andromeda/internal/ports"
)

// OTel is a ports.TelemetryPort backed by the OpenTelemetry Go SDK (ADR-011). Export is
// consent-gated at the caller: an OTel provider is constructed only where remote/local export
// is enabled (Volume 10). Metric emission and span creation flow through the SDK's providers,
// which are wired to exporters (stdout for local-first sinks, OTLP where consented).
type OTel struct {
	tracer oteltrace.Tracer
	meter  otelMeter
	tp     *sdktrace.TracerProvider
	mp     *sdkmetric.MeterProvider
}

// otelMeter is the minimal metric surface used (a counter registry keyed by name).
type otelMeter struct {
	provider *sdkmetric.MeterProvider
}

// NewOTel builds an OTel-backed TelemetryPort over the given SDK providers. The caller wires the
// exporters (e.g. stdout for local sinks) and passes the providers here.
func NewOTel(tp *sdktrace.TracerProvider, mp *sdkmetric.MeterProvider) *OTel {
	return &OTel{
		tracer: tp.Tracer("andromeda"),
		meter:  otelMeter{provider: mp},
		tp:     tp,
		mp:     mp,
	}
}

var _ ports.TelemetryPort = (*OTel)(nil)

// EmitMetric records one sample as an OTel counter add against the metric name.
func (o *OTel) EmitMetric(ctx context.Context, sample ports.MetricSample) error {
	if ctx.Err() != nil {
		return nil
	}
	m := o.meter.provider.Meter("andromeda")
	c, err := m.Float64Counter(sample.Name)
	if err != nil {
		return nil // never fail the observed operation
	}
	c.Add(ctx, sample.Value)
	return nil
}

// StartSpan opens an OTel span nested under the span in ctx.
func (o *OTel) StartSpan(ctx context.Context, name string, attrs ports.Attributes) (context.Context, ports.Span) {
	ctx, s := o.tracer.Start(ctx, name)
	for k, v := range attrs {
		s.SetAttributes(toKV(k, v))
	}
	return ctx, &otelSpan{span: s}
}

// Flush drains buffered telemetry to the exporters.
func (o *OTel) Flush(ctx context.Context) error {
	_ = o.tp.ForceFlush(ctx)
	_ = o.mp.ForceFlush(ctx)
	return nil
}

// Shutdown flushes and stops the providers.
func (o *OTel) Shutdown(ctx context.Context) error {
	_ = o.tp.Shutdown(ctx)
	_ = o.mp.Shutdown(ctx)
	return nil
}

type otelSpan struct{ span oteltrace.Span }

func (s *otelSpan) SetAttribute(key string, value ports.AttrValue) {
	s.span.SetAttributes(toKV(key, value))
}
func (s *otelSpan) RecordError(err error) { s.span.RecordError(err) }
func (s *otelSpan) End(status ports.SpanStatus) {
	switch status {
	case ports.SpanError:
		s.span.SetStatus(codes.Error, "")
	case ports.SpanOK:
		s.span.SetStatus(codes.Ok, "")
	}
	s.span.End()
}

func toKV(key string, v ports.AttrValue) attribute.KeyValue {
	switch {
	case v.String != nil:
		return attribute.String(key, *v.String)
	case v.Int != nil:
		return attribute.Int64(key, *v.Int)
	case v.Float != nil:
		return attribute.Float64(key, *v.Float)
	case v.Bool != nil:
		return attribute.Bool(key, *v.Bool)
	default:
		return attribute.String(key, "")
	}
}
