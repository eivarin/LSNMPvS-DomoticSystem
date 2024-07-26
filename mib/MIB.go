package mib

import (
	"fmt"
	"net"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/eivarin/LSNMPvS-DomoticSystem/CustomLogger"
	netfuncs "github.com/eivarin/LSNMPvS-DomoticSystem/NetFuncs"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types/CodableValues"
)

type RecPacketList struct {
	packets     *[]packet.LSNMPvS_Packet
	packetsByID map[string]packet.LSNMPvS_Packet
	lock        *sync.RWMutex
}

func NewRecPacketList() RecPacketList {
	return RecPacketList{
		packets:     &[]packet.LSNMPvS_Packet{},
		packetsByID: make(map[string]packet.LSNMPvS_Packet),
		lock:        &sync.RWMutex{},
	}
}

func (r *RecPacketList) AddPacket(p packet.LSNMPvS_Packet) packet.PacketErr {
	r.lock.Lock()
	defer r.lock.Unlock()
	mId := p.GetMessageID()
	if _, ok := r.packetsByID[p.GetMessageID()]; ok {
		return packet.ErrorDuplicateMessageId
	}
	r.packetsByID[mId] = p
	*r.packets = append(*r.packets, p)
	return 0
}

func (r *RecPacketList) RenderPacketsWithLipGloss(width int, height int) string {
	r.lock.RLock()
	defer r.lock.RUnlock()
	result := ""
	for _, packet := range *r.packets {
		result = lipgloss.JoinVertical(lipgloss.Center, result, packet.RenderPacketWithLipgloss(width))
	}
	splitted:= strings.Split(result, "\n")
	slices.Reverse(splitted)
	reversedStr := strings.Join(splitted, "\n")
	reversedStr = lipgloss.NewStyle().AlignHorizontal(1).MaxHeight(height).Render(reversedStr)
	splitted = strings.Split(reversedStr, "\n")
	slices.Reverse(splitted)
	reversedStr = strings.Join(splitted, "\n")
	return reversedStr
}

type HandlerI interface {
	HandleGet(r packet.LSNMPvS_Packet, addr *net.UDPAddr) (*packet.LSNMPvS_Packet, error, bool)
	HandleSet(r packet.LSNMPvS_Packet, addr *net.UDPAddr) (*packet.LSNMPvS_Packet, error, bool)
	HandleResponse(r packet.LSNMPvS_Packet, addr *net.UDPAddr) (*packet.LSNMPvS_Packet, error, bool)
	HandleNotification(r packet.LSNMPvS_Packet, addr *net.UDPAddr) (*packet.LSNMPvS_Packet, error, bool)
}

type MIB struct {
	Structures map[int]StructureI
	Groups     []*Group
	Tables     []*Table
	Logger     *CustomLogger.CustomLogger
	Packets    RecPacketList
	StartTime  time.Time
}

func NewMIB(logger *CustomLogger.CustomLogger, structures []StructureI) MIB {
	res := MIB{
		Structures: make(map[int]StructureI),
		Groups:     make([]*Group, 0),
		Tables:     make([]*Table, 0),
		Logger:     logger,
		StartTime:  time.Now(),
		Packets:    NewRecPacketList(),
	}
	for _, structure := range structures {
		switch s := structure.(type) {
		case *Group:
			res.AddGroup(s)
		case *Table:
			res.AddTable(s)
		}
		res.Structures[structure.GetStructureIID()] = structure
	}
	return res
}

func (m *MIB) AddGroup(structure *Group) {
	m.Groups = append(m.Groups, structure)
	m.Structures[structure.StructureIID] = structure
}

func (m *MIB) AddTable(structure *Table) {
	m.Tables = append(m.Tables, structure)
	m.Structures[structure.StructureIID] = structure
}

func (m *MIB) Get(structure, objectIID int, index *int) (types.IdValuePair, packet.PacketErr) {
	var (
		IID         *types.CompleteCodableValue
		ObjectValue *types.CompleteCodableValue
		indexList   []int
		pErr        packet.PacketErr
	)
	pErr = 0
	indexList = nil
	if index != nil {
		indexList = append(indexList, *index)
	}
	IID = types.NewCodableIID(structure, objectIID, indexList)
	if s, ok := m.Structures[structure]; ok {
		if objectIID == 0 {
			if index != nil {
				pErr = packet.ErrorInvalidIID
			} else {
				ObjectValue = types.NewCodableInt(s.Len())
			}
		} else if objectIID > 0 {
			objectLen := s.Count(objectIID)
			if *index == 0 {
				ObjectValue = types.NewCodableInt(objectLen)
			} else if *index > 0 && *index <= objectLen {
				correctedIndex := *index - 1
				ObjectValue, pErr = s.Get(objectIID, correctedIndex)
			} else {
				pErr = packet.ErrorIndexOutOfRange
			}
		} else {
			pErr = packet.ErrorObjectIdDoesntExist
		}
	}
	return types.IdValuePair{
		IID:   IID,
		Value: ObjectValue,
	}, pErr
}

