package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/yuichikadota/lazytodo/internal/domain"
)

// WorkspacePaneModel holds the state for the workspace pane
type WorkspacePaneModel struct {
	Workspaces    []*domain.Workspace
	SelectedIndex int
	IsActive      bool
	Width         int
	Height        int
	Styles        Styles
	// Editing state
	IsEditing    bool
	EditingIndex int
	EditBuffer   string
	IsAdding     bool
}

// Render renders the workspace pane
func (m WorkspacePaneModel) Render() string {
	var paneStyle lipgloss.Style
	if m.IsActive {
		paneStyle = m.Styles.ActivePane
	} else {
		paneStyle = m.Styles.InactivePane
	}

	// Calculate content dimensions
	contentWidth := m.Width - 4  // Account for border and padding
	contentHeight := m.Height - 3 // Account for border and title

	// Build content
	var content strings.Builder

	// Title
	title := m.Styles.PaneTitle.Render("Workspaces")
	content.WriteString(title)
	content.WriteString("\n")

	if len(m.Workspaces) == 0 && !m.IsAdding {
		empty := m.Styles.EmptyState.Width(contentWidth).Render("No workspaces\nPress 'a' to create")
		content.WriteString(empty)
	} else {
		// Show add input at top if adding
		if m.IsAdding {
			addLine := m.renderAddInput(contentWidth)
			content.WriteString(addLine)
			if len(m.Workspaces) > 0 {
				content.WriteString("\n")
			}
		}

		// Render workspaces
		for i, ws := range m.Workspaces {
			if i >= contentHeight-1 {
				content.WriteString("...")
				break
			}

			var line string
			if m.IsEditing && i == m.EditingIndex {
				line = m.renderEditingItem(ws, contentWidth)
			} else {
				line = m.renderWorkspaceItem(ws, i == m.SelectedIndex, contentWidth)
			}
			content.WriteString(line)
			if i < len(m.Workspaces)-1 {
				content.WriteString("\n")
			}
		}
	}

	// Apply pane style
	return paneStyle.
		Width(m.Width).
		Height(m.Height).
		Render(content.String())
}

// renderWorkspaceItem renders a single workspace item
func (m WorkspacePaneModel) renderWorkspaceItem(ws *domain.Workspace, selected bool, width int) string {
	// Icon
	icon := IconFolderOpen
	if !ws.IsExpanded {
		icon = IconFolderClosed
	}
	if ws.IsSystem() {
		icon = IconArchive
	}

	// Color based on depth and type
	var iconStyle lipgloss.Style
	if ws.IsSystem() {
		iconStyle = lipgloss.NewStyle().Foreground(ColorIconArchive)
	} else if ws.Depth == 0 {
		iconStyle = m.Styles.WorkspaceRoot
	} else {
		iconStyle = m.Styles.WorkspaceChild
	}

	// Indentation
	indent := strings.Repeat("  ", ws.Depth)

	// Tree guide
	treeGuide := ""
	if ws.Depth > 0 {
		treeGuide = m.Styles.TreeGuide.Render("├─ ")
	}

	// Build line
	prefix := " "
	if selected && m.IsActive {
		prefix = ">"
	}

	iconRendered := iconStyle.Render(icon)
	name := ws.Name

	// Truncate if too long
	maxNameLen := width - len(indent) - len(treeGuide) - 5
	if len(name) > maxNameLen && maxNameLen > 3 {
		name = name[:maxNameLen-3] + "..."
	}

	line := fmt.Sprintf("%s%s%s%s %s", prefix, indent, treeGuide, iconRendered, name)

	// Apply selection style
	if selected && m.IsActive {
		return m.Styles.SelectedItem.Render(line)
	}

	return m.Styles.UnselectedItem.Render(line)
}

// renderEditingItem renders a workspace item in editing mode
func (m WorkspacePaneModel) renderEditingItem(ws *domain.Workspace, width int) string {
	// Icon
	icon := IconFolderOpen
	iconStyle := m.Styles.WorkspaceRoot

	// Indentation
	indent := strings.Repeat("  ", ws.Depth)

	// Tree guide
	treeGuide := ""
	if ws.Depth > 0 {
		treeGuide = m.Styles.TreeGuide.Render("├─ ")
	}

	iconRendered := iconStyle.Render(icon)

	// Show edit buffer with cursor
	editText := m.EditBuffer + "_"

	line := fmt.Sprintf(">%s%s%s %s", indent, treeGuide, iconRendered, editText)
	return m.Styles.EditingItem.Render(line)
}

// renderAddInput renders the add input line
func (m WorkspacePaneModel) renderAddInput(width int) string {
	icon := IconFolderOpen
	iconStyle := m.Styles.WorkspaceRoot
	iconRendered := iconStyle.Render(icon)

	// Show edit buffer with cursor
	editText := m.EditBuffer + "_"

	line := fmt.Sprintf(">%s %s", iconRendered, editText)
	return m.Styles.EditingItem.Render(line)
}
