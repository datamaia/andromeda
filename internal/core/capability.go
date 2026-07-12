package core

// Capability is a declared, machine-checkable ability of a provider/model (Principle 2).
// The runtime keys behavior off capabilities, never off model names. This is a closed
// enumeration (Volume 5, chapter 02); extending it requires an ADR.
type Capability string

// Capability enumeration (Volume 5, chapter 02); each value is the wire-stable identifier.
const (
	CapChat                Capability = "chat"
	CapStreaming           Capability = "streaming"
	CapToolCalling         Capability = "tool_calling"
	CapParallelToolCalling Capability = "parallel_tool_calling"
	CapStructuredOutputs   Capability = "structured_outputs"
	CapReasoning           Capability = "reasoning"
	CapVision              Capability = "vision"
	CapAudioInput          Capability = "audio_input"
	CapAudioOutput         Capability = "audio_output"
	CapEmbeddings          Capability = "embeddings"
	CapTokenUsageReporting Capability = "token_usage_reporting" //nolint:gosec // G101: capability identifier literal, not a credential
	CapCostReporting       Capability = "cost_reporting"
	CapModelDiscovery      Capability = "model_discovery"
	CapCancellation        Capability = "cancellation"
	CapTokenCounting       Capability = "token_counting" // ADR-056
)

// Capabilities is the set of capabilities a model declares.
type Capabilities []Capability

// Has reports whether c is present in the set.
func (cs Capabilities) Has(c Capability) bool {
	for _, x := range cs {
		if x == c {
			return true
		}
	}
	return false
}
