package main

import "C"
import (
	"fmt"
	"marmalade/internal/devices"
	"marmalade/internal/server"

	"github.com/diamondburned/gotk4/pkg/gtk/v3"
)

var camera_indices []uint8
var is_webcam_refreshing bool

//export webcam_notify_selected
func webcam_notify_selected() {
	if is_webcam_refreshing {
		return
	}

	webcam_dropdown := UI.GetObject("webcam_dropdown").(*gtk.ComboBoxText)
	selected := webcam_dropdown.Active()
	if selected != -1 {
		index := camera_indices[selected]
		server.Config.Camera = int(index)
		update_unsaved_config(true)
	}
}

//export webcam_refresh_clicked
func webcam_refresh_clicked() {
	webcam_dropdown := UI.GetObject("webcam_dropdown").(*gtk.ComboBoxText)
	err := fill_camera_list(webcam_dropdown)
	if err != nil {
		UI.errChannel <- err
	}
}

func init_webcam_setting() {
	webcam_dropdown := UI.GetObject("webcam_dropdown").(*gtk.ComboBoxText)
	is_webcam_refreshing = false
	fill_camera_list(webcam_dropdown)
}

func fill_camera_list(input *gtk.ComboBoxText) error {
	cameras, err := devices.ListVideoCaptures()
	if err != nil {
		return err
	}

	if len(cameras) == 0 {
		input.SetActive(-1)
		input.RemoveAll()
		return nil
	}

	camera_indices = make([]uint8, 0, len(cameras))
	input.RemoveAll()
	selected_index := -1

	for i, camera := range cameras {
		camera_string := fmt.Sprintf("%d: %s", camera.Index, camera.Name)
		input.AppendText(camera_string)
		camera_indices = append(camera_indices, camera.Index)

		if camera.Index == uint8(server.Config.Camera) {
			selected_index = i
		}
	}

	is_webcam_refreshing = true

	if selected_index >= 0 {
		input.SetActive(selected_index)
		is_webcam_refreshing = false
	} else {
		is_webcam_refreshing = false
		input.SetActive(-1)
	}

	return nil
}
