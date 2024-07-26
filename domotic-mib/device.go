package domoticmib

import (
	"time"

	"github.com/eivarin/LSNMPvS-DomoticSystem/mib"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types"
	"github.com/eivarin/LSNMPvS-DomoticSystem/packet/types/CodableValues"
)

type DeviceObjects struct {
	mib.GroupObjects
	Id                *mib.Object
	Type              *mib.Object
	BeaconRate        *mib.Object
	NSensors          *mib.Object
	NActuators        *mib.Object
	DateAndTime       *mib.Object
	Uptime            *mib.Object
	LastTimeUpdated   *mib.Object
	OperationalStatus *mib.Object
	Reset             *mib.Object
}

func (d DeviceObjects) GetGroupObjects() mib.GroupObjects {
	return d.GroupObjects
}

func (d DeviceObjects) UpdateTimes(uptime *types.CompleteCodableValue) {
	d.Uptime.Lock.Lock()
	d.Uptime.Value.Value.(*CodableValues.Duration).Value = uptime.Value.(*CodableValues.Duration).Value
	d.Uptime.Lock.Unlock()
	d.DateAndTime.Lock.Lock()
	d.DateAndTime.Value.Value.(*CodableValues.Timestamp).Ts = time.Now()
	d.DateAndTime.Lock.Unlock()
}

func (d DeviceObjects) Set(objectIID, index int, value types.CompleteCodableValue) packet.PacketErr {
	res := d.GroupObjects.Set(objectIID, index, value)
	return res
}

func (d DeviceObjects) UpdateLastTimeChanged() {
	d.LastTimeUpdated.Lock.Lock()
	d.LastTimeUpdated.Value.Value.(*CodableValues.Timestamp).Ts = time.Now()
	d.LastTimeUpdated.Lock.Unlock()
}

func (d DeviceObjects) UpdateOperationalStatus(status int) {
	d.OperationalStatus.Value = *types.NewCodableInt(status)
}

func (d DeviceObjects) CheckNewValueValidity(objectIID, index int, value types.CompleteCodableValue) packet.PacketErr {
	return 0
}

type DeviceConfig struct {
	ID         string `yaml:"ID"`
	Type       string `yaml:"Type"`
	BeaconRate int    `yaml:"BeaconRate"`
	NSensors   int    `yaml:"nSensors"`
	NActuators int    `yaml:"nActuators"`
}

func NewDeviceObjects(c DeviceConfig) DeviceObjects {
	id := mib.NewObject("id", 1, "Tag identifying the device (the MacAddress, for example).", false, *types.NewCodableString(c.ID))
	t := mib.NewObject("type", 2, "Text description for the type of device (“Lights & A/C Conditioning”, for example)", false, *types.NewCodableString(c.Type))
	beaconRate := mib.NewObject("beaconRate", 3, "Frequency rate in seconds for issuing a notification message with information from this group that acts as a beacon broadcasting message to all the managers in the LAN. If value is set to zero the notifications for this group are halted.", true, *types.NewCodableInt(c.BeaconRate))
	nSensors := mib.NewObject("nSensors", 4, "Number of sensors implemented in the device and present in the sensors Table.", false, *types.NewCodableInt(c.NSensors))
	nActuators := mib.NewObject("nActuators", 5, "Number of actuators implemented in the device and present in the actuators Table.", false, *types.NewCodableInt(c.NActuators))
	dateAndTime := mib.NewObject("dateAndTime", 6, "System date and time setup in the device.", true, *types.NewCodableTimestamp(time.Now()))
	upTime := mib.NewObject("upTime", 7, "For how long the device is working since last boot/reset.", false, *types.NewCodableDuration(0))
	lastTimeUpdated := mib.NewObject("lastTimeUpdated", 8, "Date and time of the last update of any object in the device L-MIBvS.", false, *types.NewCodableTimestamp(time.Now()))
	operationalStatus := mib.NewObject("operationalStatus", 9, "The operational state of the device, where the value 0 corresponds to a standby operational state, 1 corresponds to a normal operational state and 2 or greater corresponds to an non-operational error state.", false, *types.NewCodableInt(1))
	reset := mib.NewObject("reset", 10, "Value 0 means no reset and value 1 means a reset procedure must be done.", true, *types.NewCodableInt(0))
	objects := DeviceObjects{
		Id:                &id,
		Type:              &t,
		BeaconRate:        &beaconRate,
		NSensors:          &nSensors,
		NActuators:        &nActuators,
		DateAndTime:       &dateAndTime,
		Uptime:            &upTime,
		LastTimeUpdated:   &lastTimeUpdated,
		OperationalStatus: &operationalStatus,
		Reset:             &reset,
	}
	objects.GroupObjects = mib.NewGroupObjects([]*mib.Object{&id, &t, &beaconRate, &nSensors, &nActuators, &dateAndTime, &upTime, &lastTimeUpdated, &operationalStatus, &reset})
	return objects
}

func NewDeviceGroup(c DeviceConfig) *mib.Group {
	DeviceObject := &mib.Group{
		Structure:            mib.NewStructure("device", 1, "Simple list of objects, where each object represents a characteristic from a domotics device agent"),
		Objects:              NewDeviceObjects(c),
		HasNotifications:     true,
		NotificationsObjects: []int{1, 2, 4, 5, 6, 7, 8, 9},
		NotificationRateOid:  3,
	}
	return DeviceObject
}
