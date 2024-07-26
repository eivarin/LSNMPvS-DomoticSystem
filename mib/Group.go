package mib

import (
	"time"

	netfuncs "github.com/eivarin/LSNMPvS-DomoticSystem/NetFuncs"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types/CodableValues"
)

type GroupObjectsI interface {
	Get(objectIID, index int) (*types.CompleteCodableValue, packet.PacketErr)
	GetGroupObjects() GroupObjects
	Set(objectIID int, index int, value types.CompleteCodableValue) packet.PacketErr
	Update(objectIID int, index int, value types.CompleteCodableValue)
	CheckNewValueValidity(objectIID, index int, value types.CompleteCodableValue) packet.PacketErr
}

type GroupObjects map[int][]*Object

func (g GroupObjects) Get(objectIID, index int) (*types.CompleteCodableValue, packet.PacketErr) {
	if ok := g[objectIID]; ok != nil {
		return g[objectIID][index].Value.Copy(), 0
	} else {
		return nil, packet.ErrorObjectIdDoesntExist
	}
}

func (g GroupObjects) Set(objectIID int, index int, value types.CompleteCodableValue) packet.PacketErr {
	if ok := g[objectIID]; ok != nil {
		return g[objectIID][index].Set(value)
	} else {
		return packet.ErrorObjectIdDoesntExist
	}
}

func (g GroupObjects) Update(objectIID int, index int, value types.CompleteCodableValue) {
	if groupObjects, ok := g[objectIID]; ok{
		for index >= len(groupObjects) {
			groupObjects = append(groupObjects, groupObjects[0].Copy())
		}
		groupObjects[index].Update(value)
	}
}

func NewGroupObjects(Objects []*Object) GroupObjects {
	newObjects := make(GroupObjects)
	for _, object := range Objects {
		newObjects[object.ObjectIID] = append(newObjects[object.ObjectIID], object)
	}
	return newObjects
}

type Group struct {
	Structure
	Objects              GroupObjectsI
	HasNotifications     bool
	NotificationsObjects []int
	NotificationRateOid  int
}

func (g *Group) Get(objectIID, index int) (*types.CompleteCodableValue, packet.PacketErr) {
	return g.Objects.Get(objectIID, index)
}

func (g *Group) Set(objectIID, index int, value types.CompleteCodableValue) packet.PacketErr {
	if err := g.Objects.CheckNewValueValidity(objectIID, index, value); err != 0 {
		return err
	}
	return g.Objects.Set(objectIID, index, value)
}

func (g *Group) Update(objectIID, index int, value types.CompleteCodableValue) {
	g.Objects.Update(objectIID, index, value)
}

func (g *Group) PopulateObjectIDWithLength(objectIID int, length int){
}

func (g *Group) GetStructureName() string {
	return g.Name
}

func (g *Group) GetStructureIID() int {
	return g.StructureIID
}

func (g *Group) GetDescription() string {
	return g.Description
}

func (g *Group) Len() int {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return len(g.Objects.GetGroupObjects())
}

func (g *Group) Count(objectIID int) int {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return len(g.Objects.GetGroupObjects()[objectIID])
}

func (g *Group) RenderTableWithLipGloss(width int) string {
	Titles := make([]string, 0)
	Values := make([][]string, 0)
	gos := g.Objects.GetGroupObjects()
	leng := len(gos)
	var row []string
	for i := 1; i <= leng; i++ {
		for _, object := range gos[i] {
			Titles = append(Titles, object.Name)
			row = append(row, object.Value.String())
		}
	}
	Values = append(Values, row)
	return g.renderStructureTableWithLipGloss(Titles, Values, width)
}

func (g *Group) SendNotifications(uptime *types.CompleteCodableValue) {
	Entrys := make([]types.IdValuePair, len(g.NotificationsObjects))
	for i, objectIID := range g.NotificationsObjects {
		val, _ := g.Get(objectIID, 0)
		Entrys[i] = types.IdValuePair{
			IID:   types.NewCodableIID(g.StructureIID, objectIID, nil),
			Value: val,
		}
	}
	// fmt.Printf("Sending notifications: %v\n", Entrys)
	p := packet.NewNotificationPacket(Entrys, uptime)
	encStr := p.Encode()
	netfuncs.SendBroadcast(12345, []byte(encStr))
}

func (g *Group) GetNotificationRate() time.Duration {
	g.lock.RLock()
	defer g.lock.RUnlock()
	notiRateCodable, _ := g.Get(g.NotificationRateOid, 0)
	return time.Duration(int64(notiRateCodable.Value.(*CodableValues.CodableInt).Value)) * time.Second
}

func (g *Group) GetDimensions() map[int]int {
	g.lock.RLock()
	defer g.lock.RUnlock()
	dimensions := make(map[int]int)
	for objectIID, objects := range g.Objects.GetGroupObjects() {
		dimensions[objectIID] = len(objects)
	}
	return dimensions
}