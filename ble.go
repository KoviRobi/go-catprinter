package catprinter

import (
	"context"
	"github.com/go-ble/ble"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

// For some reason, bleak reports the 0xaf30 service on my macOS, while it reports
// 0xae30 (which I believe is correct) on my Raspberry Pi. This hacky workaround
// should cover both cases.

var possibleServiceUuids = []string{
	"ae30",
	"af30",
}

const txCharacteristicUuid = "ae01"

const scanTimeout = 10 * time.Second

const waitAfterEachChunkS = 20 * time.Millisecond

const waitAfterDataSentS = 30 * time.Second

func chunkify(data []byte, chunkSize int) [][]byte {
	var chunks [][]byte
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}
	return chunks
}

func (c *Client) writeData(data []byte) error {

	chunks := chunkify(data, c.chunkSize)
	c.log("Sending %d chunks of size %d...", len(chunks), c.chunkSize)
	for i, chunk := range chunks {
		err := c.printer.WriteCharacteristic(c.characteristic, chunk, true)
		if err != nil {
			return errors.Wrap(err, "writing to characteristic, chunk "+strconv.Itoa(i))
		}
		time.Sleep(waitAfterEachChunkS)
	}
	c.log("All sent.")

	return nil
}

// ScanDevices scans for a device with the given name, or auto discovers it based on characteristics (not implemented yet).
// Returns a map with MACs as key, and device names as values.
func (c *Client) ScanDevices(name string) (map[string]string, error) {

	var devices = make(map[string]string)
	var found []string

	c.log("Looking for a BLE device named %s", name)

	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), c.Timeout))

	err := ble.Scan(ctx, true, func(a ble.Advertisement) {
		if strings.Contains(strings.Join(found, " "), a.Addr().String()) {
			return
		}
		found = append(found, a.Addr().String())
		if strings.Contains(strings.ToLower(a.LocalName()), strings.ToLower(name)) {
			devices[a.Addr().String()] = a.LocalName()
			c.log("Matches  %s %s", a.Addr().String(), a.LocalName())
			return
		}
		c.log("No match %s %s", a.Addr().String(), a.LocalName())
	}, nil)

	switch errors.Cause(err) {
	case nil:
	case context.DeadlineExceeded:
		c.log("Timeout: scan completed")
	case context.Canceled:
		c.log("Scan was canceled")
	default:
		return nil, errors.Wrap(err, "failed scan")
	}

	if len(devices) < 0 {
		return nil, ErrPrinterNotFound
	}
	return devices, nil

}

// Connect establishes a BLE connection to a printer by MAC address.
func (c *Client) Connect(mac string) error {

	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), c.Timeout))
	connect, err := ble.Dial(ctx, ble.NewAddr(mac))
	if err != nil {
		return err
	}

	profile, err := connect.DiscoverProfile(true)
	if err != nil {
		return errors.Wrap(err, "discovering profile")
	}

	var char *ble.Characteristic
	for _, service := range profile.Services {
		c.log("service %s", service.UUID.String())
		for _, characteristic := range service.Characteristics {
			c.log("  %s", characteristic.UUID.String())
			if characteristic.UUID.Equal(ble.MustParse(txCharacteristicUuid)) &&
				strings.Contains(strings.Join(possibleServiceUuids, " "), service.UUID.String()) {
				c.log("    found characteristic!")
				char = characteristic
				break
			}
		}
	}

	if char == nil {
		return ErrMissingCharacteristic
	}

	c.characteristic = char
	c.printer = connect
	c.chunkSize = c.printer.Conn().RxMTU() - 3
	return nil

}
