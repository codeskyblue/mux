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

// usbmux/SafeStreamSocket.(*SafeStreamSocket).Send
// i made a revert here!
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

		totalsent += sent
	}
}

// usbmux/SafeStreamSocket.(*SafeStreamSocket).Recv
// net.(*conn).Read
func (s *SafeStreamSocket) Recv(size int) []byte {
	fmt.Println("Recv called!")

	var msg []byte

	for len(msg) < size {
		buf := make([]byte, 0, size-len(msg))

		chunk, err := s.Sock.Read(buf)
		if err != nil {
			fmt.Println(err)
		}

		if chunk == 0 {
			panic("socket connection broken")
		}

		msg = append(msg, buf...)
	}
	return msg
}
