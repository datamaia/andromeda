package app

// Provider catalog (Volume 5). A single source of truth for the model providers Andromeda can
// talk to, shared by the CLI (`provider list`, `run`/`tui` --provider) and the TUI provider menu.
// Most entries are OpenAI Chat Completions-compatible surfaces that differ only in base URL and
// the environment variable that holds the API key, so wiring a new one is a data change, not code.
//
// Endpoints and the environment-variable conventions follow the maintainer's provider reference.
// Candidate default models are conservative, verified names; the live catalogue is authoritative
// and is surfaced by `andromeda model list` / the menu's discovery.

// ProviderKind selects which adapter builds the provider.
type ProviderKind int

const (
	// KindOpenAICompat is the generic OpenAI Chat Completions surface (base URL + bearer key).
	KindOpenAICompat ProviderKind = iota
	// KindAnthropic is the Anthropic Messages API.
	KindAnthropic
	// KindOllama is a local Ollama daemon (no key).
	KindOllama
	// KindOpenAIChatGPT is a ChatGPT-subscription OAuth session against the Codex backend.
	KindOpenAIChatGPT
)

// ProviderInfo describes one catalog entry.
type ProviderInfo struct {
	ID           string       // stable identifier used on the CLI and in config
	Display      string       // human label for menus
	Kind         ProviderKind // which adapter to build
	BaseURL      string       // default endpoint (OpenAI-compatible entries); "" uses the adapter default
	KeyEnv       string       // environment variable holding the API key ("" = no key needed)
	KeyRequired  bool         // whether a key is mandatory (local servers are not)
	Local        bool         // runs on the user's machine
	Reasoning    bool         // exposes configurable reasoning/effort (model-dependent)
	DefaultModel string       // a sensible starting model ("" = user must choose / discover)
	Note         string       // short description
	Models       []string     // curated model list, surfaced when live /models discovery is unavailable
}

// catalog is ordered for display: local first, then hosted key-based providers.
var catalog = []ProviderInfo{
	{
		ID: "ollama", Display: "Ollama (local)", Kind: KindOllama,
		BaseURL: "http://localhost:11434", Local: true, Reasoning: true,
		DefaultModel: "llama3.2", Note: "local models, no API key",
	},
	{
		ID: "vllm", Display: "vLLM (local)", Kind: KindOpenAICompat,
		BaseURL: "http://localhost:8000/v1", KeyEnv: "VLLM_API_KEY", Local: true, Reasoning: true,
		Note: "self-hosted OpenAI-compatible server",
	},
	{
		ID: "openai", Display: "OpenAI", Kind: KindOpenAICompat,
		BaseURL: "https://api.openai.com/v1", KeyEnv: "OPENAI_API_KEY", KeyRequired: true, Reasoning: true,
		Note: "OpenAI API (API key)",
	},
	{
		ID: "anthropic", Display: "Anthropic", Kind: KindAnthropic,
		KeyEnv: "ANTHROPIC_API_KEY", KeyRequired: true, Reasoning: true,
		Note: "Anthropic Messages API (API key)",
	},
	{
		ID: "openai-chatgpt", Display: "ChatGPT (subscription)", Kind: KindOpenAIChatGPT,
		Reasoning: true, DefaultModel: "gpt-5.5",
		Note: "sign in with your ChatGPT account: andromeda auth login openai-chatgpt",
	},
	{
		ID: "gemini", Display: "Google AI Studio (Gemini)", Kind: KindOpenAICompat,
		BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai", KeyEnv: "GEMINI_API_KEY",
		KeyRequired: true, Reasoning: true, DefaultModel: "gemini-2.0-flash",
		Note: "Gemini via its OpenAI-compatible layer",
	},
	{
		ID: "groq", Display: "Groq", Kind: KindOpenAICompat,
		BaseURL: "https://api.groq.com/openai/v1", KeyEnv: "GROQ_API_KEY", KeyRequired: true, Reasoning: true,
		DefaultModel: "llama-3.3-70b-versatile", Note: "GroqCloud (fast inference)",
	},
	{
		ID: "cerebras", Display: "Cerebras", Kind: KindOpenAICompat,
		BaseURL: "https://api.cerebras.ai/v1", KeyEnv: "CEREBRAS_API_KEY", KeyRequired: true, Reasoning: true,
		DefaultModel: "gpt-oss-120b", Note: "Cerebras Inference",
	},
	{
		ID: "openrouter", Display: "OpenRouter", Kind: KindOpenAICompat,
		BaseURL: "https://openrouter.ai/api/v1", KeyEnv: "OPENROUTER_API_KEY", KeyRequired: true, Reasoning: true,
		DefaultModel: "openai/gpt-oss-120b", Note: "aggregator, 400+ models",
	},
	{
		ID: "xai", Display: "xAI (Grok)", Kind: KindOpenAICompat,
		BaseURL: "https://api.x.ai/v1", KeyEnv: "XAI_API_KEY", KeyRequired: true, Reasoning: true,
		Note: "xAI Grok (API key)",
	},
	{
		ID: "huggingface", Display: "Hugging Face", Kind: KindOpenAICompat,
		BaseURL: "https://router.huggingface.co/v1", KeyEnv: "HF_TOKEN", KeyRequired: true,
		Note: "Inference Providers router",
	},
	{
		// AWS Kiro (the agentic IDE, successor to Amazon Q / CodeWhisperer) has no official public
		// model-invocation API; its models run on the AWS CodeWhisperer/Q backend behind a Kiro / AWS
		// Builder ID sign-in. In practice they are reached from third-party tools through a local
		// OpenAI-compatible gateway (e.g. kiro-gateway on :8000) that presents Kiro's session as
		// /v1/chat/completions. This entry points at that gateway; it is unofficial and depends on the
		// user running the gateway with their own Kiro credentials. Amazon Bedrock is the official
		// alternative for the same Claude/Nova models.
		ID: "kiro", Display: "AWS Kiro (gateway)", Kind: KindOpenAICompat,
		BaseURL: "http://localhost:8000/v1", KeyEnv: "KIRO_API_KEY", Local: true, Reasoning: true,
		DefaultModel: "claude-sonnet-4-5",
		Models: []string{
			"claude-sonnet-4-5", "claude-haiku-4-5", "claude-sonnet-4",
			"glm-5", "deepseek-v3.2", "minimax-m2.5", "minimax-m2.1", "qwen3-coder-next",
		},
		Note: "AWS Kiro models via a local OpenAI-compatible gateway (unofficial; needs kiro-gateway + your Kiro sign-in)",
	},
}

// Providers returns the catalog in display order.
func Providers() []ProviderInfo { return catalog }

// LookupProvider finds a catalog entry by ID.
func LookupProvider(id string) (ProviderInfo, bool) {
	for _, p := range catalog {
		if p.ID == id {
			return p, true
		}
	}
	return ProviderInfo{}, false
}
