//go:build withgtk3

package gtk3

import (
	"fmt"
	"marmalade/devices"
	"marmalade/server"

	"github.com/diamondburned/gotk4/pkg/gtk/v3"
)

var camera_indices []uint8

func create_webcam_setting(grid *gtk.Grid, err_chan chan error) {
	webcam_label := gtk.NewLabel("Webcam:")
	webcam_label.SetSizeRequest(125, 1)
	webcam_label.SetHAlign(gtk.AlignStart)
	webcam_label.SetXAlign(0)

	webcam_box := gtk.NewBox(gtk.OrientationHorizontal, 3)

	webcam_input := gtk.NewComboBoxText()
	webcam_input.SetHExpand(true)
	webcam_box.Add(webcam_input)

	is_refreshing := false

	fill_camera_list(webcam_input, &is_refreshing)

	webcam_input.Connect("notify::selected", func() {
		if is_refreshing {
			return
		}

		selected := webcam_input.Active()
		index := camera_indices[selected]
		server.Config.Camera = float64(index)
		update_unsaved_config(true)
	})

	webcam_refresh := gtk.NewButtonFromIconName("view-refresh-symbolic", 4)
	webcam_box.Add(webcam_refresh)

	webcam_refresh.Connect("clicked", func() {
		err := fill_camera_list(webcam_input, &is_refreshing)
		if err != nil {
			err_chan <- err
		}
	})

	grid.Attach(webcam_label, 0, 1, 1, 1)
	grid.Attach(webcam_box, 1, 1, 1, 1)
}

func fill_camera_list(input *gtk.ComboBoxText, is_refreshing *bool) error {
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

	*is_refreshing = true

	if selected_index >= 0 {
		input.SetActive(selected_index)
		*is_refreshing = false
	} else {
		*is_refreshing = false
		input.SetActive(-1)
	}

	return nil
}
