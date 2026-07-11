package storage

import "time"

// nowUTC returns the current time as an RFC 3339 UTC timestamp — the canonical timestamp
// format for stored rows (Volume 2, chapter 10).
func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}
