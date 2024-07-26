package domoticmib

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	ghostStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type Item struct {
	IP          string
	Name        string
	LastUpdated time.Time
}

func (i Item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.IP)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}
	str = fn(str)
	if i.Name != "" {
		str = lipgloss.JoinHorizontal(0, str, ghostStyle.Render(" • Name: "+i.Name))
	}
	if i.LastUpdated != (time.Time{}) {
		str = lipgloss.JoinHorizontal(0, str, ghostStyle.Render(" • Last Updated At: "+i.LastUpdated.Format("2006-01-02 15:04:05")))
	}
	fmt.Fprint(w, str)
}

func NewList(il []list.Item, logCommands string, index int, lines int) *list.Model {
	l := list.New(il, itemDelegate{}, 80, lines)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.InfiniteScrolling = true
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.Select(index)
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithHelp("q:", " Quit")),
			key.NewBinding(key.WithHelp("enter:", " Pick RemoteAgent to inspect")),
			key.NewBinding(key.WithHelp("p:", " Show Packets Being Received")),
			key.NewBinding(key.WithHelp(logCommands, "")),
		}
	}
	return &l
}

func NewTextInput(width int, prompt, initialValue string) textinput.Model {
	ti := textinput.New()
	ti.Prompt = prompt
	ti.Cursor.Style = lipgloss.NewStyle().Background(lipgloss.Color("241"))
	ti.Width = width
	ti.SetValue(initialValue)
	ti.Focus()
	return ti
}
