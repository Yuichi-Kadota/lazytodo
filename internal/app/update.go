package app

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yuichikadota/lazytodo/internal/domain"
	"github.com/yuichikadota/lazytodo/internal/input"
	"github.com/yuichikadota/lazytodo/internal/wal"
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

	// Todo CRUD responses
	case todoCreatedMsg:
		m.notification = "Todo created"
		m.notificationErr = false
		return m, tea.Batch(m.loadTodos(), clearNotificationAfter(2*time.Second))

	case todoUpdatedMsg:
		m.notification = "Todo updated"
		m.notificationErr = false
		return m, tea.Batch(m.loadTodos(), clearNotificationAfter(2*time.Second))

	case todoDeletedMsg:
		m.notification = "Todo deleted"
		m.notificationErr = false
		// Adjust selection if needed
		if m.selectedTodoIndex > 0 {
			m.selectedTodoIndex--
		}
		return m, tea.Batch(m.loadTodos(), clearNotificationAfter(2*time.Second))

	// Workspace CRUD responses
	case workspaceCreatedMsg:
		m.notification = "Workspace created"
		m.notificationErr = false
		return m, tea.Batch(m.loadWorkspaces(), clearNotificationAfter(2*time.Second))

	case workspaceUpdatedMsg:
		m.notification = "Workspace updated"
		m.notificationErr = false
		return m, tea.Batch(m.loadWorkspaces(), clearNotificationAfter(2*time.Second))

	case workspaceDeletedMsg:
		m.notification = "Workspace deleted"
		m.notificationErr = false
		// Adjust selection if needed
		if m.selectedWsIndex > 0 {
			m.selectedWsIndex--
		}
		return m, tea.Batch(m.loadWorkspaces(), clearNotificationAfter(2*time.Second))

	case undoMsg:
		m.notification = "Undone"
		m.notificationErr = false
		// Reload all data
		return m, tea.Batch(m.loadWorkspaces(), clearNotificationAfter(2*time.Second))

	case searchResultsMsg:
		m.searchResults = msg.todos
		m.todos = msg.todos
		m.selectedTodoIndex = 0
		return m, nil

	case todosSortedMsg:
		m.todos = msg.todos
		m.selectedTodoIndex = 0
		m.notification = "Sorted by " + msg.sortBy
		m.notificationErr = false
		return m, clearNotificationAfter(2 * time.Second)

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
			if m.activePane == PaneTodo && m.SelectedTodo() != nil {
				return m, m.deleteTodo()
			} else if m.activePane == PaneWorkspace && m.SelectedWorkspace() != nil {
				return m, m.deleteWorkspace()
			}
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
		if m.activePane == PaneTodo && m.SelectedTodo() != nil {
			m.mode = input.ModeInsert
			m.inputPrompt = "Edit: "
			m.inputAction = "edit"
			m.inputBuffer = m.SelectedTodo().Description
		} else if m.activePane == PaneWorkspace && m.SelectedWorkspace() != nil {
			m.mode = input.ModeInsert
			m.inputPrompt = "Edit: "
			m.inputAction = "edit"
			m.inputBuffer = m.SelectedWorkspace().Name
		}
		return m, nil
	case "a":
		// Add new item
		m.mode = input.ModeInsert
		m.inputPrompt = "Add: "
		m.inputAction = "add"
		m.inputBuffer = ""
		return m, nil
	case "A":
		// Add child item (only for todos)
		if m.activePane == PaneTodo && m.SelectedTodo() != nil {
			m.mode = input.ModeInsert
			m.inputPrompt = "Add child: "
			m.inputAction = "add_child"
			m.inputBuffer = ""
		}
		return m, nil
	case "d":
		// Start delete sequence
		m.pendingDelete = true
		return m, nil
	case "enter", " ":
		// Toggle status (for todos)
		if m.activePane == PaneTodo && m.SelectedTodo() != nil {
			return m, m.toggleTodoStatus()
		}
		return m, nil
	case "o":
		// Toggle expand/collapse
		if m.activePane == PaneWorkspace && m.SelectedWorkspace() != nil {
			return m, m.toggleExpand()
		}
		return m, nil
	case ">":
		// Indent (make child of sibling above)
		if m.activePane == PaneTodo && m.SelectedTodo() != nil {
			return m, m.indentTodo()
		} else if m.activePane == PaneWorkspace && m.SelectedWorkspace() != nil {
			return m, m.indentWorkspace()
		}
		return m, nil
	case "<":
		// Outdent (move up one level)
		if m.activePane == PaneTodo && m.SelectedTodo() != nil {
			return m, m.outdentTodo()
		} else if m.activePane == PaneWorkspace && m.SelectedWorkspace() != nil {
			return m, m.outdentWorkspace()
		}
		return m, nil
	case "ctrl+j":
		// Move item down
		if m.activePane == PaneTodo && m.SelectedTodo() != nil {
			return m, m.moveTodoDown()
		} else if m.activePane == PaneWorkspace && m.SelectedWorkspace() != nil {
			return m, m.moveWorkspaceDown()
		}
		return m, nil
	case "ctrl+k":
		// Move item up
		if m.activePane == PaneTodo && m.SelectedTodo() != nil {
			return m, m.moveTodoUp()
		} else if m.activePane == PaneWorkspace && m.SelectedWorkspace() != nil {
			return m, m.moveWorkspaceUp()
		}
		return m, nil

	// Undo/Redo
	case "u":
		return m, m.undo()
	case "ctrl+r":
		return m, m.redo()

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
		m.inputAction = ""
		return m, nil
	case "enter":
		// Confirm input
		if m.inputBuffer == "" {
			m.mode = input.ModeNormal
			m.inputBuffer = ""
			m.inputPrompt = ""
			m.inputAction = ""
			return m, nil
		}

		var cmd tea.Cmd
		switch m.inputAction {
		case "add":
			if m.activePane == PaneTodo {
				cmd = m.createTodo(m.inputBuffer, "")
			} else {
				cmd = m.createWorkspace(m.inputBuffer, "")
			}
		case "add_child":
			if m.activePane == PaneTodo && m.SelectedTodo() != nil {
				cmd = m.createTodo(m.inputBuffer, m.SelectedTodo().ID)
			}
		case "edit":
			if m.activePane == PaneTodo && m.SelectedTodo() != nil {
				cmd = m.updateTodo(m.inputBuffer)
			} else if m.activePane == PaneWorkspace && m.SelectedWorkspace() != nil {
				cmd = m.updateWorkspace(m.inputBuffer)
			}
		}

		m.mode = input.ModeNormal
		m.inputBuffer = ""
		m.inputPrompt = ""
		m.inputAction = ""
		return m, cmd
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
		m.isSearching = false
		m.searchResults = nil
		// Restore original todos
		return m, m.loadTodos()
	case "enter":
		// Confirm search and stay on results
		m.mode = input.ModeNormal
		m.inputPrompt = ""
		if len(m.searchResults) > 0 {
			m.todos = m.searchResults
			m.selectedTodoIndex = 0
		}
		m.isSearching = false
		return m, nil
	case "backspace":
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
		}
		// Trigger incremental search
		if m.inputBuffer != "" {
			return m, m.searchTodos(m.inputBuffer)
		}
		return m, nil
	case "j", "down":
		// Navigate search results
		if m.selectedTodoIndex < len(m.searchResults)-1 {
			m.selectedTodoIndex++
		}
		return m, nil
	case "k", "up":
		if m.selectedTodoIndex > 0 {
			m.selectedTodoIndex--
		}
		return m, nil
	default:
		// Add character to buffer
		if len(msg.String()) == 1 {
			m.inputBuffer += msg.String()
			// Trigger incremental search
			return m, m.searchTodos(m.inputBuffer)
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
		// Sort by name (description)
		m.mode = input.ModeNormal
		return m, m.sortTodos("name")
	case "d":
		// Sort by date
		m.mode = input.ModeNormal
		return m, m.sortTodos("date")
	case "u":
		// Sort by urgency
		m.mode = input.ModeNormal
		return m, m.sortTodos("urgency")
	case "s":
		// Sort by status
		m.mode = input.ModeNormal
		return m, m.sortTodos("status")
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

// Todo CRUD commands

func (m Model) createTodo(description string, parentID string) tea.Cmd {
	return func() tea.Msg {
		ws := m.SelectedWorkspace()
		if ws == nil {
			return errMsg{domain.ErrNotFound}
		}

		todo := &domain.Todo{
			WorkspaceID: ws.ID,
			Description: description,
			Status:      domain.StatusPending,
			Urgency:     domain.UrgencyMedium,
			ParentID:    parentID,
		}

		if err := m.todoRepo.Create(context.Background(), todo); err != nil {
			return errMsg{err}
		}

		return todoCreatedMsg{todo: todo}
	}
}

func (m Model) updateTodo(description string) tea.Cmd {
	return func() tea.Msg {
		todo := m.SelectedTodo()
		if todo == nil {
			return errMsg{domain.ErrNotFound}
		}

		todo.Description = description
		if err := m.todoRepo.Update(context.Background(), todo); err != nil {
			return errMsg{err}
		}

		return todoUpdatedMsg{todo: todo}
	}
}

func (m Model) deleteTodo() tea.Cmd {
	return func() tea.Msg {
		todo := m.SelectedTodo()
		if todo == nil {
			return errMsg{domain.ErrNotFound}
		}

		if err := m.todoRepo.Delete(context.Background(), todo.ID); err != nil {
			return errMsg{err}
		}

		return todoDeletedMsg{id: todo.ID}
	}
}

func (m Model) toggleTodoStatus() tea.Cmd {
	return func() tea.Msg {
		todo := m.SelectedTodo()
		if todo == nil {
			return errMsg{domain.ErrNotFound}
		}

		if todo.Status == domain.StatusPending {
			todo.Status = domain.StatusCompleted
			now := time.Now()
			todo.CompletedAt = &now
		} else {
			todo.Status = domain.StatusPending
			todo.CompletedAt = nil
		}

		if err := m.todoRepo.Update(context.Background(), todo); err != nil {
			return errMsg{err}
		}

		return todoUpdatedMsg{todo: todo}
	}
}

// Workspace CRUD commands

func (m Model) createWorkspace(name string, parentID string) tea.Cmd {
	return func() tea.Msg {
		ws := &domain.Workspace{
			Name:     name,
			ParentID: parentID,
		}

		if err := m.workspaceRepo.Create(context.Background(), ws); err != nil {
			return errMsg{err}
		}

		return workspaceCreatedMsg{workspace: ws}
	}
}

func (m Model) updateWorkspace(name string) tea.Cmd {
	return func() tea.Msg {
		ws := m.SelectedWorkspace()
		if ws == nil {
			return errMsg{domain.ErrNotFound}
		}

		ws.Name = name
		if err := m.workspaceRepo.Update(context.Background(), ws); err != nil {
			return errMsg{err}
		}

		return workspaceUpdatedMsg{workspace: ws}
	}
}

func (m Model) deleteWorkspace() tea.Cmd {
	return func() tea.Msg {
		ws := m.SelectedWorkspace()
		if ws == nil {
			return errMsg{domain.ErrNotFound}
		}

		if err := m.workspaceRepo.Delete(context.Background(), ws.ID); err != nil {
			return errMsg{err}
		}

		return workspaceDeletedMsg{id: ws.ID}
	}
}

// Indent/Outdent commands

func (m Model) indentTodo() tea.Cmd {
	return func() tea.Msg {
		todo := m.SelectedTodo()
		if todo == nil {
			return errMsg{domain.ErrNotFound}
		}

		// Find sibling above to become new parent
		var newParentID string
		for _, t := range m.todos {
			if t.ID == todo.ID {
				break
			}
			if t.Depth == todo.Depth {
				newParentID = t.ID
			}
		}

		if newParentID == "" {
			return notificationMsg{message: "Cannot indent: no sibling above", isError: true}
		}

		if err := m.todoRepo.Move(context.Background(), todo.ID, newParentID, ""); err != nil {
			return errMsg{err}
		}

		return todoUpdatedMsg{todo: todo}
	}
}

func (m Model) outdentTodo() tea.Cmd {
	return func() tea.Msg {
		todo := m.SelectedTodo()
		if todo == nil || todo.ParentID == "" {
			return notificationMsg{message: "Cannot outdent: no parent", isError: true}
		}

		// Get grandparent ID
		parent, err := m.todoRepo.GetByID(context.Background(), todo.ParentID)
		if err != nil {
			return errMsg{err}
		}

		if err := m.todoRepo.Move(context.Background(), todo.ID, parent.ParentID, ""); err != nil {
			return errMsg{err}
		}

		return todoUpdatedMsg{todo: todo}
	}
}

func (m Model) indentWorkspace() tea.Cmd {
	return func() tea.Msg {
		ws := m.SelectedWorkspace()
		if ws == nil {
			return errMsg{domain.ErrNotFound}
		}

		// Find sibling above to become new parent
		var newParentID string
		for _, w := range m.workspaces {
			if w.ID == ws.ID {
				break
			}
			if w.Depth == ws.Depth {
				newParentID = w.ID
			}
		}

		if newParentID == "" {
			return notificationMsg{message: "Cannot indent: no sibling above", isError: true}
		}

		if err := m.workspaceRepo.Move(context.Background(), ws.ID, newParentID); err != nil {
			return errMsg{err}
		}

		return workspaceUpdatedMsg{workspace: ws}
	}
}

func (m Model) outdentWorkspace() tea.Cmd {
	return func() tea.Msg {
		ws := m.SelectedWorkspace()
		if ws == nil || ws.ParentID == "" {
			return notificationMsg{message: "Cannot outdent: no parent", isError: true}
		}

		// Get grandparent ID
		parent, err := m.workspaceRepo.GetByID(context.Background(), ws.ParentID)
		if err != nil {
			return errMsg{err}
		}

		if err := m.workspaceRepo.Move(context.Background(), ws.ID, parent.ParentID); err != nil {
			return errMsg{err}
		}

		return workspaceUpdatedMsg{workspace: ws}
	}
}

// Reorder commands

func (m Model) moveTodoDown() tea.Cmd {
	return func() tea.Msg {
		todo := m.SelectedTodo()
		if todo == nil {
			return errMsg{domain.ErrNotFound}
		}

		newPosition := todo.Position + 1
		if err := m.todoRepo.Reorder(context.Background(), todo.ID, newPosition); err != nil {
			return errMsg{err}
		}

		return todoUpdatedMsg{todo: todo}
	}
}

func (m Model) moveTodoUp() tea.Cmd {
	return func() tea.Msg {
		todo := m.SelectedTodo()
		if todo == nil {
			return errMsg{domain.ErrNotFound}
		}

		if todo.Position <= 0 {
			return notificationMsg{message: "Already at top", isError: false}
		}

		newPosition := todo.Position - 1
		if err := m.todoRepo.Reorder(context.Background(), todo.ID, newPosition); err != nil {
			return errMsg{err}
		}

		return todoUpdatedMsg{todo: todo}
	}
}

func (m Model) moveWorkspaceDown() tea.Cmd {
	return func() tea.Msg {
		ws := m.SelectedWorkspace()
		if ws == nil {
			return errMsg{domain.ErrNotFound}
		}

		newPosition := ws.Position + 1
		if err := m.workspaceRepo.Reorder(context.Background(), ws.ID, newPosition); err != nil {
			return errMsg{err}
		}

		return workspaceUpdatedMsg{workspace: ws}
	}
}

func (m Model) moveWorkspaceUp() tea.Cmd {
	return func() tea.Msg {
		ws := m.SelectedWorkspace()
		if ws == nil {
			return errMsg{domain.ErrNotFound}
		}

		if ws.Position <= 0 {
			return notificationMsg{message: "Already at top", isError: false}
		}

		newPosition := ws.Position - 1
		if err := m.workspaceRepo.Reorder(context.Background(), ws.ID, newPosition); err != nil {
			return errMsg{err}
		}

		return workspaceUpdatedMsg{workspace: ws}
	}
}

// Toggle expand/collapse

func (m Model) toggleExpand() tea.Cmd {
	return func() tea.Msg {
		ws := m.SelectedWorkspace()
		if ws == nil {
			return errMsg{domain.ErrNotFound}
		}

		ws.IsExpanded = !ws.IsExpanded
		if err := m.workspaceRepo.Update(context.Background(), ws); err != nil {
			return errMsg{err}
		}

		return workspaceUpdatedMsg{workspace: ws}
	}
}

// Undo/Redo commands

func (m Model) undo() tea.Cmd {
	return func() tea.Msg {
		ops, err := m.wal.GetUndoOperations(1)
		if err != nil {
			return errMsg{err}
		}

		if len(ops) == 0 {
			return notificationMsg{message: "Nothing to undo", isError: false}
		}

		op := ops[0]
		if err := m.wal.MarkUndone(op.ID); err != nil {
			return errMsg{err}
		}

		return undoMsg{operation: op}
	}
}

func (m Model) redo() tea.Cmd {
	return func() tea.Msg {
		// For redo, we'd need to track undone operations
		// For now, just show a message
		return notificationMsg{message: "Redo not yet implemented", isError: false}
	}
}

// Search and sort commands

func (m Model) searchTodos(query string) tea.Cmd {
	return func() tea.Msg {
		if query == "" {
			return searchResultsMsg{todos: nil}
		}

		todos, err := m.todoRepo.Search(context.Background(), query, false)
		if err != nil {
			return errMsg{err}
		}

		return searchResultsMsg{todos: todos}
	}
}

func (m Model) sortTodos(sortBy string) tea.Cmd {
	return func() tea.Msg {
		if len(m.todos) == 0 {
			return todosSortedMsg{todos: m.todos, sortBy: sortBy}
		}

		// Create a copy to sort
		sorted := make([]*domain.Todo, len(m.todos))
		copy(sorted, m.todos)

		switch sortBy {
		case "name":
			sortTodosByName(sorted)
		case "date":
			sortTodosByDate(sorted)
		case "urgency":
			sortTodosByUrgency(sorted)
		case "status":
			sortTodosByStatus(sorted)
		}

		return todosSortedMsg{todos: sorted, sortBy: sortBy}
	}
}

func sortTodosByName(todos []*domain.Todo) {
	for i := 0; i < len(todos)-1; i++ {
		for j := i + 1; j < len(todos); j++ {
			if todos[i].Description > todos[j].Description {
				todos[i], todos[j] = todos[j], todos[i]
			}
		}
	}
}

func sortTodosByDate(todos []*domain.Todo) {
	for i := 0; i < len(todos)-1; i++ {
		for j := i + 1; j < len(todos); j++ {
			if todos[i].CreatedAt.After(todos[j].CreatedAt) {
				todos[i], todos[j] = todos[j], todos[i]
			}
		}
	}
}

func sortTodosByUrgency(todos []*domain.Todo) {
	for i := 0; i < len(todos)-1; i++ {
		for j := i + 1; j < len(todos); j++ {
			if todos[i].Urgency < todos[j].Urgency {
				todos[i], todos[j] = todos[j], todos[i]
			}
		}
	}
}

func sortTodosByStatus(todos []*domain.Todo) {
	for i := 0; i < len(todos)-1; i++ {
		for j := i + 1; j < len(todos); j++ {
			// Pending before completed
			if todos[i].Status > todos[j].Status {
				todos[i], todos[j] = todos[j], todos[i]
			}
		}
	}
}

// Message types for CRUD operations
type todoCreatedMsg struct{ todo *domain.Todo }
type todoUpdatedMsg struct{ todo *domain.Todo }
type todoDeletedMsg struct{ id string }
type workspaceCreatedMsg struct{ workspace *domain.Workspace }
type workspaceUpdatedMsg struct{ workspace *domain.Workspace }
type workspaceDeletedMsg struct{ id string }
type undoMsg struct{ operation *wal.Operation }
type searchResultsMsg struct{ todos []*domain.Todo }
type todosSortedMsg struct {
	todos  []*domain.Todo
	sortBy string
}
