package builtin

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite" // pure-Go SQLite driver (ADR-007), registered as "sqlite"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
)

// maxSQLiteRows bounds a result set (parity with other tools' truncation policy).
const maxSQLiteRows = 1000

// SQLiteQuery runs SQL against a SQLite database file in the workspace. Andromeda's own state
// databases are refused — their integrity is governed by ADR-028/ADR-029, not by agent SQL.
// Read statements require `read`; mutating statements (detected by classification, and only when
// read_only is false) additionally require `write`. Phase: Beta.
type SQLiteQuery struct{}

func (SQLiteQuery) Describe(context.Context) (ports.ToolDescriptor, error) {
	return ports.ToolDescriptor{
		Name: "sqlite_query", Namespace: "sqlite", Version: "1",
		Description: "Run SQL against a workspace SQLite database",
		InputSchema: []byte(`{"type":"object","required":["database","sql"],"properties":{` +
			`"database":{"type":"string"},"sql":{"type":"string"},"params":{"type":"array"},` +
			`"read_only":{"type":"boolean"}}}`),
		OutputSchema: []byte(`{"type":"object","properties":{"columns":{"type":"array"},"rows":{"type":"array"},"row_count":{"type":"integer"},"truncated":{"type":"boolean"}}}`),
		Permissions:  []core.Permission{core.PermRead, core.PermWrite}, Origin: "builtin", TrustLevel: "trusted",
	}, nil
}

type sqliteInput struct {
	Database string `json:"database"`
	SQL      string `json:"sql"`
	Params   []any  `json:"params"`
	ReadOnly *bool  `json:"read_only"`
}

func (in sqliteInput) readOnly() bool { return in.ReadOnly == nil || *in.ReadOnly }

func (SQLiteQuery) Validate(_ context.Context, input ports.JSON) (ports.ValidationResult, error) {
	var in sqliteInput
	if err := json.Unmarshal(input, &in); err != nil || in.Database == "" || strings.TrimSpace(in.SQL) == "" {
		return ports.ValidationResult{Valid: false, Findings: []string{"database and sql are required"}}, nil
	}
	if isStateDB(in.Database) {
		return ports.ValidationResult{Valid: false, Findings: []string{"Andromeda state databases cannot be queried directly"}}, nil
	}
	return ports.ValidationResult{Valid: true}, nil
}

func (SQLiteQuery) Resources(input ports.JSON) ([]ports.PermissionQuery, error) {
	var in sqliteInput
	_ = json.Unmarshal(input, &in)
	qs := []ports.PermissionQuery{{Permission: core.PermRead, Scope: core.ScopePath, Subject: in.Database}}
	if isMutation(in.SQL) {
		qs = append(qs, ports.PermissionQuery{Permission: core.PermWrite, Scope: core.ScopePath, Subject: in.Database})
	}
	return qs, nil
}

func (SQLiteQuery) Execute(ctx context.Context, req ports.ToolExecuteRequest) (ports.Stream[ports.ToolEvent], error) {
	var in sqliteInput
	_ = json.Unmarshal(req.Input, &in)
	if isStateDB(in.Database) {
		return errEvent("Andromeda state databases cannot be queried directly"), nil
	}
	mutation := isMutation(in.SQL)
	if mutation && in.readOnly() {
		return errEvent("statement modifies data but read_only is set; pass read_only=false to allow"), nil
	}

	// Defense in depth: when read_only, open with query_only(true) so any write the classifier
	// missed (e.g. a mutation hidden behind a CTE) is refused by the database itself.
	dsn := "file:" + in.Database + "?_pragma=busy_timeout(2000)"
	if in.readOnly() {
		dsn += "&_pragma=query_only(true)"
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return errEvent("could not open database: " + err.Error()), nil
	}
	defer db.Close()

	if mutation {
		res, err := db.ExecContext(ctx, in.SQL, in.Params...)
		if err != nil {
			return errEvent("exec failed: " + err.Error()), nil
		}
		affected, _ := res.RowsAffected()
		out, _ := json.Marshal(map[string]any{"columns": []string{}, "rows": [][]any{}, "row_count": affected, "truncated": false})
		return okEvent(string(out)), nil
	}

	rows, err := db.QueryContext(ctx, in.SQL, in.Params...)
	if err != nil {
		return errEvent("query failed: " + err.Error()), nil
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return errEvent("columns failed: " + err.Error()), nil
	}
	var data [][]any
	truncated := false
	for rows.Next() {
		if len(data) >= maxSQLiteRows {
			truncated = true
			break
		}
		cells := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range cells {
			ptrs[i] = &cells[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return errEvent("scan failed: " + err.Error()), nil
		}
		data = append(data, normalizeRow(cells))
	}
	if err := rows.Err(); err != nil {
		return errEvent("row iteration failed: " + err.Error()), nil
	}
	out, _ := json.Marshal(map[string]any{"columns": cols, "rows": data, "row_count": len(data), "truncated": truncated})
	return okEvent(string(out)), nil
}

func (SQLiteQuery) Cancel(context.Context, core.ULID) error { return nil }

// normalizeRow renders []byte cells (SQLite text/blobs) as strings so the JSON result is readable.
func normalizeRow(cells []any) []any {
	out := make([]any, len(cells))
	for i, c := range cells {
		if b, ok := c.([]byte); ok {
			out[i] = string(b)
		} else {
			out[i] = c
		}
	}
	return out
}

// isStateDB refuses Andromeda's own state databases by fixed filename or by residence under the
// .andromeda marker directory (ADR-028).
func isStateDB(path string) bool {
	base := filepath.Base(path)
	if base == "state.db" || base == "global.db" {
		return true
	}
	clean := filepath.ToSlash(filepath.Clean(path))
	return strings.Contains(clean, "/.andromeda/") || strings.HasPrefix(clean, ".andromeda/")
}

// isMutation classifies a statement by its leading keyword. SELECT/WITH/PRAGMA/EXPLAIN are reads;
// everything else (INSERT/UPDATE/DELETE/CREATE/DROP/ALTER/REPLACE/…) is a mutation.
func isMutation(sqlText string) bool {
	s := strings.TrimSpace(sqlText)
	// Skip a leading line comment if present.
	for strings.HasPrefix(s, "--") {
		if i := strings.IndexByte(s, '\n'); i >= 0 {
			s = strings.TrimSpace(s[i+1:])
		} else {
			s = ""
		}
	}
	kw := s
	if i := strings.IndexAny(kw, " \t\n("); i >= 0 {
		kw = kw[:i]
	}
	switch strings.ToUpper(kw) {
	case "SELECT", "WITH", "PRAGMA", "EXPLAIN", "":
		return false
	default:
		return true
	}
}
