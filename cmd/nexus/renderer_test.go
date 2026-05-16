package main

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
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
			result := renderIssueList(nil, 0, theme, 80, 20, false)
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
			result := renderIssueList(issues, tt.selectedIdx, theme, 80, 20, false)
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
			result := renderPRList(nil, 0, theme, 80, 20, false)
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
			result := renderPRList(prs, tt.selectedIdx, theme, 80, 20, false)
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
// Phase 3: Issue #15 — PR Context Panel & GH:ID Column tests
// ---------------------------------------------------------------------------

// TestRenderer_ContextPanel_WithLinkedPR verifies that when a worktree has a
// LinkedPR, the context panel shows PR-specific information: number, title,
// author, state, and labels in the correct format.
func TestRenderer_ContextPanel_WithLinkedPR(t *testing.T) {
	tests := []struct {
		name     string
		worktree domain.Worktree
		wantIn   []string
	}{
		{
			name: "shows PR context header with number",
			worktree: domain.Worktree{
				Path:   "/tmp/feat-login",
				Branch: "feat/login",
				IsClean: true,
				LinkedPR: &domain.PullRequest{
					Number: 42,
					Title:  "Add login feature",
					Author: "alice",
					State:  "OPEN",
					Labels: []string{},
				},
			},
			wantIn: []string{"Context: PR #42"},
		},
		{
			name: "shows PR title in context",
			worktree: domain.Worktree{
				Path:   "/tmp/feat-login",
				Branch: "feat/login",
				IsClean: true,
				LinkedPR: &domain.PullRequest{
					Number: 42,
					Title:  "Add login feature",
					Author: "alice",
					State:  "OPEN",
					Labels: []string{},
				},
			},
			wantIn: []string{"Add login feature"},
		},
		{
			name: "shows GH Title label and author with @ prefix",
			worktree: domain.Worktree{
				Path:   "/tmp/feat-login",
				Branch: "feat/login",
				IsClean: true,
				LinkedPR: &domain.PullRequest{
					Number: 42,
					Title:  "Add login feature",
					Author: "alice",
					State:  "OPEN",
					Labels: []string{},
				},
			},
			wantIn: []string{"GH Title:", "Author: @alice"},
		},
		{
			name: "shows status dot and state",
			worktree: domain.Worktree{
				Path:   "/tmp/feat-login",
				Branch: "feat/login",
				IsClean: true,
				LinkedPR: &domain.PullRequest{
					Number: 42,
					Title:  "Add login feature",
					Author: "alice",
					State:  "OPEN",
					Labels: []string{},
				},
			},
			wantIn: []string{"Status: ●", "OPEN"},
		},
		{
			name: "shows labels in bracket format",
			worktree: domain.Worktree{
				Path:   "/tmp/feat-auth",
				Branch: "feat/auth",
				IsClean: true,
				LinkedPR: &domain.PullRequest{
					Number: 99,
					Title:  "Auth refactor",
					Author: "bob",
					State:  "OPEN",
					Labels: []string{"enhancement", "backend"},
				},
			},
			wantIn: []string{"Labels: [enhancement][backend]"},
		},
		{
			name: "shows agent commands section",
			worktree: domain.Worktree{
				Path:   "/tmp/feat-login",
				Branch: "feat/login",
				IsClean: true,
				LinkedPR: &domain.PullRequest{
					Number: 42,
					Title:  "Add login feature",
					Author: "alice",
					State:  "MERGED",
					Labels: []string{},
				},
			},
			wantIn: []string{"AGENT COMMANDS:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.view = viewWorktrees
			model.Worktrees = []domain.Worktree{tt.worktree}
			model.selectedIdx = 0

			view := model.View()

			for _, want := range tt.wantIn {
				assert.Contains(t, view, want)
			}
		})
	}
}

