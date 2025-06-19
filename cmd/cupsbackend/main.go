package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"

	"git.massivebox.net/massivebox/go-catprinter"

	"honnef.co/go/cups/raster"
	"honnef.co/go/cups/raster/image"
)

var URI_RE = regexp.MustCompile(`((\w*):///?)(([0-9a-fA-F]{2}[:-]?){6})`)

const (
	CUPS_BACKEND_OK            int = 0 // Job completed successfully
	CUPS_BACKEND_FAILED        int = 1 // Job failed, use error-policy
	CUPS_BACKEND_AUTH_REQUIRED int = 2 // Job failed, authentication required
	CUPS_BACKEND_HOLD          int = 3 // Job failed, hold job
	CUPS_BACKEND_STOP          int = 4 // Job failed, stop queue
	CUPS_BACKEND_CANCEL        int = 5 // Job failed, cancel job
	CUPS_BACKEND_RETRY         int = 6 // Job failed, retry this job later
	CUPS_BACKEND_RETRY_CURRENT int = 7 // Job failed, retry this job immediately
)

func print(mac string) error {

	d, err := raster.NewDecoder(os.Stdin)
	if err != nil {
		return fmt.Errorf("Failed to create decoder: %w", err)
	}

	fmt.Println("INFO: Initializing...")
	c, err := catprinter.NewClient()
	if err != nil {
		return fmt.Errorf("Failed to create BLE client: %w", err)
	}
	defer c.Stop()

	opts := catprinter.NewOptions().
		SetFeed(40).
		SetBestQuality(true).
		SetDither(true).
		SetFill(true).
		SetAutoRotate(false).
		SetBlackPoint(0.5)

	fmt.Println("INFO: Connecting...")
	err = c.Connect(mac)
	if err != nil {
		return fmt.Errorf("Failed to connect: %w", err)
	}
	fmt.Println("INFO: Connected!")

	fmt.Println("INFO: Printing...")

	for p, err := d.NextPage(); p != nil; {
		if err != nil {
			return fmt.Errorf("Failed to get next page: %w", err)
		}

		fmt.Printf("INFO: %#v\n", *p.Header)

		image, err := image.Image(p)
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("Failed to convert to image: %w", err)
		}
		err = c.Print(image, opts, false)
		if err != nil {
			return fmt.Errorf("Failed to print image: %w", err)
		}
	}

	fmt.Println("INFO: All done, exiting now.")
	return nil
}

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)
	log.SetPrefix("ERROR: ")
	if len(os.Args) == 1 {
		fmt.Println(`direct catprinter "Unknown" "Catprinter"`)
		return
	} else if len(os.Args) != 6 && len(os.Args) != 7 {
		log.Printf("ERROR: Got %d args %#v\n", len(os.Args), os.Args)
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

	// Parse MAC address
	mac := os.Getenv("DEVICE_URI")
	if mac == "" {
		mac = os.Args[0]
	}
	if match := URI_RE.FindStringSubmatch(mac); match != nil {
		urimac := match[3]
		parsed := []byte("XX:XX:XX:XX:XX:XX")
		for i, j := 0, 0; i < len(parsed) && j < len(urimac); i++ {
			if parsed[i] == ':' && (urimac[j] == '-' || urimac[j] == ':') {
				j++
			} else if parsed[i] == 'X' {
				parsed[i] = urimac[j]
				j++
			}
		}
		mac = string(parsed)
	}
	fmt.Printf("ERROR: Printing to MAC %s\n", mac)

	fmt.Printf(
		"ERROR: Job ID %s User %s Title %s Copies %s Options %s File %s\n",
		jobId, user, title, copies, options, file,
	)

	if err := print(mac); err != nil {
		log.Fatal(err)
	}
}
