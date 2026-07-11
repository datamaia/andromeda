// Package logging is layer L2 infrastructure: structured, redacted, local-first logging via
// the standard library's log/slog with a JSON handler (ADR-011). Sensitive attribute keys are
// redacted at the handler so secrets never reach a log sink (Volume 9 redaction; PRD-006).
package logging

import (
	"io"
	"log/slog"
	"strings"
)

// redactedKeys are attribute keys whose values are replaced with a placeholder. Matching is
// case-insensitive and substring-based so nested variants (api_key, auth_token) are caught.
var redactedKeys = []string{
	"secret", "token", "password", "passwd", "api_key", "apikey",
	"credential", "authorization", "private_key", "access_key",
}

// Redacted is the replacement value written in place of a sensitive attribute.
const Redacted = "[REDACTED]"

// Options configures a logger.
type Options struct {
	Level  string // "debug" | "info" | "warn" | "error"
	Format string // "json" (default) | "text"
}

// New returns a slog.Logger writing to w with the given options and redaction applied.
func New(w io.Writer, opts Options) *slog.Logger {
	handlerOpts := &slog.HandlerOptions{
		Level:       parseLevel(opts.Level),
		ReplaceAttr: redact,
	}
	var h slog.Handler
	if strings.EqualFold(opts.Format, "text") {
		h = slog.NewTextHandler(w, handlerOpts)
	} else {
		h = slog.NewJSONHandler(w, handlerOpts)
	}
	return slog.New(h)
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// IsSensitiveKey reports whether an attribute key should be redacted.
func IsSensitiveKey(key string) bool {
	lk := strings.ToLower(key)
	for _, s := range redactedKeys {
		if strings.Contains(lk, s) {
			return true
		}
	}
	return false
}

func redact(_ []string, a slog.Attr) slog.Attr {
	if IsSensitiveKey(a.Key) {
		return slog.String(a.Key, Redacted)
	}
	return a
}
