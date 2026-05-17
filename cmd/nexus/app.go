package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/m00nk0d3/nexus/internal/data"
	"github.com/m00nk0d3/nexus/internal/domain"
	internalexec "github.com/m00nk0d3/nexus/internal/exec"
	"github.com/m00nk0d3/nexus/internal/tui/modal"
	"github.com/m00nk0d3/nexus/internal/tui/styles"
)

// issuesFetchedMsg carries the result of a background gh issue list call.
type issuesFetchedMsg struct {
	issues []domain.Issue // List of fetched issues, or nil on error
	err    error          // Error during fetch, if any
}

// worktreeOpDoneMsg carries the result of an add/remove worktree operation.
type worktreeOpDoneMsg struct {
	err error // Error during operation, if any
}

// worktreeSwitchedMsg carries the result of switching to a worktree.
type worktreeSwitchedMsg struct {
	err error // Error during switch, if any
}

// githubSyncedMsg carries the result of a background GitHub PR/issue sync.
type githubSyncedMsg struct {
	prs      []domain.PullRequest
	issues   []domain.Issue
	err      error
	syncedAt time.Time
}

// syncTickMsg triggers the next periodic GitHub sync.
type syncTickMsg struct{}

// browserOpenErrMsg carries an error from opening an issue or PR in the browser.
type browserOpenErrMsg struct{ err error }

// agentDoneMsg is dispatched when an AI agent process exits.
// It carries enough information to log the run and update UI state.
type agentDoneMsg struct {
	agentName string
	prompt    string
	exitCode  int
	startedAt time.Time
}

// aiderFilesFetchedMsg carries the result of listing modified files for the Aider file picker.
type aiderFilesFetchedMsg struct {
	worktreePath string
	files        []string
	err          error
}

// activeView represents the currently active main panel view.
type activeView int

const (
	viewWorktrees activeView = iota // Shows the worktree list (default)
	viewIssues                      // Shows the GitHub issues list
	viewPRs                         // Shows the GitHub pull requests list
)

// focusedPanel identifies which panel currently has keyboard focus.
type focusedPanel int

const (
	panelNav   focusedPanel = iota // Left navigation rail (default focus)
	panelList                      // Main content list
	panelCtx                       // Right context panel
	panelCount                     // Sentinel — used for modular cycling via (p+1)%panelCount
)

// Model represents the root Bubbletea model for the Nexus TUI application.
// It manages the list of git worktrees, user interactions, and active modals.
type Model struct {
	Worktrees        []domain.Worktree    // List of available git worktrees
	RepoPath         string               // Path to the repository root
	Config           *domain.Config       // Loaded application configuration
	selectedIdx      int                  // Currently selected worktree index
	activeModal      modal.Modal          // Currently open modal (if any)
	Error            string               // Error message to display (if any)
	themeIdx         int                  // Index into styles.Themes for the active theme
	view             activeView           // Currently active main panel view
	width            int                  // Terminal width in columns; 0 means use default
	height           int                  // Terminal height in rows; 0 means use default
	prs              []domain.PullRequest // Latest synced pull requests
	issues           []domain.Issue       // Latest synced issues
	lastSynced       time.Time            // When the last successful GitHub sync completed
	syncErr          error                // Error from the most recent GitHub sync attempt
	syncing          bool                 // True while a background GitHub sync is in progress
	selectedIssueIdx int                  // Currently selected issue index
	selectedPRIdx    int                  // Currently selected PR index
	focused          focusedPanel         // Which panel currently has keyboard focus
	ctxScrollOffset  int                  // Scroll position within the context panel

	// DB is optional; when non-nil, agent runs are logged to agent_history.
	db *data.DB

	// Copilot prompt state
	copilotPromptActive bool            // true while the inline Copilot prompt is open
	copilotPromptInput  textinput.Model // text input for entering the Copilot prompt

	// Claude prompt state
	claudePromptActive bool            // true while the inline Claude prompt is open
	claudePromptInput  textinput.Model // text input for entering the Claude prompt
}

