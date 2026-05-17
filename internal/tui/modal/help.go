package modal

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type helpTab int

const (
	tabKeybindings    helpTab = iota
	tabTips                   // Quick Tips
	tabTroubleshooting        // Troubleshooting
	tabAbout                  // About
	helpTabCount              // sentinel
)

var helpTabLabels = [helpTabCount]string{
	"Keybindings",
	"Quick Tips",
	"Troubleshooting",
	"About",
}

// HelpModal is a Bubbletea model for the in-app help overlay (f1 / ?).
// It renders four tabs: Keybindings, Quick Tips, Troubleshooting, and About.
type HelpModal struct {
	activeTab    helpTab
	scrollOffset int
}

// NewHelpModal creates a new HelpModal starting on the Keybindings tab.
func NewHelpModal() *HelpModal {
	return &HelpModal{}
}

// Init satisfies tea.Model.
func (m *HelpModal) Init() tea.Cmd { return nil }

// Title returns the modal title for themed overlay rendering.
func (m *HelpModal) Title() string { return "NEXUS HELP" }

// Update handles Bubbletea messages.
func (m *HelpModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.Type {
	case tea.KeyEsc:
		return m, func() tea.Msg { return ModalCancelledMsg{} }

	case tea.KeyTab:
		m.activeTab = (m.activeTab + 1) % helpTabCount
		m.scrollOffset = 0
		return m, nil

	case tea.KeyUp:
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}
		return m, nil

	case tea.KeyDown:
		m.scrollOffset++
		return m, nil

	case tea.KeyRunes:
		switch keyMsg.String() {
		case "q":
			return m, func() tea.Msg { return ModalCancelledMsg{} }
		case "l":
			m.activeTab = (m.activeTab + 1) % helpTabCount
			m.scrollOffset = 0
		case "h":
			m.activeTab = (m.activeTab + helpTabCount - 1) % helpTabCount
			m.scrollOffset = 0
		case "j":
			m.scrollOffset++
		case "k":
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
		case "1":
			m.activeTab = tabKeybindings
			m.scrollOffset = 0
		case "2":
			m.activeTab = tabTips
			m.scrollOffset = 0
		case "3":
			m.activeTab = tabTroubleshooting
			m.scrollOffset = 0
		case "4":
			m.activeTab = tabAbout
			m.scrollOffset = 0
		}
	}

	return m, nil
}

// View renders the help modal content.
func (m *HelpModal) View() string {
	var b strings.Builder

	// Tab bar
	for i := helpTab(0); i < helpTabCount; i++ {
		if i == m.activeTab {
			b.WriteString(fmt.Sprintf("[%s]", helpTabLabels[i]))
		} else {
			b.WriteString(fmt.Sprintf("  %s  ", helpTabLabels[i]))
		}
		if i < helpTabCount-1 {
			b.WriteString("  ")
		}
	}
	b.WriteString("\n\n")

	// Tab content (with scrolling applied)
	lines := strings.Split(m.tabContent(), "\n")
	start := m.scrollOffset
	if start > len(lines) {
		start = len(lines)
	}
	b.WriteString(strings.Join(lines[start:], "\n"))

	b.WriteString("\n\nTab/h/l: switch tabs   j/k: scroll   Esc/q: close")
	return b.String()
}

// tabContent returns the full unscrolled body for the active tab.
func (m *HelpModal) tabContent() string {
	switch m.activeTab {
	case tabKeybindings:
		return keybindingsContent
	case tabTips:
		return tipsContent
	case tabTroubleshooting:
		return troubleshootingContent
	case tabAbout:
		return aboutContent
	}
	return ""
}

const keybindingsContent = `NAVIGATION
  ↑/↓  j/k      Navigate within panel
  ←/→  h/l      Switch between panels / tabs
  Tab             Cycle panel focus / tab
  Enter           Open shell in worktree / Select

WORKTREE OPERATIONS
  Ctrl+N          Create new worktree
  Ctrl+D          Delete selected worktree
  s               Open shell in worktree

AI AGENTS (from Context Panel)
  a               Spawn Claude Code
  c               Spawn GitHub Copilot
  Space           Unified agent launcher

VIEWS
  w / W           Worktrees view
  i / I           Issues view
  p / P           PRs view
  t               Cycle themes

GLOBAL
  f1 / ?          Open this help modal
  g               Open selected item in GitHub
  Esc / Ctrl+C    Quit`

const tipsContent = `COMMON WORKFLOWS

1. Create a worktree from a GitHub issue
   Press Ctrl+N → pick issue → choose branch type → confirm.
   Nexus creates the branch and worktree automatically.

2. Switch to a worktree and start coding
   Navigate to the worktree with ↑/↓, then press Enter or s.
   Your shell opens directly in the worktree directory.

3. Spawn an AI agent in the right context
   Select a worktree, then press a (Claude), c (Copilot),
   or Space to open the unified launcher. Nexus suspends
   itself and resumes when the agent exits.

4. Keep GitHub data fresh
   Press r to trigger a manual sync. Nexus auto-syncs on
   startup and at the configured interval (~5 min default).

5. Open an issue or PR in the browser
   While in the Issues or PRs view, press g to open the
   selected item in your default browser via gh CLI.

6. Switch themes on the fly
   Press t to cycle through Digital Noir, Matrix, and Light.`

const troubleshootingContent = `COMMON ISSUES

GitHub authentication failure
  Symptom: "gh: not logged in" or empty issues/PRs list.
  Fix: Run  gh auth login  in your terminal and follow the
  prompts to authenticate with your GitHub account.

Agent not found (Claude / Copilot / Aider)
  Symptom: "binary not found" error when spawning an agent.
  Fix: Install the missing tool, or disable it in config:
    ~/.nexus/config.toml  →  ai_agents.claude_enabled = false
  For Claude: https://docs.anthropic.com/en/docs/claude-code
  For Copilot: gh extension install github/gh-copilot

git not in PATH
  Symptom: Nexus starts but shows no worktrees.
  Fix: Ensure git is installed and accessible in your PATH.
  Run  git --version  to verify.

Worktree list is stale
  Symptom: Deleted worktrees still appear.
  Fix: Run  git worktree prune  in your repo, then press r
  in Nexus to refresh.

Config not loading
  Symptom: Warning banner on startup.
  Fix: Check  ~/.nexus/config.toml  for TOML syntax errors.
  Delete or reset the file to restore defaults.`

const aboutContent = `NEXUS — Git Worktree Orchestrator & AI Agent Hub

Version:   v1.0.0
Repo:      github.com/m00nk0d3/nexus
License:   MIT

Built with the Charm.sh ecosystem:
  • Bubbletea  — TUI framework
  • Lipgloss   — terminal styling
  • Bubbles    — UI components

Nexus helps you manage multiple git worktrees and
launch AI coding agents with the correct filesystem
context — all from a single terminal interface.`
