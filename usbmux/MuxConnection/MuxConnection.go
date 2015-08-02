package MuxConnection

import (
	"fmt"
	"net"
	"runtime"
	"time"
	"usbmux/BinaryProtocol"
	"usbmux/MuxDevice"
	"usbmux/SafeStreamSocket"
)

type MuxConnection struct {
	socketpath string
	socket     *SafeStreamSocket.SafeStreamSocket
	proto      *BinaryProtocol.BinaryProtocol
	pkttag     int
	Devices    []*MuxDevice.MuxDevice
}

func New(socketpath string) *MuxConnection {
	var address, network string

	// get this straightened out
	if runtime.GOOS == "windows" {
		address = "127.0.0.1:27015"
		network = "ip4"
	} else {
		address = socketpath
		network = "unix"
	}

	s := SafeStreamSocket.New(network, address)

	return &MuxConnection{socketpath, s, BinaryProtocol.New(s), 1, nil}
}

func (m *MuxConnection) _getreply() (interface{}, map[string]interface{}) {
	for true {
		resp, tag, data := m.proto.GetPacket()

		if resp == BinaryProtocol.TypeResult {
			return tag, data
		}

		panic(fmt.Sprintf("Invalid packet type received: %d", resp))
	}
	return nil, nil
}

// this function is disgusting
func (m *MuxConnection) _processpacket() {
	// tag not needed?
	resp, _, data := m.proto.GetPacket()

	switch resp {
	case BinaryProtocol.TypeDeviceAdd:
		// welcome to assertion hell
		// this is literally hitler code
		m.Devices = append(m.Devices, MuxDevice.New(data["DeviceID"].(float32), data["Properties"].(byte), data["Properties"].(map[string]string)["SerialNumber"], data["Properties"].(map[string]byte)["LocationID"]))
	case BinaryProtocol.TypeDeviceRemove:
		for i, v := range m.Devices {
			if v.Devid == data["DeviceID"] {
				// deletes an element from the map
				m.Devices = append(m.Devices[:i], m.Devices[i+1:]...)
			}
		}
	case BinaryProtocol.TypeResult:
		panic(fmt.Sprintf("Unexpeted result: %d", resp))
	default:
		panic(fmt.Sprintf("Invalid packet type received %d", resp))
	}
}

func (m *MuxConnection) _exchange(req int, payload map[string]interface{}) interface{} {
	m.pkttag++

	m.proto.SendPacket(req, m.pkttag, payload)
	recvtag, data := m._getreply()

	if recvtag != m.pkttag {
		panic(fmt.Sprintf("Reply tag mismatch: expected %d, got %d", m.pkttag, recvtag))
	}

	return data["Number"].(string)
}

func (m *MuxConnection) Listen() {
	ret := m._exchange(BinaryProtocol.TypeListen, nil)
	if ret != nil {
		panic(fmt.Sprintf("Listen failed: error %d", ret))
	}
}

func (m *MuxConnection) Process(timeout time.Duration) {
	if m.proto.Connected {
		panic("Socket is connected, cannot process listener events")
	}
	var ch chan net.Conn

	ch <- m.socket.Sock
	ch <- nil
	ch <- m.socket.Sock

	select {
	case v := <-ch:
		if v == m.socket.Sock {
			m._processpacket()
		}
	// not sure if this is a correct implementation
	case <-time.After(timeout):
		err := m.socket.Sock.Close()
		if err != nil {
			panic(fmt.Sprintln("Exception in listener socket (channel timed out), ", err))
		}
		// defer close(ch)

		panic(fmt.Sprintln("Exception in listener socket (channel timed out)"))
	}
}

func (m *MuxConnection) Connect(device *MuxDevice.MuxDevice, port int) net.Conn {
	payload := map[string]interface{}{
		"DeviceID":   device.Devid,
		"PortNumber": ((port << 8) & 0xFF00) | (port >> 8),
	}

	ret := m._exchange(BinaryProtocol.TypeConnect, payload)
	if ret != 0 {
		panic(fmt.Sprintf("Connect failed: error %d", ret))
	}

	m.proto.Connected = true
	return m.socket.Sock
}

func (m *MuxConnection) close() {
	err := m.socket.Sock.Close()
	if err != nil {
		panic(err)
	}
}
