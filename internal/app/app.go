package app

import tea "github.com/charmbracelet/bubbletea"

// Model is the main application model
type Model struct {
	width  int
	height int
}

// New creates a new application model
func New() Model {
	return Model{}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View implements tea.Model
func (m Model) View() string {
	return "lazytodo - Press 'q' to quit"
}