// NewModel creates and returns a new Model instance with all required fields initialized.
func NewModel() *Model {
	cfg, err := data.LoadConfig(data.DefaultConfigPath())

	var configErr string
	if err != nil {
		cfg = domain.DefaultConfig()
		configErr = fmt.Sprintf("config load failed: %v", err)
	}

	themeIdx := 0
	for i, name := range styles.Themes {
		if name == cfg.Appearance.Theme {
			themeIdx = i
			break
		}
	}

	return &Model{
		Config:   cfg,
		themeIdx: themeIdx,
		Error:    configErr,
		focused:  panelList,
	}
}

// Init initializes the model and triggers an initial worktree list load and GitHub sync.
func (m *Model) Init() tea.Cmd {
	m.syncing = true
	return tea.Batch(m.refreshWorktreesCmd(), m.syncGitHubCmd())
}

// Update handles incoming messages and returns an updated model and command.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Route all messages to the active modal while one is open.
	if m.activeModal != nil {
		switch msg := msg.(type) {
		case modal.WorktreeCreateConfirmedMsg:
			m.activeModal = nil
			return m, m.addWorktreeCmd(msg.Branch, msg.Path)
		case modal.PRWorktreeCreateConfirmedMsg:
			m.activeModal = nil
			return m, m.checkoutPRWorktreeCmd(msg.Branch, msg.Path)
		case modal.WorktreeDeleteConfirmedMsg:
			m.activeModal = nil
			return m, m.removeWorktreeCmd(msg.Path)
		case modal.AiderLaunchMsg:
			m.activeModal = nil
			if selected, ok := m.selectedWorktree(); ok {
				return m, m.spawnAiderCmd(selected.Path, msg.Files)
			}
			return m, nil
		case modal.SpawnAgentMsg:
			m.activeModal = nil
			switch msg.AgentName {
			case modal.AgentNameCopilot:
				return m, m.spawnCopilotCmd(msg.WorktreePath, msg.Prompt)
			case modal.AgentNameClaude:
				return m, m.spawnClaudeCmd(msg.WorktreePath, msg.Prompt)
		case modal.AgentNameAider:
				return m, m.fetchAiderFilesCmd(msg.WorktreePath)
			}
			return m, nil
		case modal.ModalCancelledMsg:
			m.activeModal = nil
			return m, nil
		default:
			updated, cmd := m.activeModal.Update(msg)
			if next, ok := updated.(modal.Modal); ok {
				m.activeModal = next
			}
			return m, cmd
		}
	}

	// While the Copilot inline prompt is open, route key events to the textinput.
	// Non-key messages (e.g. agentDoneMsg, tea.WindowSizeMsg) fall through to
	// the main switch below so they are still handled correctly.
	if m.copilotPromptActive {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.Type {
			case tea.KeyEnter:
				prompt := strings.TrimSpace(m.copilotPromptInput.Value())
				if prompt == "" {
					return m, nil
				}
				m.copilotPromptActive = false
				if selected, ok := m.selectedWorktree(); ok {
					return m, m.spawnCopilotCmd(selected.Path, prompt)
				}
				m.copilotPromptInput.SetValue("")
				return m, nil
			case tea.KeyEsc:
				m.copilotPromptActive = false
				m.copilotPromptInput.SetValue("")
				return m, nil
			default:
				var cmd tea.Cmd
				m.copilotPromptInput, cmd = m.copilotPromptInput.Update(keyMsg)
				return m, cmd
			}
		}
		// Non-key message: fall through to the main switch to handle it normally.
	}

	// While the Claude inline prompt is open, route key events to the textinput.
	// Non-key messages fall through to the main switch below.
	if m.claudePromptActive {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.Type {
			case tea.KeyEnter:
				prompt := strings.TrimSpace(m.claudePromptInput.Value())
				if prompt == "" {
					return m, nil
				}
				m.claudePromptActive = false
				if selected, ok := m.selectedWorktree(); ok {
					return m, m.spawnClaudeCmd(selected.Path, prompt)
				}
				m.claudePromptInput.SetValue("")
				return m, nil
			case tea.KeyEsc:
				m.claudePromptActive = false
				m.claudePromptInput.SetValue("")
				return m, nil
			default:
				var cmd tea.Cmd
				m.claudePromptInput, cmd = m.claudePromptInput.Update(keyMsg)
				return m, cmd
			}
		}
		// Non-key message: fall through to the main switch to handle it normally.
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Dismiss any visible error overlay on the next keypress.
		m.Error = ""
		switch msg.Type {
		case tea.KeyTab:
			m.focused = (m.focused + 1) % panelCount
			return m, nil
		case tea.KeyEnter:
			switch m.view {
			case viewPRs:
				if len(m.prs) == 0 || m.selectedPRIdx >= len(m.prs) {
					return m, nil
				}
				pr := m.prs[m.selectedPRIdx]
				path := prWorktreePath(m.RepoPath, pr.Branch)
				// Guard: if any existing worktree already uses this branch, show an error.
				for _, wt := range m.Worktrees {
					if wt.Branch == pr.Branch {
						m.Error = fmt.Sprintf("Worktree for branch %q already exists at %s", pr.Branch, wt.Path)
						return m, nil
					}
				}
				m.activeModal = modal.NewPRCheckoutModal(pr, path)
				return m, nil
			default:
				if selected, ok := m.selectedWorktree(); ok {
					return m, m.switchWorktreeCmd(selected.Path)
				}
				return m, nil
			}
		case tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyCtrlN:
			return m, m.fetchIssuesCmd()
		case tea.KeyCtrlD:
			if selected, ok := m.selectedWorktree(); ok {
				m.activeModal = modal.NewDeleteModal(selected)
			}
		case tea.KeyUp:
			m.moveUp()
		case tea.KeyDown:
			m.moveDown()
		case tea.KeySpace:
			if m.view != viewWorktrees {
				m.Error = "Agent launcher is only available in the Worktrees view — press w to switch"
				return m, nil
			}
			if selected, ok := m.selectedWorktree(); ok {
				m.activeModal = modal.NewAgentLauncherModal(m.Config, selected.Path)
			} else {
				m.Error = "No worktree selected — select one first"
			}
			return m, nil
		case tea.KeyRunes:
			switch msg.String() {
			case " ":
				// Spacebar can arrive as KeyRunes " " on some terminals (e.g. Windows).
				// Mirror the KeySpace handler above.
				if m.view != viewWorktrees {
					m.Error = "Agent launcher is only available in the Worktrees view — press w to switch"
					return m, nil
				}
				if selected, ok := m.selectedWorktree(); ok {
					m.activeModal = modal.NewAgentLauncherModal(m.Config, selected.Path)
				} else {
					m.Error = "No worktree selected — select one first"
				}
				return m, nil
			case "j":
				m.moveDown()
				return m, nil
			case "k":
				m.moveUp()
				return m, nil
			case "t":
				m.themeIdx = (m.themeIdx + 1) % len(styles.Themes)
			case "w", "W":
				m.view = viewWorktrees
				m.ctxScrollOffset = 0
			case "i", "I":
				m.view = viewIssues
				m.ctxScrollOffset = 0
			case "p", "P":
				m.view = viewPRs
				m.ctxScrollOffset = 0
			case "g", "G":
				return m, m.openInBrowserCmd()
			case "s", "S":
				if m.view == viewWorktrees {
					if selected, ok := m.selectedWorktree(); ok {
						return m, m.switchWorktreeCmd(selected.Path)
					}
				}
			case "c", "C":
				if m.view != viewWorktrees {
					m.Error = "Copilot (c) is only available in the Worktrees view — press w to switch"
					return m, nil
				}
				if !m.Config.AIAgents.CopilotEnabled {
					m.Error = "Copilot is disabled — set copilot_enabled = true in ~/.nexus/config.toml"
					return m, nil
				}
				if _, ok := m.selectedWorktree(); !ok {
					m.Error = "No worktree selected — select one first"
					return m, nil
				}
				if _, err := exec.LookPath("gh"); err != nil {
					m.Error = "gh not found on $PATH — install GitHub CLI to use Copilot"
					return m, nil
				}
				ti := textinput.New()
				ti.Placeholder = "Enter Copilot prompt…"
				focusCmd := ti.Focus()
				m.copilotPromptInput = ti
				m.copilotPromptActive = true
				return m, focusCmd
			case "a", "A":
				if m.view != viewWorktrees {
					m.Error = "Claude (a) is only available in the Worktrees view — press w to switch"
					return m, nil
				}
				if !m.Config.AIAgents.ClaudeEnabled {
					m.Error = "Claude is disabled — set claude_enabled = true in ~/.nexus/config.toml"
					return m, nil
				}
				if _, ok := m.selectedWorktree(); !ok {
					m.Error = "No worktree selected — select one first"
					return m, nil
				}
				if _, err := resolveClaudeBinary(m.Config); err != nil {
					m.Error = fmt.Sprintf("claude binary not found: %v", err)
					return m, nil
				}
				ti := textinput.New()
				ti.Placeholder = "Enter Claude prompt…"
				focusCmd := ti.Focus()
				m.claudePromptInput = ti
				m.claudePromptActive = true
				return m, focusCmd
			case "f", "F":
				if m.view != viewWorktrees {
					m.Error = "Aider (f) is only available in the Worktrees view — press w to switch"
					return m, nil
				}
				if !m.Config.AIAgents.AiderEnabled {
					m.Error = "Aider is disabled — set aider_enabled = true in ~/.nexus/config.toml"
					return m, nil
				}
				selected, ok := m.selectedWorktree()
				if !ok {
					m.Error = "No worktree selected — select one first"
					return m, nil
				}
				if _, err := resolveAiderBinary(m.Config); err != nil {
					m.Error = "aider not found on $PATH — install Aider to use this feature"
					return m, nil
				}
				return m, m.fetchAiderFilesCmd(selected.Path)
			}
		}

	case issuesFetchedMsg:
		if msg.err == nil {
			m.activeModal = modal.NewCreateModal(msg.issues, m.RepoPath)
		}

	case aiderFilesFetchedMsg:
		if msg.err != nil {
			m.Error = fmt.Sprintf("Failed to list files: %v", msg.err)
			return m, nil
		}
		m.activeModal = modal.NewAiderFilePicker(msg.files)
		return m, nil

	case worktreeOpDoneMsg:
		// Refresh the worktree list after an add/remove operation.
		return m, m.refreshWorktreesCmd()

	case worktreeSwitchedMsg:
		if msg.err != nil {
			m.Error = fmt.Sprintf("Failed to switch worktree: %v", msg.err)
			return m, nil
		}
		m.Error = ""
		// Refresh worktrees after switching back
		return m, m.refreshWorktreesCmd()

	case worktreesRefreshedMsg:
		if msg.err == nil {
			m.Worktrees = msg.worktrees
			m.clampSelectedIdx()
			// Always use the main worktree (first entry) as the canonical repo path
			// so the header shows the repo name rather than the current worktree dir.
			if len(msg.worktrees) > 0 {
				m.RepoPath = msg.worktrees[0].Path
			}
		}

	case browserOpenErrMsg:
		if msg.err != nil {
			m.Error = fmt.Sprintf("Failed to open in browser: %v", msg.err)
		}

	case agentDoneMsg:
		m.copilotPromptActive = false
		m.copilotPromptInput.SetValue("")
		m.claudePromptActive = false
		m.claudePromptInput.SetValue("")
		if m.db != nil {
			entry := data.AgentHistoryEntry{
				AgentName: msg.agentName,
				Prompt:    msg.prompt,
				ExitCode:  msg.exitCode,
				StartedAt: msg.startedAt,
				EndedAt:   time.Now(),
			}
			if err := data.LogAgentRun(m.db, entry); err != nil {
				m.Error = fmt.Sprintf("failed to log agent run: %v", err)
			}
		}
		if msg.exitCode != 0 {
			m.Error = fmt.Sprintf("⚠ Agent exited with code %d", msg.exitCode)
		}
		return m, m.refreshWorktreesCmd()

	case githubSyncedMsg:
		m.syncing = false
		m.syncErr = msg.err
		if msg.err == nil {
			m.prs = msg.prs
			m.issues = msg.issues
			m.lastSynced = msg.syncedAt
			// Clamp per-view selection indices after a sync in case the list shrank.
			m.clampIssueIdx()
			m.clampPRIdx()
		}
		// On error, m.prs and m.issues intentionally retain their previous values
		// so the UI continues to show the last known good data.
		// Schedule next periodic sync tick.
		return m, tea.Tick(m.Config.GitHub.SyncInterval(), func(t time.Time) tea.Msg {
			return syncTickMsg{}
		})

	case syncTickMsg:
		m.syncing = true
		return m, m.syncGitHubCmd()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View returns a string representation of the model's current state.
