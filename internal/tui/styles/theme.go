package styles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Themes is the ordered list of available theme names used for cycling.
var Themes = []string{"digital-noir", "matrix", "light"}

// Theme holds the color palette and component styles for a named visual theme.
type Theme struct {
	Name    string
	accent  string
	bg      string
	surface string
	fg      string
	muted   string
	success string
	warning string
	danger  string
}

// NewTheme creates a Theme for the given name, defaulting to digital-noir.
func NewTheme(name string) Theme {
	switch name {
	case "matrix":
		return Theme{
			Name:    "matrix",
			accent:  "#00FF00",
			bg:      "#000000",
			surface: "#0a0a0a",
			fg:      "#00FF00",
			muted:   "#006600",
			success: "#00FF00",
			warning: "#FFFF00",
			danger:  "#FF0000",
		}
	case "light":
		return Theme{
			Name:    "light",
			accent:  "#0066CC",
			bg:      "#F0F0F0",
			surface: "#F5F5F5",
			fg:      "#1A1A1A",
			muted:   "#666666",
			success: "#008000",
			warning: "#CC6600",
			danger:  "#CC0000",
		}
	default: // digital-noir
		return Theme{
			Name:    "digital-noir",
			accent:  "#00D9FF",
			bg:      "#0a0e27",
			surface: "#0d1117",
			fg:      "#E2E8F0",
			muted:   "#4A5568",
			success: "#00FF88",
			warning: "#FFD700",
			danger:  "#FF4757",
		}
	}
}

// GetStyle returns a lipgloss.Style for the named component.
// Unknown component names return a default surface/foreground style.
func (t Theme) GetStyle(component string) lipgloss.Style {
	switch component {
	case "header":
		return lipgloss.NewStyle().
			Background(lipgloss.Color(t.accent)).
			Foreground(lipgloss.Color(t.bg)).
			Bold(true).
			Padding(0, 1)
	case "nav-rail":
		return lipgloss.NewStyle().
			Background(lipgloss.Color(t.surface)).
			Foreground(lipgloss.Color(t.fg)).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.accent)).
			Padding(0, 1)
	case "worktree-list":
		return lipgloss.NewStyle().
			Background(lipgloss.Color(t.surface)).
			Foreground(lipgloss.Color(t.fg)).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.accent)).
			Padding(0, 1)
	case "selected-row":
		return lipgloss.NewStyle().
			Background(lipgloss.Color(t.accent)).
			Foreground(lipgloss.Color(t.bg)).
			Bold(true)
	case "status-bar":
		return lipgloss.NewStyle().
			Background(lipgloss.Color(t.surface)).
			Foreground(lipgloss.Color(t.muted))
	case "modal-border":
		return lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.accent))
	case "error":
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.danger))
	case "success":
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.success))
	case "context-panel":
		return lipgloss.NewStyle().
			Background(lipgloss.Color(t.surface)).
			Foreground(lipgloss.Color(t.fg)).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(t.accent)).
			Padding(0, 1)
	case "table-header":
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.muted)).
			Bold(true)
	default:
		return lipgloss.NewStyle().
			Background(lipgloss.Color(t.surface)).
			Foreground(lipgloss.Color(t.fg))
	}
}

// StatusStyle returns a lipgloss.Style for a worktree status value.
func (t Theme) StatusStyle(status string) lipgloss.Style {
	switch strings.ToLower(status) {
	case "checked", "checked out": // reserved for future git worktree "checkedout" state
		return lipgloss.NewStyle().Foreground(lipgloss.Color(t.accent))
	case "idle", "clean":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(t.success))
	case "created", "dirty":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(t.warning))
	case "locked":
		return lipgloss.NewStyle().Foreground(lipgloss.Color(t.danger))
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color(t.muted))
	}
}

// RenderBox renders content inside a rounded-border panel with an optional title.
func (t Theme) RenderBox(title, content string) string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(t.accent)).
		Padding(0, 1)
	if title != "" {
		return style.Render(fmt.Sprintf("%s\n%s", title, content))
	}
	return style.Render(content)
}

// RenderTable renders a padded table with styled muted column headers.
func (t Theme) RenderTable(rows [][]string, headers []string) string {
	var b strings.Builder
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.muted)).
		Bold(true)

	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	for i, h := range headers {
		b.WriteString(headerStyle.Render(fmt.Sprintf("%-*s", colWidths[i], h)))
		if i < len(headers)-1 {
			b.WriteString("  ")
		}
	}
	b.WriteString("\n")

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) {
				b.WriteString(fmt.Sprintf("%-*s", colWidths[i], cell))
				if i < len(row)-1 {
					b.WriteString("  ")
				}
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}