func (m *MIB) Set(structure, objectIID int, index *int, value types.CompleteCodableValue) packet.PacketErr {
	if s, ok := m.Structures[structure]; ok {
		correctedIndex := 0
		if index != nil {
			correctedIndex = *index - 1
		}
		if objectIID == 0 {
			return packet.ErrorInvalidIID
		} else if correctedIndex < 0 || correctedIndex >= s.Count(objectIID) {
			return packet.ErrorIndexOutOfRange
		}

		return s.Set(objectIID, correctedIndex, value)
	}
	return packet.ErrorStructureDoesntExist
}

func (m *MIB) Update(r packet.LSNMPvS_Packet) (*packet.LSNMPvS_Packet, error, bool) {
	needsNewGet := false
	iidListToGet := make(types.CodableList, 0)
	for _, idValuePair := range r.GetIidValuePairList() {
		iid := idValuePair.IID.Value.(*CodableValues.IID)
		value := idValuePair.Value
		if s, ok := m.Structures[iid.Structure]; ok {
			correctedIndex := 0
			if iid.FirstIndex != nil {
				correctedIndex = *iid.FirstIndex - 1
			}
			if correctedIndex == -1 {
				s.PopulateObjectIDWithLength(iid.Object, value.Value.(*CodableValues.CodableInt).Value)
				needsNewGet = true
				numberOfObjects := s.Len()
				for i := 1; i <= numberOfObjects; i++ {
					iidListToGet.Append(types.NewCodableIID(iid.Structure, i, []int{0,0}))
				}
			} else {
				s.Update(iid.Object, correctedIndex, *value)
			}
		}
	}
	if needsNewGet {
		return packet.NewGetRequestPacket(iidListToGet), nil, true
	}
	return nil, nil, false
}

func (m *MIB) StartNotificationLoop(sub chan struct{}) {
	for _, group := range m.Groups {
		if group.HasNotifications {
			go func(g *Group) {
				for {
					time.Sleep(g.GetNotificationRate())
					uptime := m.GetUptime()
					g.SendNotifications(uptime)
					sub <- struct{}{}
				}
			}(group)
		}
	}
}

func (m *MIB) GetUptime() *types.CompleteCodableValue {
	return types.NewCodableDuration(time.Since(m.StartTime))
}

func (m *MIB) HandleRequest(r packet.LSNMPvS_Packet, remAddr net.UDPAddr, sub chan struct{}, handler HandlerI) {
	var (
		handlingErr error
		respPacket  *packet.LSNMPvS_Packet
		respond     bool
	)
	rType, verifyErr := r.VerifyAndGetType()
	duplicatePacketErr := m.Packets.AddPacket(r)
	if verifyErr == 0 && duplicatePacketErr == 0 {
		var handlingFunc func(r packet.LSNMPvS_Packet, addr *net.UDPAddr) (*packet.LSNMPvS_Packet, error, bool)
		loggingText := ""
		switch rType {
		case 'G':
			loggingText = "Received Get Packet"
			handlingFunc = handler.HandleGet
		case 'S':
			loggingText = "Received Set Packet"
			handlingFunc = handler.HandleSet
		case 'R':
			loggingText = "Received Response Packet"
			handlingFunc = handler.HandleResponse
		case 'N':
			loggingText = "Received Notification Packet"
			handlingFunc = handler.HandleNotification
		default:
			m.Logger.LogError("Packet type error that should not happen(Unhandled Valid Response Type)", "Request")
			return
		}
		m.Logger.LogDebug(loggingText, "Request")
		respPacket, handlingErr, respond = handlingFunc(r, &remAddr)
	} else {
		var pErr packet.PacketErr
		if verifyErr != 0 {
			pErr = verifyErr
		} else {
			pErr = duplicatePacketErr
		}
		respPacket, handlingErr, respond = pErr.Compile(r)
	}
	reqDescr := fmt.Sprintf("%c from %s", rType, remAddr.String())
	if handlingErr != nil {
		m.Logger.LogError("Error handling "+reqDescr+": "+handlingErr.Error(), "Request")
		return
	}
	if !respond {
		sub <- struct{}{}
		return
	}
	remAddr.Port = 12345
	err := netfuncs.Send(&remAddr, []byte(respPacket.Encode()))
	if err != nil {
		m.Logger.LogError("Error sending response to "+reqDescr+": "+err.Error(), "Request")
		return
	}
	m.Logger.LogInfo(reqDescr+" Handled Successfully", "Request")
	sub <- struct{}{}
}

func (m *MIB) GetStructureLengths() map[int]map[int]int {
	res := make(map[int]map[int]int)
	for _, structure := range m.Structures {
		res[structure.GetStructureIID()] = structure.GetDimensions()
	}
	return res
}