func (m *Model) View() string {
	baseView := renderFull(m.Worktrees, m.selectedIdx, m.RepoPath, m.themeIdx, m.view, m.width, m.height, m.syncing, m.lastSynced, m.syncErr, m.issues, m.selectedIssueIdx, m.prs, m.selectedPRIdx, m.focused, m.ctxScrollOffset)

	w, h := m.width, m.height
	if w <= 0 {
		w = defaultTermWidth
	}
	if h <= 0 {
		h = 24
	}

	// Overlay helpers — center a themed RenderBox over the full base view.
	overlay := func(title, content string) string {
		theme := styles.NewTheme(styles.Themes[m.themeIdx])
		box := theme.RenderBox(title, content)
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
	}

	if m.activeModal != nil {
		return overlay(m.activeModal.Title(), m.activeModal.View())
	}

	if m.copilotPromptActive {
		return overlay("Spawn Copilot",
			fmt.Sprintf("> %s\n\nEnter confirm  •  Esc cancel", m.copilotPromptInput.View()))
	}

	if m.claudePromptActive {
		return overlay("Spawn Claude Code",
			fmt.Sprintf("> %s\n\nEnter confirm  •  Esc cancel", m.claudePromptInput.View()))
	}

	if m.Error != "" {
		return overlay("⚠ Error", m.Error+"\n\nPress any key to dismiss")
	}

	return baseView
}

