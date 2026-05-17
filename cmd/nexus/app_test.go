package main

import (
	"errors"
	"os/exec"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/m00nk0d3/nexus/internal/data"
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

// ---------------------------------------------------------------------------
// Phase 2: Issues & PRs View tests
// ---------------------------------------------------------------------------

// TestModel_ViewSwitching verifies that pressing W/I/P (upper- and lower-case)
// switches the model's active view to the correct activeView constant.
func TestModel_ViewSwitching(t *testing.T) {
	tests := []struct {
		name     string
		key      rune
		wantView activeView
	}{
		{
			name:     "pressing W sets view to viewWorktrees",
			key:      'W',
			wantView: viewWorktrees,
		},
		{
			name:     "pressing I sets view to viewIssues",
			key:      'I',
			wantView: viewIssues,
		},
		{
			name:     "pressing P sets view to viewPRs",
			key:      'P',
			wantView: viewPRs,
		},
		{
			name:     "pressing w (lowercase) sets view to viewWorktrees",
			key:      'w',
			wantView: viewWorktrees,
		},
		{
			name:     "pressing i (lowercase) sets view to viewIssues",
			key:      'i',
			wantView: viewIssues,
		},
		{
			name:     "pressing p (lowercase) sets view to viewPRs",
			key:      'p',
			wantView: viewPRs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)

			updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}})
			m, ok := updated.(*Model)
			require.True(t, ok, "Update must return *Model")

			assert.Equal(t, tt.wantView, m.view)
		})
	}
}

// TestModel_IssueNavigation verifies that up/down navigation in viewIssues
// moves selectedIssueIdx correctly and does NOT move the worktree selectedIdx.
func TestModel_IssueNavigation(t *testing.T) {
	issues := []domain.Issue{
		{Number: 1, Title: "First"},
		{Number: 2, Title: "Second"},
		{Number: 3, Title: "Third"},
	}

	tests := []struct {
		name            string
		initialIssueIdx int
		keyType         tea.KeyType
		wantIssueIdx    int
		wantWorktreeIdx int
	}{
		{
			name:            "down key increments selectedIssueIdx",
			initialIssueIdx: 0,
			keyType:         tea.KeyDown,
			wantIssueIdx:    1,
			wantWorktreeIdx: 0,
		},
		{
			name:            "up key decrements selectedIssueIdx",
			initialIssueIdx: 1,
			keyType:         tea.KeyUp,
			wantIssueIdx:    0,
			wantWorktreeIdx: 0,
		},
		{
			name:            "up key does not go below 0 (boundary)",
			initialIssueIdx: 0,
			keyType:         tea.KeyUp,
			wantIssueIdx:    0,
			wantWorktreeIdx: 0,
		},
		{
			name:            "down key does not exceed len(issues)-1 (boundary)",
			initialIssueIdx: 2,
			keyType:         tea.KeyDown,
			wantIssueIdx:    2,
			wantWorktreeIdx: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.issues = issues
			model.view = viewIssues
			model.focused = panelList
			model.selectedIssueIdx = tt.initialIssueIdx

			updated, _ := model.Update(tea.KeyMsg{Type: tt.keyType})
			m, ok := updated.(*Model)
			require.True(t, ok, "Update must return *Model")

			assert.Equal(t, tt.wantIssueIdx, m.selectedIssueIdx, "issue index mismatch")
			assert.Equal(t, tt.wantWorktreeIdx, m.selectedIdx, "worktree idx must not change when navigating issues")
		})
	}
}

