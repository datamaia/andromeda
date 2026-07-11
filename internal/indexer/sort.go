package indexer

import (
	"sort"

	"github.com/datamaia/andromeda/internal/ports"
)

func sortHits(hits []ports.IndexHit) {
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}
		return hits[i].Path < hits[j].Path
	})
}
