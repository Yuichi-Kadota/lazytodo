package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/yuichikadota/lazytodo/internal/domain"
)

// TodoPaneModel holds the state for the todo pane
type TodoPaneModel struct {
	Todos         []*domain.Todo
	SelectedIndex int
	IsActive      bool
	Width         int
	Height        int
	WorkspaceName string
	Styles        Styles
	// Editing state
	IsEditing    bool
	EditingIndex int
	EditBuffer   string
	IsAdding     bool
}

// Render renders the todo pane
func (m TodoPaneModel) Render() string {
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
	title := m.Styles.PaneTitle.Render("Todos")
	content.WriteString(title)
	content.WriteString("\n")

	if len(m.Todos) == 0 && !m.IsAdding {
		empty := m.Styles.EmptyState.Width(contentWidth).Render("No todos yet.\nPress 'a' to add one.")
		content.WriteString(empty)
	} else {
		lineCount := 0

		// If no todos but adding, show add input
		if len(m.Todos) == 0 && m.IsAdding {
			addLine := m.renderAddInput(contentWidth)
			content.WriteString(addLine)
			lineCount++
		}

		// Render todos with add input after selected item
		for i, todo := range m.Todos {
			if lineCount >= contentHeight-1 {
				content.WriteString("...")
				break
			}

			var line string
			if m.IsEditing && i == m.EditingIndex {
				line = m.renderEditingItem(todo, contentWidth)
			} else {
				line = m.renderTodoItem(todo, i == m.SelectedIndex, contentWidth)
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

			if i < len(m.Todos)-1 || (m.IsAdding && i != m.SelectedIndex) {
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

// renderTodoItem renders a single todo item
func (m TodoPaneModel) renderTodoItem(todo *domain.Todo, selected bool, width int) string {
	// Icon based on status and urgency
	var icon string

	if todo.IsCompleted() {
		icon = IconTodoDone
	} else if todo.Urgency >= domain.UrgencyHigh {
		icon = IconTodoUrgent
	} else {
		icon = IconTodo
	}

	// Indentation for nested todos
	indent := strings.Repeat("  ", todo.Depth)
	treeGuide := ""
	if todo.Depth > 0 {
		treeGuide = "├─ "
	}

	// Build line without styles first
	prefix := " "
	if selected && m.IsActive {
		prefix = ">"
	}

	desc := todo.Description

	// Due date (plain text)
	dueDateStr := ""
	if todo.DueDate != nil {
		if todo.IsOverdue() {
			dueDateStr = " [Overdue]"
		} else if todo.IsDueToday() {
			dueDateStr = " [Today]"
		} else {
			dueDateStr = fmt.Sprintf(" [%s]", todo.DueDate.Format("Jan 2"))
		}
	}

	// Tags (plain text)
	tagsStr := ""
	tags := todo.ExtractTags()
	for _, tag := range tags {
		tagsStr += " @" + tag
	}

	// Truncate if too long
	maxDescLen := width - len(indent) - len(treeGuide) - len(dueDateStr) - len(tagsStr) - 6
	if len(desc) > maxDescLen && maxDescLen > 3 {
		desc = desc[:maxDescLen-3] + "..."
	}

	line := fmt.Sprintf("%s%s%s%s %s%s%s", prefix, indent, treeGuide, icon, desc, tagsStr, dueDateStr)

	// Apply single style at the end
	if selected && m.IsActive {
		return m.Styles.SelectedItem.Render(line)
	}

	if todo.IsCompleted() {
		return m.Styles.CompletedItem.Render(line)
	}

	return m.Styles.UnselectedItem.Render(line)
}

// renderEditingItem renders a todo item in editing mode
func (m TodoPaneModel) renderEditingItem(todo *domain.Todo, width int) string {
	icon := IconTodo

	// Indentation
	indent := strings.Repeat("  ", todo.Depth)
	treeGuide := ""
	if todo.Depth > 0 {
		treeGuide = "├─ "
	}

	// Show edit buffer with cursor
	editText := m.EditBuffer + "_"

	line := fmt.Sprintf(">%s%s%s %s", indent, treeGuide, icon, editText)
	return m.Styles.EditingItem.Render(line)
}

// renderAddInput renders the add input line
func (m TodoPaneModel) renderAddInput(width int) string {
	icon := IconTodo

	// Show edit buffer with cursor
	editText := m.EditBuffer + "_"

	line := fmt.Sprintf(">%s %s", icon, editText)
	return m.Styles.EditingItem.Render(line)
}

// formatDueDate formats a due date for display
func formatDueDate(t time.Time) string {
	now := time.Now()
	if t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day() {
		return "Today"
	}
	if t.Before(now) {
		return "Overdue"
	}
	return t.Format("Jan 2")
}
