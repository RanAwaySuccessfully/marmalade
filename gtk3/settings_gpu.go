//go:build withgtk3

package gtk3

import (
	"marmalade/devices"
	"marmalade/server"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v3"
)

var gpu_ids []string

func create_gpu_widget() *gtk.ComboBoxText {
	gpu_input := gtk.NewComboBoxText()
	gpu_input.SetHExpand(true)

	fill_gpu_list(gpu_input)

	cells := gpu_input.Cells()
	for _, cell := range cells {
		cell.SetObjectProperty("width", 50)
		cell.SetObjectProperty("height", 24)
	}

	gpu_input.Connect("changed", func() {
		selected := gpu_input.Active()

		if selected == -1 {
			return
		}

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

func fill_gpu_list(input *gtk.ComboBoxText) error {
	gpus, err := devices.ListDisplayControllers()

	gpu_ids = make([]string, 0, len(gpus))
	input.RemoveAll()
	input.AppendText("CPU")
	input.AppendText("GPU (Auto)")

	selected_index := -1

	if len(gpus) > 0 {
		for i, device := range gpus {
			camera_string := "GPU: " + device.Device
			input.AppendText(camera_string)

			replacer := strings.NewReplacer(":", "_", ".", "_")
			gpu_id := "pci-" + replacer.Replace(device.Slot)
			gpu_ids = append(gpu_ids, gpu_id)

			if gpu_id == server.Config.PrimeId {
				selected_index = i
			}
		}
	}

	if selected_index >= 0 {
		input.SetActive(selected_index + 2)
	} else if server.Config.UseGpu {

		if server.Config.PrimeId == "" {
			input.SetActive(1)
		} else {
			input.SetActive(-1)
		}

	} else {
		input.SetActive(0)
	}

	return err
}
