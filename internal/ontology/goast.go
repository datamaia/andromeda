package ontology

import (
	"bufio"
	"bytes"
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// maxCodeFiles caps how many Go files the AST pass parses, so a pathological tree cannot blow up
// memory or output. When exceeded, the code model is marked Truncated (surfaced in the manifest and
// CLI) rather than silently capped. Files are processed in sorted order, so the cut is deterministic.
const maxCodeFiles = 6000

// CodeModel is the AST-level view of a workspace's Go code: its packages, their declared types and
// functions, and the resolved relationships between them (imports, calls, implements). It is derived
// deterministically from go/parser output and is emitted to code.ttl, separately from the structural
// project.ttl. All slices are sorted so serialization is reproducible.
type CodeModel struct {
	Root       string
	ModulePath string // module path from go.mod ("" if none — cross-package resolution is then limited)
	Packages   []*CodePackage
	Calls      []CallEdge // resolved, deduped function/package-qualified call edges (sorted)
	Implements []ImplEdge // structural interface-satisfaction edges within the module (sorted)
	Files      int        // number of Go files parsed
	Truncated  bool       // the file cap was hit
}

// CodePackage is one Go package (one source directory), with its declared types and functions.
type CodePackage struct {
	Dir        string   // workspace-relative directory ("" for the module root)
	ImportPath string   // canonical import path (module path + dir)
	Name       string   // package clause name
	Files      []string // workspace-relative .go files, sorted
	Imports    []string // import paths this package depends on (deduped, sorted)
	Types      []TypeDecl
	Funcs      []FuncDecl
}

// TypeDecl is a declared type. Methods holds interface method names (for interface kinds) so the
// implements pass can match against concrete types' method sets.
type TypeDecl struct {
	Name     string
	Kind     string // "struct" | "interface" | "type"
	Exported bool
	File     string
	Methods  []string // interface method names (interfaces only), sorted
}

// FuncDecl is a declared function or method (Recv != "" for methods).
type FuncDecl struct {
	Name     string
	Recv     string // receiver type name, "" for plain functions
	Exported bool
	File     string
	Params   int
	Results  int
}

// localID is the package-local identity of a function/method: "Recv.Name" for methods, "Name" for
// plain functions.
func (f FuncDecl) localID() string {
	if f.Recv != "" {
		return f.Recv + "." + f.Name
	}
	return f.Name
}

// CallEdge is a resolved call: (FromDir, FromFunc) calls (ToDir, ToFunc). Approx marks a cross-package
// edge inferred from a package selector (no full type resolution).
type CallEdge struct {
	FromDir, FromFunc string
	ToDir, ToFunc     string
	Approx            bool
}

// ImplEdge records that a concrete type structurally satisfies an interface (method-name containment).
type ImplEdge struct {
	TypeDir, TypeName   string
	IfaceDir, IfaceName string
}

// rawCall is a call site captured during parsing, resolved against the completed package tables in a
// second pass so no ASTs need to be retained.
type rawCall struct {
	fromDir, fromFunc string
	file              string
	sel               bool   // true: pkg.Name selector; false: bare Name identifier
	pkgLocal          string // local package name for selector calls
	name              string
}

// ScanCode parses the workspace's Go source into a CodeModel. It reuses the file list from a
// structural Model (only non-test .go files), grouping files by directory into packages, then
// resolves import, call, and interface-satisfaction edges. Parsing is best-effort: a file that fails
// to parse is skipped. Deterministic given identical inputs.
func ScanCode(_ context.Context, root string, m *Model) (*CodeModel, error) {
	modPath := modulePath(root)
	cm := &CodeModel{Root: root, ModulePath: modPath}

	// Collect the Go files to parse, in the model's sorted order, honoring the cap.
	var goFiles []string
	for _, f := range m.Files {
		if f.Ext != "go" || strings.HasSuffix(f.Name, "_test.go") {
			continue
		}
		if len(goFiles) >= maxCodeFiles {
			cm.Truncated = true
			break
		}
		goFiles = append(goFiles, f.Path)
	}

	fset := token.NewFileSet()
	pkgs := map[string]*CodePackage{} // dir -> package
	importSet := map[string]map[string]struct{}{}
	fileImports := map[string]map[string]string{}                // rel file -> localName -> import path
	methodsByType := map[string]map[string]map[string]struct{}{} // dir -> type -> method-name set
	var raws []rawCall

	pkgFor := func(dir string) *CodePackage {
		p := pkgs[dir]
		if p == nil {
			p = &CodePackage{Dir: dir, ImportPath: importPathFor(modPath, dir)}
			pkgs[dir] = p
			importSet[dir] = map[string]struct{}{}
			methodsByType[dir] = map[string]map[string]struct{}{}
		}
		return p
	}
	addMethod := func(dir, typ, method string) {
		set := methodsByType[dir][typ]
		if set == nil {
			set = map[string]struct{}{}
			methodsByType[dir][typ] = set
		}
		set[method] = struct{}{}
	}

	for _, rel := range goFiles {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		src, err := os.ReadFile(abs) //nolint:gosec // workspace-relative path from a git/walk enumeration
		if err != nil {
			continue
		}
		file, err := parser.ParseFile(fset, abs, src, parser.SkipObjectResolution)
		if err != nil {
			continue // best-effort: skip files that do not parse
		}
		dir := path0(rel)
		p := pkgFor(dir)
		if p.Name == "" {
			p.Name = file.Name.Name
		}
		p.Files = append(p.Files, rel)
		cm.Files++

		// Imports (package-level, deduped) and the file's local import-name map.
		fimp := map[string]string{}
		for _, imp := range file.Imports {
			ip, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			importSet[dir][ip] = struct{}{}
			local := importBaseName(ip)
			if imp.Name != nil {
				if imp.Name.Name == "_" || imp.Name.Name == "." {
					continue // blank/dot imports have no usable local selector
				}
				local = imp.Name.Name
			}
			fimp[local] = ip
		}
		fileImports[rel] = fimp

		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				if d.Tok != token.TYPE {
					continue
				}
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					td := TypeDecl{Name: ts.Name.Name, Exported: ts.Name.IsExported(), File: rel}
					switch it := ts.Type.(type) {
					case *ast.StructType:
						td.Kind = "struct"
					case *ast.InterfaceType:
						td.Kind = "interface"
						td.Methods = interfaceMethods(it)
					default:
						td.Kind = "type"
					}
					p.Types = append(p.Types, td)
				}
			case *ast.FuncDecl:
				fn := FuncDecl{
					Name:     d.Name.Name,
					Exported: d.Name.IsExported(),
					File:     rel,
					Params:   countFields(d.Type.Params),
					Results:  countFields(d.Type.Results),
				}
				if d.Recv != nil && len(d.Recv.List) > 0 {
					fn.Recv = receiverTypeName(d.Recv.List[0].Type)
					if fn.Recv != "" {
						addMethod(dir, fn.Recv, fn.Name)
					}
				}
				p.Funcs = append(p.Funcs, fn)
				raws = append(raws, collectCalls(dir, fn.localID(), rel, d.Body)...)
			}
		}
	}

	finalizePackages(cm, pkgs, importSet)
	cm.Calls = resolveCalls(raws, cm.Packages, fileImports)
	cm.Implements = resolveImplements(cm.Packages, methodsByType)
	return cm, nil
}

