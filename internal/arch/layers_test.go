package arch

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLayerDependencyRule enforces ADR-030/ADR-033: a package may import internal packages
// only at its own layer or lower; every internal package must be classified in the manifest.
// It parses imports directly (no build, no external tooling) so it runs everywhere.
func TestLayerDependencyRule(t *testing.T) {
	root := repoRoot(t)
	internalDir := filepath.Join(root, "internal")

	pkgLayer := PackageLayer
	fset := token.NewFileSet()

	err := filepath.WalkDir(internalDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		rel, _ := filepath.Rel(internalDir, path)
		pkgName := filepath.Dir(rel)
		// Only enforce top-level internal packages (no nested subpackages yet).
		if strings.Contains(pkgName, string(filepath.Separator)) {
			pkgName = strings.SplitN(pkgName, string(filepath.Separator), 2)[0]
		}
		selfLayer, ok := pkgLayer[pkgName]
		if !ok {
			t.Errorf("package %q is not classified in the layer manifest (arch.PackageLayer)", pkgName)
			return nil
		}

		f, perr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if perr != nil {
			return perr
		}
		for _, imp := range f.Imports {
			ip := strings.Trim(imp.Path.Value, `"`)
			if !strings.HasPrefix(ip, InternalPrefix) {
				continue
			}
			dep := strings.TrimPrefix(ip, InternalPrefix)
			dep = strings.SplitN(dep, "/", 2)[0]
			depLayer, ok := pkgLayer[dep]
			if !ok {
				t.Errorf("%s imports unclassified internal package %q", rel, dep)
				continue
			}
			if depLayer > selfLayer {
				t.Errorf("layer violation: %s (L%d) imports %q (L%d) — a package may not import a higher layer",
					rel, selfLayer, dep, depLayer)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found walking up from test dir")
		}
		dir = parent
	}
}
