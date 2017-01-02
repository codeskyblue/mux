package usbmux

import (
	"fmt"
	"log"
	"net"
)

type SafeStreamSocket struct {
	Sock net.Conn
}

func NewSafeStreamSocket(network, address string) *SafeStreamSocket {
	c, err := net.Dial(network, address)
	if err != nil {
		fmt.Println(err)
	}

	return &SafeStreamSocket{c}
}

func (s *SafeStreamSocket) Send(msg []byte) {
	for totalsent := 0; totalsent < len(msg); {
		sent, err := s.Sock.Write(msg[totalsent:])
		if err != nil {
			log.Fatal(err)
		}

		if sent == 0 {
			log.Fatal("socket connection broken")
		}

		totalsent += sent
	}
}

func (s *SafeStreamSocket) Recv(size int) []byte {
	var msg []byte

	for len(msg) < size {
		buf := make([]byte, size-len(msg))

		chunk, err := s.Sock.Read(buf)
		if err != nil {
			log.Fatal(err)
		}

		if chunk == 0 {
			log.Fatal("socket connection broken")
		}

		msg = append(msg, buf...)
	}
	return msg
}
