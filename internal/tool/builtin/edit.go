package builtin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
)

// ---- fs_replace ----

// FSReplace does exact-match or regex replacement inside one file — the edit primitive. The match
// must be unique unless replace_all. Phase: MVP.
type FSReplace struct{}

// Describe returns the fs_replace tool descriptor.
func (FSReplace) Describe(context.Context) (ports.ToolDescriptor, error) {
	return ports.ToolDescriptor{
		Name: "fs_replace", Namespace: "fs", Version: "1", Description: "Replace text inside a file",
		InputSchema: []byte(`{"type":"object","required":["path","old","new"],"properties":{` +
			`"path":{"type":"string"},"old":{"type":"string"},"new":{"type":"string"},` +
			`"replace_all":{"type":"boolean"},"regex":{"type":"boolean"}}}`),
		OutputSchema: []byte(`{"type":"object","properties":{"path":{"type":"string"},"replacements":{"type":"integer"},"after_hash":{"type":"string"}}}`),
		Permissions:  []core.Permission{core.PermRead, core.PermWrite}, Origin: "builtin", TrustLevel: "trusted",
	}, nil
}

type replaceInput struct {
	Path       string `json:"path"`
	Old        string `json:"old"`
	New        string `json:"new"`
	ReplaceAll bool   `json:"replace_all"`
	Regex      bool   `json:"regex"`
}

// Validate requires path and old, and compiles the pattern when regex is set.
func (FSReplace) Validate(_ context.Context, input ports.JSON) (ports.ValidationResult, error) {
	var in replaceInput
	if err := json.Unmarshal(input, &in); err != nil || in.Path == "" || in.Old == "" {
		return ports.ValidationResult{Valid: false, Findings: []string{"path and old are required"}}, nil
	}
	if in.Regex {
		if _, err := regexp.Compile(in.Old); err != nil {
			return ports.ValidationResult{Valid: false, Findings: []string{"invalid regex: " + err.Error()}}, nil
		}
	}
	return ports.ValidationResult{Valid: true}, nil
}

// Resources requests read and write access to the target path.
func (FSReplace) Resources(input ports.JSON) ([]ports.PermissionQuery, error) {
	p := pathOf(input)
	return []ports.PermissionQuery{
		{Permission: core.PermRead, Scope: core.ScopePath, Subject: p},
		{Permission: core.PermWrite, Scope: core.ScopePath, Subject: p},
	}, nil
}

// Execute replaces the matched text in the file and returns the replacement count and new hash.
func (FSReplace) Execute(_ context.Context, req ports.ToolExecuteRequest) (ports.Stream[ports.ToolEvent], error) {
	var in replaceInput
	_ = json.Unmarshal(req.Input, &in)
	data, err := os.ReadFile(in.Path) //nolint:gosec // permission-checked by the Runtime
	if err != nil {
		return errEvent("could not read file: " + err.Error()), nil
	}
	src := string(data)

	var out string
	var count int
	if in.Regex {
		re, err := regexp.Compile(in.Old)
		if err != nil {
			return errEvent("invalid regex: " + err.Error()), nil
		}
		matches := re.FindAllStringIndex(src, -1)
		count = len(matches)
		if count == 0 {
			return errEvent("no match for pattern"), nil
		}
		if !in.ReplaceAll && count != 1 {
			return errEvent(fmt.Sprintf("pattern is not unique (%d matches); set replace_all", count)), nil
		}
		if in.ReplaceAll {
			out = re.ReplaceAllString(src, in.New)
		} else {
			out = src[:matches[0][0]] + in.New + src[matches[0][1]:]
			count = 1
		}
	} else {
		count = strings.Count(src, in.Old)
		if count == 0 {
			return errEvent("no match for old text"), nil
		}
		if !in.ReplaceAll && count != 1 {
			return errEvent(fmt.Sprintf("old text is not unique (%d matches); set replace_all", count)), nil
		}
		if in.ReplaceAll {
			out = strings.ReplaceAll(src, in.Old, in.New)
		} else {
			out = strings.Replace(src, in.Old, in.New, 1)
			count = 1
		}
	}

	if err := os.WriteFile(in.Path, []byte(out), 0o644); err != nil { //nolint:gosec // permission-checked
		return errEvent("write failed: " + err.Error()), nil
	}
	sum := sha256.Sum256([]byte(out))
	res, _ := json.Marshal(map[string]any{"path": in.Path, "replacements": count, "after_hash": hex.EncodeToString(sum[:])})
	return okEvent(string(res)), nil
}

