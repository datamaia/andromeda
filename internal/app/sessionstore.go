package app

import (
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/datamaia/andromeda/internal/pal"
	"github.com/datamaia/andromeda/internal/ports"
)

// StoredSession is one persisted conversation. The transcript is provider-neutral (ports.Message),
// so a resumed session can continue under whichever provider the user selects.
type StoredSession struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	Provider  string          `json:"provider"`
	Model     string          `json:"model"`
	Mode      string          `json:"mode"`
	UpdatedAt string          `json:"updated_at"` // RFC3339
	Messages  []ports.Message `json:"messages"`
}

// sessionsDirFn resolves the sessions directory; overridable in tests. The default is the PAL data
// directory (XDG on Unix, %LOCALAPPDATA% on Windows) under a "sessions" subdirectory.
var sessionsDirFn = defaultSessionsDir

func defaultSessionsDir() (string, error) {
	base, err := pal.NewConfigDirs().DataHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "sessions"), nil
}

// NewSessionID mints a sortable, collision-resistant id: a UTC timestamp plus a short random suffix.
func NewSessionID() string {
	var b [3]byte
	_, _ = crand.Read(b[:])
	return time.Now().UTC().Format("20060102-150405") + "-" + hex.EncodeToString(b[:])
}

// SessionTitle derives a short title from the first user message.
func SessionTitle(messages []ports.Message) string {
	for _, m := range messages {
		if m.Role != "user" {
			continue
		}
		for _, p := range m.Parts {
			if (p.Type == "" || p.Type == "text") && strings.TrimSpace(p.Text) != "" {
				t := strings.TrimSpace(p.Text)
				if len(t) > 60 {
					t = t[:60] + "…"
				}
				return strings.ReplaceAll(t, "\n", " ")
			}
		}
	}
	return "(no prompt)"
}

// SaveSession persists a session atomically (temp file + rename), owner-only since transcripts may
// hold sensitive content. Persistence is best-effort: callers ignore the error when the store is
// unavailable so the session still runs.
func SaveSession(s StoredSession) error {
	dir, err := sessionsDirFn()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, s.ID+".json")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// LoadSession reads one session by id.
func LoadSession(id string) (StoredSession, error) {
	dir, err := sessionsDirFn()
	if err != nil {
		return StoredSession{}, err
	}
	data, err := os.ReadFile(filepath.Join(dir, id+".json")) //nolint:gosec // session id addresses a file in the app data dir
	if err != nil {
		return StoredSession{}, err
	}
	var s StoredSession
	if err := json.Unmarshal(data, &s); err != nil {
		return StoredSession{}, err
	}
	return s, nil
}

// ListSessions returns every stored session, newest first (by UpdatedAt).
func ListSessions() ([]StoredSession, error) {
	dir, err := sessionsDirFn()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []StoredSession
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		if s, err := LoadSession(id); err == nil {
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt > out[j].UpdatedAt })
	return out, nil
}

// LatestSessionID returns the most recently updated session's id, or "" when none exist.
func LatestSessionID() string {
	sessions, err := ListSessions()
	if err != nil || len(sessions) == 0 {
		return ""
	}
	return sessions[0].ID
}

// RemoveSession deletes a saved session by id.
func RemoveSession(id string) error {
	dir, err := sessionsDirFn()
	if err != nil {
		return err
	}
	return os.Remove(filepath.Join(dir, id+".json"))
}

// CountTurns returns the number of user turns in a stored conversation.
func CountTurns(messages []ports.Message) int {
	n := 0
	for _, m := range messages {
		if m.Role == "user" {
			n++
		}
	}
	return n
}
