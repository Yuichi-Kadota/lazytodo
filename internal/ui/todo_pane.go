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

	if len(m.Todos) == 0 {
		empty := m.Styles.EmptyState.Width(contentWidth).Render("No todos yet.\nPress 'a' to add one.")
		content.WriteString(empty)
	} else {
		// Separate pending and completed
		var pending, completed []*domain.Todo
		for _, t := range m.Todos {
			if t.IsCompleted() {
				completed = append(completed, t)
			} else {
				pending = append(pending, t)
			}
		}

		lineCount := 0

		// Render pending todos
		for i, todo := range pending {
			if lineCount >= contentHeight-1 {
				content.WriteString("...")
				break
			}

			actualIndex := i
			line := m.renderTodoItem(todo, actualIndex == m.SelectedIndex, contentWidth)
			content.WriteString(line)
			content.WriteString("\n")
			lineCount++
		}

		// Separator if both pending and completed exist
		if len(pending) > 0 && len(completed) > 0 && lineCount < contentHeight-2 {
			sep := m.Styles.TreeGuide.Render(strings.Repeat("─", contentWidth-2))
			content.WriteString(sep)
			content.WriteString("\n")
			lineCount++
		}

		// Render completed todos
		for i, todo := range completed {
			if lineCount >= contentHeight-1 {
				content.WriteString("...")
				break
			}

			actualIndex := len(pending) + i
			line := m.renderTodoItem(todo, actualIndex == m.SelectedIndex, contentWidth)
			content.WriteString(line)
			if i < len(completed)-1 {
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

	// Build line
	prefix := " "
	if selected && m.IsActive {
		prefix = ">"
	}

	iconRendered := iconStyle.Render(icon)
	desc := todo.Description

	// Truncate if too long
	maxDescLen := width - len(indent) - len(treeGuide) - len(dueDateStr) - 6
	if len(desc) > maxDescLen && maxDescLen > 3 {
		desc = desc[:maxDescLen-3] + "..."
	}

	line := fmt.Sprintf("%s%s%s%s %s%s", prefix, indent, treeGuide, iconRendered, desc, dueDateStr)

	// Apply style
	if selected && m.IsActive {
		return m.Styles.SelectedItem.Render(line)
	}

	if todo.IsCompleted() {
		return m.Styles.CompletedItem.Render(line)
	}

	return m.Styles.UnselectedItem.Render(line)
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
