package data

import (
	"encoding/json"
	"fmt"

	"github.com/m00nk0d3/nexus/internal/domain"
	"github.com/m00nk0d3/nexus/internal/exec"
)

// GitHubRepository persists and retrieves GitHub PR and Issue data from SQLite.
type GitHubRepository struct {
	db *DB
}

// NewGitHubRepository creates a new GitHubRepository backed by the given DB.
func NewGitHubRepository(db *DB) *GitHubRepository {
	return &GitHubRepository{db: db}
}

// UpsertPRs inserts or replaces all provided pull requests in the cache.
func (r *GitHubRepository) UpsertPRs(prs []domain.PullRequest) error {
	for _, pr := range prs {
		labels, err := json.Marshal(pr.Labels)
		if err != nil {
			return fmt.Errorf("upsert prs: marshal labels for pr %d: %w", pr.Number, err)
		}

		_, err = r.db.Conn.Exec(`
			INSERT INTO github_prs (number, title, branch, author, state, is_draft, labels, synced_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT(number) DO UPDATE SET
				title     = excluded.title,
				branch    = excluded.branch,
				author    = excluded.author,
				state     = excluded.state,
				is_draft  = excluded.is_draft,
				labels    = excluded.labels,
				synced_at = CURRENT_TIMESTAMP
		`, pr.Number, pr.Title, pr.Branch, pr.Author, pr.State, pr.IsDraft, string(labels))
		if err != nil {
			return fmt.Errorf("upsert prs: %w", err)
		}
	}
	return nil
}

// GetPRs returns all cached pull requests from the database.
func (r *GitHubRepository) GetPRs() ([]domain.PullRequest, error) {
	rows, err := r.db.Conn.Query(`
		SELECT number, title, branch, author, state, is_draft, labels
		FROM github_prs
		ORDER BY number
	`)
	if err != nil {
		return nil, fmt.Errorf("get prs: %w", err)
	}
	defer rows.Close()

	var prs []domain.PullRequest
	for rows.Next() {
		var pr domain.PullRequest
		var labelsJSON string
		var isDraftInt int

		if err := rows.Scan(&pr.Number, &pr.Title, &pr.Branch, &pr.Author, &pr.State, &isDraftInt, &labelsJSON); err != nil {
			return nil, fmt.Errorf("get prs: scan row: %w", err)
		}

		pr.IsDraft = isDraftInt != 0

		if err := json.Unmarshal([]byte(labelsJSON), &pr.Labels); err != nil {
			return nil, fmt.Errorf("get prs: parse labels for pr %d: %w", pr.Number, err)
		}

		prs = append(prs, pr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get prs: rows error: %w", err)
	}

	if prs == nil {
		prs = []domain.PullRequest{}
	}
	return prs, nil
}

// UpsertIssues inserts or replaces all provided issues in the cache.
func (r *GitHubRepository) UpsertIssues(issues []domain.Issue) error {
	for _, issue := range issues {
		labels, err := json.Marshal(issue.Labels)
		if err != nil {
			return fmt.Errorf("upsert issues: marshal labels for issue %d: %w", issue.Number, err)
		}

		_, err = r.db.Conn.Exec(`
			INSERT INTO github_issues (number, title, state, labels, synced_at)
			VALUES (?, ?, '', ?, CURRENT_TIMESTAMP)
			ON CONFLICT(number) DO UPDATE SET
				title     = excluded.title,
				labels    = excluded.labels,
				synced_at = CURRENT_TIMESTAMP
		`, issue.Number, issue.Title, string(labels))
		if err != nil {
			return fmt.Errorf("upsert issues: %w", err)
		}
	}
	return nil
}

// GetIssues returns all cached issues from the database.
func (r *GitHubRepository) GetIssues() ([]domain.Issue, error) {
	rows, err := r.db.Conn.Query(`
		SELECT number, title, labels
		FROM github_issues
		ORDER BY number
	`)
	if err != nil {
		return nil, fmt.Errorf("get issues: %w", err)
	}
	defer rows.Close()

	var issues []domain.Issue
	for rows.Next() {
		var issue domain.Issue
		var labelsJSON string

		if err := rows.Scan(&issue.Number, &issue.Title, &labelsJSON); err != nil {
			return nil, fmt.Errorf("get issues: scan row: %w", err)
		}

		if err := json.Unmarshal([]byte(labelsJSON), &issue.Labels); err != nil {
			return nil, fmt.Errorf("get issues: parse labels for issue %d: %w", issue.Number, err)
		}

		issues = append(issues, issue)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get issues: rows error: %w", err)
	}

	if issues == nil {
		issues = []domain.Issue{}
	}
	return issues, nil
}

// SyncPRs fetches open PRs from GitHub via the client and upserts them into the cache.
// On CLI error, the existing cached data is preserved and the error is returned.
func (r *GitHubRepository) SyncPRs(client *exec.PRCommand) error {
	prs, err := client.ListOpenPRs()
	if err != nil {
		return fmt.Errorf("sync prs: %w", err)
	}
	return r.UpsertPRs(prs)
}

// SyncIssues fetches open issues from GitHub via the client and upserts them into the cache.
// On CLI error, the existing cached data is preserved and the error is returned.
func (r *GitHubRepository) SyncIssues(client *exec.IssueCommand) error {
	issues, err := client.ListOpenIssues()
	if err != nil {
		return fmt.Errorf("sync issues: %w", err)
	}
	return r.UpsertIssues(issues)
}
