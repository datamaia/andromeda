package builtin

import (
	"fmt"
	"strconv"
	"strings"
)

// This is a small, self-contained unified-diff engine shared by fs_diff (compute) and fs_patch
// (apply). It is line-based and workspace-local (no git, no network) so the MVP filesystem tools
// stay offline (FR-TOOL-007). The diff it emits round-trips through its own applier.

// splitLines splits content into lines, remembering whether it ended with a newline so the
// reconstruction is exact.
func splitLines(s string) (lines []string, finalNewline bool) {
	if s == "" {
		return nil, false
	}
	finalNewline = strings.HasSuffix(s, "\n")
	if finalNewline {
		s = s[:len(s)-1]
	}
	return strings.Split(s, "\n"), finalNewline
}

func joinLines(lines []string, finalNewline bool) string {
	s := strings.Join(lines, "\n")
	if finalNewline && len(lines) > 0 {
		s += "\n"
	}
	return s
}

type diffOp struct {
	tag  byte // 'e' equal, 'd' delete (in a), 'i' insert (from b)
	aIdx int
	bIdx int
}

// diffLines computes an edit sequence transforming a into b via a longest-common-subsequence
// table. Adequate for the bounded file sizes these tools accept.
func diffLines(a, b []string) []diffOp {
	n, m := len(a), len(b)
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}
	var ops []diffOp
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case a[i] == b[j]:
			ops = append(ops, diffOp{'e', i, j})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			ops = append(ops, diffOp{'d', i, j})
			i++
		default:
			ops = append(ops, diffOp{'i', i, j})
			j++
		}
	}
	for ; i < n; i++ {
		ops = append(ops, diffOp{'d', i, j})
	}
	for ; j < m; j++ {
		ops = append(ops, diffOp{'i', i, j})
	}
	return ops
}

// unifiedDiff renders a unified diff between a and b with the given context, using the standard
// `--- `/`+++ `/`@@` header shape (paths prefixed a/ and b/ like git). Returns "" when equal.
func unifiedDiff(aName, bName string, a, b []string, context int) string {
	if context < 0 {
		context = 3
	}
	ops := diffLines(a, b)
	changed := false
	for _, op := range ops {
		if op.tag != 'e' {
			changed = true
			break
		}
	}
	if !changed {
		return ""
	}

	var out strings.Builder
	fmt.Fprintf(&out, "--- a/%s\n+++ b/%s\n", aName, bName)

	// Group ops into hunks separated by runs of >2*context equal lines.
	type hunk struct{ ops []diffOp }
	var hunks []hunk
	var cur []diffOp
	equalRun := 0
	flush := func() {
		if len(cur) > 0 {
			hunks = append(hunks, hunk{cur})
			cur = nil
		}
	}
	for _, op := range ops {
		if op.tag == 'e' {
			equalRun++
		} else {
			equalRun = 0
		}
		cur = append(cur, op)
		if equalRun > 2*context && len(cur) > 0 {
			// close the current hunk, keeping trailing context
			flush()
		}
	}
	flush()

	for _, h := range hunks {
		// Trim leading/trailing equal lines beyond context.
		ops := trimHunk(h.ops, context)
		if len(ops) == 0 {
			continue
		}
		aStart, aLen, bStart, bLen := hunkBounds(ops)
		fmt.Fprintf(&out, "@@ -%s +%s @@\n", rangeStr(aStart, aLen), rangeStr(bStart, bLen))
		for _, op := range ops {
			switch op.tag {
			case 'e':
				out.WriteString(" " + a[op.aIdx] + "\n")
			case 'd':
				out.WriteString("-" + a[op.aIdx] + "\n")
			case 'i':
				out.WriteString("+" + b[op.bIdx] + "\n")
			}
		}
	}
	return out.String()
}

