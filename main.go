package main

import (
	"fmt"
	"usbmux"
)

func main() {
	// put this is usbmux's init function?
	mux := usbmux.New("")
	fmt.Println("Waiting for devices...")

	if mux.Devices == nil {
		mux.Process(0.1)
	}

	for true {
		fmt.Println("Devices:")

		for _, v := range mux.Devices {
			fmt.Println(v)
		}

		mux.Process(nil)
	}
}
