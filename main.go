// gopaper TUI app
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ieroNo47/gopaper/internal/instapaper"
	"github.com/joho/godotenv"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

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
			title := fmt.Sprintf("%s [%s]", bookmark.Title, strings.Join(tagNames, ","))
			items = append(items, item{title: title, desc: bookmark.Description})
		}
		return initListMsg(items)
	}

}

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return initList()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

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
	m := model{list: list.New(items, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "My Instapaper list"

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}
