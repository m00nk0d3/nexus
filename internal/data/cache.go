package data

import (
	"database/sql"
	"fmt"
	"time"
)

// IsCacheStale reports whether the most recent synced_at in the given table
// is older than ttl, or the table is empty (stale = needs refresh).
func IsCacheStale(db *DB, tableName string, ttl time.Duration) (bool, error) {
	var raw sql.NullString
	query := fmt.Sprintf("SELECT MAX(synced_at) FROM %s", tableName)
	if err := db.Conn.QueryRow(query).Scan(&raw); err != nil {
		return true, fmt.Errorf("is cache stale %s: %w", tableName, err)
	}
	if !raw.Valid || raw.String == "" {
		return true, nil // table is empty
	}
	t, err := time.Parse("2006-01-02 15:04:05", raw.String)
	if err != nil {
		return true, fmt.Errorf("is cache stale %s: parse time %q: %w", tableName, raw.String, err)
	}
	return time.Since(t) > ttl, nil
}
