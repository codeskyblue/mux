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
// this func is fucked dawg
func (s *SafeStreamSocket) Recv(size int) []byte {
	fmt.Println("Recv called!")

	var msg []byte
	buf := make([]byte, 0, size-len(msg))
	fmt.Println("buffer allocated")

	for len(msg) < size {
		chunk, err := s.Sock.Read(buf)
		if err != nil {
			panic(err)
		}

		if chunk == 0 {
			panic("socket connection broken")
		}

		msg = append(msg, byte(chunk))
	}
	return msg
}
