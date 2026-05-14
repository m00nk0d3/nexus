package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/m00nk0d3/nexus/internal/domain"
	"github.com/m00nk0d3/nexus/internal/exec"
	"github.com/m00nk0d3/nexus/internal/tui/modal"
)

// issuesFetchedMsg carries the result of a background gh issue list call.
type issuesFetchedMsg struct {
	issues []domain.Issue
	err    error
}

// worktreeOpDoneMsg carries the result of an add/remove worktree operation.
type worktreeOpDoneMsg struct {
	err error
}

// Model represents the root Bubbletea model for the Nexus TUI application.
type Model struct {
	Worktrees   []domain.Worktree
	RepoPath    string
	selectedIdx int
	activeModal tea.Model
}

// NewModel creates and returns a new Model instance with all required fields initialized.
func NewModel() *Model {
	return &Model{}
}

// Init initializes the model and returns an initial command.
func (m *Model) Init() tea.Cmd {
	return nil
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
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyCtrlN:
			return m, m.fetchIssuesCmd()
		case tea.KeyCtrlD:
			if len(m.Worktrees) > 0 {
				m.activeModal = modal.NewDeleteModal(m.Worktrees[m.selectedIdx])
			}
		case tea.KeyUp:
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
		case tea.KeyDown:
			if m.selectedIdx < len(m.Worktrees)-1 {
				m.selectedIdx++
			}
		}

	case issuesFetchedMsg:
		if msg.err == nil {
			m.activeModal = modal.NewCreateModal(msg.issues, m.RepoPath)
		}

	case worktreeOpDoneMsg:
		// Refresh the worktree list after an add/remove operation.
		return m, m.refreshWorktreesCmd()

	case worktreesRefreshedMsg:
		if msg.err == nil {
			m.Worktrees = msg.worktrees
		}

	case tea.WindowSizeMsg:
		// handled for future layout use
	}

	return m, nil
}

// View returns a string representation of the model's current state.
func (m *Model) View() string {
	if m.activeModal != nil {
		return m.activeModal.View()
	}

	if len(m.Worktrees) > 0 {
		return renderWorktreeList(m.Worktrees)
	}

	return "Nexus TUI"
}

// fetchIssuesCmd returns a Cmd that fetches open GitHub issues in the background.
func (m *Model) fetchIssuesCmd() tea.Cmd {
	repoPath := m.RepoPath
	return func() tea.Msg {
		cmd := exec.NewIssueCommand(repoPath)
		issues, err := cmd.ListOpenIssues()
		return issuesFetchedMsg{issues: issues, err: err}
	}
}

// addWorktreeCmd returns a Cmd that creates a new git worktree with a new branch from main.
func (m *Model) addWorktreeCmd(branch, path string) tea.Cmd {
	repoPath := m.RepoPath
	return func() tea.Msg {
		cmd := exec.NewGitCommand(repoPath)
		err := cmd.AddWorktreeNewBranch(path, branch, "main")
		return worktreeOpDoneMsg{err: err}
	}
}

// removeWorktreeCmd returns a Cmd that removes a git worktree.
func (m *Model) removeWorktreeCmd(path string) tea.Cmd {
	repoPath := m.RepoPath
	return func() tea.Msg {
		cmd := exec.NewGitCommand(repoPath)
		err := cmd.RemoveWorktree(path, true)
		return worktreeOpDoneMsg{err: err}
	}
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
		cmd := exec.NewGitCommand(repoPath)
		worktrees, err := cmd.ListWorktrees()
		return worktreesRefreshedMsg{worktrees: worktrees, err: err}
	}
}
