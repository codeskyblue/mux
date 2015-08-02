package SafeStreamSocket

import (
	"fmt"
	"net"
)

type SafeStreamSocket struct {
	Sock net.Conn
}

func New(network, address string) *SafeStreamSocket {
	ret, err := net.Dial(network, address)
	if err != nil {
		fmt.Println(err)
	}

	return &SafeStreamSocket{ret}
}

func (s *SafeStreamSocket) Send(msg []byte) {
	var totalsent int

	for totalsent < len(msg) {
		sent, err := s.Sock.Write(msg[totalsent:])
		if err != nil {
			fmt.Println(err)
		}
		if sent == 0 {
			panic("socket connection broken")
		}
		totalsent = totalsent + sent
	}
}

// no longer returns a string
func (s *SafeStreamSocket) Recv(size int) []byte {
	var msg []byte
	data := byte(size - len(msg))
	payload := []byte{data}

	for len(msg) < size {
		chunk, err := s.Sock.Read(payload)
		if err != nil {
			panic(err)
		}
		if chunk == 0 {
			panic("socket connection broken")
		}
		// msg = msg + chunk
		msg = append(msg, byte(chunk))
	}
	return msg
}
