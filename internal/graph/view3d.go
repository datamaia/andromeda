package graph

import (
	_ "embed"
	"strings"
)

// LayoutAlgorithm names the deterministic node placement the 3D viewer applies to graph.json. It is
// recorded in the manifest so a caller can tell which layout produced a given view.
const LayoutAlgorithm = "depth-ring-3d-v1"

//go:embed graph3d.html
var view3DTemplate string

// graphDataToken is the placeholder inside graph3d.html's graph-data <script> that render3D replaces
// with the real graph JSON.
const graphDataToken = "__ANDROMEDA_GRAPH_DATA__"

// render3D produces the self-contained, offline 3D viewer by embedding the graph JSON into the
// template's graph-data <script>. It is deterministic: the same graph yields byte-for-byte identical
// HTML (graph.json is already deterministic and no timestamps are added).
//
// The JSON is inserted verbatim except that any "</" is rewritten to "<\/" so a string value can
// never terminate the surrounding <script>. Go's json encoder already escapes '<' and '>' to <
// /> by default, so this is belt-and-suspenders — but it guarantees the script content stays
// valid JSON that document.getElementById('graph-data') can JSON.parse without HTML-entity decoding.
func render3D(g *Graph) []byte {
	data := strings.ReplaceAll(string(g.JSON()), "</", "<\\/")
	return []byte(strings.Replace(view3DTemplate, graphDataToken, data, 1))
}
