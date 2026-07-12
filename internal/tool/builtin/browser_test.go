package builtin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockWebDriver answers the standard W3C WebDriver commands the tool issues.
func mockWebDriver(t *testing.T, seen *[]string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*seen = append(*seen, r.Method+" "+r.URL.Path)
		switch {
		case strings.HasSuffix(r.URL.Path, "/url"):
			w.Write([]byte(`{"value":null}`))
		case strings.HasSuffix(r.URL.Path, "/source"):
			w.Write([]byte(`{"value":"<html>hi</html>"}`))
		default:
			w.Write([]byte(`{"value":{}}`))
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestBrowserControlNavigateAndRead(t *testing.T) {
	var seen []string
	srv := mockWebDriver(t, &seen)
	tool := NewBrowserControl(srv.URL, srv.Client())

	if outcome, tx := runTool(t, tool, mustJSON(map[string]any{"operation": "navigate", "session": "S1", "url": "https://example.com"})); outcome != "success" {
		t.Fatalf("navigate failed: %s", tx)
	}
	outcome, text := runTool(t, tool, mustJSON(map[string]any{"operation": "read", "session": "S1"}))
	if outcome != "success" {
		t.Fatalf("read failed: %s", text)
	}
	var res struct {
		Content string `json:"content"`
	}
	_ = json.Unmarshal([]byte(text), &res)
	if !strings.Contains(res.Content, "hi") {
		t.Fatalf("content = %q", res.Content)
	}
	// The tool must speak the standard WebDriver session command paths.
	if len(seen) != 2 || !strings.Contains(seen[0], "/session/S1/url") || !strings.Contains(seen[1], "/session/S1/source") {
		t.Fatalf("webdriver calls = %v", seen)
	}
}

func TestBrowserControlUnconfigured(t *testing.T) {
	tool := NewBrowserControl("", nil)
	if outcome, _ := runTool(t, tool, `{"operation":"read","session":"S1"}`); outcome != "error" {
		t.Fatal("unconfigured browser.control should error")
	}
}

func TestBrowserControlValidation(t *testing.T) {
	tool := NewBrowserControl("http://localhost:4444", nil)
	if vr, _ := tool.Validate(t.Context(), []byte(`{"operation":"navigate","session":"S1"}`)); vr.Valid {
		t.Fatal("navigate without url should be invalid")
	}
}