// openInBrowserCmd returns a Cmd that opens the selected issue or PR in the browser
// using the gh CLI. Returns nil when in viewWorktrees or when the relevant list is empty.
func (m *Model) openInBrowserCmd() tea.Cmd {
	switch m.view {
	case viewIssues:
		if len(m.issues) == 0 || m.selectedIssueIdx >= len(m.issues) {
			return nil
		}
		num := m.issues[m.selectedIssueIdx].Number
		cmd := exec.Command("gh", "issue", "view", fmt.Sprintf("%d", num), "--web")
		return tea.ExecProcess(cmd, func(err error) tea.Msg { return browserOpenErrMsg{err: err} })
	case viewPRs:
		if len(m.prs) == 0 || m.selectedPRIdx >= len(m.prs) {
			return nil
		}
		num := m.prs[m.selectedPRIdx].Number
		cmd := exec.Command("gh", "pr", "view", fmt.Sprintf("%d", num), "--web")
		return tea.ExecProcess(cmd, func(err error) tea.Msg { return browserOpenErrMsg{err: err} })
	default:
		return nil
	}
}

// fetchIssuesCmd returns a Cmd that fetches open GitHub issues in the background,
// allowing the user to create worktrees from issues.
func (m *Model) fetchIssuesCmd() tea.Cmd {
	repoPath := m.RepoPath
	return func() tea.Msg {
		cmd := internalexec.NewIssueCommand(repoPath)
		issues, err := cmd.ListOpenIssues()
		return issuesFetchedMsg{issues: issues, err: err}
	}
}

