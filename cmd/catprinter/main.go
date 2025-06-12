package main

import (
	"fmt"
	"git.massivebox.net/massivebox/go-catprinter"
	"github.com/pkg/errors"
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

var flags = []cli.Flag{
	&cli.StringFlag{
		Name:  "mac",
		Usage: "MAC address of the printer. Provide this or name.",
	},
	&cli.StringFlag{
		Name:  "name",
		Usage: "common name of the printer. Provide this or MAC.",
	},
	&cli.StringFlag{
		Name:      "image",
		Usage:     "path to the image file to be printed",
		Required:  true,
		TakesFile: true,
	},
	&cli.BoolFlag{
		Name:  "lowerQuality",
		Usage: "print with lower quality, but slightly faster speed",
	},
	&cli.BoolFlag{
		Name:  "autoRotate",
		Usage: "rotate image to fit printer",
	},
	&cli.BoolFlag{
		Name:  "dontDither",
		Usage: "don't dither the image",
	},
	&cli.BoolFlag{
		Name:  "fill",
		Usage: "fill/crop instead of resize",
	},
	&cli.Float64Flag{
		Name:  "blackPoint",
		Value: 0.5,
		Usage: "regulate at which point a gray pixel is printed as black",
	},
	&cli.IntFlag{
		Name:  "feed",
		Value: 40,
		Usage: "amount of paper to feed",
	},
	&cli.BoolFlag{
		Name:  "debugLog",
		Usage: "print debugging messages",
	},
	&cli.BoolFlag{
		Name:  "dumpImage",
		Usage: "save dithered image to ./image.png",
	},
	&cli.BoolFlag{
		Name:  "dumpRequest",
		Usage: "save raw data sent to printer to ./request.bin",
	},
	&cli.BoolFlag{
		Name:  "dontPrint",
		Usage: "don't actually print the image",
	},
}

func findMac(name string, c *catprinter.Client) (string, error) {
	fmt.Printf("Finding MAC by name (will take %d seconds)...", c.Timeout/time.Second)
	devices, err := c.ScanDevices(name)
	if err != nil {
		return "", err
	}
	switch len(devices) {
	case 0:
		return "", errors.New("no devices found with name " + name)
	case 1:
		for k, _ := range devices {
			return k, nil
		}
		break
	default:
		fmt.Println("Found multiple devices:")
		for m, n := range devices {
			fmt.Printf("%s\t%s", m, n)
		}
		return "", errors.New("multiple devices found with name " + name + ", please specify MAC directly")
	}
	return "", nil
}

func action(cCtx *cli.Context) error {

	var (
		mac          = cCtx.String("mac")
		name         = cCtx.String("name")
		imagePath    = cCtx.String("image")
		lowerQuality = cCtx.Bool("lowerQuality")
		autoRotate   = cCtx.Bool("autoRotate")
		dontDither   = cCtx.Bool("dontDither")
		fill         = cCtx.Bool("fill")
		blackPoint   = cCtx.Float64("blackPoint")
		feed         = cCtx.Int("feed")
		debugLog     = cCtx.Bool("debugLog")
		dumpImage    = cCtx.Bool("dumpImage")
		dumpRequest  = cCtx.Bool("dumpRequest")
		dontPrint    = cCtx.Bool("dontPrint")
	)

	fmt.Println("Initializing...")
	c, err := catprinter.NewClient()
	if err != nil {
		return err
	}
	defer c.Stop()

	c.Debug.Log = debugLog
	c.Debug.DumpImage = dumpImage
	c.Debug.DumpRequest = dumpRequest
	c.Debug.DontPrint = dontPrint

	opts := catprinter.NewOptions().
		SetFeed(feed).
		SetBestQuality(!lowerQuality).
		SetDither(!dontDither).
		SetFill(fill).
		SetAutoRotate(autoRotate).
		SetBlackPoint(float32(blackPoint))

	if (mac != "") == (name != "") {
		return errors.New("either mac or name must be provided")
	}

	if name != "" {
		mac, err = findMac(name, c)
		if err != nil {
			return err
		}
	}

	fmt.Println("Connecting...")
	err = c.Connect(mac)
	if err != nil {
		return err
	}
	fmt.Println("Connected!")

	fmt.Println("Printing...")
	err = c.PrintFile(imagePath, opts)
	if err != nil {
		return err
	}

	fmt.Println("All done, exiting now.")

	return nil
}

func main() {

	app := &cli.App{
		Name:   "catprinter",
		Usage:  "print images to some BLE thermal printers",
		Flags:  flags,
		Action: action,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
