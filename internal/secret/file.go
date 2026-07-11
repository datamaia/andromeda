package secret

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"filippo.io/age"
)

// FileBackend is the opt-in, explicitly lower-security fallback (ADR-014): secrets are stored
// as a single age-encrypted JSON map on disk, protected by a passphrase (age scrypt recipient)
// and 0600 file permissions. It is used only when no OS keychain is available and the user has
// enabled the fallback.
type FileBackend struct {
	path       string
	passphrase string
	mu         sync.Mutex
}

// NewFileBackend returns a file backend storing its encrypted blob at path, protected by
// passphrase. An empty passphrase is rejected — the fallback is never unencrypted.
func NewFileBackend(path, passphrase string) (*FileBackend, error) {
	if passphrase == "" {
		return nil, fmt.Errorf("secret: file fallback requires a passphrase")
	}
	return &FileBackend{path: path, passphrase: passphrase}, nil
}

func (f *FileBackend) load() (map[string]string, error) {
	data, err := os.ReadFile(f.path) //nolint:gosec // path is an internal, mode-0600 store
	if os.IsNotExist(err) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	id, err := age.NewScryptIdentity(f.passphrase)
	if err != nil {
		return nil, err
	}
	r, err := age.Decrypt(bytes.NewReader(data), id)
	if err != nil {
		return nil, fmt.Errorf("secret: decrypt fallback store: %w", err)
	}
	plain, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	m := map[string]string{}
	if len(plain) > 0 {
		if err := json.Unmarshal(plain, &m); err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (f *FileBackend) save(m map[string]string) error {
	rec, err := age.NewScryptRecipient(f.passphrase)
	if err != nil {
		return err
	}
	// The file fallback is the explicitly lower-security path (ADR-014); a moderate scrypt
	// work factor keeps interactive Set/Get responsive without a keychain.
	rec.SetWorkFactor(15)
	plain, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(f.path), 0o700); err != nil {
		return err
	}
	var buf bytes.Buffer
	w, err := age.Encrypt(&buf, rec)
	if err != nil {
		return err
	}
	if _, err := w.Write(plain); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return os.WriteFile(f.path, buf.Bytes(), 0o600)
}

func keyOf(service, account string) string { return service + "\x00" + account }

// Get returns the stored secret bytes for (service, account).
func (f *FileBackend) Get(service, account string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, err := f.load()
	if err != nil {
		return nil, err
	}
	v, ok := m[keyOf(service, account)]
	if !ok {
		return nil, ErrNotFound
	}
	return []byte(v), nil
}

// Set stores the secret bytes for (service, account).
func (f *FileBackend) Set(service, account string, secret []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, err := f.load()
	if err != nil {
		return err
	}
	m[keyOf(service, account)] = string(secret)
	return f.save(m)
}

// Delete removes the secret for (service, account).
func (f *FileBackend) Delete(service, account string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, err := f.load()
	if err != nil {
		return err
	}
	delete(m, keyOf(service, account))
	return f.save(m)
}

// ListAccounts returns the accounts stored under a service prefix.
func (f *FileBackend) ListAccounts(service string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, err := f.load()
	if err != nil {
		return nil, err
	}
	var out []string
	prefix := service + "\x00"
	for k := range m {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix {
			out = append(out, k[len(prefix):])
		}
	}
	return out, nil
}

// Available is always true for the file backend.
func (f *FileBackend) Available() bool { return true }