// syncGitHubCmd returns a Cmd that fetches open PRs and issues from GitHub in the background.
func (m *Model) syncGitHubCmd() tea.Cmd {
	repoPath := m.RepoPath
	return func() tea.Msg {
		issueCmd := internalexec.NewIssueCommand(repoPath)
		prCmd := internalexec.NewPRCommand(repoPath)
		issues, issErr := issueCmd.ListOpenIssues()
		prs, prErr := prCmd.ListOpenPRs()
		return githubSyncedMsg{prs: prs, issues: issues, err: errors.Join(issErr, prErr), syncedAt: time.Now()}
	}
}

// addWorktreeCmd returns a Cmd that creates a new git worktree with a new branch from main.
func (m *Model) addWorktreeCmd(branch, path string) tea.Cmd {
	repoPath := m.RepoPath
	return func() tea.Msg {
		cmd := internalexec.NewGitCommand(repoPath)
		err := cmd.AddWorktreeNewBranch(path, branch, "main")
		return worktreeOpDoneMsg{err: err}
	}
}

// checkoutPRWorktreeCmd returns a Cmd that fetches a remote PR branch and creates a worktree for it.
func (m *Model) checkoutPRWorktreeCmd(branch, path string) tea.Cmd {
	repoPath := m.RepoPath
	return func() tea.Msg {
		cmd := internalexec.NewGitCommand(repoPath)
		err := cmd.CheckoutPRWorktree(path, branch)
		return worktreeOpDoneMsg{err: err}
	}
}

