package domoticmib

import (
	"time"

	"github.com/eivarin/LSNMPvS-DomoticSystem/mib"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types/CodableValues"
)

type ActuatorsEntry struct {
	mib.TableEntry
	Id              mib.Object
	ActuatorType    mib.Object
	Status          mib.Object
	MinValue        mib.Object
	MaxValue        mib.Object
	LastControlTime mib.Object
}
func (a ActuatorsEntry) Set(objectIID int, value types.CompleteCodableValue) packet.PacketErr {
	res := a.TableEntry.Set(objectIID, value)
	if res == 0 {
		a.LastControlTime.Value.Value.(*CodableValues.Timestamp).Ts = time.Now()
	}
	return res
}

func (a ActuatorsEntry) CheckNewValueValidity(value types.CompleteCodableValue) packet.PacketErr {
	var valToCheck, maxValue, minValue int
	if value.DataType != 'I'{
		return packet.ErrorInvalidDataType
	}
	valToCheck = value.Value.(*CodableValues.CodableInt).Value
	maxValue = a.MaxValue.Value.Value.(*CodableValues.CodableInt).Value
	minValue = a.MinValue.Value.Value.(*CodableValues.CodableInt).Value
	if valToCheck > maxValue || valToCheck < minValue {
		return packet.ErrorValueOutOfRange
	}
	return 0
}

func (a ActuatorsEntry) Copy() mib.TableEntryI {
	newEntry := ActuatorsEntry{
		Id:              *a.Id.Copy(),
		ActuatorType:    *a.ActuatorType.Copy(),
		Status:          *a.Status.Copy(),
		MinValue:        *a.MinValue.Copy(),
		MaxValue:        *a.MaxValue.Copy(),
		LastControlTime: *a.LastControlTime.Copy(),
	}
	newEntry.TableEntry = mib.NewTableEntry([]*mib.Object{&newEntry.Id, &newEntry.ActuatorType, &newEntry.Status, &newEntry.MinValue, &newEntry.MaxValue, &newEntry.LastControlTime})
	return newEntry
}

func (a ActuatorsEntry) GetTableEntry() mib.TableEntry {
	return a.TableEntry
}

type ActuatorConfig struct {
	ID       string `yaml:"ID"`
	Type     string `yaml:"Type"`
	Status   int    `yaml:"Status"`
	MinValue int    `yaml:"MinValue"`
	MaxValue int    `yaml:"MaxValue"`
}

func NewActuatorsEntry(c ActuatorConfig) ActuatorsEntry {
	entry := ActuatorsEntry{
		Id:              mib.NewObject("id", 1, "Tag identifying the actuator (the MacAddress, for example).", false, *types.NewCodableString(c.ID)),
		ActuatorType:    mib.NewObject("actuatorType", 2, "Text description for the type of actuator (“Temperature”, for example).", false, *types.NewCodableString(c.Type)),
		Status:          mib.NewObject("status", 3, "Configuration value set for the actuator (value must be between minValue and maxValue).", true, *types.NewCodableInt(c.Status)),
		MinValue:        mib.NewObject("minValue", 4, "Minimum value possible for the configuration of the actuator.", false, *types.NewCodableInt(c.MinValue)),
		MaxValue:        mib.NewObject("maxValue", 5, "Maximum value possible for the configuration of the actuator.", false, *types.NewCodableInt(c.MaxValue)),
		LastControlTime: mib.NewObject("lastControlTime", 6, "Date and time when the last configuration/control operation was executed.", false, *types.NewCodableTimestampNow()),
	}
	entry.TableEntry = mib.NewTableEntry([]*mib.Object{&entry.Id, &entry.ActuatorType, &entry.Status, &entry.MinValue, &entry.MaxValue, &entry.LastControlTime})
	return entry
}

func NewActuatorsTable(c []ActuatorConfig) *mib.Table {
	actuatorsTable := &mib.Table{
		Structure: mib.NewStructure("Actuators", 3, "Table with objects to control all actuators connected to the device."),
		Columns:   NewActuatorsEntry(ActuatorConfig{}),
		Objects:   []mib.TableEntryI{},
	}
	for _, actuator := range c {
		actuatorsTable.AddRow(NewActuatorsEntry(actuator))
	}
	return actuatorsTable
}

// actuators OBJECT {
// TYPE Table
// INCLUDE id, type, status, minValue, maxValue, lastControlTime
// DESCRIPTION "Table with objects to control all actuators connected to the device."
// IID 3 }

// actuators.id OBJECT {
// TYPE String
// ACESS read-only
// DESCRIPTION "Tag identifying the actuator (the MacAddress, for example)."
// IID 3.1 }

// actuators.type OBJECT {
// TYPE String
// ACESS read-only
// DESCRIPTION "Text description for the type of actuator (“Temperature”, for example)."
// IID 3.2 }

// actuators.status OBJECT {
// TYPE Integer
// ACESS read-write
// DESCRIPTION "Configuration value set for the actuator (value must be between minValue and
// maxValue)."
// IID 3.3 }

// actuators.minValue OBJECT {
// TYPE Integer
// ACESS read-only
// DESCRIPTION "Minimum value possible for the configuration of the actuator."
// IID 3.4 }

// actuators.maxValue OBJECT {
// TYPE Integer
// ACESS read-only
// DESCRIPTION "Maximum value possible for the configuration of the actuator."
// IID 3.5 }

// actuators.lastControlTime OBJECT {
// TYPE Timestamp
// ACESS read-only
// DESCRIPTION "Date and time when the last configuration/control operation was executed."
// IID 3.6 }
