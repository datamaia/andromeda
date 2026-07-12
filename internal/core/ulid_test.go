package core

import (
	"sort"
	"strings"
	"testing"
)

func TestNewULIDShape(t *testing.T) {
	id := NewULID()
	if len(id) != 26 {
		t.Fatalf("ULID length = %d, want 26 (%q)", len(id), id)
	}
	for _, c := range id {
		if !strings.ContainsRune(crockford, c) {
			t.Fatalf("ULID %q contains non-Crockford char %q", id, c)
		}
	}
}

func TestULIDUniquenessAndMonotonicity(t *testing.T) {
	const n = 10000
	ids := make([]string, n)
	seen := make(map[string]struct{}, n)
	for i := range ids {
		ids[i] = NewULID()
		if _, dup := seen[ids[i]]; dup {
			t.Fatalf("duplicate ULID at %d: %q", i, ids[i])
		}
		seen[ids[i]] = struct{}{}
	}
	// Generated in order, they must already be sorted (monotonic within a process).
	if !sort.StringsAreSorted(ids) {
		t.Error("ULIDs generated in sequence are not lexicographically sorted")
	}
}

func TestULIDMonotonicSameMillisecond(t *testing.T) {
	a := ulidGen.next(1000)
	b := ulidGen.next(1000)
	if a >= b {
		t.Errorf("same-ms ULIDs not monotonic: %q >= %q", a, b)
	}
}
