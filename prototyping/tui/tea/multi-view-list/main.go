// multi-view tea example
// Load a list of instapaper articles on one side and display the title of the actively selected article on the other
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ieroNo47/gopaper/prototyping/instapaper"
	"github.com/joho/godotenv"
)

type item struct {
	title, desc string
	// tags        []instapaper.Tag
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type sessionState uint

const (
	listView sessionState = iota
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
	docStyle     = lipgloss.NewStyle().Margin(1, 2) // list
)

type mainModel struct {
	state   sessionState
	spinner spinner.Model
	list    list.Model
	index   int
	width   int
	height  int
}

func newModel() mainModel {
	m := mainModel{state: listView}
	m.spinner = spinner.New()
	m.list = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m.list.Title = "My Instapaper list"
	return m
}

type initListMsg []list.Item

func initList() tea.Cmd {
	return func() tea.Msg {
		client, err := instapaper.NewClient()
		if err != nil {
			log.Fatalf("Failed to init Instapaper client: %v\n", err)
		}
		bookmarks, err := client.GetBookmarks(15)
		if err != nil {
			log.Fatalf("Failed to get bookmarks: %v\n", err)
		}
		items := []list.Item{}
		for _, bookmark := range bookmarks {
			tagNames := []string{}
			for _, tag := range bookmark.Tags {
				tagNames = append(tagNames, tag.Name)
			}
			title := fmt.Sprintf("%s [%s]", bookmark.Title, strings.Join(tagNames, ","))
			items = append(items, item{title: title, desc: bookmark.Description})
		}
		return initListMsg(items)
	}

}

func (m mainModel) Init() tea.Cmd {
	// return multiple commands
	return tea.Batch(m.spinner.Tick, initList())
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-10)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			// toggle view
			if m.state == listView {
				m.state = spinnerView
			} else {
				m.state = listView
			}
		case "n":
			if m.state == spinnerView {
				m.Next()
				m.resetSpinner()
				cmds = append(cmds, m.spinner.Tick)
			}
		}
		switch m.state {
		case spinnerView:
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case initListMsg:
		cmd = m.list.SetItems(msg)
		cmds = append(cmds, cmd)
	}
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
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
	rightText := m.spinner.View()
	if len(m.list.Items()) > 0 {
		i := m.list.SelectedItem().(item)
		// rightText = m.list.Title
		rightText = i.Title()
	}

	if m.state == listView {
		s += lipgloss.JoinHorizontal(
			lipgloss.Top,
			focusedStyle.Render(m.list.View()),
			defaultStyle.Render(rightText),
		)
	} else {
		s += lipgloss.JoinHorizontal(
			lipgloss.Top,
			defaultStyle.Render(m.list.View()),
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
	if m.state == listView {
		return "list"
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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
