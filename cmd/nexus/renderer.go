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
	// panelPaddingOverhead: lipgloss Width includes padding, so pass Width(inner + panelPaddingOverhead)
	// to get a content area equal to the *Inner variable (Padding(0,1) = 1+1 = 2).
	panelPaddingOverhead = 2
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
func renderFull(worktrees []domain.Worktree, selectedIdx int, repoPath string, themeIdx int, view activeView, termWidth, termHeight int, syncing bool, lastSynced time.Time, syncErr error, issues []domain.Issue, selectedIssueIdx int, prs []domain.PullRequest, selectedPRIdx int) string {
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
	nav := renderNavRail(theme, panelHeight, view)

	var list string
	switch view {
	case viewIssues:
		list = renderIssueList(issues, selectedIssueIdx, theme, listInner, panelHeight)
	case viewPRs:
		list = renderPRList(prs, selectedPRIdx, theme, listInner, panelHeight)
	default:
		list = renderWorktreePanel(worktrees, selectedIdx, theme, listInner, panelHeight)
	}

	ctx := renderContextPanel(view, worktrees, selectedIdx, issues, selectedIssueIdx, prs, selectedPRIdx, theme, panelHeight)
	mainRow := lipgloss.JoinHorizontal(lipgloss.Top, nav, list, ctx)
	footer := renderFooterBar(theme, time.Now().UTC().Format("2006-01-02"), termWidth, syncing, lastSynced, syncErr)
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

func renderNavRail(theme styles.Theme, panelHeight int, view activeView) string {
	var b strings.Builder
	for i, item := range navItems {
		cursor := "  "
		if activeView(i) == view {
			cursor = "> "
		}
		b.WriteString(fmt.Sprintf("%s%s: %s\n", cursor, item.key, item.label))
	}
	st := theme.GetStyle("nav-rail").Width(navPanelInner + panelPaddingOverhead)
	if panelHeight > 0 {
		st = st.Height(panelHeight)
	}
	return st.Render(strings.TrimRight(b.String(), "\n"))
}

func renderWorktreePanel(worktrees []domain.Worktree, selectedIdx int, theme styles.Theme, listInner, panelHeight int) string {
	var content strings.Builder

	// fixed columns: cursor(2) + name(18) + sep(1) + sep(1) + status(8) + sep(1) + updated(10) + sep(1) + ghid(6) = 48
	const fixedRowWidth = 2 + 18 + 1 + 1 + 8 + 1 + 10 + 1 + 6 // =48 (excl path col)
	pathWidth := listInner - fixedRowWidth
	if pathWidth < minPathWidth {
		pathWidth = minPathWidth
	}

	headerStyle := theme.GetStyle("table-header")
	headerRow := fmt.Sprintf("  %-18s %-*s %-8s %-10s %-6s", "NAME", pathWidth, "PATH", "STATUS", "UPDATED", "GH:ID")
	content.WriteString(headerStyle.Render(headerRow))
	content.WriteString("\n")

	for i, wt := range worktrees {
		name := truncateStr(filepath.Base(wt.Path), 18)
		path := truncateStr(wt.Path, pathWidth)
		status := worktreeStatus(wt)
		ghID := ""
		if wt.LinkedPR != nil {
			ghID = fmt.Sprintf("%d", wt.LinkedPR.Number)
		}
		if i == selectedIdx {
			row := fmt.Sprintf("%-18s %-*s %-8s %-10s %-6s", name, pathWidth, path, status, "—", ghID)
			content.WriteString(theme.GetStyle("selected-row").Width(listInner).Render("> " + row))
		} else {
			nameCol := fmt.Sprintf("%-18s", name)
			pathCol := fmt.Sprintf("%-*s", pathWidth, path)
			statusCol := theme.StatusStyle(status).Width(8).Render(status)
			updatedCol := fmt.Sprintf("%-10s", "—") // TODO: populate from git log --format=%ai
			ghIDCol := fmt.Sprintf("%-6s", ghID)
			content.WriteString("  " + nameCol + " " + pathCol + " " + statusCol + " " + updatedCol + " " + ghIDCol)
		}
		content.WriteString("\n")
	}

	st := theme.GetStyle("worktree-list").Width(listInner + panelPaddingOverhead)
	if panelHeight > 0 {
		st = st.Height(panelHeight)
	}
	return st.Render(strings.TrimRight(content.String(), "\n"))
}

func renderContextPanel(view activeView, worktrees []domain.Worktree, worktreeIdx int, issues []domain.Issue, issueIdx int, prs []domain.PullRequest, prIdx int, theme styles.Theme, panelHeight int) string {
	var content string
	switch view {
	case viewIssues:
		if len(issues) == 0 || issueIdx < 0 || issueIdx >= len(issues) {
			content = "No issue selected.\nPress I to view issues."
		} else {
			iss := issues[issueIdx]
			labelStrs := make([]string, len(iss.Labels))
			for i, l := range iss.Labels {
				labelStrs[i] = "[" + l + "]"
			}
			labelsStr := strings.Join(labelStrs, "")
			content = fmt.Sprintf("Context: Issue #%d\n%s\n\nStatus: ● Open\nLabels: %s\n\n[g] Open in GitHub", iss.Number, iss.Title, labelsStr)
		}
	case viewPRs:
		if len(prs) == 0 || prIdx < 0 || prIdx >= len(prs) {
			content = "No PR selected.\nPress P to view PRs."
		} else {
			pr := prs[prIdx]
			state := pr.State
			if pr.IsDraft {
				state = "DRAFT"
			}
			content = fmt.Sprintf("Context: PR #%d\n%s\n\nBranch: %s\nAuthor: @%s\nStatus: %s\n\n[g] Open in GitHub", pr.Number, pr.Title, pr.Branch, pr.Author, state)
		}
	default: // viewWorktrees
		if len(worktrees) == 0 || worktreeIdx < 0 || worktreeIdx >= len(worktrees) {
			content = "No worktree selected.\nSelect a worktree to\nview context."
		} else {
			wt := worktrees[worktreeIdx]
			content = fmt.Sprintf(
				"Context: %s\nBranch: %s\nStatus: %s\n\nAGENT COMMANDS:\n[a] Spawn Claude\n[c] Spawn Copilot",
				filepath.Base(wt.Path), wt.Branch, worktreeStatus(wt),
			)
		}
	}
	st := theme.GetStyle("context-panel").Width(ctxPanelInner + panelPaddingOverhead)
	if panelHeight > 0 {
		st = st.Height(panelHeight)
	}
	return st.Render(content)
}

func renderIssueList(issues []domain.Issue, selectedIdx int, theme styles.Theme, listInner, panelHeight int) string {
	var content strings.Builder
	headerStyle := theme.GetStyle("table-header")
	titleWidth := listInner - 30
	if titleWidth < 10 {
		titleWidth = 10
	}
	headerRow := fmt.Sprintf("  %-6s %-*s %-8s %s", "#", titleWidth, "TITLE", "STATUS", "LABELS")
	content.WriteString(headerStyle.Render(headerRow))
	content.WriteString("\n")
	for i, issue := range issues {
		labels := strings.Join(issue.Labels, " ")
		title := truncateStr(issue.Title, titleWidth)
		// "Open" is hardcoded because gh issue list only returns open issues by default.
		if i == selectedIdx {
			row := fmt.Sprintf("%-6d %-*s %-8s %s", issue.Number, titleWidth, title, "Open", labels)
			content.WriteString(theme.GetStyle("selected-row").Width(listInner).Render("> " + row))
		} else {
			row := fmt.Sprintf("  %-6d %-*s %-8s %s", issue.Number, titleWidth, title, "Open", labels)
			content.WriteString(row)
		}
		content.WriteString("\n")
	}
	if len(issues) == 0 {
		content.WriteString("  No issues found.\n")
	}
	st := theme.GetStyle("worktree-list").Width(listInner + panelPaddingOverhead)
	if panelHeight > 0 {
		st = st.Height(panelHeight)
	}
	return st.Render(strings.TrimRight(content.String(), "\n"))
}

func renderPRList(prs []domain.PullRequest, selectedIdx int, theme styles.Theme, listInner, panelHeight int) string {
	var content strings.Builder
	headerStyle := theme.GetStyle("table-header")
	titleWidth := listInner - 50
	if titleWidth < 10 {
		titleWidth = 10
	}
	headerRow := fmt.Sprintf("  %-6s %-*s %-16s %-14s %s", "#", titleWidth, "TITLE", "BRANCH", "AUTHOR", "STATUS")
	content.WriteString(headerStyle.Render(headerRow))
	content.WriteString("\n")
	for i, pr := range prs {
		title := truncateStr(pr.Title, titleWidth)
		branch := truncateStr(pr.Branch, 16)
		author := truncateStr(pr.Author, 14)
		state := pr.State
		if pr.IsDraft {
			state = "DRAFT"
		}
		if i == selectedIdx {
			row := fmt.Sprintf("%-6d %-*s %-16s %-14s %s", pr.Number, titleWidth, title, branch, author, state)
			content.WriteString(theme.GetStyle("selected-row").Width(listInner).Render("> " + row))
		} else {
			row := fmt.Sprintf("  %-6d %-*s %-16s %-14s %s", pr.Number, titleWidth, title, branch, author, state)
			content.WriteString(row)
		}
		content.WriteString("\n")
	}
	if len(prs) == 0 {
		content.WriteString("  No open PRs.\n")
	}
	st := theme.GetStyle("worktree-list").Width(listInner + panelPaddingOverhead)
	if panelHeight > 0 {
		st = st.Height(panelHeight)
	}
	return st.Render(strings.TrimRight(content.String(), "\n"))
}

func renderFooterBar(theme styles.Theme, date string, termWidth int, syncing bool, lastSynced time.Time, syncErr error) string {
	left := fmt.Sprintf("%s  [%s]", footerHints, date)

	var syncStatus string
	switch {
	case syncErr != nil:
		syncStatus = "✗ sync err"
	case syncing:
		syncStatus = "⟳ syncing"
	case !lastSynced.IsZero():
		mins := int(time.Since(lastSynced).Minutes())
		if mins < 1 {
			syncStatus = "✓ synced just now"
		} else {
			syncStatus = fmt.Sprintf("✓ synced %dm ago", mins)
		}
	}

	content := left
	if syncStatus != "" {
		content = left + "  " + syncStatus
	}

	return theme.GetStyle("status-bar").Width(termWidth).Render(content)
}

func renderActionBar(theme styles.Theme, termWidth int) string {
	return theme.GetStyle("status-bar").Width(termWidth).Render(actionBarHints)
}

// truncateStr clips s to at most n runes, adding "…" if truncated.
func truncateStr(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 1 {
		return string(runes[:n])
	}
	return string(runes[:n-1]) + "…"
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