// TestRenderer_ContextPanel_NoLinkedPR verifies that when a worktree has no
// LinkedPR, the context panel shows the worktree basename, branch, and path.
func TestRenderer_ContextPanel_NoLinkedPR(t *testing.T) {
	tests := []struct {
		name     string
		worktree domain.Worktree
		wantIn   []string
		wantOut  []string
	}{
		{
			name: "shows worktree basename as context header",
			worktree: domain.Worktree{
				Path:    "/tmp/feat-search",
				Branch:  "feat/search",
				IsClean: true,
			},
			wantIn: []string{"Context: feat-search"},
		},
		{
			name: "shows branch",
			worktree: domain.Worktree{
				Path:    "/tmp/feat-search",
				Branch:  "feat/search",
				IsClean: true,
			},
			wantIn: []string{"Branch: feat/search"},
		},
		{
			name: "shows path",
			worktree: domain.Worktree{
				Path:    "/tmp/feat-search",
				Branch:  "feat/search",
				IsClean: true,
			},
			wantIn: []string{"Path: /tmp/feat-search"},
		},
		{
			name: "does not show PR-specific fields",
			worktree: domain.Worktree{
				Path:    "/tmp/feat-search",
				Branch:  "feat/search",
				IsClean: true,
			},
			wantOut: []string{"GH Title:", "Author: @", "Labels:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.view = viewWorktrees
			model.Worktrees = []domain.Worktree{tt.worktree}
			model.selectedIdx = 0

			view := model.View()

			for _, want := range tt.wantIn {
				assert.Contains(t, view, want)
			}
			for _, notWant := range tt.wantOut {
				assert.NotContains(t, view, notWant)
			}
		})
	}
}

// TestRenderer_ContextPanel_AgentCommandsAlwaysVisible verifies that the agent
// commands [a], [c], and [s] are always shown in the worktree context panel,
// regardless of whether a LinkedPR is present.
func TestRenderer_ContextPanel_AgentCommandsAlwaysVisible(t *testing.T) {
	tests := []struct {
		name     string
		worktree domain.Worktree
	}{
		{
			name: "agent commands present with LinkedPR",
			worktree: domain.Worktree{
				Path:    "/tmp/feat-auth",
				Branch:  "feat/auth",
				IsClean: true,
				LinkedPR: &domain.PullRequest{
					Number: 7,
					Title:  "Auth",
					Author: "carol",
					State:  "OPEN",
				},
			},
		},
		{
			name: "agent commands present without LinkedPR",
			worktree: domain.Worktree{
				Path:    "/tmp/feat-search",
				Branch:  "feat/search",
				IsClean: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.view = viewWorktrees
			model.Worktrees = []domain.Worktree{tt.worktree}
			model.selectedIdx = 0

			view := model.View()

			assert.Contains(t, view, "[a] Spawn Claude Code")
			assert.Contains(t, view, "[c] Spawn Copilot")
			assert.Contains(t, view, "[s] Open Shell in WT")
		})
	}
}

// TestRenderer_GHIDColumn_NoLinkedPR verifies that when a worktree has no
// LinkedPR, the GH:ID column shows "-" (not empty string).
func TestRenderer_GHIDColumn_NoLinkedPR(t *testing.T) {
	tests := []struct {
		name     string
		worktree domain.Worktree
	}{
		{
			name: "GH:ID shows dash when no PR linked (non-selected row)",
			worktree: domain.Worktree{
				Path:    "/tmp/wt-no-pr",
				Branch:  "main",
				IsClean: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.view = viewWorktrees
			model.Worktrees = []domain.Worktree{tt.worktree}
			// Set selectedIdx to -1 so the row is NOT selected, ensuring we hit the
			// non-selected branch of the rendering code.
			model.selectedIdx = -1
			model.width = 300

			view := model.View()

			// The GH:ID column should contain "- " (dash padded to 6 chars). We verify
			// the worktree row is actually rendered and the column shows the placeholder.
			assert.Contains(t, view, "wt-no-pr")
			assert.Contains(t, view, "-     ") // "- " padded to 6 chars
		})
	}
}

