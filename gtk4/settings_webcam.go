//go:build withgtk4

package gtk4

import (
	"fmt"
	"marmalade/camera"
	"marmalade/server"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

var camera_indices []uint8

func create_webcam_setting(grid *gtk.Grid, err_chan chan error) {
	webcam_label := gtk.NewLabel("Webcam:")
	webcam_label.SetSizeRequest(125, 1)
	webcam_label.SetHAlign(gtk.AlignStart)
	webcam_label.SetXAlign(0)

	webcam_box := gtk.NewBox(gtk.OrientationHorizontal, 3)

	webcam_input := gtk.NewDropDown(nil, nil)
	webcam_input.SetHExpand(true)
	webcam_box.Append(webcam_input)

	// TODO: this triggers on refresh...
	webcam_input.Connect("notify::selected", func() {
		selected := webcam_input.Selected()
		index := camera_indices[selected]
		server.Config.Camera = float64(index)
		update_unsaved_config(true)
	})

	webcam_refresh := gtk.NewButtonFromIconName("view-refresh-symbolic")
	webcam_box.Append(webcam_refresh)

	webcam_refresh.Connect("clicked", func() {
		err := fill_camera_list(webcam_input)
		if err != nil {
			err_chan <- err
		}
	})

	fill_camera_list(webcam_input)
	grid.Attach(webcam_label, 0, 1, 1, 1)
	grid.Attach(webcam_box, 1, 1, 1, 1)
}

func fill_camera_list(input *gtk.DropDown) error {
	cameras, err := camera.GetInputDevices()
	if err != nil {
		return err
	}

	if len(cameras) == 0 {
		input.SetSelected(gtk.InvalidListPosition)
		input.SetModel(nil)
		return nil
	}

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

	model := gtk.NewStringList(camera_list)
	input.SetModel(model)

	if selected_index >= 0 {
		input.SetSelected(uint(selected_index))
	} else {
		input.SetSelected(gtk.InvalidListPosition)
	}

	return nil
}
