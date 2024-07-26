package mib

import (
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types"
)

type TableEntryI interface {
	Get(int) (*types.CompleteCodableValue, packet.PacketErr)
	GetTableEntry() TableEntry
	Set(objectIID int, value types.CompleteCodableValue) packet.PacketErr
	Update(objectIID int, value types.CompleteCodableValue)
	CheckNewValueValidity(value types.CompleteCodableValue) packet.PacketErr
	Copy() TableEntryI
}

type TableEntry map[int]*Object

func (t TableEntry) Get(objectIID int) (*types.CompleteCodableValue, packet.PacketErr) {
	if ok := t[objectIID]; ok != nil {
		return t[objectIID].Value.Copy(), 0
	} else {
		return nil, packet.ErrorObjectIdDoesntExist
	}
}

func (t TableEntry) Set(objectIID int, value types.CompleteCodableValue) packet.PacketErr {
	if ok := t[objectIID]; ok != nil {
		return t[objectIID].Set(value)
	} else {
		return packet.ErrorObjectIdDoesntExist
	}
}

func (t TableEntry) Update(objectIID int, value types.CompleteCodableValue) {
	if ok := t[objectIID]; ok != nil {
		t[objectIID].Update(value)
	}
}

func (t TableEntry) CheckNewValueValidity(value types.CompleteCodableValue) packet.PacketErr {
	return 0
}

func (t TableEntry) GetTableEntry() TableEntry {
	return t
}

func (t TableEntry) Copy() TableEntry{
	newEntry := make(TableEntry)
	for k, o := range t {
		newEntry[k] = o.Copy()
	}
	return newEntry
}

func NewTableEntry(Objects []*Object) TableEntry {
	newEntry := make(TableEntry)
	for _, object := range Objects {
		newEntry[object.ObjectIID] = object
	}
	return newEntry
}

type Table struct {
	Structure
	Columns TableEntryI
	Objects []TableEntryI
}

func (t *Table) Get(objectIID, index int) (*types.CompleteCodableValue, packet.PacketErr) {
	return t.Objects[index].Get(objectIID)
}

func (t *Table) GetStructureName() string {
	return t.Name
}

func (t *Table) GetStructureIID() int {
	return t.StructureIID
}

func (t *Table) GetDescription() string {
	return t.Description
}

func (t *Table) Set(objectIID, index int, value types.CompleteCodableValue) packet.PacketErr {
	// tEntry := t.Objects[correctedIndex]
	// res := tEntry.Set(objectIID, value)
	// if res == 0 {
	// 	t.Objects[correctedIndex] = tEntry
	// }
	if err := t.Objects[index].CheckNewValueValidity(value); err != 0 {
		return err
	}
	return t.Objects[index].Set(objectIID, value)
}

func (t *Table) Update(objectIID, index int, value types.CompleteCodableValue) {
	for index >= len(t.Objects) {
		newEntry := t.Columns.Copy()
		t.AddRow(newEntry)
	}
	t.Objects[index].Update(objectIID, value)
}

func (t *Table) PopulateObjectIDWithLength(objectIID int, length int){
	for i:= t.Count(objectIID) ; i < length; i = t.Count(objectIID) {
		newEntry := t.Columns.Copy()
		t.AddRow(newEntry)
	}
}

func (t *Table) AddRow(newEntry TableEntryI) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.Objects = append(t.Objects, newEntry)
}

func (t *Table) RenderTableWithLipGloss(width int) string {
	var (
		Titles []string
		Values [][]string
	)
	columnsTableEntry := t.Columns.GetTableEntry()
	leng := len(columnsTableEntry)
	for i := 1; i <= leng; i++ {
		Titles = append(Titles, columnsTableEntry[i].Name)
	}
	for _, entry := range t.Objects {
		var row []string
		tEntry := entry.GetTableEntry()
		for j := 1; j <= leng; j++ {
			v, _ := tEntry.Get(j)
			row = append(row, v.String())
		}
		Values = append(Values, row)
	}
	return t.renderStructureTableWithLipGloss(Titles, Values, width)
}

func (t *Table) Len() int {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return len(t.Columns.GetTableEntry())
}

func (t *Table) Count(objectIID int) int {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return len(t.Objects)
}

func (t *Table) GetDimensions() map[int]int {
	t.lock.RLock()
	defer t.lock.RUnlock()
	res := make(map[int]int)
	for _, entry := range t.Columns.GetTableEntry() {
		res[entry.ObjectIID] = t.Count(entry.ObjectIID)
	}
	return res
}
