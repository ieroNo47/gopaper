// commands in bubble tea
// https://github.com/charmbracelet/bubbletea/tree/master/tutorials/commands/

package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	status int
	err    error
	url    string
}

type statusMsg int
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

func checkServer(url string) tea.Cmd {
	return func() tea.Msg {
		c := &http.Client{Timeout: 10 * time.Second}
		res, err := c.Get(url)
		if err != nil {
			return errMsg{err}
		}
		return statusMsg(res.StatusCode)
	}
}

func (m model) Init() tea.Cmd {
	return checkServer(m.url)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case statusMsg:
		m.status = int(msg)
		return m, tea.Quit

	case errMsg:
		m.err = msg
		return m, tea.Quit

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nTrouble in the bubble: %v\n\n", m.err)
	}
	s := fmt.Sprintf("Checking %s ... ", m.url)
	if m.status > 0 {
		s += fmt.Sprintf("%d %s!", m.status, http.StatusText(m.status))
	}

	return "\n" + s + "\n\n"
}

func main() {
	url := "https://charm.sh"
	if _, err := tea.NewProgram(model{url: url}).Run(); err != nil {
		log.Fatalf("There was an error: %v\n", err)
	}
}
