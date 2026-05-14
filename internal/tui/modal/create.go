package modal

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/m00nk0d3/nexus/internal/domain"
)

type createStep int

const (
	stepIssues createStep = iota
	stepType
	stepSlug
	stepConfirm
)

// BranchTypes are the valid conventional commit type prefixes for branch names.
var BranchTypes = []string{"feat", "fix", "chore", "docs", "test", "refactor"}

// CreateModal is a multi-step Bubbletea model for creating a new worktree from a GitHub issue.
// Steps: issue picker → type picker → slug editor → confirm.
type CreateModal struct {
	step      createStep
	issues    []domain.Issue
	issueIdx  int
	typeIdx   int
	slugInput textinput.Model
	repoPath  string
}

// NewCreateModal creates a new CreateModal with the given issues and repo path.
func NewCreateModal(issues []domain.Issue, repoPath string) *CreateModal {
	ti := textinput.New()
	ti.Placeholder = "slug"
	ti.CharLimit = 60

	return &CreateModal{
		step:      stepIssues,
		issues:    issues,
		repoPath:  repoPath,
		slugInput: ti,
	}
}

// Init satisfies tea.Model.
func (m *CreateModal) Init() tea.Cmd { return nil }

// Update handles Bubbletea messages, driving the multi-step state machine.
func (m *CreateModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, isKey := msg.(tea.KeyMsg)
	if !isKey {
		if m.step == stepSlug {
			var cmd tea.Cmd
			m.slugInput, cmd = m.slugInput.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	// Esc always cancels regardless of step.
	if keyMsg.Type == tea.KeyEsc {
		return m, func() tea.Msg { return ModalCancelledMsg{} }
	}

	// On the slug step, route enter to advance and everything else to the text input.
	if m.step == stepSlug {
		if keyMsg.Type == tea.KeyEnter {
			m.slugInput.Blur()
			m.step = stepConfirm
			return m, nil
		}
		var cmd tea.Cmd
		m.slugInput, cmd = m.slugInput.Update(msg)
		return m, cmd
	}

	switch keyMsg.String() {
	case "up", "k":
		m.moveUp()
	case "down", "j":
		m.moveDown()
	case "enter":
		return m.advance()
	}

	return m, nil
}

func (m *CreateModal) moveUp() {
	switch m.step {
	case stepIssues:
		if m.issueIdx > 0 {
			m.issueIdx--
		}
	case stepType:
		if m.typeIdx > 0 {
			m.typeIdx--
		}
	}
}

func (m *CreateModal) moveDown() {
	switch m.step {
	case stepIssues:
		if m.issueIdx < len(m.issues)-1 {
			m.issueIdx++
		}
	case stepType:
		if m.typeIdx < len(BranchTypes)-1 {
			m.typeIdx++
		}
	}
}

func (m *CreateModal) advance() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepIssues:
		if len(m.issues) == 0 {
			return m, nil
		}
		m.step = stepType

	case stepType:
		slug := domain.SlugFromTitle(m.SelectedIssue().Title)
		m.slugInput.SetValue(slug)
		m.slugInput.Focus()
		m.step = stepSlug

	case stepConfirm:
		branch := m.BranchName()
		path := m.WorktreePath()
		return m, func() tea.Msg {
			return WorktreeCreateConfirmedMsg{Branch: branch, Path: path}
		}
	}

	return m, nil
}

// SelectedIssue returns the currently selected issue.
func (m *CreateModal) SelectedIssue() domain.Issue {
	if len(m.issues) == 0 {
		return domain.Issue{}
	}
	return m.issues[m.issueIdx]
}

// SelectedType returns the currently selected branch type prefix.
func (m *CreateModal) SelectedType() string {
	return BranchTypes[m.typeIdx]
}

// BranchName returns the branch name following the <type>/issue-<N>-<slug> convention.
func (m *CreateModal) BranchName() string {
	issue := m.SelectedIssue()
	return fmt.Sprintf("%s/issue-%d-%s", m.SelectedType(), issue.Number, m.slugInput.Value())
}

// WorktreePath returns the filesystem path for the new worktree: ../worktrees/<type>-issue-<N>-<slug>.
func (m *CreateModal) WorktreePath() string {
	slug := strings.ReplaceAll(m.BranchName(), "/", "-")
	return filepath.Join(filepath.Dir(m.repoPath), "worktrees", slug)
}

// View renders the current step of the modal.
func (m *CreateModal) View() string {
	switch m.step {
	case stepIssues:
		return m.viewIssueList()
	case stepType:
		return m.viewTypePicker()
	case stepSlug:
		return m.viewSlugEditor()
	case stepConfirm:
		return m.viewConfirm()
	}
	return ""
}

func (m *CreateModal) viewIssueList() string {
	var b strings.Builder
	b.WriteString("Select issue:\n\n")

	for i, issue := range m.issues {
		cursor := "  "
		if i == m.issueIdx {
			cursor = "> "
		}
		b.WriteString(fmt.Sprintf("%s#%d %s\n", cursor, issue.Number, issue.Title))
	}

	if len(m.issues) == 0 {
		b.WriteString("  (no open issues)\n")
	}

	b.WriteString("\n↑/↓ navigate  •  Enter select  •  Esc cancel")
	return b.String()
}

func (m *CreateModal) viewTypePicker() string {
	issue := m.SelectedIssue()
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Issue: #%d %s\n\n", issue.Number, issue.Title))
	b.WriteString("Select branch type:\n\n")

	for i, t := range BranchTypes {
		cursor := "  "
		if i == m.typeIdx {
			cursor = "> "
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, t))
	}

	b.WriteString("\n↑/↓ navigate  •  Enter select  •  Esc cancel")
	return b.String()
}

func (m *CreateModal) viewSlugEditor() string {
	issue := m.SelectedIssue()
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Issue: #%d %s\n", issue.Number, issue.Title))
	b.WriteString(fmt.Sprintf("Type:  %s\n\n", m.SelectedType()))
	b.WriteString("Edit slug:\n")
	b.WriteString(m.slugInput.View())
	b.WriteString(fmt.Sprintf("\n\nBranch preview: %s/issue-%d-%s\n", m.SelectedType(), issue.Number, m.slugInput.Value()))
	b.WriteString("\nEnter confirm  •  Esc cancel")
	return b.String()
}

func (m *CreateModal) viewConfirm() string {
	var b strings.Builder
	b.WriteString("Create worktree:\n\n")
	b.WriteString(fmt.Sprintf("  Branch:  %s\n", m.BranchName()))
	b.WriteString(fmt.Sprintf("  Path:    %s\n", m.WorktreePath()))
	b.WriteString("\nEnter confirm  •  Esc cancel")
	return b.String()
}
