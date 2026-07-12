package memory

import (
	"database/sql"
	"sort"
	"strings"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
)

func nz(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

func toArgs(ids []core.ULID) []any {
	out := make([]any, len(ids))
	for i, id := range ids {
		out[i] = id
	}
	return out
}

func scanRecords(rows *sql.Rows) ([]ports.MemoryRecord, error) {
	var out []ports.MemoryRecord
	for rows.Next() {
		var r ports.MemoryRecord
		var prov, src *string
		if err := rows.Scan(&r.ID, &r.Layer, &r.Content, &prov, &src, &r.Status, &r.CreatedAt); err != nil {
			return nil, err
		}
		if prov != nil {
			r.Provenance = *prov
		}
		if src != nil {
			r.Source = *src
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func tokenize(s string) map[string]struct{} {
	set := map[string]struct{}{}
	for _, f := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return r != '_' && (r < 'a' || r > 'z') && (r < '0' || r > '9')
	}) {
		if len(f) > 1 {
			set[f] = struct{}{}
		}
	}
	return set
}

func overlapScore(a, b map[string]struct{}) float64 {
	if len(a) == 0 {
		return 0
	}
	var hits int
	for t := range a {
		if _, ok := b[t]; ok {
			hits++
		}
	}
	return float64(hits) / float64(len(a))
}

func sortByScore(r []ports.RankedMemory) {
	sort.SliceStable(r, func(i, j int) bool { return r[i].Score > r[j].Score })
}
