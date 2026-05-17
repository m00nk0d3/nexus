package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/m00nk0d3/nexus/internal/domain"
	"github.com/m00nk0d3/nexus/internal/tui/styles"
)

const (
	appVersion             = "1.0"
	footerHintsWorktrees   = "[Tab] Panel | [j/k] Navigate | [Enter] Select | [Space] Agents | [t] Theme | [g] GH | [esc] Quit"
	footerHintsPRs         = "[Tab] Panel | [j/k] Navigate | [Enter] Checkout | [t] Theme | [g] GH | [esc] Quit"
	footerHintsDefault     = footerHintsWorktrees
	actionBarHints         = "[c-n] New  [c-d] Delete  [c-l] Lock | [f1] Help"
	defaultTermWidth = 120
	navPanelInner    = 18
	// ctxPanelInner is no longer a constant — use computeCtxInner(termWidth) instead.
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

	// ctxMinInner / ctxMaxInner bound the dynamic context-panel content width.
	ctxMinInner = 25
	ctxMaxInner = 60
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
func renderFull(worktrees []domain.Worktree, selectedIdx int, repoPath string, themeIdx int, view activeView, termWidth, termHeight int, syncing bool, lastSynced time.Time, syncErr error, issues []domain.Issue, selectedIssueIdx int, prs []domain.PullRequest, selectedPRIdx int, focused focusedPanel, ctxScroll int) string {
	if termWidth <= 0 {
		termWidth = defaultTermWidth
	}
	theme := styles.NewTheme(styles.Themes[themeIdx])

	navOuter := navPanelInner + panelOverhead
	ctxInner := computeCtxInner(termWidth)
	ctxOuter := ctxInner + panelOverhead
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
	nav := renderNavRail(theme, panelHeight, view, focused == panelNav)

	var list string
	switch view {
	case viewIssues:
		list = renderIssueList(issues, selectedIssueIdx, worktrees, theme, listInner, panelHeight, focused == panelList)
	case viewPRs:
		list = renderPRList(prs, selectedPRIdx, theme, listInner, panelHeight, focused == panelList)
	default:
		list = renderWorktreePanel(worktrees, selectedIdx, theme, listInner, panelHeight, focused == panelList)
	}

	ctx := renderContextPanel(view, worktrees, selectedIdx, issues, selectedIssueIdx, prs, selectedPRIdx, theme, panelHeight, ctxScroll, focused == panelCtx, ctxInner)
	mainRow := lipgloss.JoinHorizontal(lipgloss.Top, nav, list, ctx)
	footer := renderFooterBar(theme, time.Now().UTC().Format("2006-01-02"), termWidth, syncing, lastSynced, syncErr, view)
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

func renderNavRail(theme styles.Theme, panelHeight int, view activeView, focused bool) string {
	var b strings.Builder
	for i, item := range navItems {
		cursor := "  "
		if activeView(i) == view {
			cursor = "> "
		}
		b.WriteString(fmt.Sprintf("%s%s: %s\n", cursor, item.key, item.label))
	}
	st := theme.GetStyle("nav-rail").Width(navPanelInner + panelPaddingOverhead)
	if !focused {
		st = theme.MutedBorder(st)
	}
	if panelHeight > 0 {
		st = st.Height(panelHeight)
	}
	return st.Render(strings.TrimRight(b.String(), "\n"))
}

func renderWorktreePanel(worktrees []domain.Worktree, selectedIdx int, theme styles.Theme, listInner, panelHeight int, focused bool) string {
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

	// Cap rendered rows so panel content never exceeds panelHeight.
	// The header row occupies 1 line, so at most panelHeight-1 data rows fit.
	// Virtual scrolling: slide the visible window so selectedIdx is always shown.
	startIdx := 0
	visible := worktrees
	if panelHeight > 0 {
		maxItems := panelHeight - 1
		if maxItems < 0 {
			maxItems = 0
		}
		if maxItems > 0 && selectedIdx >= maxItems {
			startIdx = selectedIdx - maxItems + 1
		}
		end := startIdx + maxItems
		if end > len(worktrees) {
			end = len(worktrees)
		}
		visible = worktrees[startIdx:end]
	}

	for i, wt := range visible {
		name := truncateStr(filepath.Base(wt.Path), 18)
		path := truncateStr(wt.Path, pathWidth)
		status := worktreeStatus(wt)
		ghID := "-"
		var prState string
		if wt.LinkedPR != nil {
			ghID = fmt.Sprintf("%d", wt.LinkedPR.Number)
			prState = wt.LinkedPR.State
		}
		if i+startIdx == selectedIdx {
			ghIDFormatted := fmt.Sprintf("%-6s", ghID)
			if prState != "" {
				ghIDFormatted = lipgloss.NewStyle().Foreground(prStateColor(prState)).Render(ghIDFormatted)
			}
			row := fmt.Sprintf("%-18s %-*s %-8s %-10s ", name, pathWidth, path, status, "—") + ghIDFormatted
			content.WriteString(theme.GetStyle("selected-row").Width(listInner).Render("> " + row))
		} else {
			nameCol := fmt.Sprintf("%-18s", name)
			pathCol := fmt.Sprintf("%-*s", pathWidth, path)
			statusCol := theme.StatusStyle(status).Width(8).Render(status)
			updatedCol := fmt.Sprintf("%-10s", "—") // TODO: populate from git log --format=%ai
			ghIDRaw := fmt.Sprintf("%-6s", ghID)
			var ghIDCol string
			if prState != "" {
				ghIDCol = lipgloss.NewStyle().Foreground(prStateColor(prState)).Render(ghIDRaw)
			} else {
				ghIDCol = ghIDRaw
			}
			content.WriteString("  " + nameCol + " " + pathCol + " " + statusCol + " " + updatedCol + " " + ghIDCol)
		}
		content.WriteString("\n")
	}

	st := theme.GetStyle("worktree-list").Width(listInner + panelPaddingOverhead)
	if !focused {
		st = theme.MutedBorder(st)
	}
	if panelHeight > 0 {
		st = st.Height(panelHeight).MaxHeight(panelHeight + 2)
	}
	return st.Render(strings.TrimRight(content.String(), "\n"))
}

func renderContextPanel(view activeView, worktrees []domain.Worktree, worktreeIdx int, issues []domain.Issue, issueIdx int, prs []domain.PullRequest, prIdx int, theme styles.Theme, panelHeight int, ctxScroll int, focused bool, ctxInner int) string {
	var content string
	switch view {
	case viewIssues:
		if len(issues) == 0 || issueIdx < 0 || issueIdx >= len(issues) {
			content = "No issue selected.\nPress I to view issues."
		} else {
			iss := issues[issueIdx]
			labelsStr := formatLabels(iss.Labels)
			title := wrapText(iss.Title, ctxInner)
			// "Labels: " prefix = 8 chars; wrap to remaining width to avoid re-wrap.
			labels := wrapText(labelsStr, ctxInner-8)
			body := wrapText(sanitizeBody(strings.ReplaceAll(iss.Body, "\r", "")), ctxInner)
			if body == "" {
				body = "(no description)"
			}
			statusText := "Open"
			statusDot := "●"
			if issueHasWorktree(iss.Number, worktrees) {
				statusText = "In Progress"
				statusDot = "◉"
			}
			assigneesStr := formatAssignees(iss.Assignees)
			content = fmt.Sprintf("Context: Issue #%d\n%s\n\nStatus: %s %s\nAssigned: %s\nLabels: %s\n\n%s\n\n[g] Open in GitHub", iss.Number, title, statusDot, statusText, assigneesStr, labels, body)
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
			labelsStr := formatLabels(pr.Labels)
			body := wrapText(sanitizeBody(strings.ReplaceAll(pr.Body, "\r", "")), ctxInner)
			if body == "" {
				body = "(no description)"
			}
			title := wrapText(pr.Title, ctxInner)
			branch := truncateStr(pr.Branch, ctxInner-8)  // "Branch: " prefix = 8 chars
			author := truncateStr(pr.Author, ctxInner-9)  // "Author: @" prefix = 9 chars
			// "Labels: " prefix = 8 chars; wrap to remaining width to avoid re-wrap.
			labels := wrapText(labelsStr, ctxInner-8)
			content = fmt.Sprintf("Context: PR #%d\n%s\n\nBranch: %s\nAuthor: @%s\nStatus: %s\nLabels: %s\n\n%s\n\n[g] Open in GitHub", pr.Number, title, branch, author, state, labels, body)
		}
	default: // viewWorktrees
		if len(worktrees) == 0 || worktreeIdx < 0 || worktreeIdx >= len(worktrees) {
			content = "No worktree selected.\nSelect a worktree to\nview context."
		} else {
			wt := worktrees[worktreeIdx]
			if wt.LinkedPR != nil {
				pr := wt.LinkedPR
				// "Labels: " = 8 chars; "Author: @" = 9 chars; "GH Title: " = 10 chars
				labelsStr := formatLabels(pr.Labels)
				labels := wrapText(labelsStr, ctxInner-8)
				titleTrunc := truncateStr(pr.Title, ctxInner-10) // "GH Title: " prefix = 10 chars
				statusDot := lipgloss.NewStyle().Foreground(prStateColor(pr.State)).Render("●")
				body := wrapText(sanitizeBody(strings.ReplaceAll(pr.Body, "\r", "")), ctxInner)
				if body == "" {
					body = "(no description)"
				}
				content = fmt.Sprintf(
					"Context: PR #%d\n%s\n\nGH Title: %s\nAuthor: @%s\nStatus: %s %s\nLabels: %s\n\n%s\n\nAGENT COMMANDS:\n[a] Spawn Claude Code\n[c] Spawn Copilot\n[f] Spawn Aider\n[s] Open Shell in WT",
					pr.Number, titleTrunc, pr.Title, pr.Author, statusDot, pr.State, labels, body,
				)
			} else {
				const pathLabel = "Path: "
				pathTrunc := truncateStr(wt.Path, ctxInner-len(pathLabel))
				content = fmt.Sprintf(
					"Context: %s\nBranch: %s\nPath: %s\n\nAGENT COMMANDS:\n[a] Spawn Claude Code\n[c] Spawn Copilot\n[f] Spawn Aider\n[s] Open Shell in WT",
					filepath.Base(wt.Path), wt.Branch, pathTrunc,
				)
			}
		}
	}
	st := theme.GetStyle("context-panel").Width(ctxInner + panelPaddingOverhead)
	if !focused {
		st = theme.MutedBorder(st)
	}
	if panelHeight > 0 {
		content = clipContent(content, ctxScroll, panelHeight)
		// MaxHeight(panelHeight+2): hard-cap the rendered output at panelHeight inner
		// rows + 2 border rows. MaxHeight applies AFTER borders, so this prevents
		// any lipgloss re-wrap from making the panel taller than the terminal allows.
		st = st.Height(panelHeight).MaxHeight(panelHeight + 2)
	}
	return st.Render(content)
}

func renderIssueList(issues []domain.Issue, selectedIdx int, worktrees []domain.Worktree, theme styles.Theme, listInner, panelHeight int, focused bool) string {
	var content strings.Builder
	headerStyle := theme.GetStyle("table-header")
	titleWidth := listInner - 46 // fixed: 2(cursor/sp) + 6(#) + 1 + 11(status) + 1 + 12(assigned) + 1 + 8(labels min) + 4 = 46
	if titleWidth < 10 {
		titleWidth = 10
	}
	headerRow := fmt.Sprintf("  %-6s %-*s %-11s %-12s %s", "#", titleWidth, "TITLE", "STATUS", "ASSIGNED", "LABELS")
	content.WriteString(headerStyle.Render(headerRow))
	content.WriteString("\n")
	labelsWidth := listInner - titleWidth - 35 // remaining after fixed overhead
	if labelsWidth < 8 {
		labelsWidth = 8
	}

	// Cap rendered rows and virtual-scroll so selectedIdx is always in the window.
	// The header row occupies 1 line, so at most panelHeight-1 data rows fit.
	issueStartIdx := 0
	visible := issues
	if panelHeight > 0 {
		maxItems := panelHeight - 1
		if maxItems < 0 {
			maxItems = 0
		}
		if maxItems > 0 && selectedIdx >= maxItems {
			issueStartIdx = selectedIdx - maxItems + 1
		}
		end := issueStartIdx + maxItems
		if end > len(issues) {
			end = len(issues)
		}
		visible = issues[issueStartIdx:end]
	}

	for i, issue := range visible {
		labels := truncateStr(strings.Join(issue.Labels, " "), labelsWidth)
		title := truncateStr(issue.Title, titleWidth)
		status := "Open"
		if issueHasWorktree(issue.Number, worktrees) {
			status = "In Progress"
		}
		assigned := truncateStr(formatAssignees(issue.Assignees), 12)
		if i+issueStartIdx == selectedIdx {
			row := fmt.Sprintf("%-6d %-*s %-11s %-12s %s", issue.Number, titleWidth, title, status, assigned, labels)
			content.WriteString(theme.GetStyle("selected-row").Width(listInner).Render("> " + row))
		} else {
			statusCol := theme.StatusStyle(strings.ToLower(status)).Width(11).Render(status)
			assignedCol := fmt.Sprintf("%-12s", assigned)
			prefix := fmt.Sprintf("  %-6d %-*s ", issue.Number, titleWidth, title)
			content.WriteString(prefix + statusCol + " " + assignedCol + " " + labels)
		}
		content.WriteString("\n")
	}
	if len(issues) == 0 {
		content.WriteString("  No issues found.\n")
	}
	st := theme.GetStyle("worktree-list").Width(listInner + panelPaddingOverhead)
	if !focused {
		st = theme.MutedBorder(st)
	}
	if panelHeight > 0 {
		st = st.Height(panelHeight).MaxHeight(panelHeight + 2)
	}
	return st.Render(strings.TrimRight(content.String(), "\n"))
}

func renderPRList(prs []domain.PullRequest, selectedIdx int, theme styles.Theme, listInner, panelHeight int, focused bool) string {
	var content strings.Builder
	headerStyle := theme.GetStyle("table-header")
	// Bug 2: drop AUTHOR column (visible in context panel) and reduce BRANCH to 14.
	// Fixed overhead: 2(cursor) + 6(#) + 1(sp) + 1(sp) + 14(branch) + 1(sp) = 25,
	// Dynamic branch width: ~15% of available space, clamped to [8, 18].
	branchWidth := listInner * 15 / 100
	if branchWidth < 8 {
		branchWidth = 8
	}
	if branchWidth > 18 {
		branchWidth = 18
	}
	// Fixed overhead: 2(cursor/spaces) + 6(#) + 1(sp) + 1(sp) + 1(sp) + 6(STATUS) = 17;
	// plus branchWidth + 2 spaces around it = branchWidth + 2 → total fixed = 19 + branchWidth.
	// Plus ASSIGNED column: 12 chars + 1 space = 13 → total fixed = 32 + branchWidth.
	// titleWidth fills remaining so total row = listInner.
	const prStatusMaxLen = 6
	const prAssigneeWidth = 12
	titleWidth := listInner - (11 + branchWidth + prStatusMaxLen + prAssigneeWidth + 2)
	if titleWidth < 10 {
		titleWidth = 10
	}
	headerRow := fmt.Sprintf("  %-6s %-*s %-*s %-12s %s", "#", titleWidth, "TITLE", branchWidth, "BRANCH", "ASSIGNED", "STATUS")
	content.WriteString(headerStyle.Render(headerRow))
	content.WriteString("\n")

	// Cap rendered rows and virtual-scroll so selectedIdx is always in the window.
	// The header row occupies 1 line, so at most panelHeight-1 data rows fit.
	prStartIdx := 0
	visible := prs
	if panelHeight > 0 {
		maxItems := panelHeight - 1
		if maxItems < 0 {
			maxItems = 0
		}
		if maxItems > 0 && selectedIdx >= maxItems {
			prStartIdx = selectedIdx - maxItems + 1
		}
		end := prStartIdx + maxItems
		if end > len(prs) {
			end = len(prs)
		}
		visible = prs[prStartIdx:end]
	}

	for i, pr := range visible {
		title := truncateStr(pr.Title, titleWidth)
		branch := truncateStr(pr.Branch, branchWidth)
		state := pr.State
		if pr.IsDraft {
			state = "DRAFT"
		}
		assigned := truncateStr(strings.Join(pr.Assignees, ","), prAssigneeWidth)
		if i+prStartIdx == selectedIdx {
			row := fmt.Sprintf("%-6d %-*s %-*s %-12s %s", pr.Number, titleWidth, title, branchWidth, branch, assigned, state)
			content.WriteString(theme.GetStyle("selected-row").Width(listInner).Render("> " + row))
		} else {
			row := fmt.Sprintf("  %-6d %-*s %-*s %-12s %s", pr.Number, titleWidth, title, branchWidth, branch, assigned, state)
			content.WriteString(row)
		}
		content.WriteString("\n")
	}
	if len(prs) == 0 {
		content.WriteString("  No open PRs.\n")
	}
	st := theme.GetStyle("worktree-list").Width(listInner + panelPaddingOverhead)
	if !focused {
		st = theme.MutedBorder(st)
	}
	if panelHeight > 0 {
		st = st.Height(panelHeight).MaxHeight(panelHeight + 2)
	}
	return st.Render(strings.TrimRight(content.String(), "\n"))
}

// clipContent slices content lines for bounded panel rendering.
// offset skips the first N lines; maxLines caps the visible output.
// If maxLines is 0 the content is returned unchanged.
func clipContent(content string, offset, maxLines int) string {
	if maxLines <= 0 {
		return content
	}
	lines := strings.Split(content, "\n")
	if offset > 0 {
		if offset >= len(lines) {
			offset = len(lines) - 1
		}
		lines = lines[offset:]
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	return strings.Join(lines, "\n")
}

func renderFooterBar(theme styles.Theme, date string, termWidth int, syncing bool, lastSynced time.Time, syncErr error, view activeView) string {
	hints := footerHintsDefault
	if view == viewPRs {
		hints = footerHintsPRs
	}

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

	// Build the right side (date + optional sync status) first, then truncate
	// only the hints so the sync status is never clipped on narrow terminals.
	right := fmt.Sprintf("  [%s]", date)
	if syncStatus != "" {
		right += "  " + syncStatus
	}
	maxHints := termWidth - len([]rune(right))
	if maxHints < 0 {
		maxHints = 0
	}
	content := truncateStr(hints, maxHints) + right

	return theme.GetStyle("status-bar").Width(termWidth).Render(content)
}

func renderActionBar(theme styles.Theme, termWidth int) string {
	hints := truncateStr(actionBarHints, termWidth)
	return theme.GetStyle("status-bar").Width(termWidth).Render(hints)
}

// computeCtxInner returns the inner content width for the context panel.
// It scales to ~30 % of the terminal width and is clamped to [ctxMinInner, ctxMaxInner].
func computeCtxInner(termWidth int) int {
	inner := termWidth * 30 / 100
	if inner < ctxMinInner {
		return ctxMinInner
	}
	if inner > ctxMaxInner {
		return ctxMaxInner
	}
	return inner
}

// sanitizeBody strips control characters from a PR/issue body that would
// corrupt terminal rendering (e.g. backspace 0x08, form feed 0x0C produced
// by PowerShell backtick escapes when the PR body is created via `gh pr create`
// with double-quoted strings containing markdown code spans).
// Line feeds (0x0A) are preserved; carriage returns are handled separately.
func sanitizeBody(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == '\n' || r == '\t' || r >= 0x20 {
			b.WriteRune(r)
		}
		// Drop all other control chars (0x00-0x1F except \n and \t),
		// including \b (0x08, backspace) and \f (0x0C, form feed).
	}
	return b.String()
}

// wrapText word-wraps s to at most width runes per line.
// Existing newlines are preserved; each segment is wrapped independently.
// If width <= 0 the string is returned unchanged.
func wrapText(s string, width int) string {
	if width <= 0 {
		return s
	}
	var out strings.Builder
	for i, seg := range strings.Split(s, "\n") {
		if i > 0 {
			out.WriteByte('\n')
		}
		out.WriteString(wrapLine(seg, width))
	}
	return out.String()
}

// wrapLine wraps a single newline-free string at word boundaries using display
// cell width (so multi-cell characters like emoji and CJK are measured correctly).
// Falls back to a hard break when a word exceeds width.
func wrapLine(s string, width int) string {
	if runewidth.StringWidth(s) <= width {
		return s
	}
	var out strings.Builder
	runes := []rune(s)
	for {
		// Find the rune index where display cells would exceed width.
		cells, cut := 0, len(runes)
		for i, r := range runes {
			rw := runewidth.RuneWidth(r)
			if cells+rw > width {
				cut = i
				break
			}
			cells += rw
		}
		if cut == len(runes) {
			// All remaining runes fit.
			out.WriteString(string(runes))
			break
		}
		// Prefer a word-boundary break.
		if runes[cut] == ' ' {
			out.WriteString(string(runes[:cut]))
			out.WriteByte('\n')
			runes = runes[cut+1:]
		} else {
			breakAt := -1
			for i := cut - 1; i >= 0; i-- {
				if runes[i] == ' ' {
					breakAt = i
					break
				}
			}
			if breakAt < 0 {
				// No space found — hard break at the cut point.
				out.WriteString(string(runes[:cut]))
				out.WriteByte('\n')
				runes = runes[cut:]
			} else if breakAt == 0 {
				// Segment starts with a leading space (e.g. after a previous break).
				// Skip it silently so we don't emit a spurious blank line.
				runes = runes[1:]
			} else {
				out.WriteString(string(runes[:breakAt]))
				out.WriteByte('\n')
				runes = runes[breakAt+1:]
			}
		}
		if runewidth.StringWidth(string(runes)) <= width {
			out.WriteString(string(runes))
			break
		}
	}
	return out.String()
}


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

// prStateColor returns the lipgloss color for a given PR state string.
func prStateColor(state string) lipgloss.Color {
	switch state {
	case "OPEN":
		return lipgloss.Color("#00D9FF")
	case "MERGED":
		return lipgloss.Color("#9B59B6")
	case "CLOSED":
		return lipgloss.Color("#E74C3C")
	default:
		return lipgloss.Color("#4A5568")
	}
}

// formatAssignees formats a slice of assignee logins into "@user1,@user2" format.
// Returns "-" when there are no assignees.
func formatAssignees(assignees []string) string {
	if len(assignees) == 0 {
		return "-"
	}
	parts := make([]string, len(assignees))
	for i, a := range assignees {
		parts[i] = "@" + a
	}
	return strings.Join(parts, ",")
}

// issueHasWorktree returns true if any worktree's branch contains "issue-<number>-"
// or ends with "issue-<number>", indicating a worktree was created for this issue.
func issueHasWorktree(issueNumber int, worktrees []domain.Worktree) bool {
	withDash := fmt.Sprintf("issue-%d-", issueNumber)
	atEnd := fmt.Sprintf("issue-%d", issueNumber)
	for _, wt := range worktrees {
		if strings.Contains(wt.Branch, withDash) || strings.HasSuffix(wt.Branch, atEnd) {
			return true
		}
	}
	return false
}

// formatLabels formats a slice of label strings into "[label1][label2]" format.
func formatLabels(labels []string) string {
	strs := make([]string, len(labels))
	for i, l := range labels {
		strs[i] = "[" + l + "]"
	}
	return strings.Join(strs, "")
}

