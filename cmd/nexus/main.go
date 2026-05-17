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

	opts := []tea.ProgramOption{tea.WithAltScreen()}

	// Git Bash (mintty) doesn't support Windows Console APIs that Bubbletea
	// uses by default on Windows. When MSYSTEM is set we're in a MINGW/MSYS2
	// environment, so open /dev/tty directly for proper PTY-based I/O.
	if os.Getenv("MSYSTEM") != "" {
		tty, ttyErr := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if ttyErr == nil {
			opts = append(opts, tea.WithInput(tty), tea.WithOutput(tty))
		}
	}

	p := tea.NewProgram(m, opts...)
	_, err = p.Run()
	return err
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
