package main

import "git.massivebox.net/massivebox/go-catprinter"

func main() {

	const mac = "41:c2:6f:0f:90:c7"

	c, err := catprinter.NewClient()
	if err != nil {
		panic(err)
	}

	c.Debug.Log = true

	opts := catprinter.NewOptions()
	defer c.Stop()

	if err = c.Connect(mac); err != nil {
		panic(err)
	}

	err = c.PrintFile("../demo.jpg", opts)
	if err != nil {
		panic(err)
	}

}
