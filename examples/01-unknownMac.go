package main

import (
	"git.massivebox.net/massivebox/go-catprinter"
	"log"
)

func main() {

	const name = "x6h"

	c, err := catprinter.NewClient()
	if err != nil {
		panic(err)
	}

	c.Debug.Log = true

	opts := catprinter.NewOptions()
	defer c.Stop()

	// let's find the MAC from the device name
	var mac string
	devices, err := c.ScanDevices(name)
	if err != nil {
		panic(err)
	}
	for deviceMac, deviceName := range devices {
		// you should ask the user to choose the device here, we will pretend they selected the first
		log.Println("Connecting to", deviceName, "with MAC", deviceMac)
		mac = deviceMac
		break
	}

	if err = c.Connect(mac); err != nil {
		panic(err)
	}

	err = c.PrintFile("../demo.jpg", opts)
	if err != nil {
		panic(err)
	}

}
