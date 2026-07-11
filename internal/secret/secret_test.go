package secret

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

func newFileStore(t *testing.T) *Store {
	t.Helper()
	fb, err := NewFileBackend(filepath.Join(t.TempDir(), "secrets.age"), "correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	return NewStore(fb)
}

func TestSetGetDelete(t *testing.T) {
	ctx := context.Background()
	s := newFileStore(t)
	ref := ports.SecretRef{Namespace: "providers", Name: "anthropic"}

	if err := s.Set(ctx, ref, ports.NewSecretValue([]byte("sk-abc")), ports.SecretMeta{Kind: "api_key", Provider: "anthropic"}); err != nil {
		t.Fatal(err)
	}
	v, err := s.Get(ctx, ref)
	if err != nil {
		t.Fatal(err)
	}
	if string(v.Bytes()) != "sk-abc" {
		t.Fatalf("got %q", v.Bytes())
	}
	if err := s.Delete(ctx, ref); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(ctx, ref); err == nil {
		t.Fatal("expected not-found after delete")
	}
}

func TestListReturnsRefsNotMaterial(t *testing.T) {
	ctx := context.Background()
	s := newFileStore(t)
	_ = s.Set(ctx, ports.SecretRef{Namespace: "providers", Name: "a"}, ports.NewSecretValue([]byte("1")), ports.SecretMeta{})
	_ = s.Set(ctx, ports.SecretRef{Namespace: "providers", Name: "b"}, ports.NewSecretValue([]byte("2")), ports.SecretMeta{})
	_ = s.Set(ctx, ports.SecretRef{Namespace: "other", Name: "c"}, ports.NewSecretValue([]byte("3")), ports.SecretMeta{})

	refs, err := s.List(ctx, ports.SecretScope{Namespace: "providers"})
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 2 {
		t.Fatalf("providers list = %d, want 2", len(refs))
	}
	for _, r := range refs {
		if r.Namespace != "providers" {
			t.Errorf("unexpected namespace %q", r.Namespace)
		}
	}
}

func TestGetMissingIsE_SEC_020(t *testing.T) {
	ctx := context.Background()
	s := newFileStore(t)
	_, err := s.Get(ctx, ports.SecretRef{Namespace: "x", Name: "nope"})
	var pe *ports.PortError
	if !errors.As(err, &pe) || pe.Code != "E-SEC-020" {
		t.Fatalf("want E-SEC-020, got %v", err)
	}
}

func TestFileBackendRequiresPassphrase(t *testing.T) {
	if _, err := NewFileBackend("x", ""); err == nil {
		t.Fatal("empty passphrase must be rejected")
	}
}

func TestEncryptedAtRest(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "secrets.age")
	fb, _ := NewFileBackend(path, "pw")
	s := NewStore(fb)
	_ = s.Set(ctx, ports.SecretRef{Namespace: "n", Name: "k"}, ports.NewSecretValue([]byte("TOPSECRET")), ports.SecretMeta{})

	// The on-disk blob must not contain the plaintext secret.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte("TOPSECRET")) {
		t.Fatal("plaintext secret found in the encrypted store file")
	}

	// A wrong passphrase cannot read it.
	wrong := NewStore(mustFileBackend(t, path, "wrong"))
	if _, err := wrong.Get(ctx, ports.SecretRef{Namespace: "n", Name: "k"}); err == nil {
		t.Fatal("wrong passphrase must not decrypt the store")
	}
}

func mustFileBackend(t *testing.T, path, pw string) *FileBackend {
	t.Helper()
	fb, err := NewFileBackend(path, pw)
	if err != nil {
		t.Fatal(err)
	}
	return fb
}
