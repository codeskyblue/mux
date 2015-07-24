package usbmux

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"runtime"
	"strings"
)

type SafeStreamSocket struct {
	sock net.Conn
}

func NewSafeStreamSocket(network, address string) *SafeStreamSocket {
	ret, err := net.Dial(network, address)
	if err != nil {
		fmt.Println(err)
	}

	return &SafeStreamSocket{ret}
}

func (s *SafeStreamSocket) send(msg []byte) {
	var totalsent int

	for totalsent < len(msg) {
		sent, err := s.sock.Write(msg[totalsent:])
		if err != nil {
			fmt.Println(err)
		}
		if sent == 0 {
			panic(fmt.Sprintf("socket connection broken"))
		}
		totalsent = totalsent + sent
	}
}

func (s *SafeStreamSocket) recv(size int) string {
	var msg string
	payload := []byte{byte(size - len(msg))} //ugly

	for len(msg) < size {
		var chunk, err = s.sock.Read(payload)
		if err != nil {
			fmt.Println(err)
		}
		if chunk == 0 {
			panic(fmt.Sprintf("socket connection broken"))
		}
		msg = msg + string(chunk) //ugly
	}

	return msg
}

type MuxDevice struct {
	devid    float32
	usbprod  byte
	serial   string
	location byte
}

func NewMuxDevice(devid float32, usbprod byte, serial string, location byte) *MuxDevice {
	return &MuxDevice{devid, usbprod, serial, location}
}

func (m *MuxDevice) Fields() string {
	return fmt.Sprintf("<MuxDevice: ID %g ProdID 0x%4x Serial '%s' Location 0x%x>", m.devid, m.usbprod, m.serial, m.location)
}

const (
	TypeResult       = 1
	TypeConnect      = 2
	TypeListen       = 3
	TypeDeviceAdd    = 4
	TypeDeviceRemove = 5
	Version          = 0
)

type BinaryProtocol struct {
	sock      net.Conn
	connected bool
}

func NewBinaryProtocol(socket net.Conn) *BinaryProtocol {
	return &BinaryProtocol{socket, false}
}

// still needs work
func (b *BinaryProtocol) _pack(req int, payload map[string]interface{}) []byte {
	switch req {
	case TypeConnect:
		buf := &bytes.Buffer{}
		err := binary.Write(buf, binary.LittleEndian, payload["DeviceID"].(string)+payload["PortNumber"].(string)+"\x00\x00")
		if err != nil {
			panic(fmt.Sprintln(err))
		}
		return buf.Bytes()
	case TypeListen:
		return nil
	}

	panic(fmt.Sprintf("Invalid outgoing request type %d", req))
}

func (b *BinaryProtocol) _unpack(resp int, payload interface{}) map[string]interface{} {
	switch resp {
	case TypeResult:
		buf := &bytes.Buffer{}
		err := binary.Read(buf, binary.LittleEndian, payload)
		if err != nil {
			panic(fmt.Sprintln(err))
		}
		return map[string]interface{}{
			"Number": buf.Bytes()[0],
		}
	case TypeDeviceAdd:
		buf := &bytes.Buffer{}

		err := binary.Read(buf, binary.LittleEndian, payload)
		if err != nil {
			panic(fmt.Sprintln(err))
		}
		devid, usbpid, location := buf.Bytes()[0], buf.Bytes()[1], buf.Bytes()[4]
		serial := strings.Split(string(buf.Bytes()[2]), "\\0")[0]

		return map[string]interface{}{
			"DeviceID": devid,
			"Properties": map[string]interface{}{
				"LocationID":   location,
				"SerialNumber": serial,
				"ProductID":    usbpid,
			},
		}
	case TypeDeviceRemove:
		buf := &bytes.Buffer{}

		err := binary.Read(buf, binary.LittleEndian, payload)
		if err != nil {
			panic(fmt.Sprintln(err))
		}
		devid := buf.Bytes()[0]

		return map[string]interface{}{
			"DeviceID": devid,
		}
	}
	panic(fmt.Sprintf("Invalid incoming request type %d", resp))
}

// fix variable names here
func (b *BinaryProtocol) sendpacket(req int, tag int, payload map[string]interface{}) {
	payload2 := b._pack(req, payload)

	if b.connected {
		panic(fmt.Sprintf("Mux is connected, cannot issue control packets"))
	}

	length := 16 + len(payload2)
	data := &bytes.Buffer{}
	err := binary.Write(data, binary.LittleEndian, length+Version+req+tag) // +payload2
	if err != nil {
		panic(fmt.Sprintln(err))
	}
}

