package exec

import (
	"encoding/json"
	"fmt"
	osexec "os/exec"
	"strings"

	"github.com/m00nk0d3/nexus/internal/domain"
)

// IssueCommand wraps the gh CLI for GitHub issue operations.
type IssueCommand struct {
	repoPath string
	runner   commandRunner
}

// NewIssueCommand creates a new IssueCommand using the real gh CLI.
func NewIssueCommand(repoPath string) *IssueCommand {
	return NewIssueCommandWithRunner(repoPath, runGhCommand)
}

// NewIssueCommandWithRunner creates an IssueCommand with an injected runner for testing.
func NewIssueCommandWithRunner(repoPath string, runner commandRunner) *IssueCommand {
	return &IssueCommand{repoPath: repoPath, runner: runner}
}

// ListOpenIssues returns all open GitHub issues via `gh issue list`.
func (c *IssueCommand) ListOpenIssues() ([]domain.Issue, error) {
	output, err := c.runner(c.repoPath, "issue", "list", "--json", "number,title,labels", "--state", "open", "--limit", "100")
	if err != nil {
		return nil, fmt.Errorf("list open issues: %w", err)
	}

	issues, err := parseIssueList(output)
	if err != nil {
		return nil, err
	}

	return issues, nil
}

// ghLabel is the JSON shape for a label returned by gh.
type ghLabel struct {
	Name string `json:"name"`
}

// ghIssue is the JSON shape returned by `gh issue list --json number,title,labels`.
type ghIssue struct {
	Number int       `json:"number"`
	Title  string    `json:"title"`
	Labels []ghLabel `json:"labels"`
}

func parseIssueList(raw string) ([]domain.Issue, error) {
	var gh []ghIssue
	if err := json.Unmarshal([]byte(raw), &gh); err != nil {
		return nil, fmt.Errorf("parse issue list: %w", err)
	}

	issues := make([]domain.Issue, 0, len(gh))
	for _, g := range gh {
		labels := make([]string, len(g.Labels))
		for i, l := range g.Labels {
			labels[i] = l.Name
		}
		issues = append(issues, domain.Issue{
			Number: g.Number,
			Title:  g.Title,
			Labels: labels,
		})
	}

	return issues, nil
}

// PRCommand wraps the gh CLI for GitHub pull request operations.
type PRCommand struct {
	repoPath string
	runner   commandRunner
}

// NewPRCommand creates a new PRCommand using the real gh CLI.
func NewPRCommand(repoPath string) *PRCommand {
	return NewPRCommandWithRunner(repoPath, runGhCommand)
}

// NewPRCommandWithRunner creates a PRCommand with an injected runner for testing.
func NewPRCommandWithRunner(repoPath string, runner commandRunner) *PRCommand {
	return &PRCommand{repoPath: repoPath, runner: runner}
}

const prFields = "number,title,body,headRefName,author,state,labels,isDraft"

// ListOpenPRs returns all open pull requests via `gh pr list`.
func (c *PRCommand) ListOpenPRs() ([]domain.PullRequest, error) {
	output, err := c.runner(c.repoPath, "pr", "list", "--json", prFields, "--state", "open", "--limit", "100")
	if err != nil {
		return nil, fmt.Errorf("list open prs: %w", err)
	}

	prs, err := parsePRList(output)
	if err != nil {
		return nil, err
	}

	return prs, nil
}

// GetPR returns a single pull request by number via `gh pr view`.
func (c *PRCommand) GetPR(number int) (*domain.PullRequest, error) {
	output, err := c.runner(c.repoPath, "pr", "view", fmt.Sprintf("%d", number), "--json", prFields)
	if err != nil {
		return nil, fmt.Errorf("get pr: %w", err)
	}

	pr, err := parsePR(output)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

// ghAuthor is the JSON shape for the author object returned by gh.
type ghAuthor struct {
	Login string `json:"login"`
}

// ghPR is the JSON shape returned by `gh pr list/view --json ...`.
type ghPR struct {
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	HeadRefName string    `json:"headRefName"`
	Author      ghAuthor  `json:"author"`
	State       string    `json:"state"`
	Labels      []ghLabel `json:"labels"`
	IsDraft     bool      `json:"isDraft"`
}

func ghPRToDomain(g ghPR) domain.PullRequest {
	labels := make([]string, len(g.Labels))
	for i, l := range g.Labels {
		labels[i] = l.Name
	}
	return domain.PullRequest{
		Number:  g.Number,
		Title:   g.Title,
		Body:    g.Body,
		Branch:  g.HeadRefName,
		Author:  g.Author.Login,
		State:   g.State,
		Labels:  labels,
		IsDraft: g.IsDraft,
	}
}

func parsePRList(raw string) ([]domain.PullRequest, error) {
	var gh []ghPR
	if err := json.Unmarshal([]byte(raw), &gh); err != nil {
		return nil, fmt.Errorf("parse pr list: %w", err)
	}

	prs := make([]domain.PullRequest, 0, len(gh))
	for _, g := range gh {
		prs = append(prs, ghPRToDomain(g))
	}

	return prs, nil
}

func parsePR(raw string) (*domain.PullRequest, error) {
	var g ghPR
	if err := json.Unmarshal([]byte(raw), &g); err != nil {
		return nil, fmt.Errorf("parse pr: %w", err)
	}

	pr := ghPRToDomain(g)
	return &pr, nil
}

func runGhCommand(repoPath string, args ...string) (string, error) {
	cmd := osexec.Command("gh", args...)
	cmd.Dir = repoPath

	out, err := cmd.CombinedOutput()
	output := string(out)
	if err != nil {
		trimmed := strings.TrimSpace(output)
		if trimmed != "" {
			return "", fmt.Errorf("run gh %s: %w; output: %s", strings.Join(args, " "), err, trimmed)
		}
		return "", fmt.Errorf("run gh %s: %w", strings.Join(args, " "), err)
	}

	return output, nil
}
