package ontology

// classify maps a file (by base name and lowercased extension) to a human-readable language/format
// and a coarse kind: code | doc | config | data | asset | other. It is a static, deterministic
// lookup — no content inspection.
func classify(name, ext string) (language, kind string) {
	if lang, k, ok := byName(name); ok {
		return lang, k
	}
	if e, ok := byExt[ext]; ok {
		return e.lang, e.kind
	}
	return "", "other"
}

type langKind struct {
	lang string
	kind string
}

// byName handles files identified by their whole name rather than an extension.
func byName(name string) (language, kind string, ok bool) {
	switch name {
	case "Dockerfile", "Containerfile":
		return "Dockerfile", "config", true
	case "Makefile", "GNUmakefile":
		return "Makefile", "config", true
	case "LICENSE", "LICENSE.md", "COPYING", "NOTICE":
		return "License", "doc", true
	case ".gitignore", ".gitattributes", ".dockerignore", ".editorconfig":
		return "Git/tooling", "config", true
	case "go.mod", "go.sum":
		return "Go modules", "config", true
	}
	return "", "", false
}

// byExt is the extension → (language, kind) table.
var byExt = map[string]langKind{
	// code
	"go":    {"Go", "code"},
	"js":    {"JavaScript", "code"},
	"mjs":   {"JavaScript", "code"},
	"cjs":   {"JavaScript", "code"},
	"jsx":   {"JavaScript", "code"},
	"ts":    {"TypeScript", "code"},
	"tsx":   {"TypeScript", "code"},
	"py":    {"Python", "code"},
	"rs":    {"Rust", "code"},
	"java":  {"Java", "code"},
	"kt":    {"Kotlin", "code"},
	"rb":    {"Ruby", "code"},
	"php":   {"PHP", "code"},
	"c":     {"C", "code"},
	"h":     {"C", "code"},
	"cc":    {"C++", "code"},
	"cpp":   {"C++", "code"},
	"cxx":   {"C++", "code"},
	"hpp":   {"C++", "code"},
	"cs":    {"C#", "code"},
	"swift": {"Swift", "code"},
	"scala": {"Scala", "code"},
	"sh":    {"Shell", "code"},
	"bash":  {"Shell", "code"},
	"zsh":   {"Shell", "code"},
	"sql":   {"SQL", "code"},
	"lua":   {"Lua", "code"},
	"pl":    {"Perl", "code"},
	"r":     {"R", "code"},
	"dart":  {"Dart", "code"},
	// docs
	"md":       {"Markdown", "doc"},
	"markdown": {"Markdown", "doc"},
	"rst":      {"reStructuredText", "doc"},
	"txt":      {"Text", "doc"},
	"adoc":     {"AsciiDoc", "doc"},
	"org":      {"Org", "doc"},
	"pdf":      {"PDF", "doc"},
	// config
	"json": {"JSON", "config"},
	"yaml": {"YAML", "config"},
	"yml":  {"YAML", "config"},
	"toml": {"TOML", "config"},
	"ini":  {"INI", "config"},
	"cfg":  {"INI", "config"},
	"conf": {"Config", "config"},
	"env":  {"Env", "config"},
	"lock": {"Lockfile", "config"},
	// data
	"xml":     {"XML", "data"},
	"csv":     {"CSV", "data"},
	"tsv":     {"TSV", "data"},
	"ttl":     {"Turtle", "data"},
	"proto":   {"Protobuf", "data"},
	"graphql": {"GraphQL", "data"},
	// web / markup
	"html":   {"HTML", "code"},
	"htm":    {"HTML", "code"},
	"css":    {"CSS", "code"},
	"scss":   {"CSS", "code"},
	"sass":   {"CSS", "code"},
	"less":   {"CSS", "code"},
	"vue":    {"Vue", "code"},
	"svelte": {"Svelte", "code"},
	// assets
	"png":   {"Image", "asset"},
	"jpg":   {"Image", "asset"},
	"jpeg":  {"Image", "asset"},
	"gif":   {"Image", "asset"},
	"svg":   {"Image", "asset"},
	"webp":  {"Image", "asset"},
	"ico":   {"Image", "asset"},
	"woff":  {"Font", "asset"},
	"woff2": {"Font", "asset"},
	"ttf":   {"Font", "asset"},
	"mp4":   {"Video", "asset"},
	"mp3":   {"Audio", "asset"},
	"wav":   {"Audio", "asset"},
}
