// gopaper TUI app
package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ieroNo47/gopaper/internal/instapaper"
	"github.com/joho/godotenv"
	"golang.org/x/term"
)

type sessionState uint

const (
	bookmarksView sessionState = iota
	tagsView
)

var outerStyle = lipgloss.NewStyle().
	// top margin needs to be 6 to avoid cut off issues, not sure why
	Margin(6, 0, 0, 0).
	Padding(0).
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("1")).
	MarginBackground(lipgloss.Color("1"))

var listStyle = lipgloss.NewStyle().
	Margin(0).
	Padding(0, 0, 0, 0).
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

// item is a struct that implements the list.Item interface
// it represents a single item in the list
type item struct {
	title string
	desc  string
	tags  []instapaper.Tag
}

func (i item) Title() string          { return i.title }
func (i item) Description() string    { return i.desc }
func (i item) FilterValue() string    { return i.title }
func (i item) Tags() []instapaper.Tag { return i.tags }

// initListMsg is a message type for initializing the list
type initListMsg []list.Item

// initList initializes the list with bookmarks
// it fetches the bookmarks from the instapaper client and returns a list of items
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

// model is the main model for the application
type model struct {
	list           *list.Model
	unfilteredList list.Model
	filteredList   list.Model
	table          table.Model
	help           help.Model
	state          sessionState
	log            *log.Logger
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

// filterByTagMsg is a message type for filtering the list by tag
type filterByTagMsg string

// filterByTag marks the selected tag in the table and returns a command to filter the list by that tag
func (m model) filterByTag(tag string) tea.Cmd {
	return func() tea.Msg {
		// mark tag as selected
		cursor := m.table.Cursor()
		rows := m.table.Rows()
		for i := range rows {
			if i == cursor {
				rows[i][0] = "✓"
			} else {
				rows[i][0] = " "
			}
		}
		m.table.SetRows(rows)
		return filterByTagMsg(tag)
	}
}

// clearFilterMsg is a message type for clearing the filter
type clearFilterMsg string

// clearFilter clears the tag selection and returns a command to clear the filter
func (m model) clearFilter() tea.Cmd {
	return func() tea.Msg {
		rows := m.table.Rows()
		for i := range rows {
			rows[i][0] = " "
		}
		m.table.SetRows(rows)
		return clearFilterMsg("")
	}
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return initList()
}

// Update updates the model based on received messages
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	cmds := []tea.Cmd{}

	switch msg := msg.(type) {
	// handle a key press
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
		case "enter":
			if m.state == tagsView {
				row := m.table.SelectedRow()
				var cmd tea.Cmd
				if row[0] == "✓" { // better way to check if tag is selected?
					// clear filter if tag is already selected
					cmd = m.clearFilter()
				} else {
					// filter by tag if it is not selected
					tag := row[2]
					cmd = m.filterByTag(tag)
				}
				cmds = append(cmds, cmd)
			}
		}

		// pass msg to the active view
		switch m.state {
		case bookmarksView:
			*m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
			listStyle = listStyle.BorderForeground(lipgloss.Color("5"))
			tagsStyle = tagsStyle.BorderForeground(lipgloss.Color("0"))
		case tagsView:
			m.table, cmd = m.table.Update(msg)
			cmds = append(cmds, cmd)
			listStyle = listStyle.BorderForeground(lipgloss.Color("0"))
			tagsStyle = tagsStyle.BorderForeground(lipgloss.Color("5"))
		}
	// handle a window resize event
	case tea.WindowSizeMsg:
		// TODO: Find a better way to calculate the sizes for a responsive layout
		// to properly make the outer border fit the terminal window we need to subtract the
		// border and margin sizes
		// TODO: the outer style is mostly for testing and to learn how lipgloss works, can be removed later to save some screen space
		// h is for Horizontal, not height

		// size of outer vertical and horizontal borders
		oVertical := outerStyle.GetBorderTopSize() +
			outerStyle.GetBorderBottomSize() +
			outerStyle.GetMarginTop() +
			outerStyle.GetMarginBottom()

		oHorizontal := outerStyle.GetBorderLeftSize() +
			outerStyle.GetBorderRightSize() +
			outerStyle.GetMarginLeft() +
			outerStyle.GetMarginRight()

		// size of the 'outer' parent container adjusted to be the window size - the size of the borders and margins
		outerStyle = outerStyle.Width(msg.Width - oHorizontal).Height(msg.Height - oVertical)

		// hH = help Horizontal. It is the size of the outer horizontal frame size (border + margin + padding)
		// and the help style horizontal border size
		// we subtract this from the width of the window to get the usable width for our help msg container
		hH, _ := outerStyle.GetFrameSize()
		hH -= helpStyle.GetBorderLeftSize() - helpStyle.GetBorderRightSize() - 2
		helpStyle = helpStyle.Width(msg.Width - hH)

		// lH = list Horizontal
		// lV = list Vertical
		lH, lV := outerStyle.GetFrameSize()
		// not sure why we need to subtract an extra 2 here but it works
		lH -= listStyle.GetBorderLeftSize() - listStyle.GetBorderRightSize() - 2
		lV -= listStyle.GetBorderTopSize() -
			listStyle.GetBorderBottomSize() -
			helpStyle.GetHeight() -
			helpStyle.GetVerticalFrameSize()

		// listStyle = listStyle.Width(msg.Width - lH).Height(msg.Height - lV)
		// h := outerStyle.GetHorizontalFrameSize() + listStyle.GetHorizontalFrameSize()
		// v := outerStyle.GetVerticalFrameSize() + listStyle.GetVerticalFrameSize() + helpStyle.GetVerticalFrameSize() + 5
		// // if we subtract an extra 2 or more from the width, the list contents are truncated more gracefully without
		// // wrapping to the next line and breaking the layout
		// m.list.SetSize(msg.Width-h-2, msg.Height-v)

		// tags view wip
		w := msg.Width - lH - 2
		listStyle = listStyle.Width((w * 2) / 3).Height(msg.Height - lV)
		tagsStyle = tagsStyle.Width(w / 3).Height(msg.Height - lV + 3)
		v := outerStyle.GetVerticalFrameSize() + listStyle.GetVerticalFrameSize() + helpStyle.GetVerticalFrameSize() - 5
		m.list.SetSize((w*2/3)-10, msg.Height-v)
		m.table.SetWidth((w / 3) - 5)
		m.table.SetHeight(msg.Height - v - 1)
		m.table.SetColumns([]table.Column{
			{Width: 1},
			{Width: 3},
			{Width: (w / 3) - 8},
		})
	// handle initListMsg event, which is sent when the list is initialized
	case initListMsg:
		cmd = m.list.SetItems(msg)
		cmds = append(cmds, cmd)
		m.table.SetRows(m.getTagRows())
	// handle filterByTagMsg event, which is sent when a tag is selected
	case filterByTagMsg:
		// filter the list by the selected tag
		// should some of this be in filterByTag instead?
		tag := string(msg)
		filteredItems := []list.Item{}
		for _, i := range m.unfilteredList.Items() {
			// check if tag is in the item's tags
			for _, t := range i.(item).Tags() {
				if t.Name == tag {
					filteredItems = append(filteredItems, i)
				}
			}
		}
		// switch pointer to filtered list and update it with items matching the tag
		m.list = &m.filteredList
		cmd := m.list.SetItems(filteredItems)
		cmds = append(cmds, cmd)
		cmd = m.forceRedraw()
		cmds = append(cmds, cmd)
	case clearFilterMsg:
		// switch back to the unfiltered list
		m.list = &m.unfilteredList
		cmd := m.forceRedraw()
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the model to a string to be displayed in the terminal
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

// ####### misc helper functions #######

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

// getTagRows sorts the tags by count and returns the result as a list of table rows
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
		items = append(items, table.Row{
			" ",
			strconv.Itoa(tag.value),
			tag.key})
	}

	return items
}

