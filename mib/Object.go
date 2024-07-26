package mib

import (
	"sync"

	"github.com/eivarin/LSNMPvS-DomoticSystem/packet"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types"
)

type Object struct {
	Name         string
	StructureIID int
	ObjectIID    int
	Description  string
	AllowWrite   bool
	Lock         *sync.RWMutex
	Value        types.CompleteCodableValue
}

func NewObject(Name string, ObjectIID int, Description string, AllowWrite bool, Value types.CompleteCodableValue) Object {
	return Object{
		Name:        Name,
		ObjectIID:   ObjectIID,
		Description: Description,
		AllowWrite:  AllowWrite,
		Value:       Value,
		Lock:        &sync.RWMutex{},
	}
}

func (o *Object) Get() (*types.CompleteCodableValue, packet.PacketErr) {
	o.Lock.RLock()
	defer o.Lock.RUnlock()
	return o.Value.Copy(), 0
}

func (o *Object) Set(newValue types.CompleteCodableValue) packet.PacketErr {
	if o.AllowWrite {
		o.Update(newValue)
		return 0
	}
	return packet.ErrorChangingReadOnlyValue
}

func (o *Object) Update(newValue types.CompleteCodableValue) {
	o.Lock.Lock()
	defer o.Lock.Unlock()
	o.Value = newValue
}

func (o *Object) Copy() *Object {
	o.Lock.RLock()
	defer o.Lock.RUnlock()
	vCopy := o.Value.Copy()
	return &Object{
		Name:        o.Name,
		ObjectIID:   o.ObjectIID,
		Description: o.Description,
		AllowWrite:  o.AllowWrite,
		Value:       *vCopy,
		Lock:        &sync.RWMutex{},
	}
}
