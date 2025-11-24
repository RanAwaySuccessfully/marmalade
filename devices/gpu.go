//go:build withgtk4 || withgtk3

package devices

import (
	"bufio"
	"os/exec"
	"strings"
)

type DisplayController struct {
	Slot      string
	Vendor    string
	SubVendor string
	Device    string
}

func ListDisplayControllers() ([]DisplayController, error) {
	cmd := exec.Command("lspci", "-d", "::03xx", "-vmmD")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	devices := make([]DisplayController, 0)
	device := DisplayController{}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			devices = append(devices, device)
			device = DisplayController{}
			continue
		}

		split := strings.SplitN(line, ":\t", 2)
		//fmt.Printf("[LSPCI] %s %s\n", split[0], split[1])
		switch split[0] {
		case "Slot":
			device.Slot = split[1]
		case "Vendor":
			device.Vendor = split[1]
		case "Device":
			device.Device = split[1]
		case "SVendor":
			device.SubVendor = split[1]
		}
	}

	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	return devices, nil
}
