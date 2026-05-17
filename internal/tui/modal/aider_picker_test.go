package modal

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAiderFilePicker(t *testing.T) {
	files := []string{"main.go", "go.mod"}
	picker := NewAiderFilePicker(files)
	require.NotNil(t, picker)
	assert.Equal(t, files, picker.files)
	assert.Empty(t, picker.selected)
	assert.Equal(t, 0, picker.cursor)
}

func TestAiderFilePicker_Title(t *testing.T) {
	picker := NewAiderFilePicker(nil)
	assert.Equal(t, "Aider — Select Files", picker.Title())
}

func TestAiderFilePicker_SpaceTogglesSelection(t *testing.T) {
	tests := []struct {
		name         string
		files        []string
		cursor       int
		presses      int
		wantSelected bool
	}{
		{
			name:         "one space press selects file",
			files:        []string{"main.go", "go.mod"},
			cursor:       0,
			presses:      1,
			wantSelected: true,
		},
		{
			name:         "two space presses deselects file",
			files:        []string{"main.go"},
			cursor:       0,
			presses:      2,
			wantSelected: false,
		},
		{
			name:         "space selects second file at cursor 1",
			files:        []string{"main.go", "go.mod"},
			cursor:       1,
			presses:      1,
			wantSelected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			picker := NewAiderFilePicker(tt.files)
			picker.cursor = tt.cursor

			var model tea.Model = picker
			for i := 0; i < tt.presses; i++ {
				updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
				model = updated
			}

			updated := model.(*AiderFilePicker)
			assert.Equal(t, tt.wantSelected, updated.selected[tt.cursor])
		})
	}
}

func TestAiderFilePicker_EmptySelection_ShowsError(t *testing.T) {
	picker := NewAiderFilePicker([]string{"main.go", "go.mod"})

	updated, cmd := picker.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, updated)
	assert.Nil(t, cmd, "should return no command when no files selected")

	updatedPicker := updated.(*AiderFilePicker)
	assert.Equal(t, "Select at least one file", updatedPicker.err)
}

func TestAiderFilePicker_EnterWithSelection_EmitsAiderLaunchMsg(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		selectIndices []int
		wantFiles     []string
	}{
		{
			name:          "single file selected",
			files:         []string{"main.go", "go.mod"},
			selectIndices: []int{0},
			wantFiles:     []string{"main.go"},
		},
		{
			name:          "multiple files selected",
			files:         []string{"main.go", "go.mod", "README.md"},
			selectIndices: []int{0, 2},
			wantFiles:     []string{"main.go", "README.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			picker := NewAiderFilePicker(tt.files)
			for _, idx := range tt.selectIndices {
				picker.selected[idx] = true
			}

			_, cmd := picker.Update(tea.KeyMsg{Type: tea.KeyEnter})
			require.NotNil(t, cmd, "should return a command")

			msg := cmd()
			launchMsg, ok := msg.(AiderLaunchMsg)
			require.True(t, ok, "command should emit AiderLaunchMsg")
			assert.Equal(t, tt.wantFiles, launchMsg.Files)
		})
	}
}

func TestAiderFilePicker_EscCancels(t *testing.T) {
	picker := NewAiderFilePicker([]string{"main.go"})

	_, cmd := picker.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(ModalCancelledMsg)
	assert.True(t, ok, "esc should emit ModalCancelledMsg")
}

func TestAiderFilePicker_Navigation(t *testing.T) {
	tests := []struct {
		name       string
		files      []string
		initial    int
		keys       []string
		wantCursor int
	}{
		{
			name:       "down key moves cursor down",
			files:      []string{"a.go", "b.go", "c.go"},
			initial:    0,
			keys:       []string{"down"},
			wantCursor: 1,
		},
		{
			name:       "j moves cursor down",
			files:      []string{"a.go", "b.go"},
			initial:    0,
			keys:       []string{"j"},
			wantCursor: 1,
		},
		{
			name:       "up key moves cursor up",
			files:      []string{"a.go", "b.go"},
			initial:    1,
			keys:       []string{"up"},
			wantCursor: 0,
		},
		{
			name:       "k moves cursor up",
			files:      []string{"a.go", "b.go"},
			initial:    1,
			keys:       []string{"k"},
			wantCursor: 0,
		},
		{
			name:       "down does not go past last file",
			files:      []string{"a.go", "b.go"},
			initial:    1,
			keys:       []string{"down"},
			wantCursor: 1,
		},
		{
			name:       "up does not go below zero",
			files:      []string{"a.go", "b.go"},
			initial:    0,
			keys:       []string{"up"},
			wantCursor: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			picker := NewAiderFilePicker(tt.files)
			picker.cursor = tt.initial

			var model tea.Model = picker
			for _, key := range tt.keys {
				updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
				model = updated
			}

			updatedPicker := model.(*AiderFilePicker)
			assert.Equal(t, tt.wantCursor, updatedPicker.cursor)
		})
	}
}

func TestAiderFilePicker_View_EmptyFiles(t *testing.T) {
	picker := NewAiderFilePicker(nil)
	view := picker.View()
	assert.Contains(t, view, "No modified files found")
	assert.Contains(t, view, "Esc cancel")
}

func TestAiderFilePicker_View_ShowsFilesAndHints(t *testing.T) {
	picker := NewAiderFilePicker([]string{"main.go", "go.mod"})
	view := picker.View()
	assert.Contains(t, view, "main.go")
	assert.Contains(t, view, "go.mod")
	assert.Contains(t, view, "Space toggle")
	assert.Contains(t, view, "Enter confirm")
}

func TestAiderFilePicker_View_ShowsSelectionAndError(t *testing.T) {
	picker := NewAiderFilePicker([]string{"main.go"})
	picker.selected[0] = true
	picker.err = "Select at least one file"

	view := picker.View()
	assert.Contains(t, view, "[x]")
	assert.Contains(t, view, "Select at least one file")
}

func TestAiderFilePicker_SpaceClears_ErrorMessage(t *testing.T) {
	picker := NewAiderFilePicker([]string{"main.go"})
	picker.err = "Select at least one file"

	updated, _ := picker.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	updatedPicker := updated.(*AiderFilePicker)
	assert.Empty(t, updatedPicker.err)
}
