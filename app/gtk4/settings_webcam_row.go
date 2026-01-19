package main

import "C"
import (
	"fmt"
	"marmalade/internal/devices"
	"marmalade/internal/server"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

var camera_indices []uint8
var is_webcam_refreshing bool

//export webcam_notify_selected
func webcam_notify_selected() {
	if is_webcam_refreshing {
		return
	}

	webcam_dropdown := UI.GetObject("webcam_dropdown").(*gtk.DropDown)
	selected := webcam_dropdown.Selected()
	if selected != gtk.InvalidListPosition {
		index := camera_indices[selected]
		server.Config.Camera = float64(index)
		update_unsaved_config(true)
	}
}

//export signal_webcam_refresh_clicked
func signal_webcam_refresh_clicked() {
	webcam_dropdown := UI.GetObject("webcam_dropdown").(*gtk.DropDown)
	err := fill_camera_list(webcam_dropdown)
	if err != nil {
		UI.errChannel <- err
	}
}

func init_webcam_setting() {
	webcam_dropdown := UI.GetObject("webcam_dropdown").(*gtk.DropDown)

	webcam_factory := dropdown_all_factory_create()
	webcam_dropdown.SetFactory(&webcam_factory.ListItemFactory)

	webcam_list_factory := dropdown_list_factory_create(webcam_dropdown)
	webcam_dropdown.SetListFactory(&webcam_list_factory.ListItemFactory)

	is_webcam_refreshing = false
	fill_camera_list(webcam_dropdown)
}

func fill_camera_list(input *gtk.DropDown) error {
	cameras, err := devices.ListVideoCaptures()
	if err != nil {
		return err
	}

	if len(cameras) == 0 {
		input.SetSelected(gtk.InvalidListPosition)
		input.SetModel(nil)
		return nil
	}

	camera_indices = make([]uint8, 0, len(cameras))
	camera_list := make([]string, 0, len(cameras))
	selected_index := -1

	for i, camera := range cameras {
		camera_string := fmt.Sprintf("%d: %s", camera.Index, camera.Name)
		camera_list = append(camera_list, camera_string)
		camera_indices = append(camera_indices, camera.Index)

		if camera.Index == uint8(server.Config.Camera) {
			selected_index = i
		}
	}

	is_webcam_refreshing = true

	model := gtk.NewStringList(camera_list)
	input.SetModel(model)

	if selected_index >= 0 {
		input.SetSelected(uint(selected_index))
		is_webcam_refreshing = false
	} else {
		is_webcam_refreshing = false
		input.SetSelected(gtk.InvalidListPosition)
	}

	return nil
}