// prWorktreePath derives the filesystem path for a PR worktree using the same
// convention as issue worktrees: ../worktrees/<branch-with-slashes-as-dashes>.
func prWorktreePath(repoPath, branch string) string {
	slug := strings.ReplaceAll(branch, "/", "-")
	return filepath.Join(filepath.Dir(repoPath), "worktrees", slug)
}

// removeWorktreeCmd returns a Cmd that removes a git worktree.
func (m *Model) removeWorktreeCmd(path string) tea.Cmd {
	repoPath := m.RepoPath
	return func() tea.Msg {
		cmd := internalexec.NewGitCommand(repoPath)
		err := cmd.RemoveWorktree(path, true)
		return worktreeOpDoneMsg{err: err}
	}
}

// switchWorktreeCmd returns a Cmd that launches a shell in the specified worktree directory,
// allowing the user to work within the worktree before returning to the TUI.
func (m *Model) switchWorktreeCmd(path string) tea.Cmd {
	cmd := buildShellCmd(path)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return worktreeSwitchedMsg{err: err}
	})
}

// buildCopilotCmd constructs the exec.Cmd for running gh copilot in interactive
// mode with the given prompt pre-loaded in the specified worktree directory.
// It is extracted as a top-level function to keep it unit-testable.
func buildCopilotCmd(worktreePath, prompt string) *exec.Cmd {
	cmd := exec.Command("gh", "copilot", "-i", prompt)
	cmd.Dir = worktreePath
	return cmd
}

// spawnCopilotCmd returns a Cmd that runs gh copilot suggest in the worktree
// directory and dispatches agentDoneMsg when the process exits.
func (m *Model) spawnCopilotCmd(worktreePath, prompt string) tea.Cmd {
	startedAt := time.Now()
	cmd := buildCopilotCmd(worktreePath, prompt)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		exitCode := 0
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		return agentDoneMsg{
			agentName: "copilot",
			prompt:    prompt,
			exitCode:  exitCode,
			startedAt: startedAt,
		}
	})
}

// resolveClaudeBinary returns the resolved path for the Claude binary.
// It reads cfg.AIAgents.ClaudeBinary, defaulting to "claude", then
// uses exec.LookPath to verify the binary is on the PATH.
func resolveClaudeBinary(cfg *domain.Config) (string, error) {
	bin := cfg.AIAgents.ClaudeBinary
	if bin == "" {
		bin = "claude"
	}
	return exec.LookPath(bin)
}

// resolveAiderBinary returns the resolved path for the Aider binary.
// It reads cfg.AIAgents.AiderBinary, defaulting to "aider", then
// uses exec.LookPath to verify the binary is on the PATH.
func resolveAiderBinary(cfg *domain.Config) (string, error) {
	bin := cfg.AIAgents.AiderBinary
	if bin == "" {
		bin = "aider"
	}
	return exec.LookPath(bin)
}

