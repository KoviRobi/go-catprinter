package catprinter

import (
	"github.com/disintegration/imaging"
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
	"github.com/pkg/errors"
	"image"
	"log"
	"os"
	"time"
)

var (
	ErrPrinterNotFound       = errors.New("unable to find printer, make sure it is turned on and in range")
	ErrMissingCharacteristic = errors.New("missing characteristic, make sure the MAC is correct and the printer is supported")
	ErrNotBlackWhite         = errors.New("image must be black and white (NOT only grayscale)")
	ErrInvalidImageSize      = errors.New("image must be 384px wide")
)

// Client contains information for the connection to the printer and debugging options.
type Client struct {
	device         ble.Device
	printer        ble.Client
	characteristic *ble.Characteristic
	chunkSize      int
	Timeout        time.Duration
	Debug          struct {
		Log         bool // print logs to terminal
		DumpRequest bool // saves last data sent to printer to ./request.bin
		DumpImage   bool // saves formatted image to ./image.png
		DontPrint   bool // if true, the image is not actually printed. saves paper during testing.
	}
}

// NewClient initiates a new client with sane defaults
func NewClient() (*Client, error) {
	d, err := dev.DefaultDevice()
	if err != nil {
		return nil, errors.Wrap(err, "can't create device, superuser permissions might be needed")
	}
	return NewClientFromDevice(d)
}

// NewClientFromDevice initiates a new client with a custom ble.Device and sane defaults
func NewClientFromDevice(d ble.Device) (*Client, error) {
	var c = &Client{}
	ble.SetDefaultDevice(d)
	c.device = d
	c.Timeout = scanTimeout
	return c, nil
}

// Stop closes any active connection to a printer and detaches the GATT server
func (c *Client) Stop() error {
	if err := c.Disconnect(); err != nil {
		return errors.Wrap(err, "can't disconnect printer")
	}
	return c.device.Stop()
}

// Disconnect closes any active connection to a printer
func (c *Client) Disconnect() error {
	if c.printer != nil {
		if err := c.printer.CancelConnection(); err != nil {
			return err
		}
		c.printer = nil
	}
	return nil
}

// Print prints an image to the connected printer. It also formats it and dithers if specified by opts and isAlreadyFormatted.
// Only set isAlreadyFormatted to true if the image is in black and white (NOT ONLY grayscale) and 384px wide.
func (c *Client) Print(img image.Image, opts *PrinterOptions, isAlreadyFormatted bool) error {
	if !isAlreadyFormatted {
		img = c.FormatImage(img, opts)
	}
	fmtImg, err := convertImageToBytes(img)
	if err != nil {
		return err
	}
	if opts.bestQuality {
		fmtImg = commandsPrintImg(fmtImg, opts.feed)
	} else {
		fmtImg = weakCommandsPrintImg(fmtImg, opts.feed)
	}
	if c.Debug.DumpRequest {
		err = os.WriteFile("./request.bin", fmtImg, 0644)
		if err != nil {
			log.Println("failed to save debugging request dump", err.Error())
		}
	}
	if c.Debug.DontPrint {
		log.Println("image will not be printed as Client.Debug.DontPrint is true")
		return nil
	}
	return c.writeData(fmtImg)
}

// PrintFile dithers, formats and prints an image by path to the connected printer
func (c *Client) PrintFile(path string, opts *PrinterOptions) error {
	img, err := imaging.Open(path)
	if err != nil {
		return err
	}
	return c.Print(img, opts, false)
}

func (c *Client) log(format string, a ...any) {
	if !c.Debug.Log {
		return
	}
	log.Printf(format, a...)
}
