package domoticmib

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/eivarin/LSNMPvS-DomoticSystem/CustomLogger"
	netfuncs "github.com/eivarin/LSNMPvS-DomoticSystem/NetFuncs"
	"github.com/eivarin/LSNMPvS-DomoticSystem/mib"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types/CodableValues"
)

type DomoticMIBAgent struct {
	mib.MIB
	Device          *mib.Group
	Sensors         *mib.Table
	Actuators       *mib.Table
	Name            string
	OriginalConfig  DomoticMIBAgentConfig
	updateFrequency time.Duration
}

func NewDomoticMIB(ymlConfig string) (DomoticMIBAgent, error) {
	config, err := LoadMIBAgentConfig(ymlConfig)
	logger := CustomLogger.NewCustomLogger()
	logger.LogInfo(fmt.Sprintf("Config %s Loaded", ymlConfig), "StartUP")
	device := NewDeviceGroup(config.Device)
	logger.LogInfo("Device Group Created", "StartUP")
	sensors := NewSensorsTable(config.Sensors)
	logger.LogInfo("Sensors Table Created", "StartUP")
	actuators := NewActuatorsTable(config.Actuators)
	logger.LogInfo("Actuators Table Created", "StartUP")
	if err != nil {
		return DomoticMIBAgent{}, err
	}
	agent := DomoticMIBAgent{
		MIB:             mib.NewMIB(&logger, []mib.StructureI{device, sensors, actuators}),
		Device:          device,
		Sensors:         sensors,
		Actuators:       actuators,
		updateFrequency: 1 * time.Second,
		Name:            config.Device.ID,
		OriginalConfig:  config,
	}
	return agent, nil
}

func (d *DomoticMIBAgent) UpdateName() {
	obj := d.Device.Objects.(DeviceObjects).Id
	nameValue, _ := obj.Get()
	name := nameValue.Value.(*CodableValues.CodableString).Value
	d.Name = name
}

func (d *DomoticMIBAgent) Get(structure, objectIID int, index *int) (types.IdValuePair, packet.PacketErr) {
	return d.MIB.Get(structure, objectIID, index)
}

func (d *DomoticMIBAgent) Set(structure, objectIID int, index *int, value types.CompleteCodableValue) packet.PacketErr {
	return d.MIB.Set(structure, objectIID, index, value)
}

func (d *DomoticMIBAgent) StartAgent(sub chan struct{}) {
	d.StartAgentUpdater(sub)
	d.MIB.StartNotificationLoop(sub)
	d.ListenForRequests(sub)
}

func (d *DomoticMIBAgent) StartNotifications(sub chan struct{}) {
	go func() {
		for {
			time.Sleep(d.Device.GetNotificationRate())
			d.Device.RLock()
			d.Device.SendNotifications(d.GetUptime())
			d.Logger.LogInfo("Sent Notifications", "Notification")
			sub <- struct{}{}
			d.Device.RUnlock()
		}
	}()
}

func (d *DomoticMIBAgent) StartAgentUpdater(sub chan struct{}) {
	go func() {
		timer := time.NewTimer(d.updateFrequency)
		for {
			<-timer.C
			timer.Reset(d.updateFrequency)
			d.UpdateSensorValues()
			d.UpdateDevice()
			sub <- struct{}{}
		}
	}()
}

func (d *DomoticMIBAgent) UpdateSensorValues() {
	for _, entry := range d.Sensors.Objects {
		changed, logStr := entry.(SensorsEntry).UpdateValues(d.Actuators)
		if changed {
			d.Device.Objects.(DeviceObjects).UpdateLastTimeChanged()
			d.Logger.LogInfo(logStr, "Sensor Update")
		}
	}
}

func (d *DomoticMIBAgent) UpdateDevice() {
	uptime := d.GetUptime()
	d.Device.Objects.(DeviceObjects).UpdateTimes(uptime)
}

func (d *DomoticMIBAgent) ListenForRequests(sub chan struct{}) {
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
				d.MIB.Logger.LogError(e.Error(), "Request")
			}()
			continue
		}
		go d.MIB.HandleRequest(newPacket, *addr, sub, d)
	}
}

