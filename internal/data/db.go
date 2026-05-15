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
	DB *sql.DB
}

// NewDB opens (or creates) a SQLite database at path, then runs all embedded
// SQL migrations in order. Use ":memory:" for an ephemeral in-process database.
func NewDB(path string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db := &DB{DB: sqlDB}
	if err := db.migrate(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

// Close releases the underlying database connection.
func (db *DB) Close() error {
	return db.DB.Close()
}

// migrate executes every *.sql file found in the embedded migrations directory,
// in lexicographic order (001_…, 002_…, …).
func (db *DB) migrate() error {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		if _, err := db.DB.Exec(string(content)); err != nil {
			return fmt.Errorf("exec migration %s: %w", entry.Name(), err)
		}
	}
	return nil
}
