package main

import (
	"git.massivebox.net/massivebox/go-catprinter"
	"image/jpeg"
	"os"
)

func main() {

	const mac = "41:c2:6f:0f:90:c7"

	c, err := catprinter.NewClient()
	if err != nil {
		panic(err)
	}

	c.Debug.Log = true
	defer c.Stop()

	if err = c.Connect(mac); err != nil {
		panic(err)
	}

	// note that for this we need to open the image as image.Image manually!
	file, _ := os.Open("../demo.jpg")
	defer file.Close()
	img, _ := jpeg.Decode(file)

	// for now, we will use default options
	opts := catprinter.NewOptions()
	fmtImg := c.FormatImage(img, opts)

	// now you should display your image to the user and ask for what they want to change
	// in this example, we will pretend they want to disable dithering
	opts = opts.SetDither(false)
	fmtImg = c.FormatImage(img, opts)

	// you can show the image again and make all the adjustments you see fit with opts.SetXYZ
	// when the user decides to print, we can use
	err = c.Print(fmtImg, opts, true) // NOTE THE TRUE HERE! It means the image is already formatted
	if err != nil {
		panic(err)
	}

}
