package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/muesli/termenv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	domoticmib "github.com/eivarin/LSNMPvS-DomoticSystem/domotic-mib"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types/CodableValues"
	// "github.com/charmbracelet/lipgloss"
)

type updateUImsg struct{}

type model struct {
	sub        chan struct{}
	MIB        *domoticmib.DomoticMIBManager
	quitting   bool
	windowSize tea.WindowSizeMsg
}

func runManagerRoutines(sub chan struct{}, d *domoticmib.DomoticMIBManager) tea.Cmd {
	return func() tea.Msg {
		d.StartManager(sub)
		return nil
	}
}

func waitForEvents(sub chan struct{}) tea.Cmd {
	return func() tea.Msg {
		return updateUImsg(<-sub)
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		runManagerRoutines(m.sub, m.MIB),
		waitForEvents(m.sub),
	)
}

func (m model) HandleKeyInHome(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "q":
		m.quitting = true
		return tea.Quit
	case "+":
		m.MIB.Logger.IncreaseLogLevel()
		return nil
	case "-":
		m.MIB.Logger.DecreaseLogLevel()
		return nil
	case "p":
		m.MIB.UiMode = 'p'
		return nil
	case "enter":
		i, ok := m.MIB.HomeList.SelectedItem().(domoticmib.Item)
		if !ok {
			return nil
		}
		m.MIB.CurrentAgentInUI = i.IP
		m.MIB.UiMode = 's'
		return nil
	case tea.KeyUp.String():
		if m.MIB.PickedAgentIndex > 0 {
			m.MIB.PickedAgentIndex--
		}
		return nil
	case tea.KeyDown.String():
		if m.MIB.PickedAgentIndex < len(m.MIB.RemoteAgents)-1 {
			m.MIB.PickedAgentIndex++
		}
		return nil
	default:
		return nil
	}
}

func (m model) HandleKeyInPackets(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "q":
		m.quitting = true
		return tea.Quit
	case "+":
		m.MIB.Logger.IncreaseLogLevel()
		return nil
	case "-":
		m.MIB.Logger.DecreaseLogLevel()
		return nil
	case "n":
		m.MIB.UiMode = 'n'
		return nil
	default:
		return nil
	}
}

func (m model) HandleKeyInStructure(msg tea.KeyMsg) tea.Cmd {
	if !m.MIB.WritingSetRequest {
		switch msg.String() {
		case "q":
			m.MIB.UiMode = 'n'
			return nil
		case "+":
			m.MIB.Logger.IncreaseLogLevel()
			return nil
		case "-":
			m.MIB.Logger.DecreaseLogLevel()
			return nil
		case "p":
			m.MIB.UiMode = 'p'
			return nil
		case "r":
			m.MIB.RefreshCurrentAgent()
			return nil
		case "s":
			m.MIB.WritingSetRequest = true
			m.MIB.CurrentInputStage = 1
			m.MIB.TextInputToSet.Reset()
			m.MIB.IIDToSet = CodableValues.NewIIDSingleIndex(0,0,0)
			m.MIB.ValueToSet = types.NewCodableInt(0)
			return nil
		default:
			return nil
		}
	} else {
		switch msg.String() {
		case tea.KeyEnter.String():
			switch m.MIB.CurrentInputStage {
			case 1:
				m.MIB.CurrentInputStage = 2
				m.MIB.IIDToSet.Structure, _ = strconv.Atoi(m.MIB.TextInputToSet.Value())
				m.MIB.TextInputToSet.Reset()
			case 2:
				m.MIB.CurrentInputStage = 3
				m.MIB.IIDToSet.Object, _ = strconv.Atoi(m.MIB.TextInputToSet.Value())
				m.MIB.TextInputToSet.Reset()
			case 3:
				m.MIB.CurrentInputStage = 4
				value, _ := strconv.Atoi(m.MIB.TextInputToSet.Value())
				m.MIB.IIDToSet.FirstIndex = &value
				m.MIB.TextInputToSet.Reset()
			case 4:
				m.MIB.CurrentInputStage = 0
				m.MIB.WritingSetRequest = false
				value, _ := strconv.Atoi(m.MIB.TextInputToSet.Value())
				m.MIB.ValueToSet = types.NewCodableInt(value)
				m.MIB.SendSetRequest()
			}
		case tea.KeyEsc.String():
			m.MIB.WritingSetRequest = false
			m.MIB.CurrentInputStage = 0
			m.MIB.TextInputToSet.Reset()
		}
		ti, cmd := m.MIB.TextInputToSet.Update(msg)
		m.MIB.TextInputToSet = ti
		return cmd
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msgTyped := msg.(type) {
	case tea.KeyMsg:
		switch m.MIB.UiMode {
		case 'n':
			return m, m.HandleKeyInHome(msgTyped)
		case 'p':
			return m, m.HandleKeyInPackets(msgTyped)
		case 's':
			return m, m.HandleKeyInStructure(msgTyped)
		default:
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.windowSize = msgTyped
		return m, nil
	case updateUImsg:
		return m, tea.Batch(waitForEvents(m.sub), m.MIB.HomeList.SetItems(m.MIB.GetList()))
	}
	newListModel, cmd := m.MIB.HomeList.Update(msg)
	m.MIB.HomeList = &newListModel
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "Quitting..."
	}
	return m.MIB.Render(m.windowSize.Width, m.windowSize.Height)
}

func main() {
	argsWithoutProg := os.Args[1:]
	lipgloss.SetColorProfile(termenv.TrueColor)
	if len(argsWithoutProg) == 0 {
		fmt.Println("No yml config provided")
		return
	}
	manager, err := domoticmib.NewDomoticMIBManager(argsWithoutProg[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	m := model{
		sub: make(chan struct{}),
		MIB: &manager,
	}
	pr := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := pr.Run(); err != nil {
		fmt.Println("could not start program:", err)
		os.Exit(1)
	}
}