// TestModel_PRNavigation verifies that up/down navigation in viewPRs
// moves selectedPRIdx correctly and does NOT move the worktree selectedIdx.
func TestModel_PRNavigation(t *testing.T) {
	prs := []domain.PullRequest{
		{Number: 10, Title: "PR One"},
		{Number: 11, Title: "PR Two"},
		{Number: 12, Title: "PR Three"},
	}

	tests := []struct {
		name            string
		initialPRIdx    int
		keyType         tea.KeyType
		wantPRIdx       int
		wantWorktreeIdx int
	}{
		{
			name:            "down key increments selectedPRIdx",
			initialPRIdx:    0,
			keyType:         tea.KeyDown,
			wantPRIdx:       1,
			wantWorktreeIdx: 0,
		},
		{
			name:            "up key decrements selectedPRIdx",
			initialPRIdx:    1,
			keyType:         tea.KeyUp,
			wantPRIdx:       0,
			wantWorktreeIdx: 0,
		},
		{
			name:            "up key does not go below 0 (boundary)",
			initialPRIdx:    0,
			keyType:         tea.KeyUp,
			wantPRIdx:       0,
			wantWorktreeIdx: 0,
		},
		{
			name:            "down key does not exceed len(prs)-1 (boundary)",
			initialPRIdx:    2,
			keyType:         tea.KeyDown,
			wantPRIdx:       2,
			wantWorktreeIdx: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.prs = prs
			model.view = viewPRs
			model.focused = panelList
			model.selectedPRIdx = tt.initialPRIdx

			updated, _ := model.Update(tea.KeyMsg{Type: tt.keyType})
			m, ok := updated.(*Model)
			require.True(t, ok, "Update must return *Model")

			assert.Equal(t, tt.wantPRIdx, m.selectedPRIdx, "PR index mismatch")
			assert.Equal(t, tt.wantWorktreeIdx, m.selectedIdx, "worktree idx must not change when navigating PRs")
		})
	}
}

