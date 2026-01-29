package app

import (
	"fmt"
	"strings"

	"github.com/yuichikadota/lazytodo/internal/input"
)

// View implements tea.Model
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress 'q' to quit.", m.err)
	}

	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Main content area
	contentHeight := m.height - 2 // Reserve 2 lines for status bar and input

	// For now, simple view - will be replaced with proper 2-pane layout
	b.WriteString(m.renderContent(contentHeight))
	b.WriteString("\n")

	// Input bar (if in insert/search mode)
	if m.mode == input.ModeInsert || m.mode == input.ModeSearch {
		b.WriteString(m.renderInputBar())
	} else {
		b.WriteString("\n")
	}

	// Status bar
	b.WriteString(m.renderStatusBar())

	return b.String()
}

// renderContent renders the main content area
func (m Model) renderContent(height int) string {
	var b strings.Builder

	// Welcome screen if no workspaces
	if !m.HasWorkspaces() {
		return m.renderWelcome(height)
	}

	// Simple text-based view for now
	// Will be replaced with proper 2-pane layout in Task #7
	b.WriteString("Workspaces:\n")
	for i, ws := range m.workspaces {
		prefix := "  "
		if i == m.selectedWsIndex && m.activePane == PaneWorkspace {
			prefix = "> "
		}
		b.WriteString(fmt.Sprintf("%s%s\n", prefix, ws.Name))
	}

	b.WriteString("\nTodos:\n")
	if m.SelectedWorkspace() != nil {
		if len(m.todos) == 0 {
			b.WriteString("  No todos yet. Press 'a' to add one.\n")
		} else {
			for i, todo := range m.todos {
				prefix := "  "
				if i == m.selectedTodoIndex && m.activePane == PaneTodo {
					prefix = "> "
				}
				status := "[ ]"
				if todo.IsCompleted() {
					status = "[✓]"
				}
				b.WriteString(fmt.Sprintf("%s%s %s\n", prefix, status, todo.Description))
			}
		}
	}

	return b.String()
}

// renderWelcome renders the welcome screen
func (m Model) renderWelcome(height int) string {
	var b strings.Builder

	// Center the welcome message
	padding := (height - 6) / 2
	for i := 0; i < padding; i++ {
		b.WriteString("\n")
	}

	// Welcome message
	lines := []string{
		"",
		"       Welcome to lazytodo!",
		"",
		"  Press 'A' to create your first",
		"  workspace, or '?' for help",
		"",
	}

	for _, line := range lines {
		b.WriteString(fmt.Sprintf("%s\n", line))
	}

	return b.String()
}

// renderInputBar renders the input bar
func (m Model) renderInputBar() string {
	return fmt.Sprintf("%s%s_", m.inputPrompt, m.inputBuffer)
}

// renderStatusBar renders the status bar
func (m Model) renderStatusBar() string {
	// Mode indicator
	mode := fmt.Sprintf("[%s]", m.mode.String())

	// Item count
	var itemCount string
	if m.SelectedWorkspace() != nil {
		itemCount = fmt.Sprintf("%d todos", len(m.todos))
	}

	// Workspace name
	var wsName string
	if ws := m.SelectedWorkspace(); ws != nil {
		wsName = ws.Name
	}

	// Notification
	notification := m.notification

	// Build status bar
	parts := []string{mode}
	if itemCount != "" {
		parts = append(parts, itemCount)
	}
	if wsName != "" {
		parts = append(parts, wsName)
	}
	if notification != "" {
		parts = append(parts, notification)
	}

	return strings.Join(parts, " | ")
}
