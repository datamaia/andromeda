package tui

import (
	"strings"
	"testing"
)

func TestRenderMarkdownStructures(t *testing.T) {
	st := DefaultStyles()
	in := strings.Join([]string{
		"# Title",
		"some **bold** and `code` here",
		"- first",
		"- second",
		"```go",
		"// a comment",
		"func main() {}",
		"```",
	}, "\n")
	out := renderMarkdown(in, st)

	// The heading text and list content survive (styling wraps them in ANSI, so match substrings).
	for _, want := range []string{"Title", "first", "second", "func main() {}", "a comment"} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered markdown missing %q:\n%s", want, out)
		}
	}
	// The bullet glyph replaced the "- " marker.
	if !strings.Contains(out, "•") {
		t.Error("bullets should render as •")
	}
	// The code fence markers themselves are consumed (not shown literally).
	if strings.Contains(out, "```") {
		t.Error("code fence markers should not appear in output")
	}
}

// A code block that never closes still renders its content (no dropped text).
func TestRenderMarkdownUnterminatedFence(t *testing.T) {
	out := renderMarkdown("intro\n```\nx := 1", DefaultStyles())
	if !strings.Contains(out, "x := 1") || !strings.Contains(out, "intro") {
		t.Fatalf("unterminated fence dropped content:\n%s", out)
	}
}

// Plain text passes through unharmed.
func TestRenderMarkdownPlain(t *testing.T) {
	if out := renderMarkdown("just a sentence.", DefaultStyles()); !strings.Contains(out, "just a sentence.") {
		t.Fatalf("plain text mangled: %q", out)
	}
}
