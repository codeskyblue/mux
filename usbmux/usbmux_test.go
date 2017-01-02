package usbmux

import (
	"fmt"
	"testing"
)

func TestUsbmux(t *testing.T) {
	mux := New("")
	fmt.Println(mux.Devices)

	// fmt.Println("Waiting for devices...")

	// if mux.Devices == nil {
	// 	// not necessarily right
	// 	fmt.Println("No devices dawg")
	// 	mux.Process(time.Millisecond)
	// }

	// for {
	// 	fmt.Println("Devices:")

	// 	for _, v := range mux.Devices {
	// 		fmt.Println(v)
	// 	}

	// 	// not necessarily right
	// 	mux.Process(nil)
	// }
}
