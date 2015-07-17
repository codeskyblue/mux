package main

import (
	"tcprelay"
	"usbmux"
)

func main() {
	// put this is usbmux's init function?
	mux := usbmux.NewUSBMux("")
	fmt.Println("Waiting for devices...")

	if mux.devices == nil {
		mux.process(0.1)
	}

	for true {
		fmt.Println("Devices:")
		for i, v := range mux.devices {
			fmt.Printlmn(v)
		}
		mux.process()
	}
}
