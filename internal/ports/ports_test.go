package ports

import (
	"errors"
	"testing"
)

func TestPortErrorErrorAndUnwrap(t *testing.T) {
	cause := errors.New("root cause")
	pe := &PortError{Code: "E-PROV-001", Detail: "connection refused", Cause: cause}
	if pe.Error() != "E-PROV-001: connection refused" {
		t.Errorf("unexpected Error(): %q", pe.Error())
	}
	if !errors.Is(pe, cause) {
		t.Error("expected errors.Is to find the wrapped cause")
	}

	// Falls back to the user message when Detail is empty.
	pe2 := &PortError{Code: "E-SEC-001", Message: "permission denied"}
	if pe2.Error() != "E-SEC-001: permission denied" {
		t.Errorf("unexpected Error(): %q", pe2.Error())
	}
}

func TestSecretValueZeroize(t *testing.T) {
	v := NewSecretValue([]byte("hunter2"))
	if string(v.Bytes()) != "hunter2" {
		t.Fatalf("Bytes() = %q", v.Bytes())
	}
	v.Zero()
	for _, b := range v.Bytes() {
		if b != 0 {
			t.Fatal("Zero() must clear all material bytes")
		}
	}
}
