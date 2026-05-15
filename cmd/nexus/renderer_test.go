package main

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/m00nk0d3/nexus/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderFull_ContainsHeader(t *testing.T) {
	tests := []struct {
		name     string
		repoPath string
		wantIn   []string
	}{
		{
			name:     "header includes NEXUS branding and GIT WORKTREE ORCHESTRATOR",
			repoPath: "/home/user/nexus",
			wantIn:   []string{"NEXUS", "GIT WORKTREE ORCHESTRATOR"},
		},
		{
			name:     "header includes repo path",
			repoPath: "/projects/myapp",
			wantIn:   []string{"myapp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.RepoPath = tt.repoPath

			view := model.View()

			for _, want := range tt.wantIn {
				assert.Contains(t, view, want)
			}
		})
	}
}

func TestRenderFull_ContainsNavRailItems(t *testing.T) {
	tests := []struct {
		name    string
		wantIn  []string
	}{
		{
			name:   "nav rail contains all navigation keys",
			wantIn: []string{"W", "I", "P", "T"},
		},
		{
			name:   "nav rail shows active cursor on W",
			wantIn: []string{"> W"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)

			view := model.View()

			for _, want := range tt.wantIn {
				assert.Contains(t, view, want)
			}
		})
	}
}

func TestRenderFull_ContainsWorktreeTableHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
	}{
		{
			name:    "renders Digital Noir column headers",
			headers: []string{"NAME", "PATH", "STATUS", "UPDATED", "GH:ID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.Worktrees = []domain.Worktree{
				{
					Path:      filepath.Join("worktrees", "feature-issue-4"),
					CommitSHA: "0123456789abcdef",
					IsClean:   true,
					IsLocked:  false,
				},
			}

			view := model.View()

			for _, header := range tt.headers {
				assert.Contains(t, view, header)
			}
		})
	}
}

func TestRenderFull_ContainsContextPanelPlaceholder(t *testing.T) {
	tests := []struct {
		name   string
		wantIn string
	}{
		{
			name:   "shows placeholder when no worktrees",
			wantIn: "No worktree selected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)

			view := model.View()

			assert.Contains(t, view, tt.wantIn)
		})
	}
}

func TestRenderFull_ContainsFooterKeyHints(t *testing.T) {
	tests := []struct {
		name   string
		wantIn []string
	}{
		{
			name:   "footer contains navigation and action hints",
			wantIn: []string{"[Enter] Select", "[t] Theme", "[esc] Quit"},
		},
		{
			name:   "action bar contains worktree commands",
			wantIn: []string{"[c-n] New", "[c-d] Delete", "[f1] Help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)

			view := model.View()

			for _, want := range tt.wantIn {
				assert.Contains(t, view, want)
			}
		})
	}
}

func TestRenderFull_SelectedRowDistinguishable(t *testing.T) {
	tests := []struct {
		name        string
		selectedIdx int
		wantCursor  string
	}{
		{
			name:        "selected row has > cursor on first item",
			selectedIdx: 0,
			wantCursor:  "> wt1",
		},
		{
			name:        "selected row has > cursor on second item",
			selectedIdx: 1,
			wantCursor:  "> wt2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.Worktrees = []domain.Worktree{
				{Path: "/tmp/wt1", Branch: "main", CommitSHA: "abc123", IsClean: true},
				{Path: "/tmp/wt2", Branch: "feature", CommitSHA: "def456", IsClean: false},
			}
			model.selectedIdx = tt.selectedIdx

			view := model.View()

			assert.Contains(t, view, tt.wantCursor)
		})
	}
}

func TestRenderFull_WorktreeStatusMapping(t *testing.T) {
	tests := []struct {
		name           string
		worktree       domain.Worktree
		expectedStatus string
		expectedName   string
	}{
		{
			name: "clean unlocked worktree shows Idle status",
			worktree: domain.Worktree{
				Path:    filepath.Join("tmp", "wt-clean"),
				IsClean: true, IsLocked: false,
			},
			expectedStatus: "Idle",
			expectedName:   "wt-clean",
		},
		{
			name: "dirty worktree shows Dirty status",
			worktree: domain.Worktree{
				Path:    filepath.Join("tmp", "wt-dirty"),
				IsClean: false, IsLocked: false,
			},
			expectedStatus: "Dirty",
			expectedName:   "wt-dirty",
		},
		{
			name: "locked worktree shows Locked status",
			worktree: domain.Worktree{
				Path:    filepath.Join("tmp", "wt-locked"),
				IsClean: true, IsLocked: true,
			},
			expectedStatus: "Locked",
			expectedName:   "wt-locked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.Worktrees = []domain.Worktree{tt.worktree}
			// Wide terminal so path fits without truncation.
			model.width = 300

			view := model.View()

			assert.Contains(t, view, tt.expectedName)
			assert.Contains(t, view, tt.worktree.Path)
			assert.Contains(t, view, tt.expectedStatus)
		})
	}
}

func TestRenderFooterBar_SyncStatus(t *testing.T) {
	tests := []struct {
		name       string
		syncing    bool
		lastSynced time.Time
		syncErr    error
		wantIn     string
	}{
		{
			name:    "shows syncing indicator while sync is in progress",
			syncing: true,
			wantIn:  "⟳ syncing",
		},
		{
			name:       "shows synced time when last sync succeeded",
			lastSynced: time.Now().Add(-3 * time.Minute),
			wantIn:     "✓ synced 3m ago",
		},
		{
			name:       "shows synced just now when under one minute",
			lastSynced: time.Now().Add(-30 * time.Second),
			wantIn:     "✓ synced just now",
		},
		{
			name:    "shows error indicator on sync failure",
			syncErr: errors.New("api rate limited"),
			wantIn:  "✗ sync err",
		},
		{
			name:   "shows nothing when never synced and not syncing",
			wantIn: "NEXUS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.syncing = tt.syncing
			model.lastSynced = tt.lastSynced
			model.syncErr = tt.syncErr

			view := model.View()

			assert.Contains(t, view, tt.wantIn)
		})
	}
}
