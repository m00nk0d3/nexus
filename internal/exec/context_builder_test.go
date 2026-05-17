package exec

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildContext_PopulatesAllFields(t *testing.T) {
	const worktreePath = "/repo/feature"

	runner := func(repoPath string, args ...string) (string, error) {
		if len(args) == 0 {
			return "", nil
		}
		switch args[0] {
		case "status":
			return "M internal/exec/git.go\n?? new_file.go\n", nil
		case "log":
			return "abc1234 feat: add context builder\ndef5678 fix: handle empty status\n", nil
		case "diff":
			if len(args) >= 2 && args[1] == "--name-only" {
				return "internal/exec/git.go\ninternal/domain/context.go\n", nil
			}
			return " internal/exec/git.go | 42 ++++++\n 1 file changed, 42 insertions(+)\n", nil
		}
		return "", nil
	}

	cmd := NewGitCommandWithRunner("/repo/main", runner)
	ctx, err := BuildContext(cmd, worktreePath)

	require.NoError(t, err)
	assert.Equal(t, worktreePath, ctx.Path)
	assert.Equal(t, "M internal/exec/git.go\n?? new_file.go", ctx.GitStatus)
	assert.Equal(t, "abc1234 feat: add context builder\ndef5678 fix: handle empty status", ctx.RecentLog)
	assert.Equal(t, []string{"internal/exec/git.go", "internal/domain/context.go"}, ctx.ChangedFiles)
	assert.Equal(t, "internal/exec/git.go | 42 ++++++\n 1 file changed, 42 insertions(+)", ctx.DiffSummary)
}

func TestBuildContext_EmptyRepo_ReturnsEmptyStatus(t *testing.T) {
	runner := func(repoPath string, args ...string) (string, error) {
		return "", nil
	}

	cmd := NewGitCommandWithRunner("/repo/main", runner)
	ctx, err := BuildContext(cmd, "/repo/empty")

	require.NoError(t, err)
	assert.Equal(t, "/repo/empty", ctx.Path)
	assert.Equal(t, "", ctx.GitStatus)
	assert.Equal(t, "", ctx.RecentLog)
	assert.Equal(t, []string{}, ctx.ChangedFiles)
	assert.Equal(t, "", ctx.DiffSummary)
}

func TestBuildContext_GitStatusError_ReturnsError(t *testing.T) {
	runner := func(repoPath string, args ...string) (string, error) {
		if args[0] == "status" {
			return "", errors.New("permission denied")
		}
		return "", nil
	}

	cmd := NewGitCommandWithRunner("/repo/main", runner)
	_, err := BuildContext(cmd, "/repo/feature")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "build context")
	assert.Contains(t, err.Error(), "git status")
}

func TestBuildContext_GitLogError_ReturnsError(t *testing.T) {
	runner := func(repoPath string, args ...string) (string, error) {
		if args[0] == "log" {
			return "", errors.New("not a git repo")
		}
		return "", nil
	}

	cmd := NewGitCommandWithRunner("/repo/main", runner)
	_, err := BuildContext(cmd, "/repo/feature")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "build context")
	assert.Contains(t, err.Error(), "git log")
}

func TestBuildContext_GitDiffNameOnlyError_ReturnsError(t *testing.T) {
	runner := func(repoPath string, args ...string) (string, error) {
		if args[0] == "diff" && len(args) >= 2 && args[1] == "--name-only" {
			return "", errors.New("fatal: ambiguous argument")
		}
		return "", nil
	}

	cmd := NewGitCommandWithRunner("/repo/main", runner)
	_, err := BuildContext(cmd, "/repo/feature")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "build context")
	assert.Contains(t, err.Error(), "git diff name-only")
}

func TestBuildContext_GitDiffStatError_ReturnsError(t *testing.T) {
	runner := func(repoPath string, args ...string) (string, error) {
		if args[0] == "diff" && len(args) >= 2 && args[1] == "--stat" {
			return "", errors.New("fatal: bad HEAD")
		}
		return "", nil
	}

	cmd := NewGitCommandWithRunner("/repo/main", runner)
	_, err := BuildContext(cmd, "/repo/feature")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "build context")
	assert.Contains(t, err.Error(), "git diff stat")
}
