package usbmux

import (
	"log"
	"strings"

	"github.com/Mitchell-Riley/uhaul"
)

const (
	Version = iota
	TypeResult
	TypeConnect
	TypeListen
	TypeDeviceAdd
	TypeDeviceRemove

	TYPE_RESULT        = "Result"
	TYPE_CONNECT       = "Connect"
	TYPE_LISTEN        = "Listen"
	TYPE_DEVICE_ADD    = "Attached"
	TYPE_DEVICE_REMOVE = "Detached" //???
	TYPE_PLIST         = 8
	VERSION            = 1
)

type Protocol interface {
	_pack(int, map[string]interface{}) []byte
	_unpack(int, []byte) map[string]interface{}
	SendPacket(int, int, interface{})
	GetPacket() (byte, byte, map[string]interface{})
}

type BinaryProtocol struct {
	socket    *SafeStreamSocket
	Connected bool
}

func NewBinaryProtocol(socket *SafeStreamSocket) *BinaryProtocol {
	return &BinaryProtocol{socket, false}
}

func (*BinaryProtocol) _pack(req int, payload map[string]interface{}) []byte {
	switch req {
	case TypeConnect:
		return append(uhaul.Pack("IH", payload["DeviceID"], payload["PortNumber"]), []byte{0x00, 0x00}...)
	case TypeListen:
		return nil
	default:
		log.Fatalf("Invalid outgoing request type %d", req)
		return nil
	}

}

func (*BinaryProtocol) _unpack(resp int, payload []byte) map[string]interface{} {
	switch resp {
	case TypeResult:
		return map[string]interface{}{
			"Number": uhaul.Unpack("I", payload)[0],
		}
	case TypeDeviceAdd:
		unpackedData := uhaul.Unpack("IH256sHI", payload)
		devid := unpackedData[0]
		usbpid := unpackedData[1]
		serial := strings.Split(string(unpackedData[2]), "\\0")[0]
		// pad := unpackedData[3]
		location := unpackedData[4]

		return map[string]interface{}{
			"DeviceID": devid,
			"Properties": map[string]interface{}{
				"LocationID":   location,
				"SerialNumber": serial,
				"ProductID":    usbpid,
			},
		}
	case TypeDeviceRemove:
		return map[string]interface{}{
			"DeviceID": uhaul.Unpack("I", payload)[0],
		}

	default:
		log.Fatalf("Invalid incoming request type %d", resp)
		return nil
	}

}

func (b *BinaryProtocol) SendPacket(req, tag int, payload interface{}) {
	// use a different variable here because we'll need a type assertion and loss of data can occur
	payload = b._pack(req, payload.(map[string]interface{}))

	if b.Connected {
		log.Fatal("Mux is connected, cannot issue control packets")
	}

	length := 16 + len(payload.([]byte))
	data := append(uhaul.Pack("IIII", length, Version, req, tag), payload.([]byte)...)

	b.socket.Send(data)
}

func (b *BinaryProtocol) GetPacket() (byte, byte, map[string]interface{}) {
	if b.Connected {
		log.Fatal("Mux is connected, cannot issue control packets")
	}

	dlen := uhaul.Unpack("I", b.socket.Recv(4))[0]
	body := b.socket.Recv(int(dlen - 4))

	unpackedData := uhaul.Unpack("III", body[:0xc])
	version := unpackedData[0]
	resp := unpackedData[1]
	tag := unpackedData[2]

	if version != Version {
		log.Fatalf("Version mismatch: expected %d, got %d", Version, version)
	}

	payload := b._unpack(int(resp), body[0xc:])

	return resp, tag, payload
}
