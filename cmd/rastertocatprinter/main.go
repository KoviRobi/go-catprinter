package main

import (
	"fmt"
	"log"
	"os"

	"git.massivebox.net/massivebox/go-catprinter"

	"honnef.co/go/cups/raster"
	"honnef.co/go/cups/raster/image"
)

func print(jobId, user, title, copies, options, file string) error {
	var (
		mac          string
		lowerQuality bool
		autoRotate   bool
		dontDither   bool
		fill         bool
		blackPoint   float64
		feed         int
	)

	d, err := raster.NewDecoder(os.Stdin)
	if err != nil {
		return err
	}

	fmt.Println("Initializing...")
	c, err := catprinter.NewClient()
	if err != nil {
		return err
	}
	defer c.Stop()

	opts := catprinter.NewOptions().
		SetFeed(feed).
		SetBestQuality(!lowerQuality).
		SetDither(!dontDither).
		SetFill(fill).
		SetAutoRotate(autoRotate).
		SetBlackPoint(float32(blackPoint))

	fmt.Println("Connecting...")
	err = c.Connect(mac)
	if err != nil {
		return err
	}
	fmt.Println("Connected!")

	fmt.Println("Printing...")

	for p, err := d.NextPage(); p != nil; {
		if err != nil {
			return err
		}

		image, err := image.Image(p)
		if err != nil {
			return err
		}
		err = c.Print(image, opts, true)
		if err != nil {
			return err
		}
	}

	fmt.Println("All done, exiting now.")
	return nil
}

func main() {
	if len(os.Args) != 6 || len(os.Args) != 7 {
		log.Fatalf("Usage: %s job-id user title copies options [file]\n", os.Args[0])
	}

	jobId := os.Args[1]
	user := os.Args[2]
	title := os.Args[3]
	copies := os.Args[4]
	options := os.Args[5]
	var file string
	if len(os.Args) == 7 {
		file = os.Args[6]
	}

	if err := print(jobId, user, title, copies, options, file); err != nil {
		log.Fatal(err)
	}

}