// buildClaudeCmd constructs the exec.Cmd for running the Claude CLI with the
// given prompt in the specified worktree directory.
// It is extracted as a top-level function to keep it unit-testable.
func buildClaudeCmd(worktreePath, prompt, binaryPath string) *exec.Cmd {
	cmd := exec.Command(binaryPath, prompt)
	cmd.Dir = worktreePath
	return cmd
}

// spawnClaudeCmd returns a Cmd that runs the Claude binary in the worktree
// directory and dispatches agentDoneMsg when the process exits.
func (m *Model) spawnClaudeCmd(worktreePath, prompt string) tea.Cmd {
	binaryPath, err := resolveClaudeBinary(m.Config)
	if err != nil {
		m.Error = fmt.Sprintf("claude binary not found: %v", err)
		return nil
	}
	startedAt := time.Now()
	cmd := buildClaudeCmd(worktreePath, prompt, binaryPath)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		exitCode := 0
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		return agentDoneMsg{
			agentName: "claude",
			prompt:    prompt,
			exitCode:  exitCode,
			startedAt: startedAt,
		}
	})
}

// fetchAiderFilesCmd returns a Cmd that lists modified files in the worktree
// using git ls-files, dispatching aiderFilesFetchedMsg with the result.
func (m *Model) fetchAiderFilesCmd(worktreePath string) tea.Cmd {
	return func() tea.Msg {
		cmd := internalexec.NewGitCommand(worktreePath)
		files, err := cmd.ListModifiedFiles(worktreePath)
		return aiderFilesFetchedMsg{worktreePath: worktreePath, files: files, err: err}
	}
}

// buildAiderCmd constructs the exec.Cmd for running aider with the given files
// in the specified worktree directory. Extracted as a top-level function for testability.
func buildAiderCmd(worktreePath string, files []string, binaryPath string) *exec.Cmd {
	cmd := exec.Command(binaryPath, files...)
	cmd.Dir = worktreePath
	return cmd
}

// spawnAiderCmd returns a Cmd that runs aider with the selected files in the
// worktree directory and dispatches agentDoneMsg when the process exits.
func (m *Model) spawnAiderCmd(worktreePath string, files []string) tea.Cmd {
	binaryPath, err := resolveAiderBinary(m.Config)
	if err != nil {
		m.Error = fmt.Sprintf("aider not found: %v", err)
		return nil
	}
	startedAt := time.Now()
	cmd := buildAiderCmd(worktreePath, files, binaryPath)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		exitCode := 0
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		return agentDoneMsg{
			agentName: "aider",
			exitCode:  exitCode,
			startedAt: startedAt,
		}
	})
}

// buildShellCmd constructs a platform-appropriate shell command for the given directory.
// On Windows without a SHELL env var, it uses cmd.exe with /K flag to keep the shell open.
// When SHELL is set (e.g. Git Bash), it respects that on all platforms.
func buildShellCmd(path string) *exec.Cmd {
	return buildShellCmdForOS(path, runtime.GOOS, os.Getenv("SHELL"))
}

// buildShellCmdForOS constructs a shell command for a specific OS and shell value.
// It exists to keep buildShellCmd testable across platforms.
// On Windows with no shell configured, it falls back to cmd.exe.
// When shell is set (e.g. via SHELL env var in Git Bash), it is used on any OS.
func buildShellCmdForOS(path, goos, shell string) *exec.Cmd {
	// On Windows, prefer the SHELL env var when set (e.g. Git Bash / MSYS2).
	// Only fall back to cmd.exe when no Unix-compatible shell is configured.
	if goos == "windows" && shell == "" {
		cmd := exec.Command("cmd", "/K")
		cmd.Dir = path
		return cmd
	}

	if shell == "" {
		shell = "/bin/sh"
	}

	cmd := exec.Command(shell)
	cmd.Dir = path
	return cmd
}

