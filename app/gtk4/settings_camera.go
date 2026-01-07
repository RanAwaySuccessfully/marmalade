package main

import "C"
import (
	"marmalade/app/gtk4/ui"
	"marmalade/internal/server"
	"strconv"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

//export camera_notify_expanded
func camera_notify_expanded() {
	grid := UI.GetObject("main_grid").(*gtk.Grid)
	camera_row := UI.GetObject("camera_expander").(*gtk.Expander)

	expanded := camera_row.Expanded()
	_, row, _, _ := grid.QueryChild(camera_row)
	row++

	if expanded {
		show_camera_widgets(grid, row)
	} else {
		hide_camera_widgets(grid, row)
	}
}

func init_camera_widgets() {
	UI.gtkBuilder.AddFromString(ui.SettingsCamera)

	width_input := UI.GetObject("width_input").(*gtk.Entry)
	width := strconv.FormatFloat(server.Config.Width, 'f', 0, 64)
	width_input.SetText(width)
	width_input.ConnectChanged(func() {
		update_numeric_config(width_input, &server.Config.Width)
	})

	height_input := UI.GetObject("height_input").(*gtk.Entry)
	height := strconv.FormatFloat(server.Config.Height, 'f', 0, 64)
	height_input.SetText(height)
	height_input.ConnectChanged(func() {
		update_numeric_config(height_input, &server.Config.Height)
	})

	fps_input := UI.GetObject("fps_input").(*gtk.Entry)
	fps := strconv.FormatFloat(server.Config.FPS, 'f', 0, 64)
	fps_input.SetText(fps)
	fps_input.ConnectChanged(func() {
		update_numeric_config(fps_input, &server.Config.FPS)
	})

	format_input := UI.GetObject("format_input").(*gtk.Entry)
	format_input.SetText(server.Config.Format)
	format_input.ConnectChanged(func() {
		value := format_input.Text()
		server.Config.Format = value
		update_unsaved_config(true)
	})

	camera_info := UI.GetObject("camera_info").(*gtk.Button)
	camera_info.Connect("clicked", func() {
		camera_id := uint8(server.Config.Camera)
		create_camera_info_window(camera_id)
	})

	return
}

func show_camera_widgets(grid *gtk.Grid, row int) {
	grid.InsertRow(row)
	grid.InsertRow(row + 1)
	grid.InsertRow(row + 2)
	grid.InsertRow(row + 3)

	resolution_label := UI.GetObject("resolution_label").(*gtk.Label)
	resolution_box := UI.GetObject("resolution_box").(*gtk.Box)
	grid.Attach(resolution_label, 0, row, 1, 1)
	grid.Attach(resolution_box, 1, row, 1, 1)

	fps_label := UI.GetObject("fps_label").(*gtk.Label)
	fps_input := UI.GetObject("fps_input").(*gtk.Entry)
	grid.Attach(fps_label, 0, row+1, 1, 1)
	grid.Attach(fps_input, 1, row+1, 1, 1)

	format_label := UI.GetObject("format_label").(*gtk.Label)
	format_input := UI.GetObject("format_input").(*gtk.Entry)
	grid.Attach(format_label, 0, row+2, 1, 1)
	grid.Attach(format_input, 1, row+2, 1, 1)

	camera_info := UI.GetObject("camera_info").(*gtk.Button)
	grid.Attach(camera_info, 1, row+3, 1, 1)
}

func hide_camera_widgets(grid *gtk.Grid, row int) {
	grid.RemoveRow(row + 3)
	grid.RemoveRow(row + 2)
	grid.RemoveRow(row + 1)
	grid.RemoveRow(row)
}
