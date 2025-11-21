//go:build withgtk3

package gtk3

import (
	"marmalade/camera"
	"marmalade/server"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v3"
)

var gpu_ids []string

func create_gpu_widget() *gtk.ComboBoxText {
	gpu_input := gtk.NewComboBoxText()
	gpu_input.SetHExpand(true)

	fill_gpu_list(gpu_input)

	gpu_input.Connect("notify::selected", func() {
		selected := gpu_input.Active()

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
	devices, err := camera.GetDisplayDevices()
	if err != nil {
		return err
	}

	gpu_ids = make([]string, 0, len(devices))
	input.RemoveAll()
	input.AppendText("CPU")
	input.AppendText("GPU (Auto)")

	selected_index := -1

	if len(devices) > 0 {
		for i, device := range devices {
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

	return nil
}