// TestRenderer_GHIDColumn_WithLinkedPR verifies that when a worktree has a
// LinkedPR, the GH:ID column shows the PR number.
func TestRenderer_GHIDColumn_WithLinkedPR(t *testing.T) {
	tests := []struct {
		name        string
		worktree    domain.Worktree
		wantPRNum   string
	}{
		{
			name: "GH:ID shows PR number for linked PR (non-selected row)",
			worktree: domain.Worktree{
				Path:    "/tmp/feat-login",
				Branch:  "feat/login",
				IsClean: true,
				LinkedPR: &domain.PullRequest{
					Number: 1442,
					Title:  "Login feature",
					Author: "alice",
					State:  "OPEN",
				},
			},
			wantPRNum: "1442",
		},
		{
			name: "GH:ID shows PR number for merged PR",
			worktree: domain.Worktree{
				Path:    "/tmp/fix-bug",
				Branch:  "fix/bug",
				IsClean: true,
				LinkedPR: &domain.PullRequest{
					Number: 55,
					Title:  "Bug fix",
					Author: "bob",
					State:  "MERGED",
				},
			},
			wantPRNum: "55",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)
			model.view = viewWorktrees
			model.Worktrees = []domain.Worktree{tt.worktree}
			model.selectedIdx = -1
			model.width = 300

			view := model.View()

			assert.Contains(t, view, tt.wantPRNum)
		})
	}
}

// ---------------------------------------------------------------------------
// Phase 4: clipContent & panel focus rendering tests
// ---------------------------------------------------------------------------

// TestClipContent verifies that clipContent correctly slices content lines
// with scroll offset and maxLines constraints.
func TestClipContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		offset   int
		maxLines int
		wantOut  string
	}{
		{
			name:     "no clipping needed when within bounds",
			content:  "line1\nline2\nline3",
			offset:   0,
			maxLines: 10,
			wantOut:  "line1\nline2\nline3",
		},
		{
			name:     "clips to maxLines",
			content:  "a\nb\nc\nd\ne",
			offset:   0,
			maxLines: 3,
			wantOut:  "a\nb\nc",
		},
		{
			name:     "applies offset",
			content:  "a\nb\nc\nd",
			offset:   1,
			maxLines: 10,
			wantOut:  "b\nc\nd",
		},
		{
			name:     "applies offset and clips",
			content:  "a\nb\nc\nd\ne",
			offset:   1,
			maxLines: 2,
			wantOut:  "b\nc",
		},
		{
			name:     "offset beyond content clamps to last line",
			content:  "a\nb",
			offset:   5,
			maxLines: 10,
			wantOut:  "b",
		},
		{
			name:     "zero maxLines returns content unchanged",
			content:  "a\nb\nc",
			offset:   0,
			maxLines: 0,
			wantOut:  "a\nb\nc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clipContent(tt.content, tt.offset, tt.maxLines)
			assert.Equal(t, tt.wantOut, result)
		})
	}
}

// TestRenderFull_FooterContainsTabAndJKHints verifies that the footer bar
// includes the new Tab panel and j/k navigation hints.
func TestRenderFull_FooterContainsTabAndJKHints(t *testing.T) {
	model := NewModel()
	require.NotNil(t, model)

	view := model.View()

	assert.Contains(t, view, "[Tab]")
	assert.Contains(t, view, "[j/k]")
	// Old hints must still be present
	assert.Contains(t, view, "[Enter] Select")
	assert.Contains(t, view, "[esc] Quit")
}

// ---------------------------------------------------------------------------
// End Phase 4 tests (renderer_test.go)
// ---------------------------------------------------------------------------

func TestPrStateColor(t *testing.T) {
	tests := []struct {
		state     string
		wantColor string
	}{
		{"OPEN", "#00D9FF"},
		{"MERGED", "#9B59B6"},
		{"CLOSED", "#E74C3C"},
		{"UNKNOWN", "#4A5568"},
		{"", "#4A5568"},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			got := prStateColor(tt.state)
			assert.Equal(t, lipgloss.Color(tt.wantColor), got)
		})
	}
}
