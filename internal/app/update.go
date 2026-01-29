package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yuichikadota/lazytodo/internal/input"
)

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle errors
	if m.err != nil {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case workspacesLoadedMsg:
		m.workspaces = msg.workspaces
		if len(m.workspaces) > 0 {
			return m, m.loadTodos()
		}
		return m, nil

	case todosLoadedMsg:
		m.todos = msg.todos
		m.selectedTodoIndex = 0
		return m, nil

	case errMsg:
		m.notification = msg.err.Error()
		m.notificationErr = true
		return m, nil

	case notificationMsg:
		m.notification = msg.message
		m.notificationErr = msg.isError
		if !msg.isError {
			return m, clearNotificationAfter(2 * time.Second)
		}
		return m, nil

	case clearNotificationMsg:
		if !m.notificationErr {
			m.notification = ""
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	return m, nil
}

// handleKeyMsg handles key messages based on current mode
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys (work in any mode)
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	}

	// Mode-specific handling
	switch m.mode {
	case input.ModeNormal:
		return m.handleNormalMode(msg)
	case input.ModeInsert:
		return m.handleInsertMode(msg)
	case input.ModeSearch:
		return m.handleSearchMode(msg)
	case input.ModeSort:
		return m.handleSortMode(msg)
	}

	return m, nil
}

// handleNormalMode handles keys in normal mode
func (m Model) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Handle pending delete (dd)
	if m.pendingDelete {
		m.pendingDelete = false
		if key == "d" {
			// Execute delete
			// TODO: Implement delete
			m.notification = "Deleted"
			m.notificationErr = false
			return m, clearNotificationAfter(2 * time.Second)
		}
		// Cancel delete
		return m, nil
	}

	switch key {
	case "q":
		return m, tea.Quit

	// Navigation
	case "j", "down":
		return m.moveDown(), nil
	case "k", "up":
		return m.moveUp(), nil
	case "h":
		m.activePane = PaneWorkspace
		return m, nil
	case "l":
		m.activePane = PaneTodo
		return m, nil
	case "tab", "shift+tab":
		if m.activePane == PaneWorkspace {
			m.activePane = PaneTodo
		} else {
			m.activePane = PaneWorkspace
		}
		return m, nil
	case "g":
		return m.moveToFirst(), nil
	case "G":
		return m.moveToLast(), nil

	// Actions
	case "i":
		// Edit current item
		m.mode = input.ModeInsert
		m.inputPrompt = "Edit: "
		if m.activePane == PaneWorkspace && m.SelectedWorkspace() != nil {
			m.inputBuffer = m.SelectedWorkspace().Name
		} else if m.activePane == PaneTodo && m.SelectedTodo() != nil {
			m.inputBuffer = m.SelectedTodo().Description
		}
		return m, nil
	case "a":
		// Add new item
		m.mode = input.ModeInsert
		m.inputPrompt = "Add: "
		m.inputBuffer = ""
		return m, nil
	case "A":
		// Add child item
		m.mode = input.ModeInsert
		m.inputPrompt = "Add child: "
		m.inputBuffer = ""
		return m, nil
	case "d":
		// Start delete sequence
		m.pendingDelete = true
		return m, nil
	case "enter", " ":
		// Toggle status (for todos)
		if m.activePane == PaneTodo {
			// TODO: Implement toggle
		}
		return m, nil
	case "o":
		// Toggle expand/collapse
		// TODO: Implement toggle
		return m, nil

	// Mode switches
	case "/":
		m.mode = input.ModeSearch
		m.inputPrompt = "/"
		m.inputBuffer = ""
		return m, nil
	case "s":
		m.mode = input.ModeSort
		return m, nil
	case "?":
		// Show help
		// TODO: Implement help
		return m, nil
	}

	return m, nil
}

// handleInsertMode handles keys in insert mode
func (m Model) handleInsertMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = input.ModeNormal
		m.inputBuffer = ""
		m.inputPrompt = ""
		return m, nil
	case "enter":
		// Confirm input
		// TODO: Implement save
		m.mode = input.ModeNormal
		m.notification = "Saved"
		m.notificationErr = false
		m.inputBuffer = ""
		m.inputPrompt = ""
		return m, clearNotificationAfter(2 * time.Second)
	case "backspace":
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
		}
		return m, nil
	default:
		// Add character to buffer
		if len(msg.String()) == 1 {
			m.inputBuffer += msg.String()
		}
		return m, nil
	}
}

// handleSearchMode handles keys in search mode
func (m Model) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = input.ModeNormal
		m.inputBuffer = ""
		m.inputPrompt = ""
		return m, nil
	case "enter":
		// Execute search
		// TODO: Implement search
		m.mode = input.ModeNormal
		return m, nil
	case "backspace":
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
		}
		return m, nil
	default:
		// Add character to buffer
		if len(msg.String()) == 1 {
			m.inputBuffer += msg.String()
		}
		return m, nil
	}
}

// handleSortMode handles keys in sort mode
func (m Model) handleSortMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = input.ModeNormal
		return m, nil
	case "n":
		// Sort by name
		// TODO: Implement sort
		m.mode = input.ModeNormal
		return m, nil
	case "d":
		// Sort by date
		// TODO: Implement sort
		m.mode = input.ModeNormal
		return m, nil
	case "u":
		// Sort by urgency
		// TODO: Implement sort
		m.mode = input.ModeNormal
		return m, nil
	case "s":
		// Sort by status
		// TODO: Implement sort
		m.mode = input.ModeNormal
		return m, nil
	}

	return m, nil
}

// Navigation helpers

func (m Model) moveDown() Model {
	if m.activePane == PaneWorkspace {
		if m.selectedWsIndex < len(m.workspaces)-1 {
			m.selectedWsIndex++
		}
	} else {
		if m.selectedTodoIndex < len(m.todos)-1 {
			m.selectedTodoIndex++
		}
	}
	return m
}

func (m Model) moveUp() Model {
	if m.activePane == PaneWorkspace {
		if m.selectedWsIndex > 0 {
			m.selectedWsIndex--
		}
	} else {
		if m.selectedTodoIndex > 0 {
			m.selectedTodoIndex--
		}
	}
	return m
}

func (m Model) moveToFirst() Model {
	if m.activePane == PaneWorkspace {
		m.selectedWsIndex = 0
	} else {
		m.selectedTodoIndex = 0
	}
	return m
}

func (m Model) moveToLast() Model {
	if m.activePane == PaneWorkspace {
		if len(m.workspaces) > 0 {
			m.selectedWsIndex = len(m.workspaces) - 1
		}
	} else {
		if len(m.todos) > 0 {
			m.selectedTodoIndex = len(m.todos) - 1
		}
	}
	return m
}

// clearNotificationAfter returns a command to clear notification after duration
func clearNotificationAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return clearNotificationMsg{}
	})
}
