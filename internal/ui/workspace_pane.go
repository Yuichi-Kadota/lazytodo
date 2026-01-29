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
		lineCount := 0

		// If no workspaces but adding, show add input
		if len(m.Workspaces) == 0 && m.IsAdding {
			addLine := m.renderAddInput(contentWidth)
			content.WriteString(addLine)
			lineCount++
		}

		// Render workspaces with add input after selected item
		for i, ws := range m.Workspaces {
			if lineCount >= contentHeight-1 {
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
			lineCount++

			// Show add input after selected item
			if m.IsAdding && i == m.SelectedIndex {
				content.WriteString("\n")
				addLine := m.renderAddInput(contentWidth)
				content.WriteString(addLine)
				lineCount++
			}

			if i < len(m.Workspaces)-1 || (m.IsAdding && i != m.SelectedIndex) {
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

	// Indentation
	indent := strings.Repeat("  ", ws.Depth)

	// Tree guide (plain text)
	treeGuide := ""
	if ws.Depth > 0 {
		treeGuide = "├─ "
	}

	// Build line without styles
	prefix := " "
	if selected && m.IsActive {
		prefix = ">"
	}

	name := ws.Name

	// Truncate if too long
	maxNameLen := width - len(indent) - len(treeGuide) - 5
	if len(name) > maxNameLen && maxNameLen > 3 {
		name = name[:maxNameLen-3] + "..."
	}

	line := fmt.Sprintf("%s%s%s%s %s", prefix, indent, treeGuide, icon, name)

	// Apply single style at the end
	if selected && m.IsActive {
		return m.Styles.SelectedItem.Render(line)
	}

	return m.Styles.UnselectedItem.Render(line)
}

// renderEditingItem renders a workspace item in editing mode
func (m WorkspacePaneModel) renderEditingItem(ws *domain.Workspace, width int) string {
	// Icon
	icon := IconFolderOpen

	// Indentation
	indent := strings.Repeat("  ", ws.Depth)

	// Tree guide (plain text)
	treeGuide := ""
	if ws.Depth > 0 {
		treeGuide = "├─ "
	}

	// Show edit buffer with cursor
	editText := m.EditBuffer + "_"

	line := fmt.Sprintf(">%s%s%s %s", indent, treeGuide, icon, editText)
	return m.Styles.EditingItem.Render(line)
}

// renderAddInput renders the add input line
func (m WorkspacePaneModel) renderAddInput(width int) string {
	icon := IconFolderOpen

	// Show edit buffer with cursor
	editText := m.EditBuffer + "_"

	line := fmt.Sprintf(">%s %s", icon, editText)
	return m.Styles.EditingItem.Render(line)
}