// TestModel_G_Key_OpensInBrowser verifies the [g] key opens the selected
// issue or PR in the browser (returns non-nil Cmd), and is a no-op in
// viewWorktrees or when the list is empty.
func TestModel_G_Key_OpensInBrowser(t *testing.T) {
	tests := []struct {
		name       string
		view       activeView
		issues     []domain.Issue
		prs        []domain.PullRequest
		issueIdx   int
		prIdx      int
		wantCmdNil bool
	}{
		{
			name:       "g in viewIssues with issue selected returns non-nil Cmd",
			view:       viewIssues,
			issues:     []domain.Issue{{Number: 5, Title: "Test Issue"}},
			issueIdx:   0,
			wantCmdNil: false,
		},
		{
			name:       "g in viewPRs with PR selected returns non-nil Cmd",
			view:       viewPRs,
			prs:        []domain.PullRequest{{Number: 42, Title: "My PR", Branch: "feat/awesome", Author: "alice", State: "OPEN"}},
			prIdx:      0,
			wantCmdNil: false,
		},
		{
			name:       "g in viewWorktrees is a no-op (returns nil Cmd)",
			view:       viewWorktrees,
			wantCmdNil: true,
		},
		{
			name:       "g in viewIssues with empty issues list returns nil Cmd",
			view:       viewIssues,
			issues:     []domain.Issue{},
			issueIdx:   0,
			wantCmdNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.view = tt.view
			if tt.issues != nil {
				model.issues = tt.issues
			}
			if tt.prs != nil {
				model.prs = tt.prs
			}
			model.selectedIssueIdx = tt.issueIdx
			model.selectedPRIdx = tt.prIdx

			_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})

			if tt.wantCmdNil {
				assert.Nil(t, cmd, "expected nil Cmd (no-op) but got non-nil")
			} else {
				assert.NotNil(t, cmd, "expected non-nil Cmd to open in browser")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// End Phase 2 tests (app_test.go)
// ---------------------------------------------------------------------------

// TestModel_GithubSync_ClampsIssueAndPRIdx verifies that after a successful sync
// that returns fewer items, selectedIssueIdx and selectedPRIdx are clamped to
// the new list bounds so openInBrowserCmd never panics.
func TestModel_GithubSync_ClampsIssueAndPRIdx(t *testing.T) {
	tests := []struct {
		name             string
		initialIssueIdx  int
		initialPRIdx     int
		syncIssues       []domain.Issue
		syncPRs          []domain.PullRequest
		wantIssueIdx     int
		wantPRIdx        int
	}{
		{
			name:            "issue idx clamped when sync shrinks list",
			initialIssueIdx: 4,
			initialPRIdx:    0,
			syncIssues:      []domain.Issue{{Number: 1, Title: "Only Issue"}},
			syncPRs:         []domain.PullRequest{{Number: 10, Title: "PR", Branch: "main", Author: "dev", State: "OPEN"}},
			wantIssueIdx:    0,
			wantPRIdx:       0,
		},
		{
			name:            "pr idx clamped when sync shrinks list",
			initialIssueIdx: 0,
			initialPRIdx:    5,
			syncIssues:      []domain.Issue{{Number: 1, Title: "Issue"}},
			syncPRs:         []domain.PullRequest{{Number: 10, Title: "PR", Branch: "main", Author: "dev", State: "OPEN"}},
			wantIssueIdx:    0,
			wantPRIdx:       0,
		},
		{
			name:            "idx within bounds is preserved after sync",
			initialIssueIdx: 0,
			initialPRIdx:    0,
			syncIssues:      []domain.Issue{{Number: 1, Title: "A"}, {Number: 2, Title: "B"}},
			syncPRs:         []domain.PullRequest{{Number: 10, Title: "PR", Branch: "main", Author: "dev", State: "OPEN"}},
			wantIssueIdx:    0,
			wantPRIdx:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.selectedIssueIdx = tt.initialIssueIdx
			model.selectedPRIdx = tt.initialPRIdx

			msg := githubSyncedMsg{
				issues:   tt.syncIssues,
				prs:      tt.syncPRs,
				syncedAt: time.Now(),
			}
			updated, _ := model.Update(msg)
			m, ok := updated.(*Model)
			require.True(t, ok)

			assert.Equal(t, tt.wantIssueIdx, m.selectedIssueIdx, "issue idx must be clamped")
			assert.Equal(t, tt.wantPRIdx, m.selectedPRIdx, "pr idx must be clamped")
		})
	}
}

// TestModel_BrowserOpenErrMsg_SetsError verifies that a browserOpenErrMsg with
// a non-nil error sets m.Error so the user sees feedback.
func TestModel_BrowserOpenErrMsg_SetsError(t *testing.T) {
	model := NewModel()
	require.NotNil(t, model)

	updated, _ := model.Update(browserOpenErrMsg{err: errors.New("gh: not found")})
	m, ok := updated.(*Model)
	require.True(t, ok)

	assert.Contains(t, m.Error, "Failed to open in browser")
	assert.Contains(t, m.Error, "gh: not found")
}

// TestModel_BrowserOpenErrMsg_NilErrorNoChange verifies that a nil-error
// browserOpenErrMsg does not set m.Error.
func TestModel_BrowserOpenErrMsg_NilErrorNoChange(t *testing.T) {
	model := NewModel()
	require.NotNil(t, model)

	updated, _ := model.Update(browserOpenErrMsg{err: nil})
	m, ok := updated.(*Model)
	require.True(t, ok)

	assert.Empty(t, m.Error)
}

// ---------------------------------------------------------------------------
// Phase 4: Panel focus & j/k navigation tests
// ---------------------------------------------------------------------------

// TestModel_DefaultFocus_IsNavPanel verifies that a new model starts with
// the nav panel focused (panelNav is the zero value).
func TestModel_DefaultFocus_IsNavPanel(t *testing.T) {
	model := NewModel()
	require.NotNil(t, model)
	assert.Equal(t, panelNav, model.focused)
}

// TestModel_Tab_CyclesFocusThroughPanels verifies that Tab cycles focus
// left to right: nav → list → ctx → nav.
func TestModel_Tab_CyclesFocusThroughPanels(t *testing.T) {
	tests := []struct {
		name         string
		initialFocus focusedPanel
		wantFocus    focusedPanel
	}{
		{name: "Tab from nav focuses list", initialFocus: panelNav, wantFocus: panelList},
		{name: "Tab from list focuses ctx", initialFocus: panelList, wantFocus: panelCtx},
		{name: "Tab from ctx wraps to nav", initialFocus: panelCtx, wantFocus: panelNav},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.focused = tt.initialFocus

			updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
			m, ok := updated.(*Model)
			require.True(t, ok)

			assert.Equal(t, tt.wantFocus, m.focused)
		})
	}
}

// TestModel_JK_ListFocused_NavigatesWorktrees verifies that j/k navigate
// the worktree list when the list panel is focused.
func TestModel_JK_ListFocused_NavigatesWorktrees(t *testing.T) {
	worktrees := []domain.Worktree{
		{Path: "/wt/a"},
		{Path: "/wt/b"},
		{Path: "/wt/c"},
	}

	tests := []struct {
		name       string
		key        rune
		initialIdx int
		wantIdx    int
	}{
		{name: "j moves selection down", key: 'j', initialIdx: 0, wantIdx: 1},
		{name: "k moves selection up", key: 'k', initialIdx: 1, wantIdx: 0},
		{name: "j at bottom stays", key: 'j', initialIdx: 2, wantIdx: 2},
		{name: "k at top stays", key: 'k', initialIdx: 0, wantIdx: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.Worktrees = worktrees
			model.selectedIdx = tt.initialIdx
			model.focused = panelList

			updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}})
			m, ok := updated.(*Model)
			require.True(t, ok)

			assert.Equal(t, tt.wantIdx, m.selectedIdx)
		})
	}
}

