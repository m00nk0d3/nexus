package main

import (
	"path/filepath"
	"strings"

	"github.com/m00nk0d3/nexus/internal/domain"
)

func renderWorktreeList(worktrees []domain.Worktree) string {
	var b strings.Builder

	b.WriteString("Name\tPath\tStatus\tCommit SHA\tLocked\n")

	for _, wt := range worktrees {
		status := "dirty"
		if wt.IsClean {
			status = "clean"
		}

		locked := "unlocked"
		if wt.IsLocked {
			locked = "locked"
		}

		b.WriteString(filepath.Base(wt.Path))
		b.WriteString("\t")
		b.WriteString(wt.Path)
		b.WriteString("\t")
		b.WriteString(status)
		b.WriteString("\t")
		b.WriteString(wt.CommitSHA)
		b.WriteString("\t")
		b.WriteString(locked)
		b.WriteString("\n")
	}

	return b.String()
}
