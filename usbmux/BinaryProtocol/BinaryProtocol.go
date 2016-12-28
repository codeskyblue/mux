package BinaryProtocol

import (
	"bytes"
	"encoding/binary"
	"log"
	"strings"

	"github.com/Mitchell-Riley/mux/usbmux/SafeStreamSocket"
)

const (
	Version = iota
	TypeResult
	TypeConnect
	TypeListen
	TypeDeviceAdd
	TypeDeviceRemove
)

type BinaryProtocol struct {
	socket    *SafeStreamSocket.SafeStreamSocket
	Connected bool
}

func New(socket *SafeStreamSocket.SafeStreamSocket) *BinaryProtocol {
	return &BinaryProtocol{socket, false}
}

// still needs work
func (*BinaryProtocol) _pack(req int, payload map[string]interface{}) []byte {
	switch req {
	case TypeConnect:
		buf := &bytes.Buffer{}
		err := binary.Write(buf, binary.LittleEndian, payload["DeviceID"].(byte)+payload["PortNumber"].(byte)+0x00+0x00)
		if err != nil {
			log.Fatal(err)
		}
		return buf.Bytes()
	case TypeListen:
		return nil
	}

	log.Fatalf("Invalid outgoing request type %d", req)
	return nil
}

func (*BinaryProtocol) _unpack(resp int, payload interface{}) map[string]interface{} {
	switch resp {
	case TypeResult:
		buf := &bytes.Buffer{}
		err := binary.Read(buf, binary.LittleEndian, payload)
		if err != nil {
			log.Fatal(err)
		}

		return map[string]interface{}{
			"Number": buf.Bytes()[0],
		}
	case TypeDeviceAdd:
		buf := &bytes.Buffer{}

		err := binary.Read(buf, binary.LittleEndian, payload)
		if err != nil {
			log.Fatal(err)
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
			log.Fatal(err)
		}
		devid := buf.Bytes()[0]

		return map[string]interface{}{
			"DeviceID": devid,
		}
	}
	log.Fatalf("Invalid incoming request type %d", resp)
	return nil
}

func (b *BinaryProtocol) SendPacket(req, tag int, payload interface{}) {
	payload = b._pack(req, payload.(map[string]interface{}))

	if b.Connected {
		log.Fatal("Mux is connected, cannot issue control packets")
	}

	length := 16 + len(payload.([]byte))
	buf := &bytes.Buffer{}

	err := binary.Write(buf, binary.LittleEndian, int32(length+Version+req+tag))
	if err != nil {
		log.Fatal(err)
	}

	data := append(buf.Bytes(), payload.([]byte)...)

	b.socket.Send(data)
}

// cast cast city here come the casts
// casting dlen might not be necessary but better safe than sorry?
// although most of these are harmless without any loss of precision
func (b *BinaryProtocol) GetPacket() (byte, byte, map[string]interface{}) {
	if b.Connected {
		log.Fatal("Mux is connected, cannot issue control packets")
	}

	var dlen interface{} = b.socket.Recv(4)

	byteBuf := []*bytes.Buffer{{}, {}}

	err := binary.Write(byteBuf[0], binary.LittleEndian, []uint8(dlen.([]byte)))
	if err != nil {
		log.Fatal(err)
	}

	dlen = byteBuf[0].Bytes()[0]

	body := b.socket.Recv(int(dlen.(byte)))

	err = binary.Write(byteBuf[1], binary.LittleEndian, []uint8(body)[:0xc])
	if err != nil {
		log.Fatal(err)
	}
	version, resp, tag := byteBuf[1].Bytes()[0], byteBuf[1].Bytes()[1], byteBuf[1].Bytes()[2]

	if version != Version {
		log.Fatalf("Version mismatch: expected %d, got %d", Version, version)
	}

	payload := b._unpack(int(resp), body[0xc:])

	return resp, tag, payload
}
