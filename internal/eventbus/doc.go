// Package eventbus is layer L2 infrastructure: the in-process typed publish/subscribe bus
// implementing ports.EventBusPort (ADR-012, FR-OBS-001). Publishing never blocks on slow
// subscribers — each subscriber has a bounded buffer and a per-subscription overflow policy
// (drop-oldest by default). Delivery is at-most-once in-process; durable event history is the
// EventStore's responsibility (persisted Event records, Volume 2).
package eventbus
