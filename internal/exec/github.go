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
	output, err := c.runner(c.repoPath, "issue", "list", "--json", "number,title,labels", "--state", "open")
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
