package BinaryProtocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"usbmux/SafeStreamSocket"
)

const (
	TypeResult       = 1
	TypeConnect      = 2
	TypeListen       = 3
	TypeDeviceAdd    = 4
	TypeDeviceRemove = 5
	Version          = 0
)

type BinaryProtocol struct {
	socket    *SafeStreamSocket.SafeStreamSocket
	Connected bool
}

func New(socket *SafeStreamSocket.SafeStreamSocket) *BinaryProtocol {
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
func (b *BinaryProtocol) SendPacket(req int, tag int, payload interface{}) {
	payload = b._pack(req, payload.(map[string]interface{}))

	if b.Connected {
		panic(fmt.Sprintf("Mux is connected, cannot issue control packets"))
	}

	length := 16 + len(payload.([]byte))
	data := &bytes.Buffer{}
	err := binary.Write(data, binary.LittleEndian, int32(length+Version+req+tag)) // +payload
	if err != nil {
		panic(fmt.Sprintln(err))
	}
	b.socket.Send(data.Bytes())
}

// maybe return 3 interface{} ?
func (b *BinaryProtocol) GetPacket() (interface{}, interface{}, map[string]interface{}) {
	if b.Connected {
		panic(fmt.Sprintf("Mux is connected, cannot issue control packets"))
	}

	dlen := b.socket.Recv(4)
	byteBuf := []*bytes.Buffer{{}, {}}

	err := binary.Write(byteBuf[0], binary.LittleEndian, []uint8(dlen))
	if err != nil {
		panic(fmt.Sprintln(err))
	}
	dlen = byteBuf[0].Bytes()

	var ndlen byte
	for i := range dlen {
		ndlen += dlen[i]
	}

	body := b.socket.Recv(int(ndlen) - 4)

	err = binary.Write(byteBuf[1], binary.LittleEndian, []uint8(body)[:0xc])
	if err != nil {
		panic(fmt.Sprintln(err))
	}
	version, resp, tag := byteBuf[1].Bytes()[0], byteBuf[1].Bytes()[1], byteBuf[1].Bytes()[2]

	if version != Version {
		panic(fmt.Sprintf("Version mismatch: expected %d, got %d", Version, version))
	}

	payload := b._unpack(int(resp), body[0xc:])

	return resp, tag, payload
}
