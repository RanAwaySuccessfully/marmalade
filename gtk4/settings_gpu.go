//go:build withgtk4

package gtk4

import (
	"bufio"
	"fmt"
	"marmalade/server"
	"os/exec"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type DisplayController struct {
	Slot      string
	Vendor    string
	SubVendor string
	Device    string
}

var gpu_ids []string

func create_gpu_widget() *gtk.DropDown {
	gpu_input := gtk.NewDropDown(nil, nil)
	gpu_input.SetHExpand(true)

	gpu_factory := create_custom_factory()
	gpu_input.SetFactory(&gpu_factory.ListItemFactory)

	gpu_list_factory := create_custom_list_factory(gpu_input)
	gpu_input.SetListFactory(&gpu_list_factory.ListItemFactory)

	fill_gpu_list(gpu_input)

	gpu_input.Connect("notify::selected", func() {
		selected := gpu_input.Selected()

		switch selected {
		case 0:
			server.Config.UseGpu = false
			server.Config.PrimeId = ""
		case 1:
			server.Config.UseGpu = true
			server.Config.PrimeId = ""
		default:
			server.Config.UseGpu = true
			server.Config.PrimeId = gpu_ids[selected-2]
		}

		update_unsaved_config(true)
	})

	return gpu_input
}

func fill_gpu_list(input *gtk.DropDown) error {
	devices, err := GetDisplayDevices()
	if err != nil {
		return err
	}

	gpu_ids = make([]string, 0, len(devices))
	device_list := make([]string, 2, len(devices)+2)
	device_list[0] = "CPU"
	device_list[1] = "GPU (Auto)"

	selected_index := -1

	if len(devices) > 0 {
		for i, device := range devices {
			camera_string := fmt.Sprintf("GPU: %s", device.Device)
			device_list = append(device_list, camera_string)
			gpu_ids = append(gpu_ids, device.Slot)

			if device.Slot == server.Config.PrimeId {
				selected_index = i
			}
		}
	}

	model := gtk.NewStringList(device_list)
	input.SetModel(model)

	if selected_index >= 0 {
		input.SetSelected(uint(selected_index + 2))
	} else if server.Config.UseGpu {

		if server.Config.PrimeId == "" {
			input.SetSelected(1)
		} else {
			input.SetSelected(gtk.InvalidListPosition)
		}

	} else {
		input.SetSelected(0)
	}

	return nil
}

func GetDisplayDevices() ([]DisplayController, error) {
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
