package core

import "testing"

func TestCapabilitiesHas(t *testing.T) {
	cs := Capabilities{CapChat, CapStreaming, CapToolCalling}
	if !cs.Has(CapStreaming) {
		t.Error("expected streaming to be present")
	}
	if cs.Has(CapVision) {
		t.Error("did not expect vision to be present")
	}
	if (Capabilities{}).Has(CapChat) {
		t.Error("empty set has nothing")
	}
}

func TestEnumValuesAreStable(t *testing.T) {
	// Guard the wire values that cross ports and land in stored/serialized data.
	cases := map[string]string{
		string(CapToolCalling):          "tool_calling",
		string(PermCredentialAccess):    "credential_access",
		string(ScopeWorkspace):          "workspace",
		string(DecisionAllowForSession): "allow_for_session",
		string(OutcomeDeny):             "deny",
		string(PhaseMVP):                "MVP",
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("enum value %q, want %q", got, want)
		}
	}
}
