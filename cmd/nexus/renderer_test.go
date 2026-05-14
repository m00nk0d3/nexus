package main

import (
	"path/filepath"
	"testing"

	"github.com/m00nk0d3/nexus/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelView_WithWorktrees_ContainsListHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
	}{
		{
			name:    "renders required worktree list headers",
			headers: []string{"Name", "Path", "Status", "Commit SHA", "Locked"},
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

func TestModelView_WithWorktreeRow_MapsDomainFieldsToRenderedValues(t *testing.T) {
	tests := []struct {
		name             string
		worktree         domain.Worktree
		expectedStatus   string
		expectedLockText string
	}{
		{
			name: "maps clean and unlocked worktree",
			worktree: domain.Worktree{
				Path:      filepath.Join("tmp", "wt-clean"),
				CommitSHA: "aaaaaaaaaaaaaaaa",
				IsClean:   true,
				IsLocked:  false,
			},
			expectedStatus:   "clean",
			expectedLockText: "unlocked",
		},
		{
			name: "maps dirty and locked worktree",
			worktree: domain.Worktree{
				Path:      filepath.Join("tmp", "wt-dirty"),
				CommitSHA: "bbbbbbbbbbbbbbbb",
				IsClean:   false,
				IsLocked:  true,
			},
			expectedStatus:   "dirty",
			expectedLockText: "locked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			require.NotNil(t, model)

			model.Worktrees = []domain.Worktree{tt.worktree}

			view := model.View()

			expectedName := filepath.Base(tt.worktree.Path)
			assert.Contains(t, view, expectedName)
			assert.Contains(t, view, tt.worktree.Path)
			assert.Contains(t, view, tt.expectedStatus)
			assert.Contains(t, view, tt.worktree.CommitSHA)
			assert.Contains(t, view, tt.expectedLockText)
		})
	}
}
