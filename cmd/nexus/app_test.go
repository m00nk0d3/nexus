package main

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/m00nk0d3/nexus/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModelInitialization verifies that the Model can be instantiated
// with all required fields properly initialized
func TestModelInitialization(t *testing.T) {
	tests := []struct {
		name            string
		wantModelNotNil bool
		wantHasFields   bool
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
				assert.NotNil(t, model.Config, "Config should be initialized (defaults at minimum)")
			}
		})
	}
}

// TestModelUpdate verifies that the Update method accepts tea.Msg
// and returns (tea.Model, tea.Cmd) as required by Bubbletea interface
func TestModelUpdate(t *testing.T) {
	tests := []struct {
		name          string
		msg           tea.Msg
		wantModel     bool
		wantCmdNotNil bool
		description   string
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
		name             string
		wantViewNotEmpty bool
		wantViewIsString bool
		description      string
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
		name          string
		worktrees     []interface{} // Will be converted to domain.Worktree
		selectedIdx   int
		description   string
		wantCmdNotNil bool
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
		name        string
		description string
		wantCmdNil  bool
	}{
		{
			name:        "enter on empty list returns nil command",
			description: "Should return nil Cmd when no worktrees exist",
			wantCmdNil:  true,
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

func TestModel_Enter_OutOfRangeSelectedIndex_NoOp(t *testing.T) {
	model := NewModel()
	require.NotNil(t, model)

	model.Worktrees = []domain.Worktree{
		{Path: "/home/user/repos/wt1", Branch: "main", CommitSHA: "abc123"},
	}
	model.selectedIdx = 10

	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, updatedModel)
	assert.Nil(t, cmd)
}

func TestBuildShellCmdForOS_Windows_UsesCmdKAndDir(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "windows path with spaces",
			path:     `C:\Users\dev\My Worktree`,
			expected: []string{"cmd", "/K"},
		},
		{
			name:     "windows different drive path",
			path:     `D:\repo\wt-feature`,
			expected: []string{"cmd", "/K"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := buildShellCmdForOS(tt.path, "windows", "")
			require.NotNil(t, cmd)
			assert.Equal(t, tt.expected, cmd.Args)
			assert.Equal(t, tt.path, cmd.Dir)
		})
	}
}

func TestBuildShellCmdForOS_Unix_UsesShellAndFallback(t *testing.T) {
	tests := []struct {
		name      string
		shell     string
		wantFirst string
	}{
		{
			name:      "uses provided shell",
			shell:     "/bin/zsh",
			wantFirst: "/bin/zsh",
		},
		{
			name:      "falls back to /bin/sh when shell empty",
			shell:     "",
			wantFirst: "/bin/sh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/tmp/worktree"
			cmd := buildShellCmdForOS(path, "linux", tt.shell)
			require.NotNil(t, cmd)
			require.NotEmpty(t, cmd.Args)
			assert.Equal(t, tt.wantFirst, cmd.Args[0])
			assert.Equal(t, path, cmd.Dir)
		})
	}
}

func TestGetShell_UsesEnvAndFallback(t *testing.T) {
	tests := []struct {
		name      string
		shellEnv  string
		wantShell string
	}{
		{
			name:      "uses SHELL env value when set",
			shellEnv:  "/bin/fish",
			wantShell: "/bin/fish",
		},
		{
			name:      "falls back to /bin/sh when SHELL env empty",
			shellEnv:  "",
			wantShell: "/bin/sh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SHELL", tt.shellEnv)
			assert.Equal(t, tt.wantShell, getShell())
		})
	}
}

func TestModelUpdate_WorktreeSwitchedMsg_ErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		msg          worktreeSwitchedMsg
		wantError    string
		wantCmdIsNil bool
	}{
		{
			name:         "sets model error when switch fails",
			msg:          worktreeSwitchedMsg{err: errors.New("switch failed")},
			wantError:    "Failed to switch worktree: switch failed",
			wantCmdIsNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)

			updated, cmd := model.Update(tt.msg)
			updatedModel, ok := updated.(*Model)
			require.True(t, ok)
			assert.Equal(t, tt.wantError, updatedModel.Error)
			if tt.wantCmdIsNil {
				assert.Nil(t, cmd)
			}
		})
	}
}

func TestModelUpdate_WorktreesRefreshedMsg_ClampsSelectedIndex(t *testing.T) {
	tests := []struct {
		name            string
		initialSelected int
		worktrees       []domain.Worktree
		wantSelected    int
	}{
		{
			name:            "clamps to last when selected index is too large",
			initialSelected: 5,
			worktrees: []domain.Worktree{
				{Path: "/wt/a"},
				{Path: "/wt/b"},
			},
			wantSelected: 1,
		},
		{
			name:            "normalizes negative selected index to zero",
			initialSelected: -3,
			worktrees: []domain.Worktree{
				{Path: "/wt/a"},
			},
			wantSelected: 0,
		},
		{
			name:            "resets selected index to zero for empty list",
			initialSelected: 2,
			worktrees:       nil,
			wantSelected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.selectedIdx = tt.initialSelected

			updated, cmd := model.Update(worktreesRefreshedMsg{worktrees: tt.worktrees, err: nil})
			updatedModel, ok := updated.(*Model)
			require.True(t, ok)
			assert.Equal(t, tt.wantSelected, updatedModel.selectedIdx)
			assert.Nil(t, cmd)
		})
	}
}

