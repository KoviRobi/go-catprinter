package catprinter

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"tinygo.org/x/bluetooth"
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
		n, err := c.characteristic.WriteWithoutResponse(chunk)
		if err != nil || n == 0 {
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

	adapter := bluetooth.DefaultAdapter

	err := adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		if strings.Contains(strings.Join(found, " "), device.Address.String()) {
			return
		}
		found = append(found, device.Address.String())
		if strings.Contains(strings.ToLower(device.LocalName()), strings.ToLower(name)) {
			devices[device.Address.String()] = device.LocalName()
			c.log("Matches  %s %s", device.Address.String(), device.LocalName())
			return
		}
		c.log("No match %s %s", device.Address.String(), device.LocalName())
	})

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
func (c *Client) Connect(macString string) error {
	var params bluetooth.ConnectionParams
	mac, err := bluetooth.ParseMAC(macString)
	if err != nil {
		return fmt.Errorf("Bad MAC address: %w", err)
	}
	address := bluetooth.Address{MACAddress: bluetooth.MACAddress{MAC: mac}}
	connect, err := c.adapter.Connect(address, params)
	if err != nil {
		return err
	}

	var services []bluetooth.DeviceService
	for _, uuid := range possibleServiceUuids {
		parsed, err := bluetooth.ParseUUID(uuid)
		if err != nil {
			return fmt.Errorf("Bad BLE UUID: %s; %w\n", uuid, err)
		}
		services, err = connect.DiscoverServices([]bluetooth.UUID{parsed})
		if err == nil {
			break
		}
	}
	if err != nil {
		return errors.Wrap(err, "discovering services")
	}

	var char *bluetooth.DeviceCharacteristic
	txCharacteristic, err := bluetooth.ParseUUID(txCharacteristicUuid)
	if err != nil {
		return fmt.Errorf("Bad BLE UUID: %s; %w\n", txCharacteristicUuid, err)
	}
	for _, service := range services {
		c.log("service %s", service.UUID().String())
		chars, err := service.DiscoverCharacteristics([]bluetooth.UUID{txCharacteristic})
		if err != nil {
			return fmt.Errorf("Failed to discover characteristics: %w", err)
		}
		for _, characteristic := range chars {
			c.log("  %s", characteristic.UUID().String())
			if characteristic.UUID() == txCharacteristic {
				c.log("    found characteristic!")
				char = &characteristic
				break
			}
		}
	}

	if char == nil {
		return ErrMissingCharacteristic
	}

	mtu, err := char.GetMTU()
	if err != nil {
		return fmt.Errorf("Failed to get MTU: %w", err)
	}

	c.characteristic = char
	c.printer = &connect
	c.chunkSize = int(mtu) - 3
	return nil

}
