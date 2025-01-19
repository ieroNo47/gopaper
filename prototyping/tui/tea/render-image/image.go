// example from https://github.com/mistakenelf/teacup/blob/main/examples/image/image.go
package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mistakenelf/teacup/image"
)

var outerStyle = lipgloss.NewStyle().
	// top and right margin needs to be 2 to avoid the border cut off issue
	Margin(2, 2, 0, 0).
	Padding(0).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("1")).
	MarginBackground(lipgloss.Color("1"))

// model represents the properties of the UI.
type model struct {
	image image.Model
}

// New creates a new instance of the UI.
func New() model {
	imageModel := image.New(true, true, lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"})

	return model{
		image: imageModel,
	}
}

// Init intializes the UI.
func (b model) Init() tea.Cmd {
	return nil
}

// Update handles all UI interactions.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		oVertical := outerStyle.GetBorderTopSize() +
			outerStyle.GetBorderBottomSize() +
			outerStyle.GetMarginTop() +
			outerStyle.GetMarginBottom()

		oHorizontal := outerStyle.GetBorderLeftSize() +
			outerStyle.GetBorderRightSize() +
			outerStyle.GetMarginLeft() +
			outerStyle.GetMarginRight()

		outerStyle = outerStyle.Width(msg.Width - oHorizontal).Height(msg.Height - oVertical)
		h, v := outerStyle.GetFrameSize()
		m.image.SetSize(msg.Width-h, msg.Height-v-5)
		cmds = append(cmds, m.image.SetFileName("i1.png"))

		return m, tea.Batch(cmds...)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			cmds = append(cmds, tea.Quit)
		}
	}

	m.image, cmd = m.image.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View returns a string representation of the UI.
func (m model) View() string {
	return outerStyle.Render(m.image.View())
}

func main() {
	b := New()
	p := tea.NewProgram(b, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
