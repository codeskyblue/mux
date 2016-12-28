package main

import (
	"fmt"
	"time"

	"github.com/Mitchell-Riley/mux/usbmux"
)

func main() {
	mux := usbmux.New("")

	fmt.Println("Waiting for devices...")

	if mux.Devices == nil {
		// not necessarily right
		fmt.Println("No devices dawg")
		mux.Process(time.Millisecond)
	}

	for {
		fmt.Println("Devices:")

		for _, v := range mux.Devices {
			fmt.Println(v)
		}

		var d time.Duration

		// not necessarily right
		mux.Process(d)
	}
}
