package ui

import "github.com/charmbracelet/lipgloss"

// Kanagawa Wave color palette
var (
	// Base colors
	ColorBackground = lipgloss.Color("#1F1F28") // sumiInk1
	ColorForeground = lipgloss.Color("#DCD7BA") // fujiWhite
	ColorPrimary    = lipgloss.Color("#7E9CD8") // crystalBlue
	ColorSecondary  = lipgloss.Color("#957FB8") // oniViolet
	ColorSuccess    = lipgloss.Color("#98BB6C") // springGreen
	ColorWarning    = lipgloss.Color("#E6C384") // carpYellow
	ColorError      = lipgloss.Color("#C34043") // autumnRed
	ColorMuted      = lipgloss.Color("#727169") // fujiGray

	// Pane colors
	ColorSelectedBg   = lipgloss.Color("#2D4F67") // waveBlue2
	ColorFolderRoot   = lipgloss.Color("#E6C384") // carpYellow
	ColorFolderChild  = lipgloss.Color("#957FB8") // oniViolet
	ColorTodoPending  = lipgloss.Color("#DCD7BA") // fujiWhite
	ColorTodoComplete = lipgloss.Color("#727169") // fujiGray
	ColorTodoUrgent   = lipgloss.Color("#FF5D62") // peachRed
	ColorDueToday     = lipgloss.Color("#E6C384") // carpYellow
	ColorOverdue      = lipgloss.Color("#C34043") // autumnRed

	// Tree colors
	ColorTreeGuide   = lipgloss.Color("#54546D") // sumiInk4
	ColorIconArchive = lipgloss.Color("#54546D") // sumiInk4
)

// Mode colors
var ModeColors = map[string]lipgloss.Color{
	"NORMAL": ColorPrimary,
	"INSERT": ColorSuccess,
	"SEARCH": ColorWarning,
	"SORT":   ColorSecondary,
}

// Icons (Nerd Font)
var (
	IconFolderOpen   = ""
	IconFolderClosed = ""
	IconTodo         = ""
	IconTodoDone     = ""
	IconTodoUrgent   = ""
	IconArchive      = ""
)
