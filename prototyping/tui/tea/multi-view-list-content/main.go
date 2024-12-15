// multi-view tea example
// Load a list of instapaper articles on one side and display the title of the actively selected article on the other
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/ieroNo47/gopaper/prototyping/instapaper"
	"github.com/joho/godotenv"
	"jaytaylor.com/html2text"
)

type item struct {
	title string
	desc  string
	id    int64
	text  string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }
func (i item) ID() int64           { return i.id }
func (i item) Text() string        { return i.text }
func (i *item) SetText(t string)   { i.text = t }

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
	state    sessionState
	spinner  spinner.Model
	list     list.Model
	viewport viewport.Model
	ready    bool
	index    int
	width    int
	height   int
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
			items = append(items, item{title: title, desc: bookmark.Description, id: bookmark.BookmarkID})
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
		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width-h-100, msg.Height-v-10)
			m.viewport.YPosition = v - 10
			m.viewport.HighPerformanceRendering = false
			m.viewport.SetContent("Loading...")
			m.ready = true

			// This is only necessary for high performance rendering, which in
			// most cases you won't need.
			//
			// Render the viewport one line below the header.
			m.viewport.YPosition = v - 10 + 1
		} else {
			m.viewport.Width = msg.Width - h - 100
			m.viewport.Height = msg.Height - v - 10
		}
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
			// m.spinner, cmd = m.spinner.Update(msg)
			// cmds = append(cmds, cmd)
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		case listView:
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)
			// get and set the contents of an item
			if len(m.list.Items()) > 0 {
				i := m.list.SelectedItem().(item)
				if len(i.Text()) == 0 {
					client, err := instapaper.NewClient()
					if err != nil {
						log.Fatalf("Failed to init Instapaper client: %v\n", err)
					}

					text, err := client.GetBookmarkText(i.ID())
					if err != nil {
						log.Fatalf("Failed to get item contents: %v\n", err)
					}

					t, err := html2text.FromString(text)
					if err != nil {
						log.Fatalf("Failed to convert html to text: %v\n", err)
					}

					out, err := glamour.Render(t, "dark")
					if err != nil {
						log.Fatalf("Failed to render with glamour: %v\n", err)
					}

					i.SetText(out)
					// // is there a better way or we always have to re-set the item if we modify it?
					m.list.SetItem(m.list.Index(), i)
				}
				m.viewport.SetContent(i.Text())
			}
		}
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case initListMsg:
		cmd = m.list.SetItems(msg)
		cmds = append(cmds, cmd)
		if len(m.list.Items()) > 0 {
			i := m.list.SelectedItem().(item)
			if len(i.Text()) == 0 {
				client, err := instapaper.NewClient()
				if err != nil {
					log.Fatalf("Failed to init Instapaper client: %v\n", err)
				}

				text, err := client.GetBookmarkText(i.ID())
				if err != nil {
					log.Fatalf("Failed to get item contents: %v\n", err)
				}

				t, err := html2text.FromString(text)
				if err != nil {
					log.Fatalf("Failed to convert html to text: %v\n", err)
				}

				out, err := glamour.Render(t, "dark")
				if err != nil {
					log.Fatalf("Failed to render with glamour: %v\n", err)
				}

				i.SetText(out)
				// // is there a better way or we always have to re-set the item if we modify it?
				m.list.SetItem(m.list.Index(), i)
			}
			m.viewport.SetContent(i.Text())
		}
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
	// rightText := m.spinner.View()
	// if len(m.list.Items()) > 0 {
	// 	i := m.list.SelectedItem().(item)
	// 	// rightText = m.list.Title
	// 	// rightText = i.Title()
	// 	if len(i.Text()) > 0 {
	// 		rightText = i.Text()
	// 	}
	// }
	rightText := m.viewport.View()

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
			focusedStyle.Render(rightText),
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
