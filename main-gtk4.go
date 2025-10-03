//go:build withgtk4

package main

import (
	"fmt"
	"marmalade/server"
	"marmalade/v4l2"
	"os"
	"sync"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

var mutex sync.Mutex

func main() {
	server.Config.Read()

	app := gtk.NewApplication("com.github.ranawaysuccessfully.marmalade", gio.ApplicationDefaultFlags)
	app.ConnectActivate(func() { activate(app) })

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func activate(app *gtk.Application) {
	window := gtk.NewApplicationWindow(app)
	window.SetTitle("Marmalade")
	window.SetDefaultSize(500, 150)
	window.SetVisible(true)

	grid := gtk.NewGrid()
	grid.SetColumnHomogeneous(true)
	grid.SetRowSpacing(5)
	grid.SetMarginStart(10)
	grid.SetMarginEnd(10)
	grid.SetMarginTop(5)
	grid.SetMarginBottom(5)
	window.SetChild(grid)

	button := gtk.NewButtonWithLabel("Start MediaPipe")
	button.SetHExpand(true)
	grid.Attach(button, 0, 0, 2, 1)

	err_channel := make(chan error, 1)

	button.Connect("clicked", func() {
		srv := &server.Server
		started := srv.Started()

		if started {
			srv.Stop()
			button.SetLabel("Start MediaPipe")
		} else {
			go srv.Start(err_channel)
			button.SetLabel("Stop MediaPipe")
		}
	})

	webcam_label := gtk.NewLabel("Webcam:")
	webcam_input := gtk.NewDropDown(nil, nil)
	fill_camera_list(webcam_input)
	grid.Attach(webcam_label, 0, 1, 1, 1)
	grid.Attach(webcam_input, 1, 1, 1, 1)

	camera_row := gtk.NewExpander("Camera settings")
	camera_widgets := create_camera_settings()
	grid.Attach(camera_row, 0, 2, 2, 1)

	camera_row.Connect("notify::expanded", func() {
		expanded := camera_row.Expanded()
		_, row, _, _ := grid.QueryChild(camera_row)
		row++

		if expanded {
			show_camera_settings(grid, &camera_widgets, row)
		} else {
			hide_camera_settings(grid, row)
		}
	})

	misc_row := gtk.NewExpander("Misc settings")
	misc_widgets := create_misc_settings()
	grid.Attach(misc_row, 0, 3, 2, 1)

	misc_row.Connect("notify::expanded", func() {
		expanded := misc_row.Expanded()
		_, row, _, _ := grid.QueryChild(misc_row)
		row++

		if expanded {
			show_misc_settings(grid, &misc_widgets, row)
		} else {
			hide_misc_settings(grid, row)
		}
	})

	error_window, error_label := create_error_window()
	go error_handler(button, error_window, error_label, err_channel)
}

func create_error_window() (*gtk.Window, *gtk.Label) {
	window := gtk.NewWindow()
	window.SetTitle("Marmalade - Error")
	window.SetDefaultSize(300, 100)
	window.SetResizable(false)
	window.SetHideOnClose(true)
	window.SetVisible(true)
	window.SetVisible(false)
	/*
		error_handler() is a goroutine, and if it tries to render a new window in any way shape or form, it will glitch or crash
		so we gotta make sure the window is rendered ahead of time, and it should never unload
	*/

	box := gtk.NewBox(gtk.OrientationVertical, 5)
	window.SetChild(box)

	label := gtk.NewLabel("(nothing)")
	label.SetVExpand(true)
	box.Append(label)

	button := gtk.NewButton()
	button.SetLabel("Close")
	box.Append(button)

	button.Connect("clicked", func() {
		window.SetVisible(false)
	})

	return window, label
}

func error_handler(button *gtk.Button, error_window *gtk.Window, error_label *gtk.Label, err_channel chan error) {
	for err := range err_channel {
		error_label.SetText(err.Error())
		error_window.SetVisible(true)

		srv := &server.Server
		srv.Stop()

		button.SetLabel("Start MediaPipe")
	}
}

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
	width_input := gtk.NewEntry()
	width := fmt.Sprintf("%d", int(server.Config.Width))
	width_input.SetText(width)

	height_label := gtk.NewLabel("Height:")
	height_input := gtk.NewEntry()
	height := fmt.Sprintf("%d", int(server.Config.Height))
	height_input.SetText(height)

	fps_label := gtk.NewLabel("FPS:")
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
	model_input := gtk.NewEntry()
	model := fmt.Sprintf(server.Config.Model)
	model_input.SetText(model)

	port_label := gtk.NewLabel("UDP port:")
	port_input := gtk.NewEntry()
	port := fmt.Sprintf("%d", int(server.Config.Port))
	port_input.SetText(port)

	gpu_label := gtk.NewLabel("Use GPU?")
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
