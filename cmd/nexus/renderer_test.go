package main

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/m00nk0d3/nexus/internal/domain"
	"github.com/m00nk0d3/nexus/internal/tui/styles"
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

// ---------------------------------------------------------------------------
// Phase 2: Issues & PRs View tests
// ---------------------------------------------------------------------------

// TestRenderIssueList_ContainsHeaders verifies that renderIssueList renders
// the required column headers: #, TITLE, STATUS, LABELS.
func TestRenderIssueList_ContainsHeaders(t *testing.T) {
	tests := []struct {
		name   string
		wantIn []string
	}{
		{
			name:   "renders all required column headers",
			wantIn: []string{"#", "TITLE", "STATUS", "LABELS"},
		},
	}

	theme := styles.NewTheme("digital-noir")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderIssueList(nil, 0, theme, 80, 20)
			for _, want := range tt.wantIn {
				assert.Contains(t, result, want)
			}
		})
	}
}

// TestRenderIssueList_ContainsIssueRows verifies that issue number, title, labels
// are rendered, and that the selected row has a ">" cursor.
func TestRenderIssueList_ContainsIssueRows(t *testing.T) {
	issues := []domain.Issue{
		{Number: 7, Title: "Fix the bug", Labels: []string{"bug", "p1"}},
		{Number: 8, Title: "Add feature", Labels: []string{"enhancement"}},
	}

	tests := []struct {
		name        string
		selectedIdx int
		wantIn      []string
	}{
		{
			name:        "renders issue number and title for all issues",
			selectedIdx: 0,
			wantIn:      []string{"7", "Fix the bug", "8", "Add feature"},
		},
		{
			name:        "renders issue labels",
			selectedIdx: 0,
			wantIn:      []string{"bug", "enhancement"},
		},
		{
			name:        "selected issue (idx 0) has > cursor",
			selectedIdx: 0,
			wantIn:      []string{"> "},
		},
		{
			name:        "selected issue (idx 1) has > cursor",
			selectedIdx: 1,
			wantIn:      []string{"> "},
		},
	}

	theme := styles.NewTheme("digital-noir")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderIssueList(issues, tt.selectedIdx, theme, 80, 20)
			for _, want := range tt.wantIn {
				assert.Contains(t, result, want)
			}
		})
	}
}

// TestRenderPRList_ContainsHeaders verifies that renderPRList renders
// the required column headers: #, TITLE, BRANCH, AUTHOR, STATUS.
func TestRenderPRList_ContainsHeaders(t *testing.T) {
	tests := []struct {
		name   string
		wantIn []string
	}{
		{
			name:   "renders all required column headers",
			wantIn: []string{"#", "TITLE", "BRANCH", "AUTHOR", "STATUS"},
		},
	}

	theme := styles.NewTheme("digital-noir")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderPRList(nil, 0, theme, 80, 20)
			for _, want := range tt.wantIn {
				assert.Contains(t, result, want)
			}
		})
	}
}

// TestRenderPRList_ContainsPRRows verifies that PR number, title, branch, author
// and state are rendered, and that the selected PR has a ">" cursor.
func TestRenderPRList_ContainsPRRows(t *testing.T) {
	prs := []domain.PullRequest{
		{Number: 42, Title: "My PR", Branch: "feat/awesome", Author: "alice", State: "OPEN"},
		{Number: 43, Title: "Fix PR", Branch: "fix/bug", Author: "bob", State: "OPEN"},
	}

	tests := []struct {
		name        string
		selectedIdx int
		wantIn      []string
	}{
		{
			name:        "renders PR number, title, branch, author, state",
			selectedIdx: 0,
			wantIn:      []string{"42", "My PR", "feat/awesome", "alice", "OPEN"},
		},
		{
			name:        "selected PR (idx 0) has > cursor",
			selectedIdx: 0,
			wantIn:      []string{"> "},
		},
		{
			name:        "second PR is also rendered",
			selectedIdx: 0,
			wantIn:      []string{"43", "Fix PR", "fix/bug", "bob"},
		},
	}

	theme := styles.NewTheme("digital-noir")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderPRList(prs, tt.selectedIdx, theme, 80, 20)
			for _, want := range tt.wantIn {
				assert.Contains(t, result, want)
			}
		})
	}
}

// TestRenderContextPanel_IssueContext verifies that when the active view is
// viewIssues, the context panel shows the selected issue's number, title, labels,
// and an "[g] Open in GitHub" hint.
func TestRenderContextPanel_IssueContext(t *testing.T) {
	tests := []struct {
		name        string
		issues      []domain.Issue
		selectedIdx int
		wantIn      []string
	}{
		{
			name: "shows Issue context header with number and title",
			issues: []domain.Issue{
				{Number: 14, Title: "Implement Issues View", Labels: []string{"feature"}},
			},
			selectedIdx: 0,
			wantIn:      []string{"Issue #14", "Implement Issues View"},
		},
		{
			name: "shows labels in [label] bracket format",
			issues: []domain.Issue{
				{Number: 5, Title: "Bug Report", Labels: []string{"bug", "critical"}},
			},
			selectedIdx: 0,
			wantIn:      []string{"[bug]", "[critical]"},
		},
		{
			name: "shows [g] Open in GitHub hint",
			issues: []domain.Issue{
				{Number: 1, Title: "Some Issue", Labels: nil},
			},
			selectedIdx: 0,
			wantIn:      []string{"[g] Open in GitHub"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.view = viewIssues
			model.issues = tt.issues
			model.selectedIssueIdx = tt.selectedIdx

			view := model.View()

			for _, want := range tt.wantIn {
				assert.Contains(t, view, want)
			}
		})
	}
}

