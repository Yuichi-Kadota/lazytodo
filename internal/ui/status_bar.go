package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusBarModel holds the state for the status bar
type StatusBarModel struct {
	Mode            string
	TodoCount       int
	WorkspaceName   string
	Notification    string
	IsError         bool
	Width           int
	Styles          Styles
}

// Render renders the status bar
func (m StatusBarModel) Render() string {
	// Mode indicator
	modeStyle := m.Styles.GetModeStyle(m.Mode)
	mode := modeStyle.Render(fmt.Sprintf(" %s ", m.Mode))

	// Info parts
	var infoParts []string

	// Todo count
	if m.TodoCount >= 0 {
		infoParts = append(infoParts, fmt.Sprintf("%d todos", m.TodoCount))
	}

	// Workspace name
	if m.WorkspaceName != "" {
		infoParts = append(infoParts, m.WorkspaceName)
	}

	info := strings.Join(infoParts, " │ ")

	// Notification
	var notif string
	if m.Notification != "" {
		if m.IsError {
			notif = m.Styles.ErrorNotif.Render("✗ " + m.Notification)
		} else {
			notif = m.Styles.Notification.Render("✓ " + m.Notification)
		}
	}

	// Calculate spacing
	leftPart := mode + " " + info
	leftWidth := lipgloss.Width(leftPart)
	rightWidth := lipgloss.Width(notif)
	spacing := m.Width - leftWidth - rightWidth - 2

	if spacing < 1 {
		spacing = 1
	}

	// Build status bar
	bar := leftPart + strings.Repeat(" ", spacing) + notif

	return m.Styles.StatusBar.Width(m.Width).Render(bar)
}
