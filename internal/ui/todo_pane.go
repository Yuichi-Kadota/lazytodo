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

		// Show add input at top if adding
		if m.IsAdding {
			addLine := m.renderAddInput(contentWidth)
			content.WriteString(addLine)
			content.WriteString("\n")
			lineCount++
		}

		// Render todos (all together, maintaining original order)
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
			if i < len(m.Todos)-1 {
				content.WriteString("\n")
			}
			lineCount++
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
	var iconStyle lipgloss.Style

	if todo.IsCompleted() {
		icon = IconTodoDone
		iconStyle = lipgloss.NewStyle().Foreground(ColorTodoComplete)
	} else if todo.Urgency >= domain.UrgencyHigh {
		icon = IconTodoUrgent
		iconStyle = m.Styles.TodoUrgent
	} else {
		icon = IconTodo
		iconStyle = m.Styles.TodoPending
	}

	// Indentation for nested todos
	indent := strings.Repeat("  ", todo.Depth)
	treeGuide := ""
	if todo.Depth > 0 {
		treeGuide = m.Styles.TreeGuide.Render("├─ ")
	}

	// Due date
	dueDateStr := ""
	if todo.DueDate != nil {
		if todo.IsOverdue() {
			dueDateStr = lipgloss.NewStyle().Foreground(ColorOverdue).Render(" [Overdue]")
		} else if todo.IsDueToday() {
			dueDateStr = lipgloss.NewStyle().Foreground(ColorDueToday).Render(" [Today]")
		} else {
			dueDateStr = lipgloss.NewStyle().Foreground(ColorMuted).
				Render(fmt.Sprintf(" [%s]", todo.DueDate.Format("Jan 2")))
		}
	}

	// Tags
	tagsStr := ""
	tags := todo.ExtractTags()
	if len(tags) > 0 {
		tagStyle := lipgloss.NewStyle().Foreground(ColorPrimary)
		for _, tag := range tags {
			tagsStr += tagStyle.Render(" @"+tag)
		}
	}

	// Build line
	prefix := " "
	if selected && m.IsActive {
		prefix = ">"
	}

	iconRendered := iconStyle.Render(icon)
	desc := todo.Description

	// Truncate if too long
	maxDescLen := width - len(indent) - len(treeGuide) - len(dueDateStr) - len(tagsStr) - 6
	if len(desc) > maxDescLen && maxDescLen > 3 {
		desc = desc[:maxDescLen-3] + "..."
	}

	line := fmt.Sprintf("%s%s%s%s %s%s%s", prefix, indent, treeGuide, iconRendered, desc, tagsStr, dueDateStr)

	// Apply style
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
	iconStyle := m.Styles.TodoPending
	iconRendered := iconStyle.Render(icon)

	// Indentation
	indent := strings.Repeat("  ", todo.Depth)
	treeGuide := ""
	if todo.Depth > 0 {
		treeGuide = m.Styles.TreeGuide.Render("├─ ")
	}

	// Show edit buffer with cursor
	editText := m.EditBuffer + "_"

	line := fmt.Sprintf(">%s%s%s %s", indent, treeGuide, iconRendered, editText)
	return m.Styles.EditingItem.Render(line)
}

// renderAddInput renders the add input line
func (m TodoPaneModel) renderAddInput(width int) string {
	icon := IconTodo
	iconStyle := m.Styles.TodoPending
	iconRendered := iconStyle.Render(icon)

	// Show edit buffer with cursor
	editText := m.EditBuffer + "_"

	line := fmt.Sprintf(">%s %s", iconRendered, editText)
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
