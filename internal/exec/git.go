package exec

import (
	"fmt"
	osexec "os/exec"
	"strings"

	"github.com/m00nk0d3/nexus/internal/domain"
)

const (
	worktreePrefix   = "worktree "
	headPrefix       = "HEAD "
	branchPrefix     = "branch "
	branchHeadPrefix = "refs/heads/"
	detachedMarker   = "detached"
	cleanMarker      = "clean"
	lockedMarker     = "locked"
	lockedWithReason = "locked "
)

// GitCommand represents a git command executor
type GitCommand struct {
	repoPath string
	runner   commandRunner
}

type commandRunner func(repoPath string, args ...string) (string, error)

// NewGitCommand creates a new git command executor
func NewGitCommand(repoPath string) *GitCommand {
	return NewGitCommandWithRunner(repoPath, runGitCommand)
}

// NewGitCommandWithRunner creates a git command executor with an injected runner.
func NewGitCommandWithRunner(repoPath string, runner commandRunner) *GitCommand {
	return &GitCommand{
		repoPath: repoPath,
		runner:   runner,
	}
}

// ListWorktrees returns worktrees from `git worktree list --porcelain`.
func (g *GitCommand) ListWorktrees() ([]domain.Worktree, error) {
	output, err := g.run("list worktrees", "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	worktrees, err := parseWorktreeListPorcelain(output)
	if err != nil {
		return nil, fmt.Errorf("parse worktree porcelain: %w", err)
	}

	return worktrees, nil
}

// AddWorktree adds a new worktree at path for branch.
func (g *GitCommand) AddWorktree(path, branch string) error {
	return g.runNoOutput("add worktree", "worktree", "add", path, branch)
}

// RemoveWorktree removes the worktree at path.
func (g *GitCommand) RemoveWorktree(path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)

	return g.runNoOutput("remove worktree", args...)
}

// PruneWorktrees prunes stale worktree metadata.
func (g *GitCommand) PruneWorktrees() error {
	return g.runNoOutput("prune worktrees", "worktree", "prune")
}

// LockWorktree locks the worktree at path, optionally with a reason.
func (g *GitCommand) LockWorktree(path, reason string) error {
	args := []string{"worktree", "lock"}
	if reason != "" {
		args = append(args, "--reason", reason)
	}
	args = append(args, path)

	return g.runNoOutput("lock worktree", args...)
}

// UnlockWorktree unlocks the worktree at path.
func (g *GitCommand) UnlockWorktree(path string) error {
	return g.runNoOutput("unlock worktree", "worktree", "unlock", path)
}

func (g *GitCommand) run(op string, args ...string) (string, error) {
	output, err := g.runner(g.repoPath, args...)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return output, nil
}

func (g *GitCommand) runNoOutput(op string, args ...string) error {
	_, err := g.run(op, args...)
	return err
}

func runGitCommand(repoPath string, args ...string) (string, error) {
	cmd := osexec.Command("git", args...)
	cmd.Dir = repoPath

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("run git %s: %w", strings.Join(args, " "), err)
	}

	return string(out), nil
}

func parseWorktreeListPorcelain(porcelain string) ([]domain.Worktree, error) {
	var (
		worktrees []domain.Worktree
		current   *domain.Worktree
	)

	appendCurrent := func() {
		if current != nil {
			worktrees = append(worktrees, *current)
		}
	}

	lines := strings.Split(porcelain, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if path, ok := strings.CutPrefix(line, worktreePrefix); ok {
			appendCurrent()
			current = &domain.Worktree{
				Path: path,
			}
			continue
		}

		if current == nil {
			return nil, fmt.Errorf("metadata before worktree: %q", line)
		}

		switch {
		case strings.HasPrefix(line, headPrefix):
			current.CommitSHA = strings.TrimPrefix(line, headPrefix)
		case strings.HasPrefix(line, branchPrefix):
			branchRef := strings.TrimPrefix(line, branchPrefix)
			current.Branch = strings.TrimPrefix(branchRef, branchHeadPrefix)
		case line == detachedMarker:
			if current.Branch == "" {
				current.Branch = detachedMarker
			}
		case line == cleanMarker:
			current.IsClean = true
		case line == lockedMarker || strings.HasPrefix(line, lockedWithReason):
			current.IsLocked = true
		}
	}

	appendCurrent()

	return worktrees, nil
}
