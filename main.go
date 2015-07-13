package main

import (
	"tcprelay"
	"usbmux"
)

func main() {
	mux := usbmux.NewUSBMux()
	fmt.Println("Waiting for devices...")
}
