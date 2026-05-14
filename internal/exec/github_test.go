package exec

import (
	"errors"
	"testing"

	"github.com/m00nk0d3/nexus/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIssueCommand(t *testing.T) {
	cmd := NewIssueCommand("/repo")
	assert.NotNil(t, cmd)
	assert.Equal(t, "/repo", cmd.repoPath)
	assert.NotNil(t, cmd.runner)
}

func TestListOpenIssues_ValidOutput_ReturnsIssues(t *testing.T) {
	raw := `[{"number":5,"title":"[Phase 1] Implement - Create/delete modals","labels":[{"name":"phase-1"}]},{"number":6,"title":"[Phase 1] Implement - Switch worktree","labels":[{"name":"phase-1"},{"name":"enhancement"}]}]`

	runner := func(_ string, _ ...string) (string, error) {
		return raw, nil
	}

	cmd := NewIssueCommandWithRunner("/repo", runner)
	issues, err := cmd.ListOpenIssues()

	require.NoError(t, err)
	require.Len(t, issues, 2)

	assert.Equal(t, 5, issues[0].Number)
	assert.Equal(t, "[Phase 1] Implement - Create/delete modals", issues[0].Title)
	assert.Equal(t, []string{"phase-1"}, issues[0].Labels)

	assert.Equal(t, 6, issues[1].Number)
	assert.Equal(t, []string{"phase-1", "enhancement"}, issues[1].Labels)
}

func TestListOpenIssues_EmptyList_ReturnsEmptySlice(t *testing.T) {
	runner := func(_ string, _ ...string) (string, error) {
		return "[]", nil
	}

	cmd := NewIssueCommandWithRunner("/repo", runner)
	issues, err := cmd.ListOpenIssues()

	require.NoError(t, err)
	assert.Empty(t, issues)
}

func TestListOpenIssues_NoLabels_ReturnsIssueWithEmptyLabels(t *testing.T) {
	raw := `[{"number":7,"title":"Some issue","labels":[]}]`

	runner := func(_ string, _ ...string) (string, error) {
		return raw, nil
	}

	cmd := NewIssueCommandWithRunner("/repo", runner)
	issues, err := cmd.ListOpenIssues()

	require.NoError(t, err)
	require.Len(t, issues, 1)
	assert.Empty(t, issues[0].Labels)
}

func TestListOpenIssues_RunnerError_PropagatesError(t *testing.T) {
	runner := func(_ string, _ ...string) (string, error) {
		return "", errors.New("gh: not logged in")
	}

	cmd := NewIssueCommandWithRunner("/repo", runner)
	_, err := cmd.ListOpenIssues()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "list open issues")
	assert.Contains(t, err.Error(), "gh: not logged in")
}

func TestListOpenIssues_InvalidJSON_ReturnsError(t *testing.T) {
	runner := func(_ string, _ ...string) (string, error) {
		return "not valid json", nil
	}

	cmd := NewIssueCommandWithRunner("/repo", runner)
	_, err := cmd.ListOpenIssues()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse issue list")
}

func TestListOpenIssues_PassesCorrectArgs(t *testing.T) {
	var capturedArgs []string

	runner := func(_ string, args ...string) (string, error) {
		capturedArgs = args
		return "[]", nil
	}

	cmd := NewIssueCommandWithRunner("/repo", runner)
	_, err := cmd.ListOpenIssues()

	require.NoError(t, err)
	assert.Equal(t, []string{"issue", "list", "--json", "number,title,labels", "--state", "open"}, capturedArgs)
}

func TestListOpenIssues_MapsDomainsCorrectly(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected []domain.Issue
	}{
		{
			name: "single issue with multiple labels",
			raw:  `[{"number":42,"title":"Fix the thing","labels":[{"name":"bug"},{"name":"priority-high"}]}]`,
			expected: []domain.Issue{
				{Number: 42, Title: "Fix the thing", Labels: []string{"bug", "priority-high"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := func(_ string, _ ...string) (string, error) { return tt.raw, nil }
			cmd := NewIssueCommandWithRunner("/repo", runner)

			issues, err := cmd.ListOpenIssues()

			require.NoError(t, err)
			assert.Equal(t, tt.expected, issues)
		})
	}
}
