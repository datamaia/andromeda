package eventbus

import (
	"regexp"
	"time"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
)

// nameGrammar matches the event-name grammar `<area>[.<noun>].<verb-past>` (Volume 0
// chapter 03): two or three lowercase dot-separated segments, each a lowercase identifier that
// may contain underscores.
var nameGrammar = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*){1,2}$`)

// ValidName reports whether name conforms to the event-name grammar.
func ValidName(name string) bool { return nameGrammar.MatchString(name) }

// NewEvent builds an enveloped event with version 1, a UTC timestamp, and a fresh correlation
// ID when one is not supplied. Options set correlation/session/run IDs and payload.
func NewEvent(name, producer string, opts ...Option) ports.Event {
	e := ports.Event{
		Name:      name,
		Version:   1,
		Producer:  producer,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	}
	for _, o := range opts {
		o(&e)
	}
	if e.CorrelationID == "" {
		e.CorrelationID = core.NewULID()
	}
	return e
}

// Option configures an event built by NewEvent.
type Option func(*ports.Event)

// WithCorrelation sets the correlation ID.
func WithCorrelation(id core.ULID) Option { return func(e *ports.Event) { e.CorrelationID = id } }

// WithSession sets the session ID.
func WithSession(id core.ULID) Option { return func(e *ports.Event) { e.SessionID = id } }

// WithRun sets the run ID.
func WithRun(id core.ULID) Option { return func(e *ports.Event) { e.RunID = id } }

// WithPayload sets the JSON payload.
func WithPayload(p ports.JSON) Option { return func(e *ports.Event) { e.Payload = p } }
