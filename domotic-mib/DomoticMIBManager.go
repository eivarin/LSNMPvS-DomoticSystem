package domoticmib

import (
	"net"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/eivarin/LSNMPvS-DomoticSystem/CustomLogger"
	netfuncs "github.com/eivarin/LSNMPvS-DomoticSystem/NetFuncs"
	"github.com/eivarin/LSNMPvS-DomoticSystem/mib"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types/CodableValues"
)

type RemoteAgent struct {
	MIB        *DomoticMIBAgent
	Address    string
	LastUpdate time.Time
}

func (r *RemoteAgent) GetAsItem() Item {
	r.MIB.UpdateName()
	return Item{
		Name:        r.MIB.Name,
		IP:          r.Address,
		LastUpdated: r.LastUpdate,
	}
}

type DomoticMIBManager struct {
	mib.MIB
	RemoteAgents        map[string]*RemoteAgent
	RemoteAgentsOrdered []string
	RemoteAgentsLock    *sync.RWMutex
	Logger              *CustomLogger.CustomLogger
	UiMode              byte
	CurrentAgentInUI    string
	HomeList            *list.Model
	PickedAgentIndex    int
	UpdateFrequency     time.Duration
	WritingSetRequest  	bool
	IIDToSet     	  	*CodableValues.IID
	ValueToSet   	  	*types.CompleteCodableValue
	TextInputToSet 		textinput.Model
	CurrentInputStage 	byte
}

func NewDomoticMIBManager(ymlConfig string) (DomoticMIBManager, error) {
	logger := CustomLogger.NewCustomLogger()
	logger.LogInfo("DomoticMIBManager Created", "StartUP")
	config, err := LoadMIBManagerConfig(ymlConfig)
	if err != nil {
		logger.LogError(err.Error(), "StartUP")
		return DomoticMIBManager{}, err
	}
	manager := DomoticMIBManager{
		MIB:                 mib.NewMIB(&logger, []mib.StructureI{}),
		RemoteAgents:        make(map[string]*RemoteAgent),
		RemoteAgentsOrdered: []string{},
		RemoteAgentsLock:    &sync.RWMutex{},
		Logger:              &logger,
		UiMode:              'n',
		CurrentAgentInUI:    "",
		HomeList:            NewList([]list.Item{}, logger.GetCommandString(), 0, 10),
		PickedAgentIndex:    0,
		UpdateFrequency:     120 * time.Second,
		IIDToSet:            nil,
		ValueToSet:          nil,
		TextInputToSet:      NewTextInput(20, "", ""),
		CurrentInputStage:   0,
	}
	for _, address := range config.RemoteAgentsAddresses {
		manager.AddEmptyAgent(address)
	}
	return manager, nil
}

func (m *DomoticMIBManager) AddEmptyAgent(address string) {
	m.RemoteAgentsLock.Lock()
	defer m.RemoteAgentsLock.Unlock()
	Device := NewDeviceGroup(DeviceConfig{})
	Sensors := NewSensorsTable([]SensorConfig{})
	Actuators := NewActuatorsTable([]ActuatorConfig{})
	newMIB := &DomoticMIBAgent{
		MIB:             mib.NewMIB(m.Logger, []mib.StructureI{Device, Sensors, Actuators}),
		Device:          Device,
		Sensors:         Sensors,
		Actuators:       Actuators,
		Name:            "",
		updateFrequency: 5 * time.Second,
	}
	m.RemoteAgents[address] = &RemoteAgent{
		MIB:        newMIB,
		Address:    address,
		LastUpdate: time.Now(),
	}
	m.RemoteAgentsOrdered = append(m.RemoteAgentsOrdered, address)
}

func (m *DomoticMIBManager) StartManager(sub chan struct{}) {
	m.StartManagerUpdater(sub)
	m.ListenForRequests(sub)
}

func (m *DomoticMIBManager) StartManagerUpdater(sub chan struct{}) {
	go func() {
		for {
			for addr, agent := range m.RemoteAgents {
				agent.MIB.RefreshAgent(addr)
			}
			sub <- struct{}{}
			time.Sleep(m.UpdateFrequency)
		}
	}()
}

func (d *DomoticMIBManager) ListenForRequests(sub chan struct{}) {
	addr := &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 12345,
	}
	udpListener, _ := net.ListenUDP("udp", addr)
	for {
		buffer := make([]byte, 10000)
		n, addr, _ := udpListener.ReadFromUDP(buffer)
		newPacket := packet.LSNMPvS_Packet{}
		_, e := newPacket.Decode(string(buffer[:n]))
		if e != 0 {
			go func() {
				errPacket := packet.NewErrorDecodingPacket(e)
				netfuncs.Send(addr, []byte(errPacket.Encode()))
				d.Logger.LogError(e.Error(), "Request")
			}()
			continue
		}
		go d.MIB.HandleRequest(newPacket, *addr, sub, d)
	}
}

func (m *DomoticMIBManager) GetList() []list.Item {
	m.RemoteAgentsLock.RLock()
	defer m.RemoteAgentsLock.RUnlock()
	items := make([]list.Item, 0)
	for _, agent := range m.RemoteAgentsOrdered {
		items = append(items, m.RemoteAgents[agent].GetAsItem())
	}
	return items
}

func (m *DomoticMIBManager) HandleGet(r packet.LSNMPvS_Packet, addr *net.UDPAddr) (*packet.LSNMPvS_Packet, error, bool) {
	return nil, nil, false
}

func (m *DomoticMIBManager) HandleSet(r packet.LSNMPvS_Packet, addr *net.UDPAddr) (*packet.LSNMPvS_Packet, error, bool) {
	return nil, nil, false
}