// getShell returns the user's preferred shell, or /bin/sh as a fallback.
// It reads the SHELL environment variable on Unix-like systems.
func getShell() string {
	shell := os.Getenv("SHELL")
	if shell != "" {
		return shell
	}
	return "/bin/sh"
}

// worktreesRefreshedMsg carries the result of refreshing the worktree list.
type worktreesRefreshedMsg struct {
	worktrees []domain.Worktree
	err       error
}

// refreshWorktreesCmd returns a Cmd that reloads the worktree list from git.
func (m *Model) refreshWorktreesCmd() tea.Cmd {
	repoPath := m.RepoPath
	return func() tea.Msg {
		cmd := internalexec.NewGitCommand(repoPath)
		worktrees, err := cmd.ListWorktrees()
		return worktreesRefreshedMsg{worktrees: worktrees, err: err}
	}
}

func (m *Model) selectedWorktree() (domain.Worktree, bool) {
	if len(m.Worktrees) == 0 || m.selectedIdx < 0 || m.selectedIdx >= len(m.Worktrees) {
		return domain.Worktree{}, false
	}

	return m.Worktrees[m.selectedIdx], true
}

func (m *Model) clampSelectedIdx() {
	if len(m.Worktrees) == 0 {
		m.selectedIdx = 0
		return
	}

	if m.selectedIdx < 0 {
		m.selectedIdx = 0
		return
	}

	if m.selectedIdx >= len(m.Worktrees) {
		m.selectedIdx = len(m.Worktrees) - 1
	}
}

func (m *Model) clampIssueIdx() {
	if len(m.issues) == 0 {
		m.selectedIssueIdx = 0
		return
	}
	if m.selectedIssueIdx >= len(m.issues) {
		m.selectedIssueIdx = len(m.issues) - 1
	}
}

func (m *Model) clampPRIdx() {
	if len(m.prs) == 0 {
		m.selectedPRIdx = 0
		return
	}
	if m.selectedPRIdx >= len(m.prs) {
		m.selectedPRIdx = len(m.prs) - 1
	}
}

// moveDown advances the selection within the currently focused panel.
// Nav panel: cycles the active view forward.
// Ctx panel: scrolls the context content down.
// List panel (default): moves the item cursor down.
func (m *Model) moveDown() {
	switch m.focused {
	case panelNav:
		n := int(m.view) + 1
		if n > int(viewPRs) {
			n = int(viewWorktrees)
		}
		m.view = activeView(n)
	case panelCtx:
		m.ctxScrollOffset++
	default: // panelList
		switch m.view {
		case viewIssues:
			if m.selectedIssueIdx < len(m.issues)-1 {
				m.selectedIssueIdx++
				m.ctxScrollOffset = 0
			}
		case viewPRs:
			if m.selectedPRIdx < len(m.prs)-1 {
				m.selectedPRIdx++
				m.ctxScrollOffset = 0
			}
		default:
			if m.selectedIdx < len(m.Worktrees)-1 {
				m.selectedIdx++
				m.ctxScrollOffset = 0
			}
		}
	}
}

// moveUp retreats the selection within the currently focused panel.
// Nav panel: cycles the active view backward.
// Ctx panel: scrolls the context content up.
// List panel (default): moves the item cursor up.
func (m *Model) moveUp() {
	switch m.focused {
	case panelNav:
		n := int(m.view) - 1
		if n < int(viewWorktrees) {
			n = int(viewPRs)
		}
		m.view = activeView(n)
	case panelCtx:
		if m.ctxScrollOffset > 0 {
			m.ctxScrollOffset--
		}
	default: // panelList
		switch m.view {
		case viewIssues:
			if m.selectedIssueIdx > 0 {
				m.selectedIssueIdx--
				m.ctxScrollOffset = 0
			}
		case viewPRs:
			if m.selectedPRIdx > 0 {
				m.selectedPRIdx--
				m.ctxScrollOffset = 0
			}
		default:
			if m.selectedIdx > 0 {
				m.selectedIdx--
				m.ctxScrollOffset = 0
			}
		}
	}
}