// Cancel is a no-op; the replacement completes synchronously within Execute.
func (FSReplace) Cancel(context.Context, core.ULID) error { return nil }

// ---- fs_diff ----

// FSDiff computes a unified diff between two files (or a file and provided content). Phase: MVP.
type FSDiff struct{}

// Describe returns the fs_diff tool descriptor.
func (FSDiff) Describe(context.Context) (ports.ToolDescriptor, error) {
	return ports.ToolDescriptor{
		Name: "fs_diff", Namespace: "fs", Version: "1", Description: "Compute a unified diff",
		InputSchema: []byte(`{"type":"object","required":["left","right"],"properties":{` +
			`"left":{"type":"string"},"right":{"type":"string"},"right_content":{"type":"string"},"context_lines":{"type":"integer"}}}`),
		OutputSchema: []byte(`{"type":"object","properties":{"diff":{"type":"string"},"binary":{"type":"boolean"},"truncated":{"type":"boolean"}}}`),
		Permissions:  []core.Permission{core.PermRead}, Origin: "builtin", TrustLevel: "trusted",
	}, nil
}

type diffInput struct {
	Left         string  `json:"left"`
	Right        string  `json:"right"`
	RightContent *string `json:"right_content"`
	ContextLines *int    `json:"context_lines"`
}

// Validate requires left and either right or right_content.
func (FSDiff) Validate(_ context.Context, input ports.JSON) (ports.ValidationResult, error) {
	var in diffInput
	if err := json.Unmarshal(input, &in); err != nil || in.Left == "" {
		return ports.ValidationResult{Valid: false, Findings: []string{"left is required"}}, nil
	}
	if in.Right == "" && in.RightContent == nil {
		return ports.ValidationResult{Valid: false, Findings: []string{"right or right_content is required"}}, nil
	}
	return ports.ValidationResult{Valid: true}, nil
}

// Resources requests read access to the left file and, when supplied as a path, the right file.
func (FSDiff) Resources(input ports.JSON) ([]ports.PermissionQuery, error) {
	var in diffInput
	_ = json.Unmarshal(input, &in)
	qs := []ports.PermissionQuery{{Permission: core.PermRead, Scope: core.ScopePath, Subject: in.Left}}
	if in.Right != "" {
		qs = append(qs, ports.PermissionQuery{Permission: core.PermRead, Scope: core.ScopePath, Subject: in.Right})
	}
	return qs, nil
}

// Execute computes and returns the unified diff between the two inputs.
func (FSDiff) Execute(_ context.Context, req ports.ToolExecuteRequest) (ports.Stream[ports.ToolEvent], error) {
	var in diffInput
	_ = json.Unmarshal(req.Input, &in)
	leftData, err := os.ReadFile(in.Left) //nolint:gosec // permission-checked
	if err != nil {
		return errEvent("could not read left: " + err.Error()), nil
	}
	var rightData []byte
	rightName := in.Right
	if in.RightContent != nil {
		rightData = []byte(*in.RightContent)
		if rightName == "" {
			rightName = in.Left
		}
	} else {
		rightData, err = os.ReadFile(in.Right) //nolint:gosec // permission-checked
		if err != nil {
			return errEvent("could not read right: " + err.Error()), nil
		}
	}
	if isBinary(leftData) || isBinary(rightData) {
		out, _ := json.Marshal(map[string]any{"diff": "", "binary": true, "truncated": false})
		return okEvent(string(out)), nil
	}
	ctxLines := 3
	if in.ContextLines != nil {
		ctxLines = *in.ContextLines
	}
	aLines, _ := splitLines(string(leftData))
	bLines, _ := splitLines(string(rightData))
	diff := unifiedDiff(in.Left, rightName, aLines, bLines, ctxLines)
	out, _ := json.Marshal(map[string]any{"diff": diff, "binary": false, "truncated": false})
	return okEvent(string(out)), nil
}

// Cancel is a no-op; the diff is computed synchronously within Execute.
func (FSDiff) Cancel(context.Context, core.ULID) error { return nil }

// ---- fs_patch ----

// FSPatch applies a unified diff to the workspace atomically — all hunks or none. Phase: MVP.
type FSPatch struct{}

