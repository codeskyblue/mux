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
func (s *SafeStreamSocket) Send(msg []byte) {
	// var totalsent int

	for totalsent, sent := 0, 0; totalsent < len(msg); totalsent += sent {
		sent, err := s.Sock.Write(msg[totalsent:])
		if err != nil {
			fmt.Println(err)
		}
		if sent == 0 {
			panic("socket connection broken")
		}
		// totalsent = totalsent + sent
	}
}

// usbmux/SafeStreamSocket.(*SafeStreamSocket).Recv
// net.(*conn).Read
func (s *SafeStreamSocket) Recv(size int) []byte {
	fmt.Println("Recv called!")

	var msg []byte
	buf := make([]byte, 0, size-len(msg))

	for ; len(msg) < size; msg = append(msg, buf...) {
		chunk, err := s.Sock.Read(buf)
		fmt.Println("Read called!")
		if err != nil {
			fmt.Println(err)
		}

		if chunk == 0 {
			panic("socket connection broken")
		}

		// msg = append(msg, buf...)
	}
	return msg
}
