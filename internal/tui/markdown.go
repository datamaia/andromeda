package tui

import (
	"regexp"
	"strings"
)

// renderMarkdown turns a subset of Markdown into styled terminal text: fenced code blocks (with a
// language tag and a distinct surface), ATX headings, bullet/numbered lists, blockquotes, inline
// `code`, and **bold**. It is deliberately lightweight (no external renderer) so it stays fast, matches
// the brand palette, and degrades to plain text when styling is off. Unknown syntax passes through.
func renderMarkdown(text string, st Styles) string {
	lines := strings.Split(text, "\n")
	var out []string
	inCode := false
	codeLang := ""
	var code []string

	flushCode := func() {
		if codeLang != "" {
			out = append(out, st.CodeLang.Render(codeLang))
		}
		for _, cl := range code {
			out = append(out, st.CodeBlock.Render(highlightCode(cl, st)))
		}
		code = nil
		codeLang = ""
	}

	for _, ln := range lines {
		trimmed := strings.TrimSpace(ln)
		if strings.HasPrefix(trimmed, "```") {
			if inCode {
				flushCode()
				inCode = false
			} else {
				inCode = true
				codeLang = strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
			}
			continue
		}
		if inCode {
			code = append(code, ln)
			continue
		}
		out = append(out, renderMarkdownLine(ln, st))
	}
	if inCode { // unterminated fence — render what we have
		flushCode()
	}
	return strings.Join(out, "\n")
}

var (
	reHeading  = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)
	reBullet   = regexp.MustCompile(`^(\s*)[-*+]\s+(.*)$`)
	reNumbered = regexp.MustCompile(`^(\s*)(\d+)\.\s+(.*)$`)
	reInline   = regexp.MustCompile("`([^`]+)`")
	reBold     = regexp.MustCompile(`\*\*([^*]+)\*\*`)
)

func renderMarkdownLine(ln string, st Styles) string {
	switch {
	case reHeading.MatchString(ln):
		m := reHeading.FindStringSubmatch(ln)
		return st.Heading.Render(m[2])
	case reBullet.MatchString(ln):
		m := reBullet.FindStringSubmatch(ln)
		return m[1] + st.Agent.Render("• ") + renderInline(m[2], st)
	case reNumbered.MatchString(ln):
		m := reNumbered.FindStringSubmatch(ln)
		return m[1] + st.Agent.Render(m[2]+". ") + renderInline(m[3], st)
	case strings.HasPrefix(strings.TrimSpace(ln), ">"):
		q := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(ln), ">"))
		return st.Muted.Render("│ " + q)
	default:
		return renderInline(ln, st)
	}
}

// renderInline applies inline `code` and **bold** styling. Inline code is styled first so bold
// markers inside a code span are left literal.
func renderInline(s string, st Styles) string {
	s = reInline.ReplaceAllStringFunc(s, func(m string) string {
		inner := strings.Trim(m, "`")
		return st.Code.Render(" " + inner + " ")
	})
	s = reBold.ReplaceAllStringFunc(s, func(m string) string {
		return st.Bold.Render(strings.Trim(m, "*"))
	})
	return s
}

var reComment = regexp.MustCompile(`^(\s*)(//|#)(.*)$`)

// highlightCode applies a light, overlap-free syntax highlight: whole-line comments are dimmed. Full
// per-token highlighting is intentionally omitted to avoid mangling code with interleaved ANSI codes.
func highlightCode(line string, st Styles) string {
	if m := reComment.FindStringSubmatch(line); m != nil {
		return m[1] + st.Comment.Render(m[2]+m[3])
	}
	return line
}
