// multi-view tea example
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionState uint

const (
	defaultTime              = time.Minute
	timerView   sessionState = iota
	spinnerView
	width  = 20
	height = 10
)

var (
	spinners = []spinner.Spinner{
		spinner.Line,
		spinner.Dot,
		spinner.Pulse,
	}
	modelStyle = lipgloss.NewStyle().
		// Width(width).
		// Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		BorderStyle(lipgloss.HiddenBorder())
	focusedModelStyle = lipgloss.NewStyle().
		// Width(width).
		// Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("69"))
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type mainModel struct {
	state   sessionState
	timer   timer.Model
	spinner spinner.Model
	index   int
	width   int
	height  int
}

func newModel(timeout time.Duration) mainModel {
	m := mainModel{state: timerView}
	m.timer = timer.New(timeout)
	m.spinner = spinner.New()
	return m
}

func (m mainModel) Init() tea.Cmd {
	// return multiple commands
	return tea.Batch(m.timer.Init(), m.spinner.Tick)
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			// toggle view
			if m.state == timerView {
				m.state = spinnerView
			} else {
				m.state = timerView
			}
		case "n":
			if m.state == timerView {
				m.timer = timer.New(defaultTime)
				cmds = append(cmds, m.timer.Init())
			} else {
				m.Next()
				m.resetSpinner()
				cmds = append(cmds, m.spinner.Tick)
			}
		}
		switch m.state {
		case spinnerView:
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		default:
			m.timer, cmd = m.timer.Update(msg)
			cmds = append(cmds, cmd)
		}
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case timer.TickMsg:
		m.timer, cmd = m.timer.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m mainModel) View() string {
	var s string
	model := m.currentFocusedModel()
	// find a better way to get the height of other components in view
	dummyHelp := helpStyle.Render(
		fmt.Sprintf("\ntab: focus next • n: new %s • q: exit - window: %d/%d focused:%d/%d default: %d/%d\n",
			model,
			m.height,
			m.width,
			10,
			10,
			10,
			10,
		),
	)
	defaultStyle := modelStyle.Width(m.width/2 - 2).Height(m.height - lipgloss.Height(dummyHelp) - 1)
	focusedStyle := focusedModelStyle.Width(m.width/2 - 2).Height(m.height - lipgloss.Height(dummyHelp) - 1)

	if m.state == timerView {
		s += lipgloss.JoinHorizontal(
			lipgloss.Top,
			focusedStyle.Render(fmt.Sprintf("%4s", m.timer.View())),
			defaultStyle.Render(m.spinner.View()),
		)
	} else {
		s += lipgloss.JoinHorizontal(
			lipgloss.Top,
			defaultStyle.Render(fmt.Sprintf("%4s", m.timer.View())),
			focusedStyle.Render(m.spinner.View()),
		)
	}
	helpRender := helpStyle.Render(
		fmt.Sprintf("\ntab: focus next • n: new %s • q: exit - window: %d/%d focused:%d/%d default: %d/%d\n",
			model,
			m.height,
			m.width,
			defaultStyle.GetHeight(),
			defaultStyle.GetWidth(),
			focusedStyle.GetHeight(),
			focusedStyle.GetWidth(),
		),
	)
	s += helpRender
	return s
}

func (m mainModel) currentFocusedModel() string {
	if m.state == timerView {
		return "timer"
	} else {
		return "spinner"
	}
}

func (m *mainModel) Next() {
	if m.index == len(spinners)-1 {
		m.index = 0
	} else {
		m.index++
	}
}

func (m *mainModel) resetSpinner() {
	m.spinner = spinner.New()
	m.spinner.Style = spinnerStyle
	m.spinner.Spinner = spinners[m.index]
}

func main() {
	p := tea.NewProgram(newModel(defaultTime), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
