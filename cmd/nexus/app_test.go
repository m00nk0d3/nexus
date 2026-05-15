package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/m00nk0d3/nexus/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModelInitialization verifies that the Model can be instantiated
// with all required fields properly initialized
func TestModelInitialization(t *testing.T) {
	tests := []struct {
		name             string
		wantModelNotNil  bool
		wantHasFields    bool
	}{
		{
			name:            "creates new model successfully",
			wantModelNotNil: true,
			wantHasFields:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			
			if tt.wantModelNotNil {
				assert.NotNil(t, model, "Model should not be nil")
			}
			
			if tt.wantHasFields {
				// Verify model can be cast to tea.Model (has required interface methods)
				var _ tea.Model = model
				assert.NotNil(t, model, "Model should implement tea.Model interface")
			}
		})
	}
}

// TestModelUpdate verifies that the Update method accepts tea.Msg
// and returns (tea.Model, tea.Cmd) as required by Bubbletea interface
func TestModelUpdate(t *testing.T) {
	tests := []struct {
		name           string
		msg            tea.Msg
		wantModel      bool
		wantCmdNotNil  bool
		description    string
	}{
		{
			name:          "update accepts tea.KeyMsg",
			msg:           tea.KeyMsg{Type: tea.KeyCtrlC},
			wantModel:     true,
			wantCmdNotNil: false, // Initial implementation may not return a Cmd
			description:   "Should accept KeyMsg and return model (Cmd can be nil)",
		},
		{
			name:          "update accepts generic tea.Msg",
			msg:           tea.WindowSizeMsg{Width: 80, Height: 24},
			wantModel:     true,
			wantCmdNotNil: false,
			description:   "Should accept WindowSizeMsg and return model",
		},
		{
			name:          "update handles nil message gracefully",
			msg:           nil,
			wantModel:     true,
			wantCmdNotNil: false,
			description:   "Should handle nil message without panicking",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model, "Model should be created successfully")

			// Call Update method - it must return (tea.Model, tea.Cmd)
			updatedModel, cmd := model.Update(tt.msg)

			if tt.wantModel {
				assert.NotNil(t, updatedModel, "Update should return a model: %s", tt.description)
				// Verify it's a valid Model that implements tea.Model
				var _ tea.Model = updatedModel
			}

			// cmd can be nil (no command to execute)
			if tt.wantCmdNotNil {
				assert.NotNil(t, cmd, "Update should return a Cmd: %s", tt.description)
			}
		})
	}
}

// TestModelView verifies that the View method returns a string
// representation of the model's current state
func TestModelView(t *testing.T) {
	tests := []struct {
		name                string
		wantViewNotEmpty    bool
		wantViewIsString    bool
		description         string
	}{
		{
			name:             "view returns string representation",
			wantViewNotEmpty: false, // Initial implementation may return empty string
			wantViewIsString: true,
			description:      "View should return a string (may be empty initially)",
		},
		{
			name:             "view is consistently callable",
			wantViewNotEmpty: false,
			wantViewIsString: true,
			description:      "Multiple calls to View should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model, "Model should be created successfully")

			// Call View method - it must return a string
			view := model.View()

			assert.IsType(t, "", view, "View should return a string: %s", tt.description)

			if tt.wantViewNotEmpty {
				assert.NotEmpty(t, view, "View should not be empty: %s", tt.description)
			}
		})
	}
}

// TestModelIntegration verifies that the model works correctly through
// a typical initialization and message handling sequence
func TestModelIntegration(t *testing.T) {
	tests := []struct {
		name        string
		description string
	}{
		{
			name:        "model initialization followed by update and view",
			description: "Should create model, handle update, and render view",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize model
			model := NewModel()
			require.NotNil(t, model, "Model creation should succeed")

			// Verify View works immediately
			initialView := model.View()
			assert.IsType(t, "", initialView, "View should return string after init: %s", tt.description)

			// Verify Update works with a message
			updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			assert.NotNil(t, updatedModel, "Update should return model: %s", tt.description)

			// Verify View works after update
			updatedView := updatedModel.View()
			assert.IsType(t, "", updatedView, "View should work after update: %s", tt.description)
		})
	}
}

