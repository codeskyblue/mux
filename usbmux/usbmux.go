package usbmux

import (
	"net"
	"time"
)

type USBMux struct {
	socketpath string
	listener   *Connection
	Devices    []*Device
	version    int
}

func New(socketpath string) *USBMux {
	if socketpath == "" {
		socketpath = "/var/run/usbmuxd"
	}

	u := &USBMux{socketpath: socketpath, listener: NewConnection(socketpath)}

	u.listener.Listen()
	u.Devices = u.listener.Devices
	return u
}

func (u *USBMux) Process(timeout time.Duration) {
	u.listener.Process(timeout)
}

func (u *USBMux) Connect(device *Device, port int) net.Conn {
	connector := NewConnection(u.socketpath)
	return connector.Connect(device, port)
}