func (m *DomoticMIBManager) HandleResponse(r packet.LSNMPvS_Packet, addr *net.UDPAddr) (*packet.LSNMPvS_Packet, error, bool) {
	if !r.TryLogErrors(m.Logger) {
		addr.Port = 12345
		remAgent := m.RemoteAgents[addr.String()]
		remAgent.LastUpdate = time.Now()
		p, err, respond := remAgent.MIB.Update(r)
		remAgent.MIB.UpdateName()
		return p, err, respond
	} else {
		return nil, nil, false	
	}
}

func (m *DomoticMIBManager) HandleNotification(r packet.LSNMPvS_Packet, addr *net.UDPAddr) (*packet.LSNMPvS_Packet, error, bool) {
	addr.Port = 12345
	addrStr := addr.String()
	remAgent, ok := m.RemoteAgents[addrStr]
	if !ok {
		m.AddEmptyAgent(addrStr)
		remAgent = m.RemoteAgents[addrStr]
		remAgent.MIB.RefreshAgent(addrStr)
	}
	p, err, respond := remAgent.MIB.Update(r)
	remAgent.LastUpdate = time.Now()
	remAgent.MIB.UpdateName()
	return p, err, respond
}

func (m *DomoticMIBManager) RefreshCurrentAgent() {
	remAgent := m.RemoteAgents[m.CurrentAgentInUI]
	remAgent.MIB.RefreshAgent(remAgent.Address)
}

func (m *DomoticMIBManager) SendSetRequest() {
	iidCodableList := types.CodableList{}
	iidCodableList.Append(types.NewCodableIID(m.IIDToSet.Structure, m.IIDToSet.Object, []int{*m.IIDToSet.FirstIndex}))
	valueCodableList := types.CodableList{}
	valueCodableList.Append(m.ValueToSet)
	p := packet.NewSetResponsePacket(iidCodableList, valueCodableList)
	netfuncs.SendStrAddr(m.CurrentAgentInUI, []byte(p.Encode()))
}

func (m *DomoticMIBManager) Render(width, height int) string {
	cmdStr := m.Logger.GetCommandString()
	switch m.UiMode {
	case 'n':
		smallerHeight := (height-8)/2
		l := m.GetList()
		m.HomeList = NewList(l, cmdStr, m.PickedAgentIndex, smallerHeight-2)
		listStr := m.HomeList.View()
		title := lipgloss.NewStyle().Align(lipgloss.Center).Render("Domotic MIB Manager - Home")
		color := lipgloss.Color("99")
		bigBoxColor := lipgloss.Color("208")
		renderedCmds := lipgloss.NewStyle().Align(lipgloss.Center).Foreground(lipgloss.Color("248")).Render(strings.Join([]string{"q: Exit", "p: View Packets", "↑/↓: Navigate", "Enter: Inspect Remote Agent"}, " • "))
		renderedListStr := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(color).Width(width-6).Height(smallerHeight).Align(lipgloss.Center).Render(listStr)
		renderedLogs := m.Logger.RenderLogsWithLipGloss(width-4, height-20)
		renderedBox := lipgloss.NewStyle().Align(lipgloss.Center).Border(lipgloss.RoundedBorder()).BorderForeground(bigBoxColor).Width(width-2).Height(smallerHeight).Render(lipgloss.JoinVertical(lipgloss.Center, renderedListStr, renderedLogs))
		return lipgloss.JoinVertical(lipgloss.Center, title, renderedBox, renderedCmds)
	case 'p':
		return m.RenderPacketsWithLipgloss(width, height, []string{"q: Exit", "n: Back"})
	case 's':
		var commands []string
		renderedMIB := m.RemoteAgents[m.CurrentAgentInUI].MIB.RenderMIBWithLipgloss(width, height, []string{}, false)
		if m.WritingSetRequest {
			commands = []string{"Enter: Confirm", "Esc: Cancel"}
			title := ""
			switch m.CurrentInputStage {
			case 1:
				title = "Enter Structure ID"
			case 2:
				title = "Enter Object ID"
			case 3:
				title = "Enter Index"
			case 4:
				title = "Enter Value"
			}
			StructTitle := lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Width(width).Align(lipgloss.Center).Border(lipgloss.NormalBorder(), false, false, true).BorderForeground(lipgloss.Color("208")).Render(title)
			renderedMIB = lipgloss.JoinVertical(lipgloss.Center, renderedMIB, StructTitle, lipgloss.NewStyle().Padding(0,1).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("208")).Render(m.TextInputToSet.View()))
			} else {
			commands = []string{"q: Exit", "n: Back", "s: Set Value", "r: Refresh"}
		}
		renderedCommands := lipgloss.NewStyle().Align(lipgloss.Center).Foreground(lipgloss.Color("248")).Render(strings.Join(commands, " • "))
		return lipgloss.JoinVertical(lipgloss.Center, renderedMIB, renderedCommands)
	default:
		return ""
	}
}

func (d *DomoticMIBManager) RenderPacketsWithLipgloss(width int, height int, controls []string) string {
	commandsStyle := lipgloss.NewStyle().Align(lipgloss.Center).Foreground(lipgloss.Color("248"))
	comStr := commandsStyle.Render(strings.Join(controls, " • "))
	title := lipgloss.NewStyle().Align(lipgloss.Center).Render("Domotic MIB Manager - Packets")
	rendered := d.MIB.Packets.RenderPacketsWithLipGloss(width-4, height-4)
	return lipgloss.JoinVertical(lipgloss.Center, title, lipgloss.NewStyle().Width(width-2).Height(height-4).Align(lipgloss.Bottom).Border(lipgloss.RoundedBorder()).Render(rendered), comStr)
}
