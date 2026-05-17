package exec

import (
	"fmt"

	"github.com/m00nk0d3/nexus/internal/domain"
)

const defaultLogCount = 10

// BuildContext collects the current git state of the worktree at path and
// returns a populated WorktreeContext ready to pass to AI agents.
func BuildContext(gitCmd *GitCommand, path string) (*domain.WorktreeContext, error) {
	status, err := gitCmd.GitStatus(path)
	if err != nil {
		return nil, fmt.Errorf("build context: %w", err)
	}

	log, err := gitCmd.GitLog(path, defaultLogCount)
	if err != nil {
		return nil, fmt.Errorf("build context: %w", err)
	}

	changedFiles, err := gitCmd.GitDiffNameOnly(path)
	if err != nil {
		return nil, fmt.Errorf("build context: %w", err)
	}

	diffStat, err := gitCmd.GitDiffStat(path)
	if err != nil {
		return nil, fmt.Errorf("build context: %w", err)
	}

	return &domain.WorktreeContext{
		Path:         path,
		GitStatus:    status,
		RecentLog:    log,
		ChangedFiles: changedFiles,
		DiffSummary:  diffStat,
	}, nil
}
