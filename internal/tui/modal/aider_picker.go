package modal

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// AiderFilePicker is a Bubbletea modal for selecting files to pass to Aider.
// Space toggles file selection, Enter confirms (emitting AiderLaunchMsg),
// and Esc cancels (emitting ModalCancelledMsg).
type AiderFilePicker struct {
	files    []string
	selected map[int]bool
	cursor   int
	err      string
}

// NewAiderFilePicker creates a new AiderFilePicker with the given file list.
func NewAiderFilePicker(files []string) *AiderFilePicker {
	return &AiderFilePicker{
		files:    files,
		selected: make(map[int]bool),
	}
}

// Init satisfies tea.Model.
func (m *AiderFilePicker) Init() tea.Cmd { return nil }

// Update handles key messages to drive the file picker state machine.
func (m *AiderFilePicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "esc":
		return m, func() tea.Msg { return ModalCancelledMsg{} }
	case " ":
		m.selected[m.cursor] = !m.selected[m.cursor]
		m.err = ""
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.files)-1 {
			m.cursor++
		}
	case "enter":
		var selectedFiles []string
		for i, file := range m.files {
			if m.selected[i] {
				selectedFiles = append(selectedFiles, file)
			}
		}
		if len(selectedFiles) == 0 {
			m.err = "Select at least one file"
			return m, nil
		}
		files := selectedFiles
		return m, func() tea.Msg { return AiderLaunchMsg{Files: files} }
	}

	return m, nil
}

// Title returns the modal title for themed overlay rendering.
func (m *AiderFilePicker) Title() string { return "Aider — Select Files" }

// View renders the file picker list with selection state and key hints.
func (m *AiderFilePicker) View() string {
	var b strings.Builder

	if len(m.files) == 0 {
		b.WriteString("No modified files found.\n\n")
		b.WriteString("Esc cancel")
		return b.String()
	}

	b.WriteString("Select files for Aider context:\n\n")

	for i, file := range m.files {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		check := "[ ]"
		if m.selected[i] {
			check = "[x]"
		}
		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, check, file))
	}

	if m.err != "" {
		b.WriteString(fmt.Sprintf("\n⚠ %s\n", m.err))
	}

	b.WriteString("\nSpace toggle  •  ↑/↓ navigate  •  Enter confirm  •  Esc cancel")
	return b.String()
}
