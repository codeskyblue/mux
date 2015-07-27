package MuxDevice

import (
	"fmt"
)

type MuxDevice struct {
	Devid    float32
	usbprod  byte
	serial   string
	location byte
}

func New(devid float32, usbprod byte, serial string, location byte) *MuxDevice {
	return &MuxDevice{devid, usbprod, serial, location}
}

func (m *MuxDevice) Fields() string {
	return fmt.Sprintf("<MuxDevice: ID %g ProdID 0x%4x Serial '%s' Location 0x%x>", m.Devid, m.usbprod, m.serial, m.location)
}
