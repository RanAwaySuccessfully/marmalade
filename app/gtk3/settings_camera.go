package main

import (
	"marmalade/internal/server"
	"strconv"

	"github.com/diamondburned/gotk4/pkg/gtk/v3"
)

type cameraWidgets struct {
	width_label  *gtk.Label
	width_input  *gtk.Entry
	height_label *gtk.Label
	height_input *gtk.Entry
	fps_label    *gtk.Label
	fps_input    *gtk.Entry
	format_label *gtk.Label
	format_input *gtk.Entry
	camera_info  *gtk.Button
}

func create_camera_settings(grid *gtk.Grid, window *gtk.ApplicationWindow) {
	camera_row := gtk.NewExpander("Camera settings")
	camera_row.StyleContext().AddClass("boldText")
	camera_row.SetMarginTop(5)
	camera_row.SetMarginBottom(5)

	camera_widgets := create_camera_widgets()
	grid.Attach(camera_row, 0, 2, 2, 1)

	camera_row.Connect("notify::expanded", func() {
		expanded := camera_row.Expanded()
		row := query_child_row(grid, camera_row)
		row++

		if expanded {
			show_camera_widgets(grid, &camera_widgets, row)
		} else {
			hide_camera_widgets(grid, row)
			set_window_size(window)
		}
	})
}

func create_camera_widgets() cameraWidgets {
	width_label := gtk.NewLabel("Width:")
	width_label.SetHAlign(gtk.AlignStart)

	width_input := gtk.NewEntry()
	width := strconv.Itoa(server.Config.Width) // convert int to string
	width_input.SetText(width)
	width_input.SetPlaceholderText("1280")

	width_input.ConnectChanged(func() {
		update_numeric_config(width_input, &server.Config.Width)
	})

	height_label := gtk.NewLabel("Height:")
	height_label.SetHAlign(gtk.AlignStart)

	height_input := gtk.NewEntry()
	height := strconv.Itoa(server.Config.Height) // convert int to string
	height_input.SetText(height)
	height_input.SetPlaceholderText("720")

	height_input.Connect("changed", func() {
		update_numeric_config(height_input, &server.Config.Height)
	})

	fps_label := gtk.NewLabel("Frame rate (FPS):")
	fps_label.SetHAlign(gtk.AlignStart)

	fps_input := gtk.NewEntry()
	fps := strconv.Itoa(server.Config.FPS) // convert int to string
	fps_input.SetText(fps)
	fps_input.SetPlaceholderText("30")

	fps_input.Connect("changed", func() {
		update_numeric_config(fps_input, &server.Config.FPS)
	})

	format_label := gtk.NewLabel("Format:")
	format_label.SetHAlign(gtk.AlignStart)
	format_input := gtk.NewEntry()
	format_input.SetText(server.Config.Format)
	format_input.SetPlaceholderText("YUYV")

	format_input.Connect("changed", func() {
		value := format_input.Text()
		server.Config.Format = value
		update_unsaved_config(true)
	})

	camera_info := gtk.NewButtonWithLabel("View supported settings")
	camera_info.Connect("clicked", func() {
		camera_id := uint8(server.Config.Camera)
		create_camera_info_window(camera_id)
	})

	widgets := cameraWidgets{
		width_label,
		width_input,
		height_label,
		height_input,
		fps_label,
		fps_input,
		format_label,
		format_input,
		camera_info,
	}

	return widgets
}

func show_camera_widgets(grid *gtk.Grid, widgets *cameraWidgets, row int) {
	grid.InsertRow(row)
	grid.InsertRow(row + 1)
	grid.InsertRow(row + 2)
	grid.InsertRow(row + 3)
	grid.InsertRow(row + 4)

	grid.Attach(widgets.width_label, 0, row, 1, 1)
	grid.Attach(widgets.width_input, 1, row, 1, 1)

	grid.Attach(widgets.height_label, 0, row+1, 1, 1)
	grid.Attach(widgets.height_input, 1, row+1, 1, 1)

	grid.Attach(widgets.fps_label, 0, row+2, 1, 1)
	grid.Attach(widgets.fps_input, 1, row+2, 1, 1)

	grid.Attach(widgets.format_label, 0, row+3, 1, 1)
	grid.Attach(widgets.format_input, 1, row+3, 1, 1)

	grid.Attach(widgets.camera_info, 1, row+4, 1, 1)

	grid.ShowAll()
}

func hide_camera_widgets(grid *gtk.Grid, row int) {
	grid.RemoveRow(row + 4)
	grid.RemoveRow(row + 3)
	grid.RemoveRow(row + 2)
	grid.RemoveRow(row + 1)
	grid.RemoveRow(row)
}
