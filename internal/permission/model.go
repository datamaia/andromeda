package permission

import (
	"strings"
	"time"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
)

// Effect is a grant/rule effect. It maps onto the three-value evaluation vocabulary.
type Effect string

// EffectAllow, EffectDeny, and EffectAsk are the grant/rule effects.
const (
	EffectAllow Effect = "allow"
	EffectDeny  Effect = "deny"
	EffectAsk   Effect = "ask"
)

// Grant is a standing permission row (Volume 2 Permission entity). A grant applies when it is
// unrevoked, unexpired, its permission equals the query's, its scope encloses the subject
// context, and its selector matches the query's resource.
type Grant struct {
	ID               core.ULID
	Permission       core.Permission
	Scope            core.PermissionScope
	Selector         string // resource pattern: "*", exact, or a "/prefix/**" glob
	Effect           Effect
	SubjectSession   core.ULID // set for session-scoped grants
	SubjectWorkspace core.ULID // set for workspace-scoped grants
	ValidUntil       string    // RFC 3339 UTC; empty = no expiry
	Revoked          bool
	CreatedAt        string
}

// active reports whether the grant is neither revoked nor expired as of now.
func (g Grant) active(now time.Time) bool {
	if g.Revoked {
		return false
	}
	if g.ValidUntil != "" {
		t, err := time.Parse(time.RFC3339, g.ValidUntil)
		if err == nil && now.After(t) {
			return false
		}
	}
	return true
}

// enclosesSubject reports whether the grant's scope covers the query's subject context.
func (g Grant) enclosesSubject(q ports.PermissionQuery) bool {
	switch g.Scope {
	case core.ScopeSession:
		return g.SubjectSession == "" || g.SubjectSession == q.SessionID
	case core.ScopeWorkspace:
		return g.SubjectWorkspace == "" || g.SubjectWorkspace == q.WorkspaceID
	default:
		// Resource-oriented scopes (path, host, tool, ...) are matched by selector only.
		return true
	}
}

// matches reports whether the grant applies to the query.
func (g Grant) matches(q ports.PermissionQuery, now time.Time) bool {
	return g.active(now) &&
		g.Permission == q.Permission &&
		g.enclosesSubject(q) &&
		matchSelector(g.Selector, q.Subject)
}

// matchSelector implements selector matching: "*" matches anything; a pattern ending in "/**"
// matches by path prefix; a pattern ending in "/*" matches one path segment beyond the prefix;
// otherwise an exact string match is required. An empty subject matches only "*".
func matchSelector(pattern, subject string) bool {
	switch {
	case pattern == "*":
		return true
	case strings.HasSuffix(pattern, "/**"):
		prefix := strings.TrimSuffix(pattern, "**")
		return strings.HasPrefix(subject, prefix) || subject == strings.TrimSuffix(prefix, "/")
	case strings.HasSuffix(pattern, "/*"):
		prefix := strings.TrimSuffix(pattern, "*")
		if !strings.HasPrefix(subject, prefix) {
			return false
		}
		rest := strings.TrimPrefix(subject, prefix)
		return rest != "" && !strings.Contains(rest, "/")
	default:
		return pattern == subject
	}
}

// resolve applies the precedence order (ADR-121) over the matching candidates and returns the
// evaluation outcome and the deciding grant IDs for the winning tier.
func resolve(candidates []Grant) (core.DecisionOutcome, []core.ULID) {
	var asks, allows []core.ULID
	var denies []core.ULID
	for _, c := range candidates {
		switch c.Effect {
		case EffectDeny:
			denies = append(denies, c.ID)
		case EffectAsk:
			asks = append(asks, c.ID)
		case EffectAllow:
			allows = append(allows, c.ID)
		}
	}
	switch {
	case len(denies) > 0:
		return core.OutcomeDeny, denies
	case len(asks) > 0:
		return core.OutcomeAsk, asks
	case len(allows) > 0:
		return core.OutcomeAllow, allows
	default:
		return core.OutcomeAsk, nil // default is ask (interactive) / deny (non-interactive)
	}
}