// forceRedraw sends a window size message to force a recalculation of the layout
// this is needed to ensure that the content is visible when the main bookmarks list is filtered
func (m model) forceRedraw() tea.Cmd {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatalf("Failed to get terminal size: %v\n", err)
	}

	return func() tea.Msg {
		return tea.WindowSizeMsg{
			Width:  w,
			Height: h,
		}
	}
}

// main function, inits and runs the tea
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Open a log file to write debug messages
	f, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Set initial tag table column to display loading msg
	columns := []table.Column{
		{Width: 1},
		{Width: 3},
		{Width: 10},
	}

	m := model{
		state: bookmarksView,
		// two separate lists, unfiltered will have all bookmarks
		// filtered will be used to display bookmarks filtered by tag or other state like in progress, archived, etc.
		unfilteredList: list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		filteredList:   list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		help:           help.New(),
		table: table.New(
			table.WithFocused(true),
			table.WithColumns(columns),
			table.WithHeight(5),
			table.WithRows(
				[]table.Row{{" ", " ", "Loading..."}})),
		log: log.New(f, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
	}

	m.filteredList.SetShowTitle(false)
	m.filteredList.SetShowStatusBar(false)
	m.filteredList.SetShowHelp(false)
	m.unfilteredList.SetShowTitle(false)
	m.unfilteredList.SetShowStatusBar(false)
	m.unfilteredList.SetShowHelp(false)
	// set the initial list to be the unfiltered list
	m.list = &m.unfilteredList
	// pass m as pointer because we changed Update to have a pointer receiver
	p := tea.NewProgram(&m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}
