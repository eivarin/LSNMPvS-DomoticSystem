package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	domoticmib "github.com/eivarin/LSNMPvS-DomoticSystem/domotic-mib"
	"github.com/muesli/termenv"
)

type updateUImsg struct{}

// Simulate a process that sends events at an irregular interval in real time.
// In this case, we'll send events on the channel at a random interval between
// 100 to 1000 milliseconds. As a command, Bubble Tea will run this
// asynchronously.
func runAgentRoutines(sub chan struct{}, d *domoticmib.DomoticMIBAgent) tea.Cmd {
	return func() tea.Msg {
		d.StartAgent(sub)
		return nil
	}
}

func waitForEvents(sub chan struct{}) tea.Cmd {
	return func() tea.Msg {
		return updateUImsg(<-sub)
	}
}

type model struct {
	sub        chan struct{}
	mibAgent   domoticmib.DomoticMIBAgent
	quitting   bool
	windowSize tea.WindowSizeMsg
	uiMode     byte
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		runAgentRoutines(m.sub, &m.mibAgent),
		waitForEvents(m.sub),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msgTyped := msg.(type) {
	case tea.KeyMsg:
		switch msgTyped.String() {
		case "q":
			m.quitting = true
			return m, tea.Quit
		case "+":
			m.mibAgent.Logger.IncreaseLogLevel()
			return m, nil
		case "-":
			m.mibAgent.Logger.DecreaseLogLevel()
			return m, nil
		default:
			if m.uiMode == 'n' && msgTyped.String() == "p" {
				m.uiMode = 'p'
			} else if m.uiMode == 'p' && msgTyped.String() == "n" {
				m.uiMode = 'n'
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.windowSize = msgTyped
		return m, nil
	case updateUImsg:
		return m, waitForEvents(m.sub)
	default:
		return m, nil
	}
}

func (m model) View() string {
	if m.quitting {
		return "Quitting..."
	}
	loglevelstr := m.mibAgent.Logger.GetCommandString()
	switch m.uiMode {
	case 'n':
		return m.mibAgent.RenderMIBWithLipgloss(m.windowSize.Width, m.windowSize.Height, []string{loglevelstr, "q: Quit", "p: View Receiving Packets"}, true)
	case 'p':
		return m.mibAgent.RenderPacketsWithLipgloss(m.windowSize.Width, m.windowSize.Height, []string{loglevelstr, "q: Quit", "n: View MIB"})
	default:
		return ""
	}
}

func main() {
	argsWithoutProg := os.Args[1:]
	lipgloss.SetColorProfile(termenv.TrueColor)
	if len(argsWithoutProg) == 0 {
		fmt.Println("No yml config provided")
		return
	}
	agent, err := domoticmib.NewDomoticMIB(argsWithoutProg[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	m := model{
		sub:      make(chan struct{}),
		mibAgent: agent,
		quitting: false,
		uiMode:   'n',
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
		os.Exit(1)
	}
}