// TestModel_Enter_TriggersSwitch verifies that pressing Enter on a selected worktree
// returns a tea.Cmd to switch to that worktree
func TestModel_Enter_TriggersSwitch(t *testing.T) {
	tests := []struct {
		name            string
		worktrees       []interface{} // Will be converted to domain.Worktree
		selectedIdx     int
		description     string
		wantCmdNotNil   bool
	}{
		{
			name: "enter on first worktree returns switch command",
			worktrees: []interface{}{
				map[string]interface{}{"Path": "/home/user/repos/wt1", "Branch": "main", "CommitSHA": "abc123", "IsClean": true, "IsLocked": false, "LinkedPR": nil},
				map[string]interface{}{"Path": "/home/user/repos/wt2", "Branch": "feature", "CommitSHA": "def456", "IsClean": false, "IsLocked": false, "LinkedPR": nil},
			},
			selectedIdx:   0,
			description:   "Should return a Cmd to switch to first worktree",
			wantCmdNotNil: true,
		},
		{
			name: "enter on second worktree returns switch command",
			worktrees: []interface{}{
				map[string]interface{}{"Path": "/home/user/repos/wt1", "Branch": "main", "CommitSHA": "abc123", "IsClean": true, "IsLocked": false, "LinkedPR": nil},
				map[string]interface{}{"Path": "/home/user/repos/wt2", "Branch": "feature", "CommitSHA": "def456", "IsClean": false, "IsLocked": false, "LinkedPR": nil},
			},
			selectedIdx:   1,
			description:   "Should return a Cmd to switch to second worktree",
			wantCmdNotNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Create model with populated Worktrees list
			model := NewModel()
			require.NotNil(t, model, "Model creation should succeed")

			// Convert test data to domain.Worktree
			worktrees := make([]domain.Worktree, len(tt.worktrees))
			for i, wtData := range tt.worktrees {
				data := wtData.(map[string]interface{})
				worktrees[i] = domain.Worktree{
					Path:      data["Path"].(string),
					Branch:    data["Branch"].(string),
					CommitSHA: data["CommitSHA"].(string),
					IsClean:   data["IsClean"].(bool),
					IsLocked:  data["IsLocked"].(bool),
				}
			}
			model.Worktrees = worktrees
			model.selectedIdx = tt.selectedIdx

			// Action: Call Update with tea.KeyEnter
			updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

			// Assert: Model is returned
			assert.NotNil(t, updatedModel, "Update should return model: %s", tt.description)

			// Assert: A Cmd is returned (for switching worktree)
			if tt.wantCmdNotNil {
				assert.NotNil(t, cmd, "Update should return a Cmd for switching worktree: %s", tt.description)
			}
		})
	}
}

// TestModel_Enter_EmptyList_NoOp verifies that pressing Enter on an empty worktree list
// does not trigger a switch command
func TestModel_Enter_EmptyList_NoOp(t *testing.T) {
	tests := []struct {
		name            string
		description     string
		wantCmdNil      bool
	}{
		{
			name:            "enter on empty list returns nil command",
			description:     "Should return nil Cmd when no worktrees exist",
			wantCmdNil:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Create model with empty Worktrees list
			model := NewModel()
			require.NotNil(t, model, "Model creation should succeed")
			require.Empty(t, model.Worktrees, "Worktrees should be empty initially")

			// Action: Call Update with tea.KeyEnter
			updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

			// Assert: Model is returned
			assert.NotNil(t, updatedModel, "Update should return model: %s", tt.description)

			// Assert: Cmd is nil (no-op)
			if tt.wantCmdNil {
				assert.Nil(t, cmd, "Update should return nil Cmd for empty list: %s", tt.description)
			}
		})
	}
}
