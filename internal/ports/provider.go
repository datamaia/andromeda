package ports

import (
	"context"

	"github.com/datamaia/andromeda/internal/core"
)

// ProviderPort is the single contract every model provider adapter implements (Principle 1).
// One value represents one configured Provider; model selection travels in requests. The
// Provider Layer router also implements this port. Contract owner: Volume 5. Errors: E-PROV.
type ProviderPort interface {
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
	ChatStream(ctx context.Context, req ChatRequest) (Stream[ChatEvent], error)
	Embed(ctx context.Context, req EmbedRequest) (EmbedResponse, error)
	DiscoverModels(ctx context.Context) ([]ModelDescriptor, error)
	Capabilities(ctx context.Context, model string) (CapabilitySet, error)
	CountTokens(ctx context.Context, req TokenCountRequest) (TokenCount, error)
}

// CapabilitySet is the declared capability set for one model (Volume 5, closed enum).
type CapabilitySet = core.Capabilities

// ChatRequest is one inference request: messages, tool declarations, and parameters.
type ChatRequest struct {
	Model    string
	Messages []Message
	Tools    []ToolDeclaration
	Params   ModelParams
}

// ChatResponse is a complete non-streaming response.
type ChatResponse struct {
	Message   Message
	ToolCalls []ToolCall
	Usage     Usage
}

// ChatEvent is a tagged union delivered by ChatStream (content delta, tool-call delta,
// usage, terminal). Exactly one terminal event ends the stream with final usage.
type ChatEvent struct {
	Kind         string // "content" | "tool_call" | "usage" | "terminal"
	ContentDelta string
	ToolCall     *ToolCall
	Usage        *Usage
	Terminal     bool
}

// Message is one conversation message (role and typed content parts). Full shape: Volume 5.
type Message struct {
	Role  string
	Parts []ContentPart
}

// ContentPart is a piece of message content (text, image reference, etc.).
type ContentPart struct {
	Type string
	Text string
}

// ToolDeclaration advertises a tool to a provider for tool calling.
type ToolDeclaration struct {
	Name        string
	Description string
	InputSchema JSON
}

// ToolCall is a provider's request to invoke a tool.
type ToolCall struct {
	ID    string
	Name  string
	Input JSON
}

// ModelParams carries inference parameters (temperature, max tokens, etc.).
type ModelParams struct {
	Temperature *float64
	MaxTokens   *int
	Extra       map[string]any
}

// Usage is token/cost accounting for a request.
type Usage struct {
	InputTokens     int
	OutputTokens    int
	ReasoningTokens int
	CostBasis       string // e.g. "reported" | "estimated" | "unavailable"
}

// EmbedRequest is a batch embedding request.
type EmbedRequest struct {
	Model  string
	Inputs []string
}

// EmbedResponse carries the produced vectors and usage.
type EmbedResponse struct {
	Vectors [][]float32
	Usage   Usage
}

// ModelDescriptor describes a model a provider exposes.
type ModelDescriptor struct {
	ID           string
	DisplayName  string
	Capabilities CapabilitySet
	ContextLimit int
}

// TokenCountRequest asks for a token count of content against a model.
type TokenCountRequest struct {
	Model    string
	Messages []Message
}

// TokenCount is the counted number of tokens.
type TokenCount struct {
	Tokens int
}
