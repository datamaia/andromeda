package secret

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/datamaia/andromeda/internal/ports"
)

// ErrNotFound indicates a reference has no stored material.
var ErrNotFound = errors.New("secret: reference not found")

// Backend is the minimal credential-storage surface satisfied by both the PAL keychain store
// and the age FileBackend.
type Backend interface {
	Get(service, account string) ([]byte, error)
	Set(service, account string, secret []byte) error
	Delete(service, account string) error
	Available() bool
}

const (
	servicePrefix = "andromeda:"
	indexService  = "andromeda.__index__"
)

// Store implements ports.SecretStorePort over a Backend, maintaining a per-namespace reference
// index (with metadata) so List works even on keychains that cannot enumerate.
type Store struct {
	backend Backend
}

// NewStore returns a Secret Store over the given backend.
func NewStore(b Backend) *Store { return &Store{backend: b} }

var _ ports.SecretStorePort = (*Store)(nil)

func svc(ns string) string { return servicePrefix + ns }

// Get resolves a reference to material.
func (s *Store) Get(ctx context.Context, ref ports.SecretRef) (ports.SecretValue, error) {
	if err := ctx.Err(); err != nil {
		return ports.SecretValue{}, err
	}
	b, err := s.backend.Get(svc(ref.Namespace), ref.Name)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ports.SecretValue{}, secErr("E-SEC-020", "secret not found", err)
		}
		return ports.SecretValue{}, secErr("E-SEC-021", "secret backend error", err)
	}
	return ports.NewSecretValue(b), nil
}

// Set creates or replaces material under a reference, with metadata.
func (s *Store) Set(ctx context.Context, ref ports.SecretRef, value ports.SecretValue, meta ports.SecretMeta) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := s.backend.Set(svc(ref.Namespace), ref.Name, value.Bytes()); err != nil {
		return secErr("E-SEC-021", "secret backend error", err)
	}
	return s.indexPut(ref, meta)
}

// Delete removes material and its index entry.
func (s *Store) Delete(ctx context.Context, ref ports.SecretRef) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := s.backend.Delete(svc(ref.Namespace), ref.Name); err != nil && !errors.Is(err, ErrNotFound) {
		return secErr("E-SEC-021", "secret backend error", err)
	}
	return s.indexDelete(ref)
}

// List enumerates references (never material) in a scope.
func (s *Store) List(ctx context.Context, scope ports.SecretScope) ([]ports.SecretRef, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	idx, err := s.loadIndex(scope.Namespace)
	if err != nil {
		return nil, secErr("E-SEC-021", "secret index error", err)
	}
	out := make([]ports.SecretRef, 0, len(idx))
	for name := range idx {
		out = append(out, ports.SecretRef{Namespace: scope.Namespace, Name: name})
	}
	return out, nil
}

// --- reference index (per namespace: name -> metadata) ---

func (s *Store) loadIndex(ns string) (map[string]ports.SecretMeta, error) {
	raw, err := s.backend.Get(indexService, ns)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return map[string]ports.SecretMeta{}, nil
		}
		return nil, err
	}
	m := map[string]ports.SecretMeta{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (s *Store) saveIndex(ns string, m map[string]ports.SecretMeta) error {
	raw, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return s.backend.Set(indexService, ns, raw)
}

func (s *Store) indexPut(ref ports.SecretRef, meta ports.SecretMeta) error {
	m, err := s.loadIndex(ref.Namespace)
	if err != nil {
		return err
	}
	m[ref.Name] = meta
	return s.saveIndex(ref.Namespace, m)
}

func (s *Store) indexDelete(ref ports.SecretRef) error {
	m, err := s.loadIndex(ref.Namespace)
	if err != nil {
		return err
	}
	delete(m, ref.Name)
	return s.saveIndex(ref.Namespace, m)
}

// secErr builds a redacted PortError in the E-SEC family (never carrying secret material).
func secErr(code, msg string, cause error) error {
	return &ports.PortError{Code: code, Category: "security", Severity: "error", Message: msg, Cause: cause}
}
