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
	appVersion       = "1.0"
	footerHints      = "[↑/↓] Navigate | [Enter] Select | [t] Theme | [g] Open in GH | [esc] Quit"
	actionBarHints   = "[c-n] New  [c-d] Delete  [c-l] Lock | [f1] Help"
	defaultTermWidth = 120
	navPanelInner    = 18
	ctxPanelInner    = 50
	// panelOverhead: 1 border-left + 1 pad-left + 1 pad-right + 1 border-right
	panelOverhead = 4
	// headerOverhead: 1 pad-left + 1 pad-right (no border on header/status-bar)
	headerOverhead = 2
	minPathWidth   = 5
	// fixedChromeRows: 1 header + 1 footer + 1 action bar + 2 panel borders (top+bottom)
	fixedChromeRows = 5
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

// renderFull builds the complete 3-pane TUI layout.
// termWidth is the terminal column count; 0 falls back to defaultTermWidth.
// termHeight is the terminal row count; 0 disables explicit panel height.
func renderFull(worktrees []domain.Worktree, selectedIdx int, repoPath string, themeIdx, termWidth, termHeight int) string {
	if termWidth <= 0 {
		termWidth = defaultTermWidth
	}
	theme := styles.NewTheme(styles.Themes[themeIdx])

	navOuter := navPanelInner + panelOverhead
	ctxOuter := ctxPanelInner + panelOverhead
	listOuter := termWidth - navOuter - ctxOuter
	if listOuter < minPathWidth+panelOverhead {
		listOuter = minPathWidth + panelOverhead
	}
	listInner := listOuter - panelOverhead
	headerInner := termWidth - headerOverhead

	// panelHeight is the inner content height for all three side panels.
	// 0 means let lipgloss size naturally (used in tests / zero-height terminals).
	panelHeight := 0
	if termHeight > fixedChromeRows {
		panelHeight = termHeight - fixedChromeRows
	}

	header := renderHeader(repoPath, theme, headerInner)
	nav := renderNavRail(theme, panelHeight)
	list := renderWorktreePanel(worktrees, selectedIdx, theme, listInner, panelHeight)
	ctx := renderContextPanel(worktrees, selectedIdx, theme, panelHeight)
	mainRow := lipgloss.JoinHorizontal(lipgloss.Top, nav, list, ctx)
	footer := renderFooterBar(theme, time.Now().UTC().Format("2006-01-02"), termWidth)
	actionBar := renderActionBar(theme, termWidth)

	return lipgloss.JoinVertical(lipgloss.Left, header, mainRow, footer, actionBar)
}

func renderHeader(repoPath string, theme styles.Theme, innerWidth int) string {
	if repoPath == "" {
		repoPath = "./"
	}
	text := fmt.Sprintf(
		"NEXUS v%s: GIT WORKTREE ORCHESTRATOR | Repo: %s | Local Path: %s",
		appVersion, filepath.Base(repoPath), repoPath,
	)
	return theme.GetStyle("header").Width(innerWidth).Render(text)
}

func renderNavRail(theme styles.Theme, panelHeight int) string {
	var b strings.Builder
	for i, item := range navItems {
		cursor := "  "
		if i == 0 {
			cursor = "> "
		}
		b.WriteString(fmt.Sprintf("%s%s: %s\n", cursor, item.key, item.label))
	}
	st := theme.GetStyle("nav-rail").Width(navPanelInner)
	if panelHeight > 0 {
		st = st.Height(panelHeight)
	}
	return st.Render(strings.TrimRight(b.String(), "\n"))
}

func renderWorktreePanel(worktrees []domain.Worktree, selectedIdx int, theme styles.Theme, listInner, panelHeight int) string {
	headers := []string{"NAME", "PATH", "STATUS", "UPDATED", "GH:ID"}
	var content strings.Builder

	// fixed columns: name(18) + status(8) + updated(10) + ghid(6) + 4 separators = 46
	// cursor prefix is 2 chars ("> " or "  ")
	const fixedRowWidth = 18 + 1 + 8 + 1 + 10 + 1 + 6 + 4 // =49 (incl separators, excl path+sep)
	pathWidth := listInner - 2 - fixedRowWidth
	if pathWidth < minPathWidth {
		pathWidth = minPathWidth
	}

	headerStyle := theme.GetStyle("table-header")
	content.WriteString(headerStyle.Render(strings.Join(headers, "   ")))
	content.WriteString("\n")

	for i, wt := range worktrees {
		name := filepath.Base(wt.Path)
		status := worktreeStatus(wt)
		ghID := ""
		if wt.LinkedPR != nil {
			ghID = fmt.Sprintf("%d", *wt.LinkedPR)
		}
		row := fmt.Sprintf("%-18s %-*s %-8s %-10s %-6s", name, pathWidth, wt.Path, status, "—", ghID)
		if i == selectedIdx {
			content.WriteString(theme.GetStyle("selected-row").Render("> " + row))
		} else {
			content.WriteString("  " + row)
		}
		content.WriteString("\n")
	}

	st := theme.GetStyle("worktree-list").Width(listInner)
	if panelHeight > 0 {
		st = st.Height(panelHeight)
	}
	return st.Render(strings.TrimRight(content.String(), "\n"))
}

func renderContextPanel(worktrees []domain.Worktree, selectedIdx int, theme styles.Theme, panelHeight int) string {
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
	st := theme.GetStyle("context-panel").Width(ctxPanelInner)
	if panelHeight > 0 {
		st = st.Height(panelHeight)
	}
	return st.Render(content)
}

func renderFooterBar(theme styles.Theme, date string, termWidth int) string {
	return theme.GetStyle("status-bar").Width(termWidth).Render(
		fmt.Sprintf("%s  [%s]", footerHints, date),
	)
}

func renderActionBar(theme styles.Theme, termWidth int) string {
	return theme.GetStyle("status-bar").Width(termWidth).Render(actionBarHints)
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

