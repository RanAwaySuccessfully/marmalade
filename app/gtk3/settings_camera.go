package main

import "C"
import (
	"marmalade/app/gtk3/ui"
	"marmalade/internal/server"
	"strconv"

	"github.com/diamondburned/gotk4/pkg/gtk/v3"
)

//export camera_notify_expanded
func camera_notify_expanded() {
	camera_row := UI.GetObject("camera_expander").(*gtk.Expander)
	expanded := camera_row.Expanded()

	width_label := UI.GetObject("width_label").(*gtk.Label)
	width_input := UI.GetObject("width_input").(*gtk.Entry)
	height_label := UI.GetObject("height_label").(*gtk.Label)
	height_input := UI.GetObject("height_input").(*gtk.Entry)
	fps_label := UI.GetObject("fps_label").(*gtk.Label)
	fps_input := UI.GetObject("fps_input").(*gtk.Entry)
	format_label := UI.GetObject("format_label").(*gtk.Label)
	format_input := UI.GetObject("format_input").(*gtk.Entry)
	camera_info := UI.GetObject("camera_info").(*gtk.Button)

	if expanded {
		width_label.SetVisible(true)
		width_input.SetVisible(true)
		height_label.SetVisible(true)
		height_input.SetVisible(true)
		fps_label.SetVisible(true)
		fps_input.SetVisible(true)
		format_label.SetVisible(true)
		format_input.SetVisible(true)
		camera_info.SetVisible(true)
	} else {
		width_label.SetVisible(false)
		width_input.SetVisible(false)
		height_label.SetVisible(false)
		height_input.SetVisible(false)
		fps_label.SetVisible(false)
		fps_input.SetVisible(false)
		format_label.SetVisible(false)
		format_input.SetVisible(false)
		camera_info.SetVisible(false)
	}

	set_window_size()
}

func init_camera_widgets() {
	UI.gtkBuilder.AddFromString(ui.SettingsCamera)

	width := ""
	if server.Config.Width != 0 {
		width = strconv.Itoa(server.Config.Width) // convert int to string
	}

	width_input := UI.GetObject("width_input").(*gtk.Entry)
	width_input.SetText(width)
	width_input.ConnectChanged(func() {
		update_numeric_config(width_input, &server.Config.Width)
	})

	height := ""
	if server.Config.Height != 0 {
		height = strconv.Itoa(server.Config.Height) // convert int to string
	}

	height_input := UI.GetObject("height_input").(*gtk.Entry)
	height_input.SetText(height)
	height_input.ConnectChanged(func() {
		update_numeric_config(height_input, &server.Config.Height)
	})

	fps := ""
	if server.Config.FPS != 0 {
		fps = strconv.Itoa(server.Config.FPS) // convert int to string
	}

	fps_input := UI.GetObject("fps_input").(*gtk.Entry)
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
		err := create_camera_info_window(camera_id)
		if err != nil {
			UI.errChannel <- err
		}
	})

	width_label := UI.GetObject("width_label").(*gtk.Label)
	height_label := UI.GetObject("height_label").(*gtk.Label)
	fps_label := UI.GetObject("fps_label").(*gtk.Label)
	format_label := UI.GetObject("format_label").(*gtk.Label)

	grid := UI.GetObject("main_grid").(*gtk.Grid)
	grid.Attach(width_label, 0, 11, 1, 1)
	grid.Attach(width_input, 1, 11, 1, 1)
	grid.Attach(height_label, 0, 12, 1, 1)
	grid.Attach(height_input, 1, 12, 1, 1)
	grid.Attach(fps_label, 0, 13, 1, 1)
	grid.Attach(fps_input, 1, 13, 1, 1)
	grid.Attach(format_label, 0, 14, 1, 1)
	grid.Attach(format_input, 1, 14, 1, 1)
	grid.Attach(camera_info, 1, 15, 1, 1)

	return
}