// TestRenderContextPanel_PRContext verifies that when the active view is
// viewPRs, the context panel shows the selected PR's number, title, branch,
// author, state, and an "[g] Open in GitHub" hint.
func TestRenderContextPanel_PRContext(t *testing.T) {
	tests := []struct {
		name      string
		prs       []domain.PullRequest
		prIdx     int
		wantIn    []string
	}{
		{
			name: "shows PR context header with number and title",
			prs: []domain.PullRequest{
				{Number: 42, Title: "Add search feature", Branch: "feat/search", Author: "alice", State: "OPEN"},
			},
			prIdx:  0,
			wantIn: []string{"PR #42", "Add search feature"},
		},
		{
			name: "shows branch and author in context panel",
			prs: []domain.PullRequest{
				{Number: 10, Title: "Fix bug", Branch: "fix/null-ptr", Author: "bob", State: "OPEN"},
			},
			prIdx:  0,
			wantIn: []string{"fix/null-ptr", "bob"},
		},
		{
			name: "shows PR state in context panel",
			prs: []domain.PullRequest{
				{Number: 55, Title: "Open PR", Branch: "feat/draft", Author: "carol", State: "OPEN"},
			},
			prIdx:  0,
			wantIn: []string{"OPEN"},
		},
		{
			name: "shows [g] Open in GitHub hint",
			prs: []domain.PullRequest{
				{Number: 1, Title: "Some PR", Branch: "main", Author: "dev", State: "OPEN"},
			},
			prIdx:  0,
			wantIn: []string{"[g] Open in GitHub"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.view = viewPRs
			model.prs = tt.prs
			model.selectedPRIdx = tt.prIdx

			view := model.View()

			for _, want := range tt.wantIn {
				assert.Contains(t, view, want)
			}
		})
	}
}

// TestRenderFull_IssueViewShowsIssueList verifies that when view is viewIssues,
// renderFull (via model.View()) shows issue data in the main list panel.
func TestRenderFull_IssueViewShowsIssueList(t *testing.T) {
	tests := []struct {
		name   string
		issues []domain.Issue
		wantIn []string
	}{
		{
			name: "issue view renders issue number and title",
			issues: []domain.Issue{
				{Number: 99, Title: "Test Issue"},
			},
			wantIn: []string{"99", "Test Issue"},
		},
		{
			name: "issue view renders issue column headers",
			issues: []domain.Issue{
				{Number: 1, Title: "Any Issue"},
			},
			wantIn: []string{"TITLE", "STATUS"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.view = viewIssues
			model.issues = tt.issues

			view := model.View()

			for _, want := range tt.wantIn {
				assert.Contains(t, view, want)
			}
		})
	}
}

// TestRenderFull_PRViewShowsPRList verifies that when view is viewPRs,
// renderFull (via model.View()) shows PR data in the main list panel.
func TestRenderFull_PRViewShowsPRList(t *testing.T) {
	tests := []struct {
		name   string
		prs    []domain.PullRequest
		wantIn []string
	}{
		{
			name: "PR view renders PR number, title, and branch",
			prs: []domain.PullRequest{
				{Number: 88, Title: "Test PR", Branch: "feat/test", Author: "dev", State: "OPEN"},
			},
			wantIn: []string{"88", "Test PR", "feat/test"},
		},
		{
			name: "PR view renders PR column headers",
			prs: []domain.PullRequest{
				{Number: 1, Title: "Any PR", Branch: "main", Author: "dev", State: "OPEN"},
			},
			wantIn: []string{"BRANCH", "AUTHOR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.view = viewPRs
			model.prs = tt.prs

			view := model.View()

			for _, want := range tt.wantIn {
				assert.Contains(t, view, want)
			}
		})
	}
}

// TestRenderFull_WorktreesViewStillWorks verifies that when view is viewWorktrees,
// the existing worktree table still renders correctly.
func TestRenderFull_WorktreesViewStillWorks(t *testing.T) {
	tests := []struct {
		name      string
		worktrees []domain.Worktree
		wantIn    []string
	}{
		{
			name: "worktrees view shows worktree column headers",
			worktrees: []domain.Worktree{
				{Path: "/tmp/wt-main", Branch: "main", IsClean: true},
			},
			wantIn: []string{"NAME", "PATH", "STATUS"},
		},
		{
			name: "worktrees view shows worktree name",
			worktrees: []domain.Worktree{
				{Path: "/tmp/wt-main", Branch: "main", IsClean: true},
			},
			wantIn: []string{"wt-main"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.view = viewWorktrees
			model.Worktrees = tt.worktrees

			view := model.View()

			for _, want := range tt.wantIn {
				assert.Contains(t, view, want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// End Phase 2 tests (renderer_test.go)
// ---------------------------------------------------------------------------
