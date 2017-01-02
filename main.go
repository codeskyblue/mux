package main

import "flag"

func main() {
	HOST := "localhost"

	threaded := flag.Bool("threaded", false, "use threading to handle multiple connections at once")
	bufsize := flag.Int("bufsize", 128, "specify buffer size for socket forwarding")
	sockpath := flag.String("socket", "", "specify the path of the usbmuxd socket")
	flag.Parse()

	serverclass := &TCPRelay{}

	var ports []int
}