func trimHunk(ops []diffOp, context int) []diffOp {
	first, last := -1, -1
	for i, op := range ops {
		if op.tag != 'e' {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	if first < 0 {
		return nil
	}
	lo := first - context
	if lo < 0 {
		lo = 0
	}
	hi := last + context
	if hi >= len(ops) {
		hi = len(ops) - 1
	}
	return ops[lo : hi+1]
}

func hunkBounds(ops []diffOp) (aStart, aLen, bStart, bLen int) {
	aStart, bStart = -1, -1
	for _, op := range ops {
		if op.tag == 'e' || op.tag == 'd' {
			if aStart < 0 {
				aStart = op.aIdx
			}
			aLen++
		}
		if op.tag == 'e' || op.tag == 'i' {
			if bStart < 0 {
				bStart = op.bIdx
			}
			bLen++
		}
	}
	// Unified diff line numbers are 1-based; a zero-length side starts at the preceding index.
	if aStart < 0 {
		aStart = 0
	}
	if bStart < 0 {
		bStart = 0
	}
	return aStart + 1, aLen, bStart + 1, bLen
}

func rangeStr(start, length int) string {
	if length == 1 {
		return strconv.Itoa(start)
	}
	if length == 0 {
		return strconv.Itoa(start-1) + ",0"
	}
	return strconv.Itoa(start) + "," + strconv.Itoa(length)
}

// parsedHunk is one hunk extracted from a unified diff.
type parsedHunk struct {
	aStart int // 1-based
	lines  []string // each begins with ' ', '-', or '+'
}

// fileDiff is the target path (from the +++ header, stripping a b/ prefix) and hunks for one file.
type fileDiff struct {
	path  string
	hunks []parsedHunk
}

// parseUnified splits a (possibly multi-file) unified diff into per-file hunk sets.
func parseUnified(diff string) ([]fileDiff, error) {
	var files []fileDiff
	var cur *fileDiff
	var hunk *parsedHunk
	lines := strings.Split(diff, "\n")
	for idx := 0; idx < len(lines); idx++ {
		line := lines[idx]
		switch {
		case strings.HasPrefix(line, "--- "):
			// start of a new file; next line should be +++
			continue
		case strings.HasPrefix(line, "+++ "):
			if cur != nil {
				files = append(files, *cur)
			}
			path := strings.TrimPrefix(line, "+++ ")
			path = strings.TrimPrefix(path, "b/")
			cur = &fileDiff{path: strings.TrimSpace(path)}
			hunk = nil
		case strings.HasPrefix(line, "@@"):
			if cur == nil {
				return nil, fmt.Errorf("hunk before file header")
			}
			aStart, err := parseHunkHeader(line)
			if err != nil {
				return nil, err
			}
			cur.hunks = append(cur.hunks, parsedHunk{aStart: aStart})
			hunk = &cur.hunks[len(cur.hunks)-1]
		default:
			if hunk == nil {
				continue // preamble / blank line
			}
			if line == "" && idx == len(lines)-1 {
				continue // trailing newline artifact
			}
			if len(line) > 0 && (line[0] == ' ' || line[0] == '-' || line[0] == '+') {
				hunk.lines = append(hunk.lines, line)
			}
		}
	}
	if cur != nil {
		files = append(files, *cur)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no file headers found in diff")
	}
	return files, nil
}

// parseHunkHeader reads the -aStart of an @@ -a,b +c,d @@ header.
func parseHunkHeader(line string) (int, error) {
	fields := strings.Fields(line)
	if len(fields) < 3 || !strings.HasPrefix(fields[1], "-") {
		return 0, fmt.Errorf("malformed hunk header: %q", line)
	}
	spec := strings.TrimPrefix(fields[1], "-")
	if i := strings.IndexByte(spec, ','); i >= 0 {
		spec = spec[:i]
	}
	n, err := strconv.Atoi(spec)
	if err != nil {
		return 0, fmt.Errorf("malformed hunk header: %q", line)
	}
	return n, nil
}

// applyHunks applies a file's hunks to src, returning the patched content. It verifies that the
// context and deleted lines match; a mismatch returns an error naming the failing hunk so the
// caller can reject atomically.
func applyHunks(src string, hunks []parsedHunk) (string, error) {
	lines, finalNewline := splitLines(src)
	var out []string
	cursor := 0 // 0-based index into lines
	for hi, h := range hunks {
		start := h.aStart - 1
		if start < 0 {
			start = 0
		}
		if start > len(lines) {
			return "", fmt.Errorf("hunk %d starts past end of file", hi+1)
		}
		// Copy unchanged lines up to the hunk start.
		if start < cursor {
			return "", fmt.Errorf("hunk %d overlaps a previous hunk", hi+1)
		}
		out = append(out, lines[cursor:start]...)
		cursor = start
		for _, hl := range h.lines {
			op, text := hl[0], hl[1:]
			switch op {
			case ' ':
				if cursor >= len(lines) || lines[cursor] != text {
					return "", fmt.Errorf("hunk %d context mismatch at line %d", hi+1, cursor+1)
				}
				out = append(out, text)
				cursor++
			case '-':
				if cursor >= len(lines) || lines[cursor] != text {
					return "", fmt.Errorf("hunk %d delete mismatch at line %d", hi+1, cursor+1)
				}
				cursor++
			case '+':
				out = append(out, text)
			}
		}
	}
	out = append(out, lines[cursor:]...)
	return joinLines(out, finalNewline), nil
}
