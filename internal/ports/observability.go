package ports

import "context"

// EventBusPort is the in-process typed publish/subscribe channel of ADR-012. The envelope,
// ordering, delivery semantics, and per-family overflow policy are Volume 10's. Errors: E-OBS.
type EventBusPort interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(ctx context.Context, sel TopicSelector, opts SubscribeOptions) (Subscription, error)
}

// Subscription is a live subscriber with a bounded per-subscriber buffer (ADR-012).
type Subscription interface {
	Events() Stream[Event]
	Close() error
}

// Event is one enveloped event. The full envelope is Volume 10's (FR-OBS-001); this shape
// fixes the fields ports exchange.
type Event struct {
	Name          string // "<area>[.<noun>].<verb-past>"
	Version       int
	Producer      string
	CorrelationID ULID
	SessionID     ULID
	RunID         ULID
	Timestamp     string // RFC 3339 UTC
	Payload       JSON
}

// TopicSelector selects events by exact name or area prefix.
type TopicSelector struct {
	Names    []string
	Prefixes []string
}

// SubscribeOptions tunes a subscription within Volume 10 bounds.
type SubscribeOptions struct {
	BufferSize        int
	ReplayFromPersist bool
}

// TelemetryPort records metrics, traces, and spans under ADR-011 (OpenTelemetry, local-first
// sinks, consent-gated export). Contract owner: Volume 10. Errors: E-OBS. Telemetry failure
// MUST never fail the operation being observed.
type TelemetryPort interface {
	EmitMetric(ctx context.Context, sample MetricSample) error
	StartSpan(ctx context.Context, name string, attrs Attributes) (context.Context, Span)
	Flush(ctx context.Context) error
}

// Span is one tracing span nested under the span in its parent context.
type Span interface {
	SetAttribute(key string, value AttrValue)
	RecordError(err error)
	End(status SpanStatus)
}

// MetricSample is one sample against a registered Metric definition (Volume 2).
type MetricSample struct {
	Name       string
	Value      float64
	Unit       string
	Attributes Attributes
}

// Attributes is a set of structured key/value attributes on a span or metric.
type Attributes map[string]AttrValue

// AttrValue is a typed attribute value (string, int, float, or bool).
type AttrValue struct {
	String *string
	Int    *int64
	Float  *float64
	Bool   *bool
}

// SpanStatus mirrors the Trace recorded-status vocabulary.
type SpanStatus string

const (
	SpanOK          SpanStatus = "ok"
	SpanError       SpanStatus = "error"
	SpanInterrupted SpanStatus = "interrupted"
)
