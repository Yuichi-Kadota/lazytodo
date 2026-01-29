package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yuichikadota/lazytodo/internal/app"
)

func main() {
	model := app.New(app.Config{})

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	)

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Close database connection
	if m, ok := finalModel.(app.Model); ok {
		m.Close()
	}
}
