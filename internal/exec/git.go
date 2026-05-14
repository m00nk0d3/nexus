package exec

// GitCommand represents a git command executor
type GitCommand struct {
	repoPath string
}

// NewGitCommand creates a new git command executor
func NewGitCommand(repoPath string) *GitCommand {
	return &GitCommand{repoPath: repoPath}
}
