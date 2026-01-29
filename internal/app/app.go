package app

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yuichikadota/lazytodo/internal/domain"
	"github.com/yuichikadota/lazytodo/internal/input"
	"github.com/yuichikadota/lazytodo/internal/repository"
)

// Pane represents which pane is active
type Pane int

const (
	PaneWorkspace Pane = iota
	PaneTodo
)

// Model is the main application model
type Model struct {
	// Window dimensions
	width  int
	height int

	// Current mode and pane
	mode       input.Mode
	activePane Pane

	// Data
	workspaces       []*domain.Workspace
	todos            []*domain.Todo
	selectedWsIndex  int
	selectedTodoIndex int

	// Input state
	inputBuffer string
	inputPrompt string

	// Notification
	notification    string
	notificationErr bool

	// Dependencies
	db            *repository.DB
	workspaceRepo *repository.WorkspaceRepository
	todoRepo      *repository.TodoRepository

	// Pending delete (for dd confirmation)
	pendingDelete bool

	// Error state
	err error
}

// Config holds the application configuration
type Config struct {
	DBPath string
}

// New creates a new application model
func New(cfg Config) Model {
	m := Model{
		mode:       input.ModeNormal,
		activePane: PaneWorkspace,
	}

	// Initialize database
	dbPath := cfg.DBPath
	if dbPath == "" {
		dbPath = repository.DefaultDBPath()
	}

	db, err := repository.NewDB(dbPath)
	if err != nil {
		m.err = err
		return m
	}

	// Run migrations
	if err := db.Migrate(); err != nil {
		m.err = err
		return m
	}

	m.db = db
	m.workspaceRepo = repository.NewWorkspaceRepository(db)
	m.todoRepo = repository.NewTodoRepository(db)

	return m
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	if m.err != nil {
		return nil
	}
	return m.loadWorkspaces()
}

// loadWorkspaces returns a command to load workspaces
func (m Model) loadWorkspaces() tea.Cmd {
	return func() tea.Msg {
		workspaces, err := m.workspaceRepo.GetAll(context.Background())
		if err != nil {
			return errMsg{err}
		}
		return workspacesLoadedMsg{workspaces}
	}
}

// loadTodos returns a command to load todos for the selected workspace
func (m Model) loadTodos() tea.Cmd {
	if len(m.workspaces) == 0 || m.selectedWsIndex >= len(m.workspaces) {
		return nil
	}

	wsID := m.workspaces[m.selectedWsIndex].ID
	return func() tea.Msg {
		todos, err := m.todoRepo.GetByWorkspace(context.Background(), wsID, false)
		if err != nil {
			return errMsg{err}
		}
		return todosLoadedMsg{todos}
	}
}

// Message types
type errMsg struct{ err error }
type workspacesLoadedMsg struct{ workspaces []*domain.Workspace }
type todosLoadedMsg struct{ todos []*domain.Todo }
type notificationMsg struct {
	message string
	isError bool
}
type clearNotificationMsg struct{}

// SelectedWorkspace returns the currently selected workspace
func (m Model) SelectedWorkspace() *domain.Workspace {
	if len(m.workspaces) == 0 || m.selectedWsIndex >= len(m.workspaces) {
		return nil
	}
	return m.workspaces[m.selectedWsIndex]
}

// SelectedTodo returns the currently selected todo
func (m Model) SelectedTodo() *domain.Todo {
	if len(m.todos) == 0 || m.selectedTodoIndex >= len(m.todos) {
		return nil
	}
	return m.todos[m.selectedTodoIndex]
}

// HasWorkspaces returns true if there are workspaces
func (m Model) HasWorkspaces() bool {
	return len(m.workspaces) > 0
}

// HasTodos returns true if there are todos in the current workspace
func (m Model) HasTodos() bool {
	return len(m.todos) > 0
}

// Close closes the database connection
func (m *Model) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}
