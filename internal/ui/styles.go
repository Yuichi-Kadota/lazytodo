package ui

import "github.com/charmbracelet/lipgloss"

// Layout constants
const (
	WorkspacePaneRatio = 0.3  // 30%
	TodoPaneRatio      = 0.7  // 70%
	MinPaneWidth       = 20
)

// Styles holds all the application styles
type Styles struct {
	// Base styles
	App lipgloss.Style

	// Pane styles
	ActivePane   lipgloss.Style
	InactivePane lipgloss.Style
	PaneTitle    lipgloss.Style

	// List item styles
	SelectedItem   lipgloss.Style
	UnselectedItem lipgloss.Style
	CompletedItem  lipgloss.Style

	// Workspace styles
	WorkspaceRoot  lipgloss.Style
	WorkspaceChild lipgloss.Style

	// Todo styles
	TodoPending  lipgloss.Style
	TodoComplete lipgloss.Style
	TodoUrgent   lipgloss.Style

	// Status bar styles
	StatusBar   lipgloss.Style
	ModeNormal  lipgloss.Style
	ModeInsert  lipgloss.Style
	ModeSearch  lipgloss.Style
	ModeSort    lipgloss.Style
	Notification lipgloss.Style
	ErrorNotif   lipgloss.Style

	// Input bar styles
	InputBar lipgloss.Style

	// Tree styles
	TreeGuide lipgloss.Style

	// Empty state
	EmptyState lipgloss.Style
}

// NewStyles creates a new Styles instance
func NewStyles() Styles {
	return Styles{
		App: lipgloss.NewStyle().
			Background(ColorBackground).
			Foreground(ColorForeground),

		ActivePane: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1),

		InactivePane: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(0, 1),

		PaneTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorForeground).
			Padding(0, 1),

		SelectedItem: lipgloss.NewStyle().
			Background(ColorSelectedBg).
			Foreground(ColorForeground).
			Bold(true),

		UnselectedItem: lipgloss.NewStyle().
			Foreground(ColorForeground),

		CompletedItem: lipgloss.NewStyle().
			Foreground(ColorTodoComplete).
			Strikethrough(true),

		WorkspaceRoot: lipgloss.NewStyle().
			Foreground(ColorFolderRoot),

		WorkspaceChild: lipgloss.NewStyle().
			Foreground(ColorFolderChild),

		TodoPending: lipgloss.NewStyle().
			Foreground(ColorTodoPending),

		TodoComplete: lipgloss.NewStyle().
			Foreground(ColorTodoComplete),

		TodoUrgent: lipgloss.NewStyle().
			Foreground(ColorTodoUrgent).
			Bold(true),

		StatusBar: lipgloss.NewStyle().
			Background(ColorBackground).
			Foreground(ColorForeground).
			Padding(0, 1),

		ModeNormal: lipgloss.NewStyle().
			Background(ColorPrimary).
			Foreground(ColorBackground).
			Bold(true).
			Padding(0, 1),

		ModeInsert: lipgloss.NewStyle().
			Background(ColorSuccess).
			Foreground(ColorBackground).
			Bold(true).
			Padding(0, 1),

		ModeSearch: lipgloss.NewStyle().
			Background(ColorWarning).
			Foreground(ColorBackground).
			Bold(true).
			Padding(0, 1),

		ModeSort: lipgloss.NewStyle().
			Background(ColorSecondary).
			Foreground(ColorBackground).
			Bold(true).
			Padding(0, 1),

		Notification: lipgloss.NewStyle().
			Foreground(ColorSuccess),

		ErrorNotif: lipgloss.NewStyle().
			Foreground(ColorError),

		InputBar: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(ColorMuted).
			Padding(0, 1),

		TreeGuide: lipgloss.NewStyle().
			Foreground(ColorTreeGuide),

		EmptyState: lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true).
			Align(lipgloss.Center),
	}
}

// GetModeStyle returns the style for a given mode
func (s Styles) GetModeStyle(mode string) lipgloss.Style {
	switch mode {
	case "INSERT":
		return s.ModeInsert
	case "SEARCH":
		return s.ModeSearch
	case "SORT":
		return s.ModeSort
	default:
		return s.ModeNormal
	}
}
