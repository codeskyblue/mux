package tcprelay

import (
	"fmt"
	"net"
	"tcprelay/SocketRelay"
	"usbmux"
)

type TCPRelay struct {
	server net.Conn
}

func (t *TCPRelay) handle() {
	// server_address = LocalAddr() or RemoteAddr()
	fmt.Printf("Incoming connection to %d", t.server.RemoteAddr()[1])

	mux := usbmux.NewUSBMux("")

	fmt.Println("Waiting for devices...")

	if mux.Devices == nil {
		mux.Process(1.0)
		fmt.Println("No device found")
		t.server.Close()
		return
	}

	dev := mux.Devices[0]
	fmt.Printf("Connecting to device %s", dev.Fields())

	dsock := mux.Connect(dev, t.server.RemoteAddr().String())
	lsock := t.server

	fmt.Println("Connection established, relaying data")

	fwd := SocketRelay.New(dsock, lsock, 1022)
	fwd.handle()

	defer dsock.close()
	defer lsock.close()

	fmt.Println("Connection closed")
}

type TCPServer struct {
}

type ThreadedTCPServer struct {
}
