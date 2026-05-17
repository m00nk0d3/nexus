package domain

// WorktreeContext holds the current git state of a worktree for passing to AI agents.
type WorktreeContext struct {
	Path         string
	Branch       string
	GitStatus    string   // output of git status --short
	RecentLog    string   // last 10 commits: sha + message
	ChangedFiles []string // from git diff --name-only HEAD
	DiffSummary  string   // git diff --stat HEAD
}