// TestModel_JK_NavFocused_ChangesView verifies that j/k cycle through views
// when the nav panel is focused.
func TestModel_JK_NavFocused_ChangesView(t *testing.T) {
	tests := []struct {
		name        string
		key         rune
		initialView activeView
		wantView    activeView
	}{
		{name: "j from worktrees → issues", key: 'j', initialView: viewWorktrees, wantView: viewIssues},
		{name: "j from issues → PRs", key: 'j', initialView: viewIssues, wantView: viewPRs},
		{name: "j from PRs wraps → worktrees", key: 'j', initialView: viewPRs, wantView: viewWorktrees},
		{name: "k from PRs → issues", key: 'k', initialView: viewPRs, wantView: viewIssues},
		{name: "k from issues → worktrees", key: 'k', initialView: viewIssues, wantView: viewWorktrees},
		{name: "k from worktrees wraps → PRs", key: 'k', initialView: viewWorktrees, wantView: viewPRs},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.view = tt.initialView
			model.focused = panelNav

			updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}})
			m, ok := updated.(*Model)
			require.True(t, ok)

			assert.Equal(t, tt.wantView, m.view)
		})
	}
}

// TestModel_JK_CtxFocused_ChangesScrollOffset verifies that j/k change the
// context panel scroll offset when the ctx panel is focused.
func TestModel_JK_CtxFocused_ChangesScrollOffset(t *testing.T) {
	tests := []struct {
		name          string
		key           rune
		initialOffset int
		wantOffset    int
	}{
		{name: "j increments offset", key: 'j', initialOffset: 0, wantOffset: 1},
		{name: "j increments from non-zero", key: 'j', initialOffset: 3, wantOffset: 4},
		{name: "k decrements offset", key: 'k', initialOffset: 2, wantOffset: 1},
		{name: "k at zero stays at zero", key: 'k', initialOffset: 0, wantOffset: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.focused = panelCtx
			model.ctxScrollOffset = tt.initialOffset

			updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}})
			m, ok := updated.(*Model)
			require.True(t, ok)

			assert.Equal(t, tt.wantOffset, m.ctxScrollOffset)
		})
	}
}

// TestModel_WindowSizeMsg_StoresDimensions verifies that WindowSizeMsg updates
// width and height fields on the model.
func TestModel_WindowSizeMsg_StoresDimensions(t *testing.T) {
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



// TestModelUpdate_SKeyOpensShellInWorktree verifies that pressing "s" in
// viewWorktrees with a selected worktree triggers switchWorktreeCmd.
func TestModelUpdate_SKeyOpensShellInWorktree(t *testing.T) {
tests := []struct {
name       string
view       activeView
worktrees  []domain.Worktree
wantCmdNil bool
}{
{
name: "s key triggers switchWorktreeCmd when in worktrees view",
view: viewWorktrees,
worktrees: []domain.Worktree{
{Path: "/tmp/my-wt", Branch: "feat/my-branch", IsClean: true},
},
wantCmdNil: false,
},
{
name:       "s key does nothing when worktree list is empty",
view:       viewWorktrees,
worktrees:  nil,
wantCmdNil: true,
},
{
name: "s key does nothing in issues view",
view: viewIssues,
worktrees: []domain.Worktree{
{Path: "/tmp/my-wt", Branch: "feat/my-branch", IsClean: true},
},
wantCmdNil: true,
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
model := NewModel()
require.NotNil(t, model)
model.view = tt.view
model.Worktrees = tt.worktrees

_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
if tt.wantCmdNil {
assert.Nil(t, cmd)
} else {
assert.NotNil(t, cmd, "expected a non-nil cmd from switchWorktreeCmd")
}
})
}
}

// ---------------------------------------------------------------------------
// Phase 3: GitHub Copilot launcher tests
// ---------------------------------------------------------------------------

