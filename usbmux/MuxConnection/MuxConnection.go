package MuxConnection

import (
	"log"
	"net"
	"runtime"
	"time"

	"github.com/Mitchell-Riley/mux/usbmux/BinaryProtocol"
	"github.com/Mitchell-Riley/mux/usbmux/MuxDevice"
	"github.com/Mitchell-Riley/mux/usbmux/SafeStreamSocket"
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

	if runtime.GOOS == "windows" {
		network = "tcp"
		address = "127.0.0.1:27015"
	} else {
		network = "unix"
		address = socketpath
	}

	s := SafeStreamSocket.New(network, address)

	var devices []*MuxDevice.MuxDevice
	return &MuxConnection{socketpath, s, BinaryProtocol.New(s), 1, devices}
}

func (m *MuxConnection) _getreply() (byte, map[string]interface{}) {
	for {
		resp, tag, data := m.proto.GetPacket()

		if resp == BinaryProtocol.TypeResult {
			return tag, data
		}

		log.Fatalf("Invalid packet type received: %d", resp)
	}
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
		log.Fatalf("Unexpeted result: %d", resp)
	default:
		log.Fatalf("Invalid packet type received %d", resp)
	}
}

// if payload is nil, what happens?
// need to intialize it maybe?
func (m *MuxConnection) _exchange(req int, payload map[string]interface{}) int {
	mytag := m.pkttag
	m.pkttag++

	m.proto.SendPacket(req, mytag, payload)
	recvtag, data := m._getreply()

	if int(recvtag) != mytag {
		log.Fatalf("Reply tag mismatch: expected %d, got %d", mytag, recvtag)
	}

	return data["Number"].(int)
}

func (m *MuxConnection) Listen() {
	ret := m._exchange(BinaryProtocol.TypeListen, nil)
	if ret != 0 {
		log.Fatalf("Listen failed: error %d", ret)
	}
}

func (m *MuxConnection) Process(timeout time.Duration) {
	if m.proto.Connected {
		log.Fatal("Socket is connected, cannot process listener events")
	}

	m._processpacket()

	if _, ok := <-time.After(timeout); ok == true {
		err := m.socket.Sock.Close()
		if err != nil {
			log.Fatalf("Exception in listener socket (channel timed out), ", err)
		}
		// defer m.socket.Sock.Close()
	}
}

func (m *MuxConnection) Connect(device *MuxDevice.MuxDevice, port int) net.Conn {
	ret := m._exchange(BinaryProtocol.TypeConnect,
		map[string]interface{}{
			"DeviceID":   device.Devid,
			"PortNumber": ((port << 8) & 0xFF00) | (port >> 8),
		})

	if ret != 0 {
		log.Fatalf("Connect failed: error %d", ret)
	}

	m.proto.Connected = true
	return m.socket.Sock
}

func (m *MuxConnection) close() {
	err := m.socket.Sock.Close()
	if err != nil {
		log.Fatal(err)
	}
}
