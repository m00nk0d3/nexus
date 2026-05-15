package data

import (
	"database/sql"
	"embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// DB wraps a SQLite connection and owns its lifecycle.
type DB struct {
	Conn *sql.DB
}

// NewDB opens (or creates) a SQLite database at path, then runs all embedded
// SQL migrations in order. Use ":memory:" for an ephemeral in-process database.
func NewDB(path string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db := &DB{Conn: sqlDB}
	if err := db.migrate(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

// Close releases the underlying database connection.
func (db *DB) Close() error {
	return db.Conn.Close()
}

// migrate executes every *.sql file found in the embedded migrations directory,
// in lexicographic order (001_…, 002_…, …). Each successfully applied migration
// is recorded in the schema_migrations table so it is never re-executed on
// subsequent startups.
func (db *DB) migrate() error {
	if _, err := db.Conn.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		filename   TEXT PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()

		var applied string
		err := db.Conn.QueryRow("SELECT filename FROM schema_migrations WHERE filename = ?", name).Scan(&applied)
		if err == nil {
			continue // already applied
		}
		if err != sql.ErrNoRows {
			return fmt.Errorf("check migration %s: %w", name, err)
		}

		content, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := db.Conn.Exec(string(content)); err != nil {
			return fmt.Errorf("exec migration %s: %w", name, err)
		}
		if _, err := db.Conn.Exec("INSERT INTO schema_migrations (filename) VALUES (?)", name); err != nil {
			return fmt.Errorf("record migration %s: %w", name, err)
		}
	}
	return nil
}