// newTestDB is a test helper that opens an in-memory SQLite DB for tests
// that need to exercise the DB logging path on the Model.
func newTestDB(t *testing.T) (*data.DB, error) {
	t.Helper()
	db, err := data.NewDB(":memory:")
	if err != nil {
		return nil, err
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, nil
}

// TestModel_C_Key_TriggersCopilotPrompt verifies that pressing 'c' with
// CopilotEnabled=true and a selected worktree activates the copilot prompt
// input. When CopilotEnabled=false or no worktree exists, it is a no-op.
func TestModel_C_Key_TriggersCopilotPrompt(t *testing.T) {
	tests := []struct {
		name             string
		copilotEnabled   bool
		hasWorktree      bool
		wantPromptActive bool
	}{
		{
			name:             "c key with disabled config does nothing",
			copilotEnabled:   false,
			hasWorktree:      true,
			wantPromptActive: false,
		},
		{
			name:             "c key with no worktree does nothing",
			copilotEnabled:   true,
			hasWorktree:      false,
			wantPromptActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			model.Config.AIAgents.CopilotEnabled = tt.copilotEnabled
			model.view = viewWorktrees
			if tt.hasWorktree {
				model.Worktrees = []domain.Worktree{
					{Path: "/tmp/wt", Branch: "main", CommitSHA: "abc"},
				}
			}

			updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
			updatedModel, ok := updated.(*Model)
			require.True(t, ok)

			assert.False(t, updatedModel.copilotPromptActive,
				"copilotPromptActive should be false")
		})
	}

	// Separate sub-test for the "gh on PATH" happy path, skipped if gh is absent.
	t.Run("c key with enabled config and selected worktree activates prompt", func(t *testing.T) {
		if _, err := exec.LookPath("gh"); err != nil {
			t.Skip("gh not on PATH; skipping test that requires gh CLI")
		}

		model := NewModel()
		model.Config.AIAgents.CopilotEnabled = true
		model.view = viewWorktrees
		model.Worktrees = []domain.Worktree{
			{Path: "/tmp/wt", Branch: "main", CommitSHA: "abc"},
		}

		updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
		updatedModel, ok := updated.(*Model)
		require.True(t, ok)

		assert.True(t, updatedModel.copilotPromptActive,
			"copilotPromptActive should be true when gh is on PATH")
		assert.NotNil(t, cmd, "textinput.Init() should return a non-nil cmd")
	})

	// Verify that 'c' in a non-worktree view does nothing even when enabled.
	t.Run("c key in issues view does nothing even when enabled", func(t *testing.T) {
		model := NewModel()
		model.Config.AIAgents.CopilotEnabled = true
		model.view = viewIssues
		model.Worktrees = []domain.Worktree{
			{Path: "/tmp/wt", Branch: "main", CommitSHA: "abc"},
		}

		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
		updatedModel, ok := updated.(*Model)
		require.True(t, ok)

		assert.False(t, updatedModel.copilotPromptActive,
			"copilotPromptActive should stay false when not in worktrees view")
	})
}

// TestBuildCopilotCmd_BuildsCorrectCommand verifies that buildCopilotCmd
// produces the right exec.Cmd args and working directory.
func TestBuildCopilotCmd_BuildsCorrectCommand(t *testing.T) {
	tests := []struct {
		name         string
		worktreePath string
		prompt       string
		wantArgs     []string
		wantDir      string
	}{
		{
			name:         "simple prompt",
			worktreePath: "/tmp/my-worktree",
			prompt:       "fix the null pointer",
			wantArgs:     []string{"gh", "copilot", "suggest", "fix the null pointer"},
			wantDir:      "/tmp/my-worktree",
		},
		{
			name:         "multi-word prompt",
			worktreePath: "/repo/feat-branch",
			prompt:       "add unit tests for auth handler",
			wantArgs:     []string{"gh", "copilot", "suggest", "add unit tests for auth handler"},
			wantDir:      "/repo/feat-branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := buildCopilotCmd(tt.worktreePath, tt.prompt)
			require.NotNil(t, cmd)
			assert.Equal(t, tt.wantArgs, cmd.Args)
			assert.Equal(t, tt.wantDir, cmd.Dir)
		})
	}
}

