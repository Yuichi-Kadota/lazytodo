package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yuichikadota/lazytodo/internal/input"
	"github.com/yuichikadota/lazytodo/internal/ui"
)

var styles = ui.NewStyles()

// View implements tea.Model
func (m Model) View() string {
	if m.err != nil {
		return m.renderError()
	}

	if m.width == 0 {
		return "Loading..."
	}

	// Show help screen if active
	if m.showHelp {
		return m.renderHelp()
	}

	// Check for welcome screen
	if !m.HasWorkspaces() {
		return m.renderWelcome()
	}

	var b strings.Builder

	// Calculate dimensions
	contentHeight := m.height - 2 // Reserve for status bar
	if m.mode == input.ModeSearch {
		contentHeight-- // Reserve for search bar
	}

	// Render two panes
	panes := m.renderPanes(contentHeight)
	b.WriteString(panes)
	b.WriteString("\n")

	// Search bar (only for search mode)
	if m.mode == input.ModeSearch {
		inputBar := m.renderInputBar()
		b.WriteString(inputBar)
		b.WriteString("\n")
	}

	// Status bar
	statusBar := m.renderStatusBar()
	b.WriteString(statusBar)

	return b.String()
}

// renderError renders the error screen
func (m Model) renderError() string {
	errorStyle := lipgloss.NewStyle().
		Foreground(ui.ColorError).
		Bold(true).
		Padding(2, 4)

	return errorStyle.Render(fmt.Sprintf("Error: %v\n\nPress 'q' to quit.", m.err))
}

// renderWelcome renders the welcome screen
func (m Model) renderWelcome() string {
	// Center box
	boxWidth := 40
	boxHeight := 10

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorPrimary).
		Padding(2, 4).
		Width(boxWidth).
		Align(lipgloss.Center)

	icon := lipgloss.NewStyle().
		Foreground(ui.ColorPrimary).
		Bold(true).
		Render("")

	title := lipgloss.NewStyle().
		Foreground(ui.ColorForeground).
		Bold(true).
		Render("Welcome to lazytodo!")

	hint := lipgloss.NewStyle().
		Foreground(ui.ColorMuted).
		Render("Press 'a' to create your first\nworkspace, or '?' for help")

	content := fmt.Sprintf("%s\n\n%s\n\n%s", icon, title, hint)
	boxContent := box.Render(content)

	// Center on screen
	horizontalPad := (m.width - boxWidth) / 2
	verticalPad := (m.height - boxHeight) / 2

	if horizontalPad < 0 {
		horizontalPad = 0
	}
	if verticalPad < 0 {
		verticalPad = 0
	}

	return lipgloss.NewStyle().
		Padding(verticalPad, horizontalPad).
		Render(boxContent) + "\n" + m.renderStatusBar()
}

// renderPanes renders the two-pane layout
func (m Model) renderPanes(height int) string {
	// Calculate widths (30:70 ratio)
	wsWidth := int(float64(m.width) * ui.WorkspacePaneRatio)
	todoWidth := m.width - wsWidth

	// Ensure minimum widths
	if wsWidth < ui.MinPaneWidth {
		wsWidth = ui.MinPaneWidth
		todoWidth = m.width - wsWidth
	}

	// Determine editing state
	isWsEditing := m.mode == input.ModeInsert && m.activePane == PaneWorkspace && m.inputAction == "edit"
	isWsAdding := m.mode == input.ModeInsert && m.activePane == PaneWorkspace && m.inputAction == "add"
	isTodoEditing := m.mode == input.ModeInsert && m.activePane == PaneTodo && m.inputAction == "edit"
	isTodoAdding := m.mode == input.ModeInsert && m.activePane == PaneTodo && (m.inputAction == "add" || m.inputAction == "add_child")

	// Render workspace pane
	wsPane := ui.WorkspacePaneModel{
		Workspaces:    m.workspaces,
		SelectedIndex: m.selectedWsIndex,
		IsActive:      m.activePane == PaneWorkspace,
		Width:         wsWidth,
		Height:        height,
		Styles:        styles,
		IsEditing:     isWsEditing,
		EditingIndex:  m.selectedWsIndex,
		EditBuffer:    m.inputBuffer,
		IsAdding:      isWsAdding,
	}

	// Render todo pane
	todoPane := ui.TodoPaneModel{
		Todos:         m.todos,
		SelectedIndex: m.selectedTodoIndex,
		IsActive:      m.activePane == PaneTodo,
		Width:         todoWidth,
		Height:        height,
		WorkspaceName: func() string {
			if ws := m.SelectedWorkspace(); ws != nil {
				return ws.Name
			}
			return ""
		}(),
		Styles:       styles,
		IsEditing:    isTodoEditing,
		EditingIndex: m.selectedTodoIndex,
		EditBuffer:   m.inputBuffer,
		IsAdding:     isTodoAdding,
	}

	// Join horizontally
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		wsPane.Render(),
		todoPane.Render(),
	)
}

// renderInputBar renders the input bar
func (m Model) renderInputBar() string {
	inputBar := ui.InputBarModel{
		Prompt: m.inputPrompt,
		Value:  m.inputBuffer,
		Width:  m.width,
		Styles: styles,
	}

	return inputBar.Render()
}

// renderStatusBar renders the status bar
func (m Model) renderStatusBar() string {
	statusBar := ui.StatusBarModel{
		Mode:          m.mode.String(),
		TodoCount:     len(m.todos),
		WorkspaceName: func() string {
			if ws := m.SelectedWorkspace(); ws != nil {
				return ws.Name
			}
			return ""
		}(),
		Notification: m.notification,
		IsError:      m.notificationErr,
		Width:        m.width,
		Styles:       styles,
	}

	return statusBar.Render()
}

// renderHelp renders the help screen
func (m Model) renderHelp() string {
	helpContent := `
 lazytodo - Keyboard Shortcuts

 NAVIGATION
   j/k        Move down/up
   h/l        Switch pane left/right
   Tab        Switch pane
   g/G        Jump to first/last item

 EDITING
   a          Add new item
   A          Add child item
   i          Edit current item
   dd         Delete current item
   Enter/Space Toggle todo status
   o          Toggle expand/collapse

 TREE OPERATIONS
   >          Indent (make child)
   <          Outdent (move up level)
   Ctrl+j/k   Move item down/up

 SEARCH
   /          Enter search mode
   Esc        Exit search mode

 OTHER
   ?          Toggle this help
   u          Undo
   q          Quit

 Press ? to close this help
`

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorPrimary).
		Padding(1, 2).
		Align(lipgloss.Left)

	content := box.Render(helpContent)

	// Center on screen
	boxWidth := lipgloss.Width(content)
	boxHeight := lipgloss.Height(content)

	horizontalPad := (m.width - boxWidth) / 2
	verticalPad := (m.height - boxHeight) / 2

	if horizontalPad < 0 {
		horizontalPad = 0
	}
	if verticalPad < 0 {
		verticalPad = 0
	}

	return lipgloss.NewStyle().
		Padding(verticalPad, horizontalPad).
		Render(content)
}
