package graph

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

// graphDataRE extracts the JSON embedded in the viewer's graph-data script.
var graphDataRE = regexp.MustCompile(`(?s)<script id="graph-data" type="application/json">(.*?)</script>`)

func TestDepthOf(t *testing.T) {
	cases := map[string]int{
		"":                      0, // project root
		"main.go":               1, // top-level file
		"internal":              1, // top-level dir
		"internal/util":         2,
		"internal/util/util.go": 3,
		"a/b/c/d.go":            4,
	}
	for path, want := range cases {
		if got := depthOf(path); got != want {
			t.Errorf("depthOf(%q) = %d, want %d", path, got, want)
		}
	}
}

func TestBuildPopulatesDepthAndLanguage(t *testing.T) {
	m, _ := sampleModel(t)
	g := Build(m)
	byID := map[string]Node{}
	for _, n := range g.Nodes {
		byID[n.ID] = n
	}
	if n := byID["project"]; n.Depth != 0 {
		t.Errorf("project depth = %d, want 0", n.Depth)
	}
	if n := byID["f/main.go"]; n.Depth != 1 || n.Language != "Go" {
		t.Errorf("main.go node = %+v, want depth 1 language Go", n)
	}
	if n := byID["f/internal/util/util.go"]; n.Depth != 3 || n.Language != "Go" {
		t.Errorf("util.go node = %+v, want depth 3 language Go", n)
	}
}

func TestRender3DEmbedsParseableJSON(t *testing.T) {
	m, _ := sampleModel(t)
	g := Build(m)
	g.populateSummaries(m.Root)
	html := string(render3D(g))

	match := graphDataRE.FindStringSubmatch(html)
	if match == nil {
		t.Fatal("no graph-data script found in rendered HTML")
	}
	// The placeholder must have been replaced.
	if strings.Contains(html, graphDataMarker) {
		t.Fatal("graph-data placeholder was not replaced")
	}
	var embedded Graph
	if err := json.Unmarshal([]byte(match[1]), &embedded); err != nil {
		t.Fatalf("embedded JSON does not parse: %v", err)
	}
	// HUD counts must equal the real graph.
	if len(embedded.Nodes) != len(g.Nodes) || len(embedded.Edges) != len(g.Edges) {
		t.Errorf("embedded counts (%d nodes, %d edges) != graph (%d, %d)",
			len(embedded.Nodes), len(embedded.Edges), len(g.Nodes), len(g.Edges))
	}
	if !strings.Contains(html, LayoutAlgorithm) {
		t.Errorf("rendered HTML does not name the layout algorithm %q", LayoutAlgorithm)
	}
}

func TestRender3DIsDeterministic(t *testing.T) {
	m, _ := sampleModel(t)
	a := render3D(Build(m))
	b := render3D(Build(m))
	if string(a) != string(b) {
		t.Fatal("render3D is not deterministic")
	}
}

func TestRender3DHasNoExternalReferences(t *testing.T) {
	m, _ := sampleModel(t)
	html := strings.ToLower(string(render3D(Build(m))))
	for _, bad := range []string{"http://", "https://", "cdn", "src=", "integrity=", "<link", "@import"} {
		if strings.Contains(html, bad) {
			t.Errorf("rendered viewer references external resource %q — it must be fully offline", bad)
		}
	}
}

func TestManifestReproducibilityFields(t *testing.T) {
	m, _ := sampleModel(t)
	g := Build(m)
	raw := g.manifest(g.JSON())

	var man map[string]any
	if err := json.Unmarshal(raw, &man); err != nil {
		t.Fatalf("manifest does not parse: %v", err)
	}
	for _, key := range []string{"graphHash", "layoutAlgorithm", "layoutHash", "generatorVersion"} {
		if v, ok := man[key]; !ok || v == "" {
			t.Errorf("manifest missing/empty %q", key)
		}
	}
	if man["layoutAlgorithm"] != LayoutAlgorithm {
		t.Errorf("layoutAlgorithm = %v, want %q", man["layoutAlgorithm"], LayoutAlgorithm)
	}
	// Deterministic: same graph → identical manifest bytes.
	if string(g.manifest(g.JSON())) != string(raw) {
		t.Fatal("manifest is not deterministic")
	}
}
