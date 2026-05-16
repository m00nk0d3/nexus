package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
	activeModal      tea.Model            // Currently open modal (if any)
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
		case modal.WorktreeDeleteConfirmedMsg:
			m.activeModal = nil
			return m, m.removeWorktreeCmd(msg.Path)
		case modal.ModalCancelledMsg:
			m.activeModal = nil
			return m, nil
		default:
			updated, cmd := m.activeModal.Update(msg)
			m.activeModal = updated
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			m.focused = (m.focused + 1) % panelCount
			return m, nil
		case tea.KeyEnter:
			if selected, ok := m.selectedWorktree(); ok {
				return m, m.switchWorktreeCmd(selected.Path)
			}
			return m, nil
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
		case tea.KeyRunes:
			switch msg.String() {
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
			}
		}

	case issuesFetchedMsg:
		if msg.err == nil {
			m.activeModal = modal.NewCreateModal(msg.issues, m.RepoPath)
		}

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
	var baseView string
	if m.activeModal != nil {
		baseView = m.activeModal.View()
	} else {
		baseView = renderFull(m.Worktrees, m.selectedIdx, m.RepoPath, m.themeIdx, m.view, m.width, m.height, m.syncing, m.lastSynced, m.syncErr, m.issues, m.selectedIssueIdx, m.prs, m.selectedPRIdx, m.focused, m.ctxScrollOffset)
	}

	if m.Error == "" {
		return baseView
	}

	return fmt.Sprintf("Error: %s\n\n%s", m.Error, baseView)
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

// buildShellCmd constructs a platform-appropriate shell command for the given directory.
// On Windows, it uses cmd.exe with /K flag to keep the shell open.
// On Unix-like systems, it uses the SHELL environment variable, defaulting to /bin/sh.
func buildShellCmd(path string) *exec.Cmd {
	return buildShellCmdForOS(path, runtime.GOOS, getShell())
}

// buildShellCmdForOS constructs a shell command for a specific OS and shell value.
// It exists to keep buildShellCmd testable across platforms.
func buildShellCmdForOS(path, goos, shell string) *exec.Cmd {
	if goos == "windows" {
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
