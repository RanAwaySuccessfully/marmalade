package main

import "C"
import (
	"marmalade/app/gtk3/ui"
	"marmalade/internal/devices"
	"marmalade/internal/server"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v3"
)

var gpu_ids []string

//export mediapipe_notify_expanded
func mediapipe_notify_expanded() {
	misc_row := UI.GetObject("mediapipe_expander").(*gtk.Expander)
	expanded := misc_row.Expanded()

	facem_label := UI.GetObject("facem_label").(*gtk.Label)
	facem_input := UI.GetObject("facem_input").(*gtk.Entry)
	handm_label := UI.GetObject("handm_label").(*gtk.Label)
	handm_input := UI.GetObject("handm_input").(*gtk.Entry)
	device_label := UI.GetObject("device_label").(*gtk.Label)
	device_input := UI.GetObject("device_input").(*gtk.ComboBoxText)

	if expanded {
		facem_label.SetVisible(true)
		facem_input.SetVisible(true)
		handm_label.SetVisible(true)
		handm_input.SetVisible(true)
		device_label.SetVisible(true)
		device_input.SetVisible(true)
	} else {
		facem_label.SetVisible(false)
		facem_input.SetVisible(false)
		handm_label.SetVisible(false)
		handm_input.SetVisible(false)
		device_label.SetVisible(false)
		device_input.SetVisible(false)
	}

	set_window_size()
}

func init_mediapipe_widgets() {
	UI.gtkBuilder.AddFromString(ui.SettingsMediaPipe)

	facem_input := UI.GetObject("facem_input").(*gtk.Entry)
	facem_input.SetText(server.Config.ModelFace)
	facem_input.ConnectChanged(func() {
		value := facem_input.Text()
		server.Config.ModelFace = value
		update_unsaved_config(true)
	})

	handm_input := UI.GetObject("handm_input").(*gtk.Entry)
	handm_input.SetText(server.Config.ModelHand)
	handm_input.ConnectChanged(func() {
		value := handm_input.Text()
		server.Config.ModelHand = value
		update_unsaved_config(true)
	})

	init_gpu_widget()

	facem_label := UI.GetObject("facem_label").(*gtk.Label)
	handm_label := UI.GetObject("handm_label").(*gtk.Label)
	device_label := UI.GetObject("device_label").(*gtk.Label)
	device_input := UI.GetObject("device_input").(*gtk.ComboBoxText)

	grid := UI.GetObject("main_grid").(*gtk.Grid)
	grid.Attach(facem_label, 0, 21, 1, 1)
	grid.Attach(facem_input, 1, 21, 1, 1)
	grid.Attach(handm_label, 0, 22, 1, 1)
	grid.Attach(handm_input, 1, 22, 1, 1)
	grid.Attach(device_label, 0, 23, 1, 1)
	grid.Attach(device_input, 1, 23, 1, 1)
}

func init_gpu_widget() {
	gpu_input := UI.GetObject("device_input").(*gtk.ComboBoxText)

	fill_gpu_list(gpu_input)

	cells := gpu_input.Cells()
	for _, cell := range cells {
		cell.SetObjectProperty("width", 50)
		cell.SetObjectProperty("height", 24)
	}

	gpu_input.Connect("notify::selected", func() {
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
