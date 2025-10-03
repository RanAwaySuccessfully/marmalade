//go:build withgtk4

package gtk4

import (
	"marmalade/server"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func Activate(app *gtk.Application) {
	server.Config.Read()

	window := gtk.NewApplicationWindow(app)
	titlebar := gtk.NewHeaderBar()

	display := window.Widget.Display()
	css := gtk.NewCSSProvider()
	css.LoadFromPath("css/style.css")
	gtk.StyleContextAddProviderForDisplay(display, css, 0)

	window.SetTitlebar(titlebar)
	window.SetTitle("Marmalade")
	window.SetResizable(false)
	window.SetDefaultSize(500, 150)
	window.SetVisible(true)

	about_button := gtk.NewButtonFromIconName("help-about-symbolic")
	titlebar.PackStart(about_button)
	about_button.Connect("clicked", create_about_dialog)

	grid := gtk.NewGrid()
	grid.SetRowSpacing(7)
	grid.SetColumnSpacing(0)
	grid.SetMarginStart(30)
	grid.SetMarginEnd(30)
	grid.SetMarginTop(15)
	grid.SetMarginBottom(20)
	window.SetChild(grid)

	/* WEBCAM */

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
	webcam_label.SetSizeRequest(125, 1)
	webcam_label.SetHAlign(gtk.AlignStart)
	webcam_label.SetXAlign(0)

	webcam_box := gtk.NewBox(gtk.OrientationHorizontal, 3)

	webcam_input := gtk.NewDropDown(nil, nil)
	webcam_input.SetHExpand(true)
	webcam_box.Append(webcam_input)

	webcam_refresh := gtk.NewButtonFromIconName("view-refresh-symbolic")
	webcam_box.Append(webcam_refresh)

	webcam_refresh.Connect("notify::expanded", func() {
		fill_camera_list(webcam_input)
	})

	fill_camera_list(webcam_input)
	grid.Attach(webcam_label, 0, 1, 1, 1)
	grid.Attach(webcam_box, 1, 1, 1, 1)

	/* CAMERA SETTINGS */

	camera_row := gtk.NewExpander("Camera settings")
	camera_row.AddCSSClass("boldText")
	camera_row.SetMarginTop(5)
	camera_row.SetMarginBottom(5)

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
			window.SetDefaultSize(500, 150)
		}
	})

	/* MISC SETTINGS */

	misc_row := gtk.NewExpander("Misc settings")
	misc_row.AddCSSClass("boldText")
	misc_row.SetMarginTop(5)
	misc_row.SetMarginBottom(5)

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
			window.SetDefaultSize(500, 150)
		}
	})

	/* ERROR HANDLING */

	check_venv_folder(window)

	error_window, error_label := create_error_window()
	go error_handler(button, error_window, error_label, err_channel)
}

func create_about_dialog() {
	var authors []string
	authors = append(authors, "RanAwaySuccessfully")

	dialog := gtk.NewAboutDialog()
	dialog.SetProgramName("Marmalade")
	dialog.SetComments("API server for MediaPipe, mimicking VTube Studio for iPhone")
	dialog.SetLogoIconName("dialog-question")
	dialog.SetWebsite("https://github.com/RanAwaySuccessfully/marmalade")
	dialog.SetVersion("version A1")
	dialog.SetAuthors(authors)
	dialog.SetVisible(true)
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
		error_handler() runs inside a goroutine, and if it tries to render a new window in any way shape or form, it will glitch or crash
		so we gotta make sure the window is rendered ahead of time, and it should never unload
	*/

	box := gtk.NewBox(gtk.OrientationVertical, 5)
	box.SetMarginStart(10)
	box.SetMarginEnd(10)
	box.SetMarginTop(5)
	box.SetMarginBottom(7)
	window.SetChild(box)

	label := gtk.NewLabel("(nothing)")
	label.SetWrap(true)
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
