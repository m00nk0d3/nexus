package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the root Bubbletea model for the Nexus TUI application.
// It holds the current state of the application.
type Model struct {
	// Add fields here as the application grows
}

// NewModel creates and returns a new Model instance with all required fields initialized.
func NewModel() *Model {
	return &Model{}
}

// Init initializes the model and returns an initial command.
// Currently returns nil as there are no initialization commands needed.
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and returns an updated model and command.
// It handles keyboard input (e.g., Ctrl+C) and other messages from Bubbletea.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			// Return quit command
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		// Handle window resize
		// For now, just accept the message and continue
	}

	return m, nil
}

// View returns a string representation of the model's current state.
// This is rendered to the terminal on each update.
func (m *Model) View() string {
	return "Nexus TUI"
}