func (d *DomoticMIBAgent) HandleGet(r packet.LSNMPvS_Packet, addr *net.UDPAddr) (*packet.LSNMPvS_Packet, error, bool) {
	structureLengths := d.GetStructureLengths()
	list, allOK := r.GetUncompressedIdValuePairList(structureLengths)
	if !allOK {
		return packet.PacketErr(packet.ErrorInvalidGroupIndexes).Compile(r)
	}
	respList := make([]types.IdValuePair, len(list))
	for i, idValuePair := range list {
		iid := idValuePair.IID.Value.(*CodableValues.IID)
		value, pErr := d.Get(iid.Structure, iid.Object, iid.FirstIndex)
		if pErr != 0 {
			return pErr.Compile(r)
		}
		respList[i] = value
	}
	return r.NewResponsePacket(respList, d.GetUptime()), nil, true
}

func (d *DomoticMIBAgent) HandleSet(r packet.LSNMPvS_Packet, addr *net.UDPAddr) (*packet.LSNMPvS_Packet, error, bool) {
	structureLengths := d.GetStructureLengths()
	list, allOK := r.GetUncompressedIdValuePairList(structureLengths)
	if !allOK {
		return packet.PacketErr(packet.ErrorInvalidGroupIndexes).Compile(r)
	}
	respList := make([]types.IdValuePair, len(list))
	for i, idValuePair := range list {
		iid := idValuePair.IID.Value.(*CodableValues.IID)
		pErr := d.Set(iid.Structure, iid.Object, iid.FirstIndex, *idValuePair.Value)
		if pErr != 0 {
			return pErr.Compile(r)
		}
		respList[i] = idValuePair
	}
	d.Device.Objects.(DeviceObjects).UpdateLastTimeChanged()
	return r.NewResponsePacket(respList, d.GetUptime()), nil, true
}

func (d *DomoticMIBAgent) HandleResponse(r packet.LSNMPvS_Packet, addr *net.UDPAddr) (*packet.LSNMPvS_Packet, error, bool) {
	return nil, nil, false
}

func (d *DomoticMIBAgent) HandleNotification(r packet.LSNMPvS_Packet, addr *net.UDPAddr) (*packet.LSNMPvS_Packet, error, bool) {
	return nil, nil, false
}

func (d *DomoticMIBAgent) RenderMIBWithLipgloss(width int, height int, controls []string, renderLogs bool) string {
	structures := []mib.StructureI{d.Device, d.Sensors, d.Actuators}
	title := lipgloss.NewStyle().Align(lipgloss.Center).Render("Domotic MIB Agent - " + d.Name)
	commandsStyle := lipgloss.NewStyle().Align(lipgloss.Center).Foreground(lipgloss.Color("248"))
	comStr := commandsStyle.Render(strings.Join(controls, " • "))
	rendered := ""
	for _, structure := range structures {
		rendered = lipgloss.JoinVertical(lipgloss.Center, rendered, structure.RenderTableWithLipGloss(width-4))
	}
	lines := height - 32
	if renderLogs {
		rendered = lipgloss.JoinVertical(lipgloss.Left, rendered, d.Logger.RenderLogsWithLipGloss(width-4, lines))
	}
	return lipgloss.JoinVertical(lipgloss.Center, title, lipgloss.NewStyle().Width(width-2).Align(0.5).Border(lipgloss.RoundedBorder()).Render(rendered), comStr)
}

func (d *DomoticMIBAgent) RenderPacketsWithLipgloss(width int, height int, controls []string) string {
	commandsStyle := lipgloss.NewStyle().Align(lipgloss.Center).Foreground(lipgloss.Color("248"))
	comStr := commandsStyle.Render(strings.Join(controls, " • "))
	title := lipgloss.NewStyle().Align(lipgloss.Center).Render("Domotic MIB Agent - " + d.Name + " - Packets")
	rendered := d.MIB.Packets.RenderPacketsWithLipGloss(width-4, height-4)
	return lipgloss.JoinVertical(lipgloss.Center, title, lipgloss.NewStyle().Width(width-2).Height(height-4).Align(lipgloss.Bottom).Border(lipgloss.RoundedBorder()).Render(rendered), comStr)
}

func (d *DomoticMIBAgent) RefreshAgent(addr string) {
	iidList := types.CodableList{}
	for i := 1; i <= 3; i++ {
		s := d.MIB.Structures[i]
		for j := 1; j <= s.Len(); j++ {
			iidList.Add(i, types.NewCodableIID(i, j, []int{0}))
		}
	}
	p := packet.NewGetRequestPacket(iidList)
	netfuncs.SendStrAddr(addr, []byte(p.Encode()))
}