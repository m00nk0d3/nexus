package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/m00nk0d3/nexus/internal/data"
)

func run() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	m := NewModel()
	m.RepoPath = cwd

	// Open DB best-effort: non-fatal if it fails (e.g. no write permission).
	// When db is nil, agent runs are not logged but everything else works.
	dbPath := data.DefaultDBPath()
	// Ensure the parent directory exists before opening the DB.
	_ = os.MkdirAll(filepath.Dir(dbPath), 0o755)
	if db, err := data.NewDB(dbPath); err == nil {
		m.db = db
		defer db.Close()
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
