package logging

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestRedactsSensitiveKeys(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, Options{Level: "info", Format: "json"})
	log.Info("auth", "api_key", "sk-secret-123", "user", "maia")

	var rec map[string]any
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatalf("log line is not JSON: %v (%s)", err, buf.String())
	}
	if rec["api_key"] != Redacted {
		t.Errorf("api_key = %v, want %s", rec["api_key"], Redacted)
	}
	if rec["user"] != "maia" {
		t.Errorf("non-sensitive field was altered: %v", rec["user"])
	}
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, Options{Level: "warn"})
	log.Info("should be filtered")
	if buf.Len() != 0 {
		t.Errorf("info log emitted at warn level: %s", buf.String())
	}
	log.Warn("kept")
	if buf.Len() == 0 {
		t.Error("warn log was dropped")
	}
}

func TestIsSensitiveKey(t *testing.T) {
	for _, k := range []string{"password", "API_KEY", "auth_token", "access_key_id"} {
		if !IsSensitiveKey(k) {
			t.Errorf("%q should be sensitive", k)
		}
	}
	if IsSensitiveKey("username") {
		t.Error("username should not be sensitive")
	}
}
