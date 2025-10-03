//go:build withgtk4

package gtk4

import (
	"fmt"
	"marmalade/server"
	"marmalade/v4l2"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func fill_camera_list(input *gtk.DropDown) error {
	cameras, err := v4l2.GetInputDevices()
	if err != nil {
		return err
	}

	var camera_list []string
	selected_index := -1

	for i, camera := range cameras {
		camera_string := fmt.Sprintf("%d: %s", camera.Index, camera.Name)
		camera_list = append(camera_list, camera_string)

		if camera.Index == int(server.Config.Camera) {
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

type cameraWidgets struct {
	width_label  *gtk.Label
	width_input  *gtk.Entry
	height_label *gtk.Label
	height_input *gtk.Entry
	fps_label    *gtk.Label
	fps_input    *gtk.Entry
}

func create_camera_settings() cameraWidgets {
	width_label := gtk.NewLabel("Width:")
	width_label.SetHAlign(gtk.AlignStart)

	width_input := gtk.NewEntry()
	width := fmt.Sprintf("%d", int(server.Config.Width))
	width_input.SetText(width)

	height_label := gtk.NewLabel("Height:")
	height_label.SetHAlign(gtk.AlignStart)

	height_input := gtk.NewEntry()
	height := fmt.Sprintf("%d", int(server.Config.Height))
	height_input.SetText(height)

	fps_label := gtk.NewLabel("FPS:")
	fps_label.SetHAlign(gtk.AlignStart)

	fps_input := gtk.NewEntry()
	fps := fmt.Sprintf("%d", int(server.Config.FPS))
	fps_input.SetText(fps)

	widgets := cameraWidgets{
		width_label,
		width_input,
		height_label,
		height_input,
		fps_label,
		fps_input,
	}

	return widgets
}

func show_camera_settings(grid *gtk.Grid, widgets *cameraWidgets, row int) {
	grid.InsertRow(row)
	grid.InsertRow(row + 1)
	grid.InsertRow(row + 2)

	grid.Attach(widgets.width_label, 0, row, 1, 1)
	grid.Attach(widgets.width_input, 1, row, 1, 1)

	grid.Attach(widgets.height_label, 0, row+1, 1, 1)
	grid.Attach(widgets.height_input, 1, row+1, 1, 1)

	grid.Attach(widgets.fps_label, 0, row+2, 1, 1)
	grid.Attach(widgets.fps_input, 1, row+2, 1, 1)
}

func hide_camera_settings(grid *gtk.Grid, row int) {
	grid.RemoveRow(row + 2)
	grid.RemoveRow(row + 1)
	grid.RemoveRow(row)
}

type miscWidgets struct {
	model_label *gtk.Label
	model_input *gtk.Entry
	port_label  *gtk.Label
	port_input  *gtk.Entry
	gpu_label   *gtk.Label
	gpu_input   *gtk.Switch
}

func create_misc_settings() miscWidgets {
	model_label := gtk.NewLabel("Model filename:")
	model_label.SetHAlign(gtk.AlignStart)
	model_input := gtk.NewEntry()
	model := fmt.Sprintf(server.Config.Model)
	model_input.SetText(model)

	port_label := gtk.NewLabel("UDP port:")
	port_label.SetHAlign(gtk.AlignStart)
	port_input := gtk.NewEntry()
	port := fmt.Sprintf("%d", int(server.Config.Port))
	port_input.SetText(port)

	gpu_label := gtk.NewLabel("Use GPU?")
	gpu_label.SetHAlign(gtk.AlignStart)
	gpu_input := gtk.NewSwitch()
	gpu_input.SetState(server.Config.UseGpu)

	widgets := miscWidgets{
		model_label,
		model_input,
		port_label,
		port_input,
		gpu_label,
		gpu_input,
	}

	return widgets
}

func show_misc_settings(grid *gtk.Grid, widgets *miscWidgets, row int) {
	grid.InsertRow(row)
	grid.InsertRow(row + 1)
	grid.InsertRow(row + 2)

	grid.Attach(widgets.model_label, 0, row, 1, 1)
	grid.Attach(widgets.model_input, 1, row, 1, 1)

	grid.Attach(widgets.port_label, 0, row+1, 1, 1)
	grid.Attach(widgets.port_input, 1, row+1, 1, 1)

	grid.Attach(widgets.gpu_label, 0, row+2, 1, 1)
	grid.Attach(widgets.gpu_input, 1, row+2, 1, 1)
}

func hide_misc_settings(grid *gtk.Grid, row int) {
	grid.RemoveRow(row + 2)
	grid.RemoveRow(row + 1)
	grid.RemoveRow(row)
}
