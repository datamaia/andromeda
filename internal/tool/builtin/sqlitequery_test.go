package builtin

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func makeDB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "data.db")
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.ExecContext(context.Background(),
		`CREATE TABLE items(id INTEGER PRIMARY KEY, name TEXT); INSERT INTO items(name) VALUES('a'),('b');`); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestSQLiteQueryRead(t *testing.T) {
	path := makeDB(t)
	outcome, text := runTool(t, SQLiteQuery{}, mustJSON(map[string]any{"database": path, "sql": "SELECT id, name FROM items ORDER BY id"}))
	if outcome != "success" {
		t.Fatalf("outcome = %s (%s)", outcome, text)
	}
	var res struct {
		Columns  []string `json:"columns"`
		Rows     [][]any  `json:"rows"`
		RowCount int      `json:"row_count"`
	}
	_ = json.Unmarshal([]byte(text), &res)
	if len(res.Columns) != 2 || res.RowCount != 2 || res.Rows[0][1] != "a" {
		t.Fatalf("result = %+v", res)
	}
}

func TestSQLiteQueryMutationRefusedWhenReadOnly(t *testing.T) {
	path := makeDB(t)
	// Default read_only=true: a DELETE is refused by classification before any DB work.
	if outcome, _ := runTool(t, SQLiteQuery{}, mustJSON(map[string]any{"database": path, "sql": "DELETE FROM items"})); outcome != "error" {
		t.Fatal("DELETE under default read_only should be refused")
	}
	// The row is still present.
	_, text := runTool(t, SQLiteQuery{}, mustJSON(map[string]any{"database": path, "sql": "SELECT count(*) FROM items"}))
	var res struct {
		Rows [][]any `json:"rows"`
	}
	_ = json.Unmarshal([]byte(text), &res)
	if len(res.Rows) == 0 || jsonNumber(res.Rows[0][0]) != 2 {
		t.Fatalf("rows should be intact: %s", text)
	}
}

func TestSQLiteQueryMutationAllowedWhenWritable(t *testing.T) {
	path := makeDB(t)
	outcome, text := runTool(t, SQLiteQuery{}, mustJSON(map[string]any{"database": path, "sql": "DELETE FROM items WHERE name='a'", "read_only": false}))
	if outcome != "success" {
		t.Fatalf("mutation with read_only=false should succeed: %s", text)
	}
	var res struct {
		RowCount int `json:"row_count"`
	}
	_ = json.Unmarshal([]byte(text), &res)
	if res.RowCount != 1 {
		t.Fatalf("row_count (affected) = %d, want 1", res.RowCount)
	}
}

// TestSQLiteQueryQueryOnlyBlocksHiddenWrite proves the DB-level query_only guard catches a
// mutation the leading-keyword classifier treats as a read (WITH ... DELETE).
func TestSQLiteQueryQueryOnlyBlocksHiddenWrite(t *testing.T) {
	path := makeDB(t)
	sqlText := "WITH x AS (SELECT id FROM items) DELETE FROM items WHERE id IN (SELECT id FROM x)"
	// Classified as read (starts with WITH), so it runs the query path — but query_only refuses it.
	if outcome, tx := runTool(t, SQLiteQuery{}, mustJSON(map[string]any{"database": path, "sql": sqlText})); outcome != "error" {
		t.Fatalf("hidden write should be blocked by query_only, got %s", tx)
	}
	_, text := runTool(t, SQLiteQuery{}, mustJSON(map[string]any{"database": path, "sql": "SELECT count(*) FROM items"}))
	var res struct {
		Rows [][]any `json:"rows"`
	}
	_ = json.Unmarshal([]byte(text), &res)
	if jsonNumber(res.Rows[0][0]) != 2 {
		t.Fatalf("rows should be intact after blocked write: %s", text)
	}
}

func TestSQLiteQueryRefusesStateDB(t *testing.T) {
	for _, name := range []string{"state.db", "global.db", "/proj/.andromeda/state.db"} {
		vr, _ := SQLiteQuery{}.Validate(context.Background(), []byte(`{"database":"`+name+`","sql":"SELECT 1"}`))
		if vr.Valid {
			t.Fatalf("%s should be refused as a state database", name)
		}
	}
}

func TestSQLiteQueryResourcesAddWriteForMutation(t *testing.T) {
	qs, _ := SQLiteQuery{}.Resources([]byte(`{"database":"d.db","sql":"UPDATE t SET x=1"}`))
	var write bool
	for _, q := range qs {
		if q.Permission == "write" {
			write = true
		}
	}
	if !write {
		t.Fatal("a mutating statement must request write")
	}
}

// jsonNumber coerces a JSON-decoded numeric cell (float64) to int for comparison.
func jsonNumber(v any) int {
	if f, ok := v.(float64); ok {
		return int(f)
	}
	return -1
}
