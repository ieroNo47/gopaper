// gopaper TUI app
package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ieroNo47/gopaper/internal/instapaper"
	"github.com/joho/godotenv"
)

type sessionState uint

const (
	bookmarksView sessionState = iota
	tagsView
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
	BorderForeground(lipgloss.Color("5")).
	MarginBackground(lipgloss.Color("5"))

var tagsStyle = lipgloss.NewStyle().
	Margin(0).
	Padding(0).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("0")).
	MarginBackground(lipgloss.Color("0"))

var helpStyle = lipgloss.NewStyle().
	Margin(0).
	Padding(0).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("4")).
	MarginBackground(lipgloss.Color("4"))

type item struct {
	title string
	desc  string
	tags  []instapaper.Tag
}

func (i item) Title() string          { return i.title }
func (i item) Description() string    { return i.desc }
func (i item) FilterValue() string    { return i.title }
func (i item) Tags() []instapaper.Tag { return i.tags }

type initListMsg []list.Item

func initList() tea.Cmd {
	return func() tea.Msg {
		client, err := instapaper.NewClient()
		if err != nil {
			log.Fatalf("Failed to init Instapaper client: %v\n", err)
		}
		bookmarks, err := client.GetBookmarks(50)
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
			items = append(items, item{title: title, desc: description, tags: bookmark.Tags})
		}
		return initListMsg(items)
	}

}

type model struct {
	list  list.Model
	table table.Model
	help  help.Model
	state sessionState
}

func (m model) FullHelp() [][]key.Binding {
	if m.state == bookmarksView {
		return m.list.FullHelp()
	} else {
		return m.table.KeyMap.FullHelp()
	}
}

func (m model) ShortHelp() []key.Binding {
	if m.state == bookmarksView {
		return m.list.ShortHelp()
	} else {
		return m.table.KeyMap.ShortHelp()
	}
}

func (m model) Init() tea.Cmd {
	return initList()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			// switch focus between bookmarks and tags view
			if m.state == bookmarksView {
				m.state = tagsView
			} else {
				m.state = bookmarksView
			}
		}
		// pass msg to the current view
		switch m.state {
		case bookmarksView:
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
			listStyle = listStyle.BorderForeground(lipgloss.Color("5"))
			tagsStyle = tagsStyle.BorderForeground(lipgloss.Color("0"))
		case tagsView:
			m.table, cmd = m.table.Update(msg)
			cmds = append(cmds, cmd)
			listStyle = listStyle.BorderForeground(lipgloss.Color("0"))
			tagsStyle = tagsStyle.BorderForeground(lipgloss.Color("5"))
		}
	case tea.WindowSizeMsg:
		// TODO: Find a better way to calculate the sizes for a responsive layout
		// to properly make the outer border fit the terminal window we need to subtract the
		// border and margin sizes
		// TODO: the outer style is mostly for testing and to learn how lipgloss works, can be removed later to save some screen space
		// h is for Horizontal, not height
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

		// listStyle = listStyle.Width(msg.Width - lH).Height(msg.Height - lV)
		// h := outerStyle.GetHorizontalFrameSize() + listStyle.GetHorizontalFrameSize()
		// v := outerStyle.GetVerticalFrameSize() + listStyle.GetVerticalFrameSize() + helpStyle.GetVerticalFrameSize() + 5
		// // if we subtract an extra 2 or more from the width, the list contents are truncated more gracefully without
		// // wrapping to the next line and breaking the layout
		// m.list.SetSize(msg.Width-h-2, msg.Height-v)

		// tags view wip
		w := msg.Width - lH - 2
		listStyle = listStyle.Width((w * 2) / 3).Height(msg.Height - lV)
		tagsStyle = tagsStyle.Width(w / 3).Height(msg.Height - lV)
		v := outerStyle.GetVerticalFrameSize() + listStyle.GetVerticalFrameSize() + helpStyle.GetVerticalFrameSize() + 5
		m.list.SetSize((w*2/3)-10, msg.Height-v)
		m.table.SetWidth((w / 3) - 5)
		m.table.SetHeight(msg.Height - v - 1)
		m.table.SetColumns([]table.Column{
			{Width: (w / 3) - 5},
		})
	case initListMsg:
		cmd = m.list.SetItems(msg)
		cmds = append(cmds, cmd)
		m.table.SetRows(m.getTagRows())
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	// return listStyle.Render(m.list.View())
	listWithTagsView := lipgloss.JoinHorizontal(
		lipgloss.Bottom,
		listStyle.Render(m.list.View()),
		tagsStyle.Render(m.table.View()),
	)
	view := lipgloss.JoinVertical(
		lipgloss.Bottom,
		listWithTagsView,
		helpStyle.Render(m.help.View(m)),
	)
	return outerStyle.Render(view)
}

// misc helper functions

// getTags returns a map of tags and their counts from the list of downloaded bookmarks
// the current version of the instapaper api does not support fetching tags
func (m model) getTags() map[string]int {
	tags := map[string]int{}
	for _, i := range m.list.Items() {
		for _, tag := range i.(item).Tags() {
			tags[tag.Name]++
		}
	}
	return tags
}

// todo: cleanup and move elsewhere
type tagKeyValue struct {
	key   string
	value int
}

func (m model) getTagRows() []table.Row {
	tags := m.getTags()
	// sort by count
	kv := make([]tagKeyValue, 0, len(tags))
	for k, v := range tags {
		kv = append(kv, tagKeyValue{k, v})
	}
	sort.Slice(kv, func(i, j int) bool {
		return kv[i].value > kv[j].value
	})
	items := []table.Row{}
	for _, tag := range kv {
		items = append(items, table.Row{fmt.Sprintf("(%d) %s", tag.value, tag.key)})
	}

	return items
}

// main function, inits and runs the tea
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	columns := []table.Column{
		{Width: 10},
	}

	m := model{
		state: bookmarksView,
		list:  list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		help:  help.New(),
		table: table.New(
			table.WithFocused(true),
			table.WithColumns(columns),
			table.WithHeight(5),
			table.WithRows(
				[]table.Row{{"Loading..."}})),
	}
	// m.list.Title = "My Instapaper list"
	m.list.SetShowTitle(false)
	m.list.SetShowStatusBar(false)
	m.list.SetShowHelp(false)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}
