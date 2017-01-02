package usbmux

import (
	"fmt"
)

type Device struct {
	Devid    float32
	usbprod  byte
	serial   string
	location byte
}

func NewDevice(devid float32, usbprod byte, serial string, location byte) *Device {
	return &Device{devid, usbprod, serial, location}
}

func (d *Device) Fields() string {
	return fmt.Sprintf("<Device: ID %g ProdID 0x%4x Serial '%s' Location 0x%x>", d.Devid, d.usbprod, d.serial, d.location)
}
