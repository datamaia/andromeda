package ontology

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// codeTree builds a tiny two-package Go module exercising intra-package calls, a cross-package call,
// a dropped stdlib call, and an interface implementation.
func codeTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/testmod\n\ngo 1.21\n")
	writeFile(t, root, "a/a.go", `package a

import (
	"strings"

	"example.com/testmod/b"
)

type Reader interface {
	Read() string
}

func Foo() string {
	Bar()
	_ = strings.ToUpper("x")
	return b.Helper()
}

func Bar() string { return "bar" }
`)
	writeFile(t, root, "b/b.go", `package b

func Helper() string { return "help" }

type MyReader struct{}

func (m MyReader) Read() string { return "" }
`)
	// A test file must be ignored by the code graph.
	writeFile(t, root, "b/b_test.go", "package b\n\nfunc TestIgnored() {}\n")
	return root
}

func scanCode(t *testing.T, root string) *CodeModel {
	t.Helper()
	m, err := Scan(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	cm, err := ScanCode(context.Background(), root, m)
	if err != nil {
		t.Fatal(err)
	}
	return cm
}

func TestScanCodeModel(t *testing.T) {
	cm := scanCode(t, codeTree(t))
	if cm.ModulePath != "example.com/testmod" {
		t.Errorf("module path = %q", cm.ModulePath)
	}
	if len(cm.Packages) != 2 {
		t.Fatalf("packages = %d, want 2 (test files excluded): %+v", len(cm.Packages), cm.Packages)
	}
	s := cm.stats()
	if s.Types != 2 || s.Funcs != 4 {
		t.Errorf("types=%d funcs=%d, want 2 and 4", s.Types, s.Funcs)
	}
	if s.FilesParsed != 2 {
		t.Errorf("filesParsed=%d, want 2 (b_test.go excluded)", s.FilesParsed)
	}
}

func TestScanCodeResolvedCallsOnly(t *testing.T) {
	cm := scanCode(t, codeTree(t))
	var intra, cross int
	for _, e := range cm.Calls {
		if e.FromDir == "a" && e.FromFunc == "Foo" && e.ToDir == "a" && e.ToFunc == "Bar" && !e.Approx {
			intra++
		}
		if e.FromDir == "a" && e.FromFunc == "Foo" && e.ToDir == "b" && e.ToFunc == "Helper" && e.Approx {
			cross++
		}
		// The stdlib call strings.ToUpper must never resolve to an edge.
		if e.ToDir == "strings" || e.ToFunc == "ToUpper" {
			t.Errorf("unresolved/stdlib call leaked as edge: %+v", e)
		}
	}
	if intra != 1 {
		t.Errorf("intra-package call a.Foo->a.Bar count = %d, want 1", intra)
	}
	if cross != 1 {
		t.Errorf("cross-package call a.Foo->b.Helper count = %d, want 1", cross)
	}
}

func TestScanCodeImplements(t *testing.T) {
	cm := scanCode(t, codeTree(t))
	found := false
	for _, e := range cm.Implements {
		if e.TypeDir == "b" && e.TypeName == "MyReader" && e.IfaceDir == "a" && e.IfaceName == "Reader" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected b.MyReader implements a.Reader; got %+v", cm.Implements)
	}
}

func TestCodeTTLDeterministicAndWellFormed(t *testing.T) {
	cm := scanCode(t, codeTree(t))
	var a, b bytes.Buffer
	if err := writeCodeTTL(&a, cm); err != nil {
		t.Fatal(err)
	}
	if err := writeCodeTTL(&b, cm); err != nil {
		t.Fatal(err)
	}
	if a.String() != b.String() {
		t.Fatal("code.ttl is not deterministic")
	}
	ttl := a.String()
	for _, want := range []string{
		"am:Package a rdfs:Class",
		"am:calls a rdf:Property",
		"am:implements a rdf:Property",
		"<pkg/a> a am:Package ;",
		`am:importPath "example.com/testmod/a"`,
		"<type/a/Reader> a am:Interface ;",
		"<type/b/MyReader> a am:Struct ;",
		"<func/b/MyReader.Read> a am:Method ;",
		"am:receiver \"MyReader\"",
		"<func/a/Foo> am:calls <func/a/Bar> .",
		"<func/a/Foo> am:callsApprox <func/b/Helper> .",
		"<type/b/MyReader> am:implements <type/a/Reader> .",
		"am:definedIn <f/a/a.go>", // links code nodes back to structural file nodes
	} {
		if !strings.Contains(ttl, want) {
			t.Errorf("code.ttl missing %q", want)
		}
	}
	// No absolute paths / timestamps leak.
	if strings.Contains(ttl, cm.Root) {
		t.Error("code.ttl leaks the absolute workspace path")
	}
}

func TestGenerateWritesBothOntologies(t *testing.T) {
	root := codeTree(t)
	m, cm, err := Generate(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if m == nil || cm == nil {
		t.Fatal("Generate returned nil model(s)")
	}
	man := m.manifest(cm, "deadbeef")
	for _, want := range []string{`"code":`, `"packages": 2`, `"implements": 1`, `"codeHash": "deadbeef"`} {
		if !strings.Contains(string(man), want) {
			t.Errorf("manifest missing %q:\n%s", want, man)
		}
	}
	// Structure-only manifest keeps the original shape (no code section).
	if strings.Contains(string(m.manifest(nil, "")), `"code"`) {
		t.Error("structure-only manifest should not contain a code section")
	}
}
