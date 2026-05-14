package data

// Database represents the SQLite database connection
type Database struct {
	path string
}

// NewDatabase creates a new database instance
func NewDatabase(path string) *Database {
	return &Database{path: path}
}
