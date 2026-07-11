// Package telemetry is layer L2 infrastructure: a local-first ports.TelemetryPort (ADR-011).
// This foundation records metrics and spans to in-process sinks; wiring the OpenTelemetry Go
// SDK and OTLP export (consent-gated, Volume 10) is a later increment. Telemetry failure never
// fails the operation being observed: all methods are non-blocking and swallow sink errors.
package telemetry

import (
	"context"
	"sync"

	"github.com/datamaia/andromeda/internal/ports"
)

// Recorder is a local TelemetryPort implementation with an in-memory metric registry.
type Recorder struct {
	mu      sync.Mutex
	metrics map[string]*metric
}

// New returns an empty Recorder.
func New() *Recorder {
	return &Recorder{metrics: map[string]*metric{}}
}

var _ ports.TelemetryPort = (*Recorder)(nil)

type metric struct {
	count int64
	sum   float64
	last  float64
}

// EmitMetric records one sample against a metric name.
func (r *Recorder) EmitMetric(ctx context.Context, sample ports.MetricSample) error {
	if ctx.Err() != nil {
		return nil // never fail the observed operation
	}
	r.mu.Lock()
	m := r.metrics[sample.Name]
	if m == nil {
		m = &metric{}
		r.metrics[sample.Name] = m
	}
	m.count++
	m.sum += sample.Value
	m.last = sample.Value
	r.mu.Unlock()
	return nil
}

// StartSpan opens a span nested under the span in ctx and returns the derived context.
func (r *Recorder) StartSpan(ctx context.Context, name string, attrs ports.Attributes) (context.Context, ports.Span) {
	parent := spanFromContext(ctx)
	s := &span{name: name, attrs: cloneAttrs(attrs)}
	if parent != nil {
		s.depth = parent.depth + 1
	}
	return context.WithValue(ctx, spanKey{}, s), s
}

// Flush drains buffered telemetry to local sinks. The local recorder has nothing to flush.
func (r *Recorder) Flush(ctx context.Context) error { return nil }

// MetricSnapshot returns the count and running sum for a metric (for tests and diagnostics).
func (r *Recorder) MetricSnapshot(name string) (count int64, sum float64, ok bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, present := r.metrics[name]
	if !present {
		return 0, 0, false
	}
	return m.count, m.sum, true
}

type spanKey struct{}

func spanFromContext(ctx context.Context) *span {
	s, _ := ctx.Value(spanKey{}).(*span)
	return s
}

type span struct {
	mu     sync.Mutex
	name   string
	attrs  ports.Attributes
	depth  int
	status ports.SpanStatus
	err    error
	ended  bool
}

func (s *span) SetAttribute(key string, value ports.AttrValue) {
	s.mu.Lock()
	if s.attrs == nil {
		s.attrs = ports.Attributes{}
	}
	s.attrs[key] = value
	s.mu.Unlock()
}

func (s *span) RecordError(err error) {
	s.mu.Lock()
	s.err = err
	s.mu.Unlock()
}

func (s *span) End(status ports.SpanStatus) {
	s.mu.Lock()
	if !s.ended {
		s.ended = true
		s.status = status
	}
	s.mu.Unlock()
}

func cloneAttrs(a ports.Attributes) ports.Attributes {
	if a == nil {
		return ports.Attributes{}
	}
	out := make(ports.Attributes, len(a))
	for k, v := range a {
		out[k] = v
	}
	return out
}