func TestModelView_ShowsErrorMessage(t *testing.T) {
	model := NewModel()
	require.NotNil(t, model)
	model.Error = "Failed to switch worktree: boom"

	view := model.View()
	assert.Contains(t, view, "Error: Failed to switch worktree: boom")
	assert.Contains(t, view, "GIT WORKTREE ORCHESTRATOR")
}

func TestModel_T_KeyCyclesTheme(t *testing.T) {
	tests := []struct {
		name         string
		initialIdx   int
		pressCount   int
		wantThemeIdx int
	}{
		{
			name:         "first press increments from digital-noir to matrix",
			initialIdx:   0,
			pressCount:   1,
			wantThemeIdx: 1,
		},
		{
			name:         "second press increments from matrix to light",
			initialIdx:   1,
			pressCount:   1,
			wantThemeIdx: 2,
		},
		{
			name:         "wraps from light back to digital-noir",
			initialIdx:   2,
			pressCount:   1,
			wantThemeIdx: 0,
		},
		{
			name:         "three presses cycles through all themes and returns to start",
			initialIdx:   0,
			pressCount:   3,
			wantThemeIdx: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.themeIdx = tt.initialIdx

			for i := 0; i < tt.pressCount; i++ {
				updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
				var ok bool
				model, ok = updated.(*Model)
				require.True(t, ok)
			}

			assert.Equal(t, tt.wantThemeIdx, model.themeIdx)
		})
	}
}

// TestModel_Init_ReturnsSyncCmd verifies that Init() returns a non-nil Cmd,
// meaning it schedules the initial background GitHub sync in addition to
// refreshing the worktree list.
func TestModel_Init_ReturnsSyncCmd(t *testing.T) {
	model := NewModel()
	require.NotNil(t, model)

	cmd := model.Init()

	assert.NotNil(t, cmd, "Init() must return a Cmd to trigger GitHub sync")
}

// TestModel_GithubSyncedMsg_StoresPRsAndIssues verifies that receiving a
// githubSyncedMsg via Update() correctly stores the synced data into the model.
func TestModel_GithubSyncedMsg_StoresPRsAndIssues(t *testing.T) {
	tests := []struct {
		name            string
		msg             githubSyncedMsg
		wantPRLen       int
		wantPRNumber    int
		wantIssueLen    int
		wantIssueNumber int
		wantLastSynced  bool
		wantSyncErr     string
		wantSyncing     bool
	}{
		{
			name: "stores prs and issues from sync message",
			msg: githubSyncedMsg{
				prs:      []domain.PullRequest{{Number: 42}},
				issues:   []domain.Issue{{Number: 7}},
				syncedAt: time.Now(),
			},
			wantPRLen:       1,
			wantPRNumber:    42,
			wantIssueLen:    1,
			wantIssueNumber: 7,
			wantLastSynced:  true,
			wantSyncing:     false,
		},
		{
			name:        "stores sync error without crashing",
			msg:         githubSyncedMsg{err: errors.New("api down")},
			wantSyncErr: "api down",
			wantSyncing: false,
		},
		{
			name:        "sets syncing=false after sync completes",
			msg:         githubSyncedMsg{prs: nil, issues: nil},
			wantSyncing: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)

			updated, _ := model.Update(tt.msg)
			m, ok := updated.(*Model)
			require.True(t, ok, "Update must return *Model")

			if tt.wantPRLen > 0 {
				require.Len(t, m.prs, tt.wantPRLen, "prs slice length mismatch")
				assert.Equal(t, tt.wantPRNumber, m.prs[0].Number, "PR number mismatch")
			}

			if tt.wantIssueLen > 0 {
				require.Len(t, m.issues, tt.wantIssueLen, "issues slice length mismatch")
				assert.Equal(t, tt.wantIssueNumber, m.issues[0].Number, "issue number mismatch")
			}

			if tt.wantLastSynced {
				assert.False(t, m.lastSynced.IsZero(), "lastSynced must be set to a non-zero time")
			}

			if tt.wantSyncErr != "" {
				require.NotNil(t, m.syncErr, "syncErr must not be nil")
				assert.Contains(t, m.syncErr.Error(), tt.wantSyncErr, "syncErr message mismatch")
			}

			assert.Equal(t, tt.wantSyncing, m.syncing, "syncing flag mismatch")
		})
	}
}

// TestModel_SyncTickMsg_TriggersSyncCmd verifies that receiving a syncTickMsg
// via Update() returns a non-nil Cmd to schedule the next background GitHub sync.
func TestModel_SyncTickMsg_TriggersSyncCmd(t *testing.T) {
	model := NewModel()
	require.NotNil(t, model)

	_, cmd := model.Update(syncTickMsg{})

	assert.NotNil(t, cmd, "syncTickMsg must trigger a sync Cmd")
}

func TestModel_WindowSizeMsg_StoresTerminalWidth(t *testing.T) {
	tests := []struct {
		name       string
		msgWidth   int
		msgHeight  int
		wantWidth  int
		wantHeight int
	}{
		{
			name:       "stores width from WindowSizeMsg",
			msgWidth:   160,
			msgHeight:  50,
			wantWidth:  160,
			wantHeight: 50,
		},
		{
			name:       "stores minimum width from WindowSizeMsg",
			msgWidth:   80,
			msgHeight:  24,
			wantWidth:  80,
			wantHeight: 24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)

			updated, _ := model.Update(tea.WindowSizeMsg{Width: tt.msgWidth, Height: tt.msgHeight})
			m, ok := updated.(*Model)
			require.True(t, ok)

			assert.Equal(t, tt.wantWidth, m.width)
			assert.Equal(t, tt.wantHeight, m.height)
		})
	}
}
