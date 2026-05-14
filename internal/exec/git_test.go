package exec

import (
	"errors"
	"testing"

	"github.com/m00nk0d3/nexus/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGitCommand(t *testing.T) {
	tests := []struct {
		name     string
		repoPath string
		expected string
	}{
		{
			name:     "creates git command executor",
			repoPath: "/home/user/repo",
			expected: "/home/user/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewGitCommand(tt.repoPath)
			assert.NotNil(t, cmd)
			assert.Equal(t, tt.expected, cmd.repoPath)
		})
	}
}

func TestParseWorktreeListPorcelain(t *testing.T) {
	tests := []struct {
		name        string
		porcelain   string
		expected    []domain.Worktree
		expectError string
	}{
		{
			name: "minimal valid porcelain parse into one worktree",
			porcelain: "worktree /repo/main\n" +
				"HEAD 1111111111111111111111111111111111111111\n" +
				"branch refs/heads/main\n",
			expected: []domain.Worktree{
				{
					Path:      "/repo/main",
					CommitSHA: "1111111111111111111111111111111111111111",
					Branch:    "main",
				},
			},
		},
		{
			name: "multiple worktrees with detached fallback locked marker and cleanliness marker",
			porcelain: "worktree /repo/main\n" +
				"HEAD aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n" +
				"branch refs/heads/main\n" +
				"clean\n" +
				"\n" +
				"worktree /repo/feature\n" +
				"HEAD bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\n" +
				"detached\n" +
				"locked reason\n",
			expected: []domain.Worktree{
				{
					Path:      "/repo/main",
					CommitSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Branch:    "main",
					IsClean:   true,
				},
				{
					Path:      "/repo/feature",
					CommitSHA: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
					Branch:    "detached",
					IsLocked:  true,
				},
			},
		},
		{
			name: "metadata line before worktree returns error",
			porcelain: "HEAD cccccccccccccccccccccccccccccccccccccccc\n" +
				"worktree /repo/main\n",
			expectError: "metadata before worktree",
		},
		{
			name: "unknown metadata ignored",
			porcelain: "worktree /repo/main\n" +
				"HEAD dddddddddddddddddddddddddddddddddddddddd\n" +
				"branch refs/heads/main\n" +
				"custom-field value\n",
			expected: []domain.Worktree{
				{
					Path:      "/repo/main",
					CommitSHA: "dddddddddddddddddddddddddddddddddddddddd",
					Branch:    "main",
				},
			},
		},
		{
			name: "non heads branch ref is preserved as-is",
			porcelain: "worktree /repo/remote\n" +
				"HEAD eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee\n" +
				"branch refs/remotes/origin/main\n",
			expected: []domain.Worktree{
				{
					Path:      "/repo/remote",
					CommitSHA: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
					Branch:    "refs/remotes/origin/main",
				},
			},
		},
		{
			name: "final worktree is appended without trailing newline",
			porcelain: "worktree /repo/main\n" +
				"HEAD ffffffffffffffffffffffffffffffffffffffff\n" +
				"branch refs/heads/main\n" +
				"\n" +
				"worktree /repo/feature\n" +
				"HEAD 9999999999999999999999999999999999999999\n" +
				"branch refs/heads/feature",
			expected: []domain.Worktree{
				{
					Path:      "/repo/main",
					CommitSHA: "ffffffffffffffffffffffffffffffffffffffff",
					Branch:    "main",
				},
				{
					Path:      "/repo/feature",
					CommitSHA: "9999999999999999999999999999999999999999",
					Branch:    "feature",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseWorktreeListPorcelain(tt.porcelain)

			if tt.expectError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestGitCommand_ListWorktrees(t *testing.T) {
	tests := []struct {
		name      string
		repoPath  string
		output    string
		runErr    error
		want      []domain.Worktree
		wantArgs  []string
		expectErr string
	}{
		{
			name:     "invokes git worktree list porcelain and parses output",
			repoPath: "/repo/main",
			output: "worktree /repo/main\n" +
				"HEAD eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee\n" +
				"branch refs/heads/main\n",
			want: []domain.Worktree{
				{
					Path:      "/repo/main",
					CommitSHA: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
					Branch:    "main",
				},
			},
			wantArgs: []string{"worktree", "list", "--porcelain"},
		},
		{
			name:      "returns runner error",
			repoPath:  "/repo/main",
			runErr:    errors.New("git failed"),
			wantArgs:  []string{"worktree", "list", "--porcelain"},
			expectErr: "git failed",
		},
		{
			name: "returns parse error with context",
			repoPath: "/repo/main",
			output: "HEAD badbadbadbadbadbadbadbadbadbadbadbadbadb\n" +
				"worktree /repo/main\n",
			wantArgs:  []string{"worktree", "list", "--porcelain"},
			expectErr: "parse worktree porcelain: metadata before worktree",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var called bool
			var calledRepoPath string
			var calledArgs []string

			runner := func(repoPath string, args ...string) (string, error) {
				called = true
				calledRepoPath = repoPath
				calledArgs = append([]string{}, args...)
				return tt.output, tt.runErr
			}

			cmd := NewGitCommandWithRunner(tt.repoPath, runner)

			actual, err := cmd.ListWorktrees()

			require.True(t, called, "expected command runner to be invoked")
			assert.Equal(t, tt.repoPath, calledRepoPath)
			assert.Equal(t, tt.wantArgs, calledArgs)

			if tt.expectErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, actual)
		})
	}
}

func TestGitCommand_AddWorktree(t *testing.T) {
	tests := []struct {
		name      string
		repoPath  string
		path      string
		branch    string
		runErr    error
		wantArgs  []string
		expectErr string
	}{
		{
			name:     "invokes git worktree add with path and branch",
			repoPath: "/repo/main",
			path:     "/repo/feature",
			branch:   "feature-branch",
			wantArgs: []string{"worktree", "add", "/repo/feature", "feature-branch"},
		},
		{
			name:      "returns runner error",
			repoPath:  "/repo/main",
			path:      "/repo/feature",
			branch:    "feature-branch",
			runErr:    errors.New("git failed"),
			wantArgs:  []string{"worktree", "add", "/repo/feature", "feature-branch"},
			expectErr: "git failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var called bool
			var calledRepoPath string
			var calledArgs []string

			runner := func(repoPath string, args ...string) (string, error) {
				called = true
				calledRepoPath = repoPath
				calledArgs = append([]string{}, args...)
				return "", tt.runErr
			}

			cmd := NewGitCommandWithRunner(tt.repoPath, runner)

			err := cmd.AddWorktree(tt.path, tt.branch)

			require.True(t, called, "expected command runner to be invoked")
			assert.Equal(t, tt.repoPath, calledRepoPath)
			assert.Equal(t, tt.wantArgs, calledArgs)

			if tt.expectErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestGitCommand_RemoveWorktree(t *testing.T) {
	tests := []struct {
		name      string
		repoPath  string
		path      string
		force     bool
		runErr    error
		wantArgs  []string
		expectErr string
	}{
		{
			name:     "invokes git worktree remove without force when force false",
			repoPath: "/repo/main",
			path:     "/repo/feature",
			force:    false,
			wantArgs: []string{"worktree", "remove", "/repo/feature"},
		},
		{
			name:     "invokes git worktree remove with force when force true",
			repoPath: "/repo/main",
			path:     "/repo/feature",
			force:    true,
			wantArgs: []string{"worktree", "remove", "--force", "/repo/feature"},
		},
		{
			name:      "returns runner error",
			repoPath:  "/repo/main",
			path:      "/repo/feature",
			force:     true,
			runErr:    errors.New("git failed"),
			wantArgs:  []string{"worktree", "remove", "--force", "/repo/feature"},
			expectErr: "git failed",
		},
		{
			name:     "places force flag before dash-prefixed path",
			repoPath: "/repo/main",
			path:     "--path-like-argument",
			force:    true,
			wantArgs: []string{"worktree", "remove", "--force", "--path-like-argument"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var called bool
			var calledRepoPath string
			var calledArgs []string

			runner := func(repoPath string, args ...string) (string, error) {
				called = true
				calledRepoPath = repoPath
				calledArgs = append([]string{}, args...)
				return "", tt.runErr
			}

			cmd := NewGitCommandWithRunner(tt.repoPath, runner)

			err := cmd.RemoveWorktree(tt.path, tt.force)

			require.True(t, called, "expected command runner to be invoked")
			assert.Equal(t, tt.repoPath, calledRepoPath)
			assert.Equal(t, tt.wantArgs, calledArgs)

			if tt.expectErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestGitCommand_PruneWorktrees(t *testing.T) {
	tests := []struct {
		name      string
		repoPath  string
		runErr    error
		wantArgs  []string
		expectErr string
	}{
		{
			name:     "invokes git worktree prune",
			repoPath: "/repo/main",
			wantArgs: []string{"worktree", "prune"},
		},
		{
			name:      "returns runner error",
			repoPath:  "/repo/main",
			runErr:    errors.New("git failed"),
			wantArgs:  []string{"worktree", "prune"},
			expectErr: "git failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var called bool
			var calledRepoPath string
			var calledArgs []string

			runner := func(repoPath string, args ...string) (string, error) {
				called = true
				calledRepoPath = repoPath
				calledArgs = append([]string{}, args...)
				return "", tt.runErr
			}

			cmd := NewGitCommandWithRunner(tt.repoPath, runner)

			err := cmd.PruneWorktrees()

			require.True(t, called, "expected command runner to be invoked")
			assert.Equal(t, tt.repoPath, calledRepoPath)
			assert.Equal(t, tt.wantArgs, calledArgs)

			if tt.expectErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestGitCommand_LockWorktree(t *testing.T) {
	tests := []struct {
		name      string
		repoPath  string
		path      string
		reason    string
		runErr    error
		wantArgs  []string
		expectErr string
	}{
		{
			name:     "invokes git worktree lock without reason when empty",
			repoPath: "/repo/main",
			path:     "/repo/feature",
			reason:   "",
			wantArgs: []string{"worktree", "lock", "/repo/feature"},
		},
		{
			name:     "invokes git worktree lock with reason when provided",
			repoPath: "/repo/main",
			path:     "/repo/feature",
			reason:   "active development",
			wantArgs: []string{"worktree", "lock", "--reason", "active development", "/repo/feature"},
		},
		{
			name:      "returns runner error",
			repoPath:  "/repo/main",
			path:      "/repo/feature",
			reason:    "active development",
			runErr:    errors.New("git failed"),
			wantArgs:  []string{"worktree", "lock", "--reason", "active development", "/repo/feature"},
			expectErr: "git failed",
		},
		{
			name:     "keeps reason as single argument and path last",
			repoPath: "/repo/main",
			path:     "/repo/feature",
			reason:   "budget variance analysis",
			wantArgs: []string{"worktree", "lock", "--reason", "budget variance analysis", "/repo/feature"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var called bool
			var calledRepoPath string
			var calledArgs []string

			runner := func(repoPath string, args ...string) (string, error) {
				called = true
				calledRepoPath = repoPath
				calledArgs = append([]string{}, args...)
				return "", tt.runErr
			}

			cmd := NewGitCommandWithRunner(tt.repoPath, runner)

			err := cmd.LockWorktree(tt.path, tt.reason)

			require.True(t, called, "expected command runner to be invoked")
			assert.Equal(t, tt.repoPath, calledRepoPath)
			assert.Equal(t, tt.wantArgs, calledArgs)

			if tt.expectErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestGitCommand_UnlockWorktree(t *testing.T) {
	tests := []struct {
		name      string
		repoPath  string
		path      string
		runErr    error
		wantArgs  []string
		expectErr string
	}{
		{
			name:     "invokes git worktree unlock with path",
			repoPath: "/repo/main",
			path:     "/repo/feature",
			wantArgs: []string{"worktree", "unlock", "/repo/feature"},
		},
		{
			name:      "returns runner error",
			repoPath:  "/repo/main",
			path:      "/repo/feature",
			runErr:    errors.New("git failed"),
			wantArgs:  []string{"worktree", "unlock", "/repo/feature"},
			expectErr: "git failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var called bool
			var calledRepoPath string
			var calledArgs []string

			runner := func(repoPath string, args ...string) (string, error) {
				called = true
				calledRepoPath = repoPath
				calledArgs = append([]string{}, args...)
				return "", tt.runErr
			}

			cmd := NewGitCommandWithRunner(tt.repoPath, runner)

			err := cmd.UnlockWorktree(tt.path)

			require.True(t, called, "expected command runner to be invoked")
			assert.Equal(t, tt.repoPath, calledRepoPath)
			assert.Equal(t, tt.wantArgs, calledArgs)

			if tt.expectErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
