package domoticmib

import (
	"strconv"
	"time"

	"github.com/eivarin/LSNMPvS-DomoticSystem/mib"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types/CodableValues"
)

type SensorsEntry struct {
	mib.TableEntry
	id               mib.Object
	sensorType       mib.Object
	status           mib.Object
	minValue         mib.Object
	maxValue         mib.Object
	lastSamplingTime mib.Object
	virtual          struct {
		gradientChange  bool
		factor          int
		actuatorGetInfo struct {
			Object int
			Index  int
		}
	}
}

func (s SensorsEntry) CheckNewValueValidity(value types.CompleteCodableValue) packet.PacketErr {
	return 0
}

func (s SensorsEntry) Copy() mib.TableEntryI {
	newEntry := SensorsEntry{
		id:               *s.id.Copy(),
		sensorType:       *s.sensorType.Copy(),
		status:           *s.status.Copy(),
		minValue:         *s.minValue.Copy(),
		maxValue:         *s.maxValue.Copy(),
		lastSamplingTime: *s.lastSamplingTime.Copy(),
	}
	newEntry.virtual.gradientChange = s.virtual.gradientChange
	newEntry.virtual.factor = s.virtual.factor
	newEntry.virtual.actuatorGetInfo.Object = s.virtual.actuatorGetInfo.Object
	newEntry.virtual.actuatorGetInfo.Index = s.virtual.actuatorGetInfo.Index
	newEntry.TableEntry = mib.NewTableEntry([]*mib.Object{&newEntry.id, &newEntry.sensorType, &newEntry.status, &newEntry.minValue, &newEntry.maxValue, &newEntry.lastSamplingTime})
	return newEntry
}

func (s SensorsEntry) GetTableEntry() mib.TableEntry {
	return s.TableEntry
}

type SensorConfig struct {
	ID       string `yaml:"ID"`
	Type     string `yaml:"Type"`
	Status   int    `yaml:"Status"`
	MinValue int    `yaml:"MinValue"`
	MaxValue int    `yaml:"MaxValue"`
	Virtual  struct {
		GradientChange  bool `yaml:"GradientChange"`
		Factor          int  `yaml:"Factor"`
		ActuatorGetInfo struct {
			Object int `yaml:"Object"`
			Index  int `yaml:"Index"`
		} `yaml:"ActuatorGetInfo"`
	} `yaml:"Virtual"`
}

func NewSensorsEntry(c SensorConfig) SensorsEntry {
	entry := SensorsEntry{
		id:               mib.NewObject("id", 1, "Tag identifying the sensor (the MacAddress, for example).", false, *types.NewCodableString(c.ID)),
		sensorType:       mib.NewObject("type", 2, "Text description for the type of sensor (“Light”, for example).", false, *types.NewCodableString(c.Type)),
		status:           mib.NewObject("status", 3, "Last value sampled by the sensor in percentage of the interval between minValue and maxValue.", false, *types.NewCodableInt(c.Status)),
		minValue:         mib.NewObject("minValue", 4, "Minimum value possible for the sampling values of the sensor.", false, *types.NewCodableInt(c.MinValue)),
		maxValue:         mib.NewObject("maxValue", 5, "Maximum value possible for the sampling values of the sensor.", false, *types.NewCodableInt(c.MaxValue)),
		lastSamplingTime: mib.NewObject("lastSamplingTime", 6, "Time elapsed since the last sample was obtained by the sensor.", false, *types.NewCodableTimestampNow()),
	}
	entry.virtual.gradientChange = c.Virtual.GradientChange
	entry.virtual.factor = c.Virtual.Factor
	entry.virtual.actuatorGetInfo.Object = c.Virtual.ActuatorGetInfo.Object
	entry.virtual.actuatorGetInfo.Index = c.Virtual.ActuatorGetInfo.Index
	entry.TableEntry = mib.NewTableEntry([]*mib.Object{&entry.id, &entry.sensorType, &entry.status, &entry.minValue, &entry.maxValue, &entry.lastSamplingTime})
	return entry
}

func NewSensorsTable(c []SensorConfig) *mib.Table {
	sensorsTable := &mib.Table{
		Structure: mib.NewStructure("sensors", 2, "Table with information for all types of sensors connected to the device."),
		Columns:   NewSensorsEntry(SensorConfig{}),
		Objects:   []mib.TableEntryI{},
	}
	for _, sensor := range c {
		sensorsTable.AddRow(NewSensorsEntry(sensor))
	}
	return sensorsTable
}

func (s SensorsEntry) UpdateValues(Actuators *mib.Table) (bool, string) {
	aValue, _ := Actuators.Get(s.virtual.actuatorGetInfo.Object, s.virtual.actuatorGetInfo.Index-1)
	actuatorValue := aValue.Value.(*CodableValues.CodableInt).Value
	status := s.status
	status.Lock.Lock()
	currentValue := s.status.Value.Value.(*CodableValues.CodableInt)
	oldValue := currentValue.Value
	changed := false
	if s.virtual.gradientChange {
		if currentValue.Value < actuatorValue {
			currentValue.Value += s.virtual.factor
			changed = true
		} else if currentValue.Value > actuatorValue {
			currentValue.Value -= s.virtual.factor
			changed = true
		}
	} else {
		currentValue.Value = actuatorValue*s.virtual.factor
		changed = true
	}
	status.Lock.Unlock()
	logStr := ""
	changed = changed && oldValue != currentValue.Value
	if changed {
		logStr = s.id.Value.String() + " updated: " + strconv.Itoa(oldValue) + " -> " + strconv.Itoa(currentValue.Value)
		s.lastSamplingTime.Lock.Lock()
		s.lastSamplingTime.Value.Value.(*CodableValues.Timestamp).Ts = time.Now()
		s.lastSamplingTime.Lock.Unlock()
	}
	return changed, logStr
}

// sensors.id OBJECT {
// TYPE String
// ACESS read-only
// DESCRIPTION "Tag identifying the sensor (the MacAddress, for example)."
// IID 2.1 }

// sensors.type OBJECT {
// TYPE String
// ACESS read-only
// DESCRIPTION "Text description for the type of sensor (“Light”, for example)."
// IID 2.2 }

// sensors.status OBJECT {
// TYPE Integer
// ACESS read-only
// DESCRIPTION "Last value sampled by the sensor in percentage of the interval between
// minValue and maxValue."
// IID 2.3 }

// sensors.minValue OBJECT {
// TYPE Integer
// ACESS read-only
// DESCRIPTION "Minimum value possible for the sampling values of the sensor."
// IID 2.4 }

// sensors.maxValue OBJECT {
// TYPE Integer
// ACESS read-only
// DESCRIPTION "Maximum value possible for the sampling values of the sensor."
// IID 2.5 }

// sensors.lastSamplingTime OBJECT {
// TYPE Timestamp
// ACESS read-only
// DESCRIPTION "Time elapsed since the last sample was obtained by the sensor."
// IID 2.6 }
