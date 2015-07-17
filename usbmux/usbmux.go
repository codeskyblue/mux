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
	sock net.Conn // ??
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
			fmt.Println("socket connection broken")
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
			fmt.Println("socket connection broken")
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
func (b *BinaryProtocol) _pack(req int, payload map[string]string) []byte {
	switch req {
	case TypeConnect:
		buf := &bytes.Buffer{}
		binary.Write(buf, binary.LittleEndian, payload["DeviceID"]+payload["PortNumber"]+"\x00\x00")
		return buf.Bytes()
	case TypeListen:
		return nil
	}

	fmt.Printf("Invalid outgoing request type %d", req)
	return nil
}

func (b *BinaryProtocol) _unpack(resp int, payload interface{}) map[string]interface{} {
	switch resp {
	case TypeResult:
		buf := &bytes.Buffer{}
		binary.Read(buf, binary.LittleEndian, payload)
		return map[string]interface{}{
			"Number": buf.Bytes()[0],
		}
	case TypeDeviceAdd:
		buf := &bytes.Buffer{}
		binary.Read(buf, binary.LittleEndian, payload)
		devid, usbpid, pad, location := buf.Bytes()[0], buf.Bytes()[1], buf.Bytes()[3], buf.Bytes()[4]
		serial := strings.Split(string(buf.Bytes()[2]), "\\0")[0] //ugly
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
		binary.Read(buf, binary.LittleEndian, payload)
		devid := buf.Bytes()[0]
		return map[string]interface{}{
			"DeviceID": devid,
		}
	default:
		fmt.Printf("Invalid incoming request type %d", resp)
	}

	return nil
}

func (b *BinaryProtocol) sendpacket(req int, tag int, payload map[string]string) {
	pLoad := b._pack(req, payload)

	if b.connected {
		fmt.Println("Mux is connected, cannot issue control packets")
	}
	length := 16 + len(pLoad)

	data := &bytes.Buffer{}
	binary.Write(data, binary.LittleEndian, length+Version+req+tag+payload)
	b.sock.Write(data)
}

func (b *BinaryProtocol) getpacket() (interface{}, interface{}, map[string]string) {
	if b.connected {
		fmt.Println("Mux is connected, cannot issue control packets")
	}

	buf := []byte{}
	dlen, err := b.sock.Read(buf)
	if err != nil {
		fmt.Println(err)
	}

	byteBuf := []*bytes.Buffer{{}, {}, {}}
	binary.Write(byteBuf[0], binary.LittleEndian, dlen[0])
	// body :=
	var _, _ = b.sock.Read(byteBuf[0].Bytes() - []byte{4}) //ugly
	// version, resp, tag :=
	binary.Write(byteBuf[1], binary.LittleEndian, byteBuf[0].Bytes()[:0xc])

	// payload ==
	var _ = binary.Write(byteBuf[2], binary.LittleEndian, byteBuf[0].Bytes()[0xc:])
	// return resp, tag, payload
	return nil, nil, byteBuf[2]
}

type PlistProtocol struct {
	*BinaryProtocol
}

type MuxConnection struct {
	socketpath string
	socket     *SafeStreamSocket
	proto      interface{}
	pkttag     int
	devices    []*MuxDevice
}

func NewMuxConnection(socketpath string, protoclass interface{}) *MuxConnection {
	var address, network string

	if runtime.GOOS == "windows" {
		address = "127.0.0.1:27015"
		network = "ip4" // i think
	} else {
		network = "unix"
		address = socketpath
	}

	s := NewSafeStreamSocket(network, address)

	switch protoclass.(type) {
	case BinaryProtocol:
		return &MuxConnection{socketpath, s, NewBinaryProtocol(s.sock), 1, nil}
	case PlistProtocol: // use NewPlistProtocol when it's implemented
		return &MuxConnection{socketpath, s, PlistProtocol{NewBinaryProtocol(s.sock)}, 1, nil}
	}
	return nil
}

func (m *MuxConnection) _getreply() (interface{}, map[string]string) {
	for true {
		resp, tag, data := m.proto.(*BinaryProtocol).getpacket()

		if resp == TypeResult {
			return tag, data
		}

		fmt.Printf("Invalid packet type received: %d", resp)
	}
	return nil, nil
}

func (m *MuxConnection) _processpacket() {
	resp, tag, data := m.proto.(*BinaryProtocol).getpacket()

	switch resp {
	case TypeDeviceAdd:
		append(m.devices, NewMuxDevice(data["DeviceID"], data["Properties"], data["Properties"]["SerialNumber"], data["Properties"]["LocationID"]))
	case TypeDeviceRemove:
		for i, v := range m.devices {
			if v.devid == data["DeviceID"] {
				delete(m.devices, v)
			}
		}
	case TypeResult:
		fmt.Printf("Unexpeted result: %d", resp)
	default:
		fmt.Printf("Invalid packet type received %d", resp)
	}
}

func (m *MuxConnection) _exchange(req int, payload map[string]string) {
	mytag := m.pkttag
	m.pkttag++

	m.proto.(*BinaryProtocol).sendpacket(req, m.pkttag, payload)
	recvtag, data := m._getreply()

	if recvtag != mytag {
		fmt.Printf("Reply tag mismatch: expected %d, got %d", mytag, recvtag)
	}

	return data["Number"]
}

func (m *MuxConnection) listen() {
	ret := m._exchange(TypeListen, nil)
	if ret != 0 {
		fmt.Println("Listen failed: error ", ret)
	}
}

func (m *MuxConnection) process(timeout interface{}) {
	if m.proto.(*BinaryProtocol).connected {
		fmt.Println("Socket is connected, cannot process listener events")
	}
}

func (m *MuxConnection) connect(device *MuxDevice, port int) net.Conn {
	payload := map[string]int{
		"DeviceID":   device.devid,
		"PortNumber": ((port << 8) & 0xFF00) | (port >> 8),
	}

	ret := m._exchange(TypeConnect, payload)
	if ret != 0 {
		fmt.Printf("Connect failed: error %d", ret)
	}

	m.proto.(*BinaryProtocol).connected = true
	return m.socket.sock
}

func (m *MuxConnection) close() {
	m.socket.sock.Close()
}

type USBMux struct {
	socketpath string
	protoclass interface{}
	listener   *MuxConnection
	devices    []string
	version    int
}

func NewUSBMux(socketpath string) *USBMux {
	if socketpath == "" {
		if runtime.GOOS == "darwin" {
			socketpath = "/var/run/usbmuxd"
		} else {
			socketpath = "var/run/usbmuxd"
		}
	}

	b := &BinaryProtocol{}
	u := &USBMux{socketpath, b, NewMuxConnection(socketpath, b), nil, 0}

	u.devices = u.listener.devices
	u.listener.listen()
	return u
}

func (u *USBMux) process(timeout interface{}) {
	u.listener.process(timeout)
}

func (u *USBMux) connect(device, port string) net.Conn {
	connector := NewMuxConnection(u.socketpath, BinaryProtocol{})
	return connector.connect(device, port)
}
