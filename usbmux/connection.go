package usbmux

import (
	"log"
	"net"
	"runtime"
	"time"
)

type Connection struct {
	socketpath string
	socket     *SafeStreamSocket
	proto      *BinaryProtocol
	pkttag     int
	Devices    []*Device
}

func NewConnection(socketpath string) *Connection {
	var address, network string

	if runtime.GOOS == "windows" {
		network = "tcp"
		address = "127.0.0.1:27015"
	} else {
		network = "unix"
		address = socketpath
	}

	s := NewSafeStreamSocket(network, address)

	devices := []*Device{}
	return &Connection{socketpath, s, NewBinaryProtocol(s), 1, devices}
}

func (c *Connection) _getreply() (byte, map[string]interface{}) {
	for {
		resp, tag, data := c.proto.GetPacket()

		if resp == TypeResult {
			return tag, data
		}

		log.Fatalf("Invalid packet type received: %d", resp)
	}
}

// this function is disgusting
func (c *Connection) _processpacket() {
	// tag not needed?
	resp, _, data := c.proto.GetPacket()

	switch resp {
	case TypeDeviceAdd:
		// welcome to assertion hell
		// this is literally hitler code
		c.Devices = append(c.Devices,
			NewDevice(
				data["DeviceID"].(float32),
				data["Properties"].(map[string]byte)["ProductID"],
				data["Properties"].(map[string]string)["SerialNumber"],
				data["Properties"].(map[string]byte)["LocationID"]))
	case TypeDeviceRemove:
		for i, v := range c.Devices {
			if v.Devid == data["DeviceID"] {
				// deletes an element from the map
				c.Devices = append(c.Devices[:i], c.Devices[i+1:]...)
			}
		}
	case TypeResult:
		log.Fatalf("Unexpeted result: %d", resp)
	default:
		log.Fatalf("Invalid packet type received %d", resp)
	}
}

// if payload is nil, what happens?
// need to intialize it maybe?
func (c *Connection) _exchange(req int, payload map[string]interface{}) int {
	mytag := c.pkttag
	c.pkttag++

	c.proto.SendPacket(req, mytag, payload)
	recvtag, data := c._getreply()

	if int(recvtag) != mytag {
		log.Fatalf("Reply tag mismatch: expected %d, got %d", mytag, recvtag)
	}

	return data["Number"].(int)
}

func (c *Connection) Listen() {
	ret := c._exchange(TypeListen, nil)
	if ret != 0 {
		log.Fatalf("Listen failed: error %d", ret)
	}
}

func (c *Connection) Process(timeout time.Duration) {
	if c.proto.Connected {
		log.Fatal("Socket is connected, cannot process listener events")
	}

	// if you can read from the socket, c._processpacket()
	// else, close and fatalf
	c._processpacket()

	// defer c.socket.Sock.Close()
	// log.Fatal("Exception in listener socket")
}

func (c *Connection) Connect(device *Device, port int) net.Conn {
	ret := c._exchange(TypeConnect,
		map[string]interface{}{
			"DeviceID":   device.Devid,
			"PortNumber": ((port << 8) & 0xFF00) | (port >> 8),
		})

	if ret != 0 {
		log.Fatalf("Connect failed: error %d", ret)
	}

	c.proto.Connected = true
	return c.socket.Sock
}

func (c *Connection) close() {
	err := c.socket.Sock.Close()
	if err != nil {
		log.Fatal(err)
	}
}