// finalizePackages sorts each package's members and imports, then assembles the sorted package slice.
func finalizePackages(cm *CodeModel, pkgs map[string]*CodePackage, importSet map[string]map[string]struct{}) {
	for dir, p := range pkgs {
		p.Imports = sortedSet(importSet[dir])
		sort.Strings(p.Files)
		sort.Slice(p.Types, func(i, j int) bool { return p.Types[i].Name < p.Types[j].Name })
		sort.Slice(p.Funcs, func(i, j int) bool { return p.Funcs[i].localID() < p.Funcs[j].localID() })
		cm.Packages = append(cm.Packages, p)
	}
	sort.Slice(cm.Packages, func(i, j int) bool { return cm.Packages[i].Dir < cm.Packages[j].Dir })
}

// resolveCalls turns raw call sites into deduped, sorted edges. Bare identifiers resolve to a plain
// function in the same package; package selectors resolve to an exported function in an imported
// module-internal package (marked Approx). Everything else — builtins, stdlib, value-method calls
// (which need full type resolution) — is dropped, which is what keeps the graph low-noise.
func resolveCalls(raws []rawCall, packages []*CodePackage, fileImports map[string]map[string]string) []CallEdge {
	plainFuncs := map[string]map[string]struct{}{}    // dir -> set of plain func names
	exportedFuncs := map[string]map[string]struct{}{} // dir -> set of exported plain func names
	dirByImport := map[string]string{}
	for _, p := range packages {
		plain := map[string]struct{}{}
		exp := map[string]struct{}{}
		for _, f := range p.Funcs {
			if f.Recv != "" {
				continue
			}
			plain[f.Name] = struct{}{}
			if f.Exported {
				exp[f.Name] = struct{}{}
			}
		}
		plainFuncs[p.Dir] = plain
		exportedFuncs[p.Dir] = exp
		dirByImport[p.ImportPath] = p.Dir
	}

	seen := map[string]struct{}{}
	var edges []CallEdge
	add := func(e CallEdge) {
		key := e.FromDir + "::" + e.FromFunc + "->" + e.ToDir + "::" + e.ToFunc
		if _, dup := seen[key]; dup {
			return
		}
		seen[key] = struct{}{}
		edges = append(edges, e)
	}

	for _, rc := range raws {
		if !rc.sel {
			if _, ok := plainFuncs[rc.fromDir][rc.name]; ok {
				add(CallEdge{FromDir: rc.fromDir, FromFunc: rc.fromFunc, ToDir: rc.fromDir, ToFunc: rc.name})
			}
			continue
		}
		ip := fileImports[rc.file][rc.pkgLocal]
		if ip == "" {
			continue
		}
		toDir, ok := dirByImport[ip]
		if !ok {
			continue // not a module-internal package we parsed
		}
		if _, ok := exportedFuncs[toDir][rc.name]; ok {
			add(CallEdge{FromDir: rc.fromDir, FromFunc: rc.fromFunc, ToDir: toDir, ToFunc: rc.name, Approx: true})
		}
	}
	sort.Slice(edges, func(i, j int) bool { return callKey(edges[i]) < callKey(edges[j]) })
	return edges
}

