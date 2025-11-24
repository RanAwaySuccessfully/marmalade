//go:build withgtk4

package gtk4

import (
	"marmalade/devices"
	"marmalade/server"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

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
	gpus, err := devices.ListDisplayControllers()

	gpu_ids = make([]string, 0, len(gpus))
	device_list := make([]string, 2, len(gpus)+2)
	device_list[0] = "CPU"
	device_list[1] = "GPU (Auto)"

	selected_index := -1

	if len(gpus) > 0 {
		for i, device := range gpus {
			camera_string := "GPU: " + device.Device
			device_list = append(device_list, camera_string)

			replacer := strings.NewReplacer(":", "_", ".", "_")
			gpu_id := "pci-" + replacer.Replace(device.Slot)
			gpu_ids = append(gpu_ids, gpu_id)

			if gpu_id == server.Config.PrimeId {
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

	return err
}
