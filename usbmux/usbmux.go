package usbmux

import (
	"net"
	"time"
	"usbmux/MuxConnection"
	"usbmux/MuxDevice"
)

type USBMux struct {
	socketpath string
	// protoclass *BinaryProtocol
	listener *MuxConnection.MuxConnection
	Devices  []*MuxDevice.MuxDevice
	version  int
}

func New(socketpath string) *USBMux {
	if socketpath == "" {
		socketpath = "/var/run/usbmuxd"
	}

	u := &USBMux{socketpath: socketpath, listener: MuxConnection.New(socketpath)}
	u.listener.Listen()
	u.Devices = u.listener.Devices
	return u
}

func (u *USBMux) Process(timeout time.Duration) {
	u.listener.Process(timeout)
}

func (u *USBMux) Connect(device *MuxDevice.MuxDevice, port int) net.Conn {
	connector := MuxConnection.New(u.socketpath)
	return connector.Connect(device, port)
}
