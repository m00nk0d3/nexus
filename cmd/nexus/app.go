package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

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

// Model represents the root Bubbletea model for the Nexus TUI application.
// It manages the list of git worktrees, user interactions, and active modals.
type Model struct {
	Worktrees   []domain.Worktree // List of available git worktrees
	RepoPath    string            // Path to the repository root
	Config      *domain.Config    // Loaded application configuration
	selectedIdx int               // Currently selected worktree index
	activeModal tea.Model         // Currently open modal (if any)
	Error       string            // Error message to display (if any)
	themeIdx    int               // Index into styles.Themes for the active theme
	activeNav   int               // Index of the active nav rail section (0=W,1=I,2=P,3=T)
	width       int               // Terminal width in columns; 0 means use default
	height      int               // Terminal height in rows; 0 means use default
}

// NewModel creates and returns a new Model instance with all required fields initialized.
func NewModel() *Model {
	cfg, err := data.LoadConfig(data.DefaultConfigPath())
	if err != nil {
		cfg = domain.DefaultConfig()
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
	}
}

// Init initializes the model and triggers an initial worktree list load.
func (m *Model) Init() tea.Cmd {
	return m.refreshWorktreesCmd()
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
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
		case tea.KeyDown:
			if m.selectedIdx < len(m.Worktrees)-1 {
				m.selectedIdx++
			}
		case tea.KeyRunes:
			if msg.String() == "t" {
				m.themeIdx = (m.themeIdx + 1) % len(styles.Themes)
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
		baseView = renderFull(m.Worktrees, m.selectedIdx, m.RepoPath, m.themeIdx, m.activeNav, m.width, m.height)
	}

	if m.Error == "" {
		return baseView
	}

	return fmt.Sprintf("Error: %s\n\n%s", m.Error, baseView)
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
