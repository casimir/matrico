package data

import (
	"fmt"
)

type Device struct {
	DeviceID   string
	Properties map[string]interface{}
}

func (n *Device) Label() string                 { return "Device" }
func (n *Device) Key() string                   { return "deviceId" }
func (n *Device) KeyVal() interface{}           { return n.DeviceID }
func (n *Device) Props() map[string]interface{} { return n.Properties }

func NewDevice(deviceId string) *Device {
	return &Device{DeviceID: deviceId}
}

func NewDeviceFromToken(d *DataGraph, token string) (*Device, error) {
	q := fmt.Sprintf(
		`MATCH (d:Device) WHERE d.token = '%s' RETURN d.deviceId`,
		token,
	)
	res, err := d.Query(q)
	defer res.Close()
	if err != nil {
		return nil, err
	}
	if !res.Next() {
		return nil, nil
	}

	r := res.Record()
	deviceID, _ := r.Get("d.deviceId")
	return &Device{DeviceID: deviceID.(string)}, nil
}