// this function is disgusting
// maybe return 3 interface{} ?
func (b *BinaryProtocol) getpacket() (interface{}, interface{}, interface{}) {
	if b.connected {
		panic(fmt.Sprintf("Mux is connected, cannot issue control packets"))
	}

	buf := make([]byte, 4)
	byteBuf := []*bytes.Buffer{{}, {}, {}}

	err := binary.Write(byteBuf[0], binary.LittleEndian, dlen)
	if err != nil {
		panic(fmt.Sprintln(err))
	}

	var ndlen byte
	for i := range dlen {
		ndlen += dlen[i]
	}

	// byteBuf[1] == body
	// var _, _ = b.sock.Read(byteBuf[0].Bytes() - []byte{4}) //ugly

	err = binary.Write(byteBuf[1], binary.LittleEndian, body[:0xc])
	if err != nil {
		panic(fmt.Sprintln(err))
	}
	version, resp, tag := byteBuf[1].Bytes()[0], byteBuf[1].Bytes()[1], byteBuf[1].Bytes()[2]

	if version != Version {
		panic(fmt.Sprintf("Version mismatch: expected %d, got %d", Version, version))
	}

	payload := b._unpack(int(resp), byteBuf[2].Bytes()[0xc:])

	return resp, tag, payload
}

type PlistProtocol struct {
	*BinaryProtocol
}

type MuxConnection struct {
	socketpath string
	socket     *SafeStreamSocket
	proto      *BinaryProtocol
	pkttag     int
	devices    []*MuxDevice
}

func NewMuxConnection(socketpath string) *MuxConnection {
	var address, network string

	if runtime.GOOS == "windows" {
		address = "127.0.0.1:27015"
		network = "ip4" // i think
	} else {
		network = "unix"
		address = socketpath
	}

	s := NewSafeStreamSocket(network, address)

	return &MuxConnection{socketpath, s, NewBinaryProtocol(s.sock), 1, nil}
}

func (m *MuxConnection) _getreply() (interface{}, map[string]string) {
	for true {
		resp, tag, data := m.proto.getpacket()

		if resp == TypeResult {
			return tag, data.(map[string]string)
		}

		panic(fmt.Sprintf("Invalid packet type received: %d", resp))
	}
	return nil, nil
}

// this function is disgusting
func (m *MuxConnection) _processpacket() {
	// tag not needed?
	resp, _, data := m.proto.getpacket()

	switch resp {
	case TypeDeviceAdd:
		// welcome to assertion hell
		// this is literally hitler code
		m.devices = append(m.devices, NewMuxDevice(data.(map[string]interface{})["DeviceID"].(float32), data.(map[string]interface{})["Properties"].(byte), data.(map[string]interface{})["Properties"].(map[string]string)["SerialNumber"], data.(map[string]interface{})["Properties"].(map[string]byte)["LocationID"]))
	case TypeDeviceRemove:
		for i, v := range m.devices {
			if v.devid == data.(map[string]interface{})["DeviceID"] {
				// deletes an element from the map
				m.devices = append(m.devices[:i], m.devices[i+1:]...)
			}
		}
	case TypeResult:
		panic(fmt.Sprintf("Unexpeted result: %d", resp))
	default:
		panic(fmt.Sprintf("Invalid packet type received %d", resp))
	}
}

func (m *MuxConnection) _exchange(req int, payload map[string]interface{}) interface{} {
	m.pkttag++

	m.proto.sendpacket(req, m.pkttag, payload)
	recvtag, data := m._getreply()

	if recvtag != m.pkttag {
		panic(fmt.Sprintf("Reply tag mismatch: expected %d, got %d", m.pkttag, recvtag))
	}

	return data["Number"]
}

func (m *MuxConnection) listen() {
	ret := m._exchange(TypeListen, nil)
	if ret != nil {
		panic(fmt.Sprintf("Listen failed: error %d", ret))
	}
}

func (m *MuxConnection) process(timeout interface{}) {
	if m.proto.connected {
		panic(fmt.Sprintf("Socket is connected, cannot process listener events"))
	}
	// not really a thing
	m._processpacket()
}

func (m *MuxConnection) connect(device *MuxDevice, port int) net.Conn {
	payload := map[string]interface{}{
		"DeviceID":   device.devid,
		"PortNumber": ((port << 8) & 0xFF00) | (port >> 8),
	}

	ret := m._exchange(TypeConnect, payload)
	if ret != 0 {
		panic(fmt.Sprintf("Connect failed: error %d", ret))
	}

	m.proto.connected = true
	return m.socket.sock
}

func (m *MuxConnection) close() {
	err := m.socket.sock.Close()
	if err != nil {
		panic(fmt.Sprintln(err))
	}
}

type USBMux struct {
	socketpath string
	// protoclass *BinaryProtocol
	listener *MuxConnection
	Devices  []*MuxDevice
	version  int
}

func NewUSBMux(socketpath string) *USBMux {
	if socketpath == "" {
		socketpath = "/var/run/usbmuxd"
	}

	u := &USBMux{socketpath, NewMuxConnection(socketpath), nil, 0}
	u.Devices = u.listener.devices
	u.listener.listen()
	return u
}

func (u *USBMux) Process(timeout interface{}) {
	u.listener.process(timeout)
}

func (u *USBMux) Connect(device *MuxDevice, port int) net.Conn {
	connector := NewMuxConnection(u.socketpath)
	return connector.connect(device, port)
}
