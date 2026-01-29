package ui

// InputBarModel holds the state for the input bar
type InputBarModel struct {
	Prompt  string
	Value   string
	Width   int
	Styles  Styles
}

// Render renders the input bar
func (m InputBarModel) Render() string {
	content := m.Prompt + m.Value + "_"

	return m.Styles.InputBar.
		Width(m.Width).
		Render(content)
}
