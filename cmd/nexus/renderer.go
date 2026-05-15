package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/m00nk0d3/nexus/internal/domain"
	"github.com/m00nk0d3/nexus/internal/tui/styles"
)

const (
	appVersion     = "1.0"
	footerHints    = "[+/↓] Navigate | [Enter] Select | [t] Theme | [g] Open in GH | [esc] Quit"
	actionBarHints = "[c-n] New  [c-d] Delete  [c-l] Lock | [f1] Help"
)

type navItem struct {
	key   string
	label string
}

var navItems = []navItem{
	{"W", "WORKTREES"},
	{"I", "ISSUES"},
	{"P", "PRs"},
	{"T", "THEMES"},
}

// renderFull builds the complete 3-pane TUI layout using the theme at themeIdx.
func renderFull(worktrees []domain.Worktree, selectedIdx int, repoPath string, themeIdx int) string {
	theme := styles.NewTheme(styles.Themes[themeIdx])

	header := renderHeader(repoPath, theme)
	nav := renderNavRail(theme)
	list := renderWorktreePanel(worktrees, selectedIdx, theme)
	ctx := renderContextPanel(worktrees, selectedIdx, theme)
	mainRow := lipgloss.JoinHorizontal(lipgloss.Top, nav, list, ctx)
	footer := renderFooterBar(theme, time.Now().UTC().Format("2006-01-02"))
	actionBar := renderActionBar(theme)

	return lipgloss.JoinVertical(lipgloss.Left, header, mainRow, footer, actionBar)
}

func renderHeader(repoPath string, theme styles.Theme) string {
	if repoPath == "" {
		repoPath = "./"
	}
	text := fmt.Sprintf(
		"NEXUS v%s: GIT WORKTREE ORCHESTRATOR | Repo: %s | Local Path: %s",
		appVersion, filepath.Base(repoPath), repoPath,
	)
	return theme.GetStyle("header").Width(120).Render(text)
}

func renderNavRail(theme styles.Theme) string {
	var b strings.Builder
	for i, item := range navItems {
		cursor := "  "
		if i == 0 {
			cursor = "> "
		}
		b.WriteString(fmt.Sprintf("%s%s: %s\n", cursor, item.key, item.label))
	}
	return theme.GetStyle("nav-rail").Width(18).Render(strings.TrimRight(b.String(), "\n"))
}

func renderWorktreePanel(worktrees []domain.Worktree, selectedIdx int, theme styles.Theme) string {
	headers := []string{"NAME", "PATH", "STATUS", "UPDATED", "GH:ID"}
	var content strings.Builder

	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4A5568")).Bold(true)
	content.WriteString(headerStyle.Render(strings.Join(headers, "   ")))
	content.WriteString("\n")

	for i, wt := range worktrees {
		name := filepath.Base(wt.Path)
		status := worktreeStatus(wt)
		ghID := ""
		if wt.LinkedPR != nil {
			ghID = fmt.Sprintf("%d", *wt.LinkedPR)
		}
		row := fmt.Sprintf("%-18s %-30s %-10s %-10s %-6s", name, wt.Path, status, "—", ghID)
		if i == selectedIdx {
			content.WriteString(theme.GetStyle("selected-row").Render("> " + row))
		} else {
			content.WriteString("  " + row)
		}
		content.WriteString("\n")
	}

	return theme.GetStyle("worktree-list").Width(72).Render(strings.TrimRight(content.String(), "\n"))
}

func renderContextPanel(worktrees []domain.Worktree, selectedIdx int, theme styles.Theme) string {
	var content string
	if len(worktrees) == 0 || selectedIdx < 0 || selectedIdx >= len(worktrees) {
		content = "No worktree selected.\nSelect a worktree to\nview context."
	} else {
		wt := worktrees[selectedIdx]
		content = fmt.Sprintf(
			"Context: %s\nBranch: %s\nStatus: %s\n\nAGENT COMMANDS:\n[a] Spawn Claude\n[c] Spawn Copilot",
			filepath.Base(wt.Path), wt.Branch, worktreeStatus(wt),
		)
	}
	return theme.GetStyle("context-panel").Width(34).Render(content)
}

func renderFooterBar(theme styles.Theme, date string) string {
	return theme.GetStyle("status-bar").Width(120).Render(
		fmt.Sprintf("%s  [%s]", footerHints, date),
	)
}

func renderActionBar(theme styles.Theme) string {
	return theme.GetStyle("status-bar").Width(120).Render(actionBarHints)
}

// worktreeStatus maps domain fields to a display status string.
func worktreeStatus(wt domain.Worktree) string {
	if wt.IsLocked {
		return "Locked"
	}
	if wt.IsClean {
		return "Idle"
	}
	return "Dirty"
}

// renderWorktreeList is kept as a thin wrapper for callers that don't need the
// full 3-pane layout (e.g. legacy paths). It uses the default theme.
func renderWorktreeList(worktrees []domain.Worktree) string {
	return renderFull(worktrees, -1, "", 0)
}