func callKey(e CallEdge) string {
	return e.FromDir + "::" + e.FromFunc + "->" + e.ToDir + "::" + e.ToFunc
}

// resolveImplements matches concrete (non-interface) types against interfaces by method-name
// containment, within the module. It is a deterministic structural heuristic (no signature checking),
// bounded to the interfaces and types actually declared, so it stays useful without exploding.
func resolveImplements(packages []*CodePackage, methodsByType map[string]map[string]map[string]struct{}) []ImplEdge {
	type iface struct {
		dir, name string
		methods   []string
	}
	var ifaces []iface
	for _, p := range packages {
		for _, t := range p.Types {
			if t.Kind == "interface" && len(t.Methods) > 0 {
				ifaces = append(ifaces, iface{p.Dir, t.Name, t.Methods})
			}
		}
	}
	var edges []ImplEdge
	for _, p := range packages {
		for _, t := range p.Types {
			if t.Kind == "interface" {
				continue
			}
			mset := methodsByType[p.Dir][t.Name]
			if len(mset) == 0 {
				continue
			}
			for _, i := range ifaces {
				if i.dir == p.Dir && i.name == t.Name {
					continue
				}
				if hasAll(mset, i.methods) {
					edges = append(edges, ImplEdge{TypeDir: p.Dir, TypeName: t.Name, IfaceDir: i.dir, IfaceName: i.name})
				}
			}
		}
	}
	sort.Slice(edges, func(i, j int) bool { return implKey(edges[i]) < implKey(edges[j]) })
	return edges
}

func implKey(e ImplEdge) string {
	return e.TypeDir + "::" + e.TypeName + "=>" + e.IfaceDir + "::" + e.IfaceName
}

func hasAll(set map[string]struct{}, names []string) bool {
	for _, n := range names {
		if _, ok := set[n]; !ok {
			return false
		}
	}
	return true
}

// collectCalls walks a function body and records bare-identifier and package-selector call sites.
func collectCalls(dir, fromFunc, file string, body *ast.BlockStmt) []rawCall {
	if body == nil {
		return nil
	}
	var out []rawCall
	ast.Inspect(body, func(n ast.Node) bool {
		ce, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := ce.Fun.(type) {
		case *ast.Ident:
			out = append(out, rawCall{fromDir: dir, fromFunc: fromFunc, file: file, name: fun.Name})
		case *ast.SelectorExpr:
			if x, ok := fun.X.(*ast.Ident); ok {
				out = append(out, rawCall{fromDir: dir, fromFunc: fromFunc, file: file, sel: true, pkgLocal: x.Name, name: fun.Sel.Name})
			}
		}
		return true
	})
	return out
}

// --- small AST/model helpers --------------------------------------------

func interfaceMethods(it *ast.InterfaceType) []string {
	var names []string
	if it.Methods == nil {
		return names
	}
	for _, f := range it.Methods.List {
		for _, nm := range f.Names { // named methods only; embedded interfaces are skipped
			names = append(names, nm.Name)
		}
	}
	sort.Strings(names)
	return names
}

func receiverTypeName(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.StarExpr:
		return receiverTypeName(t.X)
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr: // generic receiver T[P]
		return receiverTypeName(t.X)
	case *ast.IndexListExpr: // generic receiver T[P, Q]
		return receiverTypeName(t.X)
	}
	return ""
}

func countFields(fl *ast.FieldList) int {
	if fl == nil {
		return 0
	}
	n := 0
	for _, f := range fl.List {
		if len(f.Names) == 0 {
			n++
		} else {
			n += len(f.Names)
		}
	}
	return n
}

// importPathFor returns the canonical import path of a workspace directory given the module path.
func importPathFor(modPath, dir string) string {
	if modPath == "" {
		return dir
	}
	if dir == "" {
		return modPath
	}
	return modPath + "/" + dir
}

// importBaseName is the conventional local name of an import path (its last segment).
func importBaseName(ip string) string {
	if i := strings.LastIndexByte(ip, '/'); i >= 0 {
		return ip[i+1:]
	}
	return ip
}

// modulePath reads the module path from the workspace's go.mod, or "" when there is none.
func modulePath(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "go.mod")) //nolint:gosec // fixed file at the workspace root
	if err != nil {
		return ""
	}
	sc := bufio.NewScanner(bytes.NewReader(data))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if rest, ok := strings.CutPrefix(line, "module "); ok {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

func sortedSet(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