// TestModel_CopilotPrompt_EscCancels verifies that pressing Esc while the
// copilot prompt is active clears copilotPromptActive without spawning.
func TestModel_CopilotPrompt_EscCancels(t *testing.T) {
	model := NewModel()
	model.copilotPromptActive = true
	model.Worktrees = []domain.Worktree{{Path: "/tmp/wt", Branch: "main", CommitSHA: "abc"}}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updatedModel, ok := updated.(*Model)
	require.True(t, ok)

	assert.False(t, updatedModel.copilotPromptActive,
		"Esc should deactivate the copilot prompt")
}

// TestModel_CopilotPrompt_EscClearsInputValue verifies that Esc also resets
// the text input value so it starts fresh on the next invocation.
func TestModel_CopilotPrompt_EscClearsInputValue(t *testing.T) {
	model := NewModel()
	model.copilotPromptActive = true
	model.copilotPromptInput.SetValue("some typed text")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updatedModel, ok := updated.(*Model)
	require.True(t, ok)

	assert.Equal(t, "", updatedModel.copilotPromptInput.Value(),
		"Esc should clear the copilot prompt input value")
}

// TestModel_CopilotPrompt_EnterWithEmptyPrompt_DoesNotSpawn verifies that
// pressing Enter with an empty (or whitespace-only) prompt is a no-op.
func TestModel_CopilotPrompt_EnterWithEmptyPrompt_DoesNotSpawn(t *testing.T) {
	model := NewModel()
	model.copilotPromptActive = true
	// Leave input value empty (default)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updatedModel, ok := updated.(*Model)
	require.True(t, ok)

	// With an empty prompt, state should not change (prompt stays active)
	// and no cmd should be spawned.
	assert.True(t, updatedModel.copilotPromptActive,
		"empty-prompt Enter should leave copilotPromptActive true")
	assert.Nil(t, cmd, "empty-prompt Enter should not return a spawn Cmd")
}

// TestModel_AgentDoneMsg_ClearsPrompt verifies that receiving agentDoneMsg
// clears the copilot prompt state regardless of exit code.
func TestModel_AgentDoneMsg_ClearsPrompt(t *testing.T) {
	tests := []struct {
		name     string
		exitCode int
	}{
		{name: "exit code 0 clears prompt", exitCode: 0},
		{name: "non-zero exit code still clears prompt", exitCode: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			model.copilotPromptActive = true

			updated, _ := model.Update(agentDoneMsg{
				agentName: "copilot",
				prompt:    "test prompt",
				exitCode:  tt.exitCode,
			})
			updatedModel, ok := updated.(*Model)
			require.True(t, ok)

			assert.False(t, updatedModel.copilotPromptActive,
				"agentDoneMsg should clear copilotPromptActive")
		})
	}
}

// TestModel_AgentDoneMsg_LogsToDBWhenAvailable verifies that agentDoneMsg
// triggers a DB log call when model.db is set (non-nil).
func TestModel_AgentDoneMsg_LogsToDBWhenAvailable(t *testing.T) {
	// We test the logging path by supplying a real in-memory DB.
	db, err := newTestDB(t)
	require.NoError(t, err)

	model := NewModel()
	model.copilotPromptActive = true
	model.db = db

	updated, _ := model.Update(agentDoneMsg{
		agentName: "copilot",
		prompt:    "fix bug",
		exitCode:  0,
	})
	updatedModel, ok := updated.(*Model)
	require.True(t, ok)

	// No error should be set on the model.
	assert.Empty(t, updatedModel.Error, "DB log should not set an error on success")

	// Verify the row was actually written.
	var count int
	require.NoError(t, db.Conn.QueryRow(
		"SELECT COUNT(*) FROM agent_history WHERE agent_name = 'copilot'",
	).Scan(&count))
	assert.Equal(t, 1, count, "one agent_history row should have been inserted")
}

// TestModel_View_ShowsCopilotPromptWhenActive verifies that View() returns a
// string containing the prompt UI when copilotPromptActive is true.
func TestModel_View_ShowsCopilotPromptWhenActive(t *testing.T) {
	model := NewModel()
	model.copilotPromptActive = true

	view := model.View()

	assert.Contains(t, view, "Spawn Copilot",
		"View should show the Copilot prompt header when active")
	assert.Contains(t, view, "Esc cancel",
		"View should show the cancel hint when copilot prompt is active")
}

