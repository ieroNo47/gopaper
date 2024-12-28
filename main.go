// gopaper TUI app
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ieroNo47/gopaper/internal/instapaper"
	"github.com/joho/godotenv"
)

var outerStyle = lipgloss.NewStyle().
	// top and right margin needs to be 2 to avoid the border cut off issue
	Margin(2, 2, 0, 0).
	Padding(0).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("1")).
	MarginBackground(lipgloss.Color("1"))

var listStyle = lipgloss.NewStyle().
	Margin(0).
	Padding(0, 10, 0, 0).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("2")).
	MarginBackground(lipgloss.Color("2"))

var helpStyle = lipgloss.NewStyle().
	Margin(0).
	Padding(0).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("3")).
	MarginBackground(lipgloss.Color("3"))

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

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
			title := bookmark.Title
			description := fmt.Sprintf("%s | %.0f%%", strings.Join(tagNames, ","), bookmark.Progress*100)
			items = append(items, item{title: title, desc: description})
		}
		return initListMsg(items)
	}

}

type model struct {
	list list.Model
	help help.Model
}

func (m model) FullHelp() [][]key.Binding {
	return m.list.FullHelp()
}

func (m model) ShortHelp() []key.Binding {
	return m.list.ShortHelp()
}

func (m model) Init() tea.Cmd {
	return initList()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		// TODO: Find a better way to calculate the sizes for a responsive layout
		// to properly make the outer border fit the terminal window we need to subtract the
		// border and margin sizes
		oVertical := outerStyle.GetBorderTopSize() +
			outerStyle.GetBorderBottomSize() +
			outerStyle.GetMarginTop() +
			outerStyle.GetMarginBottom()

		oHorizontal := outerStyle.GetBorderLeftSize() +
			outerStyle.GetBorderRightSize() +
			outerStyle.GetMarginLeft() +
			outerStyle.GetMarginRight()

		outerStyle = outerStyle.Width(msg.Width - oHorizontal).Height(msg.Height - oVertical)

		hH, _ := outerStyle.GetFrameSize()
		hH -= helpStyle.GetBorderLeftSize() - helpStyle.GetBorderRightSize() - 2
		helpStyle = helpStyle.Width(msg.Width - hH)

		lH, lV := outerStyle.GetFrameSize()
		// not sure why we need to subtract an extra 2 here but it works
		lH -= listStyle.GetBorderLeftSize() - listStyle.GetBorderRightSize() - 2
		// not sure why we need to subtract an extra 5 here but it works. Maybe because the height is not set?
		lV -= listStyle.GetBorderTopSize() -
			listStyle.GetBorderBottomSize() -
			helpStyle.GetHeight() -
			helpStyle.GetVerticalFrameSize() - 5

		listStyle = listStyle.Width(msg.Width - lH).Height(msg.Height - lV)

		h := outerStyle.GetHorizontalFrameSize() + listStyle.GetHorizontalFrameSize()
		v := outerStyle.GetVerticalFrameSize() + listStyle.GetVerticalFrameSize() + helpStyle.GetVerticalFrameSize() + 5
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case initListMsg:
		cmd = m.list.SetItems(msg)
		cmds = append(cmds, cmd)
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	m.help = m.list.Help
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	// return listStyle.Render(m.list.View())
	view := lipgloss.JoinVertical(
		lipgloss.Bottom,
		listStyle.Render(m.list.View()),
		helpStyle.Render(m.help.View(m)),
	)
	// view := lipgloss.JoinVertical(
	// 	lipgloss.Center,
	// 	listStyle.Render("lorem ipsum"),
	// 	helpStyle.Render("sit amet consectetur adipiscing elit"),
	// )
	return outerStyle.Render(view)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	m := model{list: list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0), help: help.New()}
	// m.list.Title = "My Instapaper list"
	m.list.SetShowTitle(false)
	m.list.SetShowStatusBar(false)
	m.list.SetShowHelp(false)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}