// Describe returns the fs_patch tool descriptor.
func (FSPatch) Describe(context.Context) (ports.ToolDescriptor, error) {
	return ports.ToolDescriptor{
		Name: "fs_patch", Namespace: "fs", Version: "1", Description: "Apply a unified diff atomically",
		InputSchema: []byte(`{"type":"object","required":["diff"],"properties":{` +
			`"diff":{"type":"string"},"check_only":{"type":"boolean"},"root":{"type":"string"}}}`),
		OutputSchema: []byte(`{"type":"object","properties":{"applied":{"type":"boolean"},"files":{"type":"array"},"rejected_hunks":{"type":"array"}}}`),
		Permissions:  []core.Permission{core.PermWrite}, Origin: "builtin", TrustLevel: "trusted",
	}, nil
}

type patchInput struct {
	Diff      string `json:"diff"`
	CheckOnly bool   `json:"check_only"`
	Root      string `json:"root"`
}

// Validate requires a non-empty, parseable unified diff.
func (FSPatch) Validate(_ context.Context, input ports.JSON) (ports.ValidationResult, error) {
	var in patchInput
	if err := json.Unmarshal(input, &in); err != nil || strings.TrimSpace(in.Diff) == "" {
		return ports.ValidationResult{Valid: false, Findings: []string{"diff is required"}}, nil
	}
	if _, err := parseUnified(in.Diff); err != nil {
		return ports.ValidationResult{Valid: false, Findings: []string{"invalid diff: " + err.Error()}}, nil
	}
	return ports.ValidationResult{Valid: true}, nil
}

// Resources requests write access to every file named in the diff, rooted at root.
func (FSPatch) Resources(input ports.JSON) ([]ports.PermissionQuery, error) {
	var in patchInput
	_ = json.Unmarshal(input, &in)
	files, err := parseUnified(in.Diff)
	if err != nil {
		return nil, err
	}
	qs := make([]ports.PermissionQuery, 0, len(files))
	for _, f := range files {
		qs = append(qs, ports.PermissionQuery{Permission: core.PermWrite, Scope: core.ScopePath, Subject: filepath.Join(in.Root, f.path)})
	}
	return qs, nil
}

// Execute applies the unified diff atomically, writing all patched files or none.
func (FSPatch) Execute(_ context.Context, req ports.ToolExecuteRequest) (ports.Stream[ports.ToolEvent], error) {
	var in patchInput
	_ = json.Unmarshal(req.Input, &in)
	files, err := parseUnified(in.Diff)
	if err != nil {
		return errEvent("invalid diff: " + err.Error()), nil
	}

	// Compute every file's patched content first; a single failing hunk rejects the whole patch.
	type pending struct {
		path    string
		content string
	}
	var results []map[string]any
	var toWrite []pending
	for _, f := range files {
		target := filepath.Join(in.Root, f.path)
		var src string
		if data, err := os.ReadFile(target); err == nil { //nolint:gosec // permission-checked
			src = string(data)
		} // a missing file is treated as empty (new-file creation)
		patched, err := applyHunks(src, f.hunks)
		if err != nil {
			out, _ := json.Marshal(map[string]any{"applied": false, "files": results, "rejected_hunks": []string{f.path + ": " + err.Error()}})
			return errEvent(string(out)), nil
		}
		results = append(results, map[string]any{"path": f.path, "op": "modify", "result": "ok"})
		toWrite = append(toWrite, pending{path: target, content: patched})
	}

	if in.CheckOnly {
		out, _ := json.Marshal(map[string]any{"applied": false, "files": results, "rejected_hunks": []string{}})
		return okEvent(string(out)), nil
	}

	for _, w := range toWrite {
		if err := os.MkdirAll(filepath.Dir(w.path), 0o750); err != nil {
			return errEvent("mkdir failed: " + err.Error()), nil
		}
		if err := os.WriteFile(w.path, []byte(w.content), 0o644); err != nil { //nolint:gosec // permission-checked
			return errEvent("write failed: " + err.Error()), nil
		}
	}
	out, _ := json.Marshal(map[string]any{"applied": true, "files": results, "rejected_hunks": []string{}})
	return okEvent(string(out)), nil
}

// Cancel is a no-op; the patch is applied synchronously within Execute.
func (FSPatch) Cancel(context.Context, core.ULID) error { return nil }

// isBinary reports whether data looks non-textual (contains a NUL byte in its head).
func isBinary(data []byte) bool {
	n := len(data)
	if n > 8000 {
		n = 8000
	}
	for i := 0; i < n; i++ {
		if data[i] == 0 {
			return true
		}
	}
	return false
}
