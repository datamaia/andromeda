package graph

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

//go:embed viewer.html
var viewerHTML []byte

// Serve starts a localhost HTTP server that renders the workspace graph in a self-contained,
// dependency-free force-directed viewer. It binds 127.0.0.1:<port> (port 0 picks a free port),
// invokes onReady with the resolved URL once listening, and serves until ctx is cancelled.
// graph.json is read live from the graph output directory, so re-running `andromeda graph` is
// reflected on the next refresh.
func Serve(ctx context.Context, root string, port int, onReady func(url string)) error {
	dir := Dir(root)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(viewerHTML)
	})
	mux.HandleFunc("/graph.json", func(w http.ResponseWriter, _ *http.Request) {
		data, err := os.ReadFile(filepath.Join(dir, "graph.json")) //nolint:gosec // fixed path under the workspace marker dir
		if err != nil {
			http.Error(w, "no graph yet — run `andromeda graph build`", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	})

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return err
	}
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	if onReady != nil {
		onReady("http://" + ln.Addr().String())
	}
	go func() {
		<-ctx.Done()
		_ = srv.Close()
	}()
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
