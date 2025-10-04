//go:build withgtk4

package gtk4

import (
	"fmt"
	"marmalade/server"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type cameraWidgets struct {
	width_label  *gtk.Label
	width_input  *gtk.Entry
	height_label *gtk.Label
	height_input *gtk.Entry
	fps_label    *gtk.Label
	fps_input    *gtk.Entry
}

func create_camera_settings(grid *gtk.Grid, window *gtk.ApplicationWindow) {
	camera_row := gtk.NewExpander("Camera settings")
	camera_row.AddCSSClass("boldText")
	camera_row.SetMarginTop(5)
	camera_row.SetMarginBottom(5)

	camera_widgets := create_camera_widgets()
	grid.Attach(camera_row, 0, 2, 2, 1)

	camera_row.Connect("notify::expanded", func() {
		expanded := camera_row.Expanded()
		_, row, _, _ := grid.QueryChild(camera_row)
		row++

		if expanded {
			show_camera_widgets(grid, &camera_widgets, row)
		} else {
			hide_camera_widgets(grid, row)
			window.SetDefaultSize(500, 150)
		}
	})
}

func create_camera_widgets() cameraWidgets {
	width_label := gtk.NewLabel("Width:")
	width_label.SetHAlign(gtk.AlignStart)

	width_input := gtk.NewEntry()
	width := fmt.Sprintf("%d", int(server.Config.Width))
	width_input.SetText(width)

	width_input.Connect("changed", func() {
		update_numeric_config(width_input, &server.Config.Width)
	})

	height_label := gtk.NewLabel("Height:")
	height_label.SetHAlign(gtk.AlignStart)

	height_input := gtk.NewEntry()
	height := fmt.Sprintf("%d", int(server.Config.Height))
	height_input.SetText(height)

	height_input.Connect("changed", func() {
		update_numeric_config(height_input, &server.Config.Height)
	})

	fps_label := gtk.NewLabel("FPS:")
	fps_label.SetHAlign(gtk.AlignStart)

	fps_input := gtk.NewEntry()
	fps := fmt.Sprintf("%d", int(server.Config.FPS))
	fps_input.SetText(fps)

	fps_input.Connect("changed", func() {
		update_numeric_config(fps_input, &server.Config.FPS)
	})

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

func show_camera_widgets(grid *gtk.Grid, widgets *cameraWidgets, row int) {
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

func hide_camera_widgets(grid *gtk.Grid, row int) {
	grid.RemoveRow(row + 2)
	grid.RemoveRow(row + 1)
	grid.RemoveRow(row)
}
