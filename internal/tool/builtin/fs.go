// Package builtin provides the built-in tools (Volume 6, chapter 03). Each tool implements
// ports.ToolPort and, where its input names filesystem resources, tool.ResourceScoped so the
// Tool Runtime evaluates path-level permissions.
package builtin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/streams"
)

// ---- fs_read ----

// FSRead reads a UTF-8 text file. Phase: MVP.
type FSRead struct{}

func (FSRead) Describe(context.Context) (ports.ToolDescriptor, error) {
	return ports.ToolDescriptor{
		Name: "fs_read", Namespace: "fs", Version: "1", Description: "Read a text file",
		InputSchema:  []byte(`{"type":"object","required":["path"],"properties":{"path":{"type":"string"}}}`),
		OutputSchema: []byte(`{"type":"object","properties":{"content":{"type":"string"}}}`),
		Permissions:  []core.Permission{core.PermRead}, Origin: "builtin", TrustLevel: "trusted",
	}, nil
}

func (FSRead) Validate(_ context.Context, input ports.JSON) (ports.ValidationResult, error) {
	var in struct{ Path string }
	if err := json.Unmarshal(input, &in); err != nil || in.Path == "" {
		return ports.ValidationResult{Valid: false, Findings: []string{"path is required"}}, nil
	}
	return ports.ValidationResult{Valid: true}, nil
}

func (FSRead) Resources(input ports.JSON) ([]ports.PermissionQuery, error) {
	p := pathOf(input)
	return []ports.PermissionQuery{{Permission: core.PermRead, Scope: core.ScopePath, Subject: p}}, nil
}

func (FSRead) Execute(_ context.Context, req ports.ToolExecuteRequest) (ports.Stream[ports.ToolEvent], error) {
	p := pathOf(req.Input)
	data, err := os.ReadFile(p) //nolint:gosec // path is permission-checked by the Runtime
	if err != nil {
		return errEvent("could not read file: " + err.Error()), nil
	}
	out, _ := json.Marshal(map[string]string{"content": string(data)})
	return okEvent(string(out)), nil
}

func (FSRead) Cancel(context.Context, core.ULID) error { return nil }

// ---- fs_write ----

// FSWrite writes a text file, creating parent directories. Phase: MVP.
type FSWrite struct{}

func (FSWrite) Describe(context.Context) (ports.ToolDescriptor, error) {
	return ports.ToolDescriptor{
		Name: "fs_write", Namespace: "fs", Version: "1", Description: "Write a text file",
		InputSchema: []byte(`{"type":"object","required":["path","content"],"properties":{"path":{"type":"string"},"content":{"type":"string"}}}`),
		Permissions: []core.Permission{core.PermWrite}, Origin: "builtin", TrustLevel: "trusted",
	}, nil
}

func (FSWrite) Validate(_ context.Context, input ports.JSON) (ports.ValidationResult, error) {
	var in struct {
		Path    string
		Content *string
	}
	if err := json.Unmarshal(input, &in); err != nil || in.Path == "" || in.Content == nil {
		return ports.ValidationResult{Valid: false, Findings: []string{"path and content are required"}}, nil
	}
	return ports.ValidationResult{Valid: true}, nil
}

func (FSWrite) Resources(input ports.JSON) ([]ports.PermissionQuery, error) {
	return []ports.PermissionQuery{{Permission: core.PermWrite, Scope: core.ScopePath, Subject: pathOf(input)}}, nil
}

func (FSWrite) Execute(_ context.Context, req ports.ToolExecuteRequest) (ports.Stream[ports.ToolEvent], error) {
	var in struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	_ = json.Unmarshal(req.Input, &in)
	if err := os.MkdirAll(filepath.Dir(in.Path), 0o755); err != nil {
		return errEvent("mkdir failed: " + err.Error()), nil
	}
	if err := os.WriteFile(in.Path, []byte(in.Content), 0o644); err != nil { //nolint:gosec // permission-checked
		return errEvent("write failed: " + err.Error()), nil
	}
	return okEvent(`{"written":true}`), nil
}

func (FSWrite) Cancel(context.Context, core.ULID) error { return nil }

// ---- fs_search ----

// FSSearch does a case-insensitive substring search across text files under a root. Phase: MVP.
type FSSearch struct{}

func (FSSearch) Describe(context.Context) (ports.ToolDescriptor, error) {
	return ports.ToolDescriptor{
		Name: "fs_search", Namespace: "fs", Version: "1", Description: "Search files for a substring",
		InputSchema: []byte(`{"type":"object","required":["query"],"properties":{"query":{"type":"string"},"root":{"type":"string"}}}`),
		Permissions: []core.Permission{core.PermRead}, Origin: "builtin", TrustLevel: "trusted",
	}, nil
}

func (FSSearch) Validate(_ context.Context, input ports.JSON) (ports.ValidationResult, error) {
	var in struct{ Query string }
	if err := json.Unmarshal(input, &in); err != nil || in.Query == "" {
		return ports.ValidationResult{Valid: false, Findings: []string{"query is required"}}, nil
	}
	return ports.ValidationResult{Valid: true}, nil
}

func (FSSearch) Resources(input ports.JSON) ([]ports.PermissionQuery, error) {
	var in struct{ Root string }
	_ = json.Unmarshal(input, &in)
	root := in.Root
	if root == "" {
		root = "."
	}
	return []ports.PermissionQuery{{Permission: core.PermRead, Scope: core.ScopePath, Subject: root}}, nil
}

func (FSSearch) Execute(_ context.Context, req ports.ToolExecuteRequest) (ports.Stream[ports.ToolEvent], error) {
	var in struct {
		Query string `json:"query"`
		Root  string `json:"root"`
	}
	_ = json.Unmarshal(req.Input, &in)
	root := in.Root
	if root == "" {
		root = "."
	}
	needle := strings.ToLower(in.Query)
	var matches []string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == ".andromeda" {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := d.Info()
		if err != nil || info.Size() > 1<<20 {
			return nil
		}
		data, err := os.ReadFile(path) //nolint:gosec // permission-checked
		if err != nil {
			return nil
		}
		for i, line := range strings.Split(string(data), "\n") {
			if strings.Contains(strings.ToLower(line), needle) {
				matches = append(matches, path+":"+itoa(i+1))
			}
		}
		return nil
	})
	out, _ := json.Marshal(map[string]any{"matches": matches, "count": len(matches)})
	return okEvent(string(out)), nil
}

func (FSSearch) Cancel(context.Context, core.ULID) error { return nil }

// ---- helpers ----

func pathOf(input ports.JSON) string {
	var in struct {
		Path string `json:"path"`
	}
	_ = json.Unmarshal(input, &in)
	return in.Path
}

func okEvent(text string) ports.Stream[ports.ToolEvent] {
	return streams.Slice([]ports.ToolEvent{
		{Kind: "output", Text: text},
		{Kind: "terminal", Terminal: true, Outcome: "success", Text: text},
	})
}

func errEvent(text string) ports.Stream[ports.ToolEvent] {
	return streams.Slice([]ports.ToolEvent{{Kind: "terminal", Terminal: true, Outcome: "error", Text: text}})
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
