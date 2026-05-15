package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func run() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	m := NewModel()
	m.RepoPath = cwd

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
