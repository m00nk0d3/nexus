package domain

// PullRequest represents a GitHub pull request.
type PullRequest struct {
	Number  int
	Title   string
	Branch  string
	Author  string
	State   string // "OPEN", "MERGED", "CLOSED"
	Labels  []string
	IsDraft bool
}
