//go:build withgtk4

package gtk4

import (
	"marmalade/server"
	"regexp"
	"strconv"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func Activate(app *gtk.Application) {
	server.Config.Read()

	window := gtk.NewApplicationWindow(app)
	titlebar := gtk.NewHeaderBar()

	display := window.Widget.Display()
	css := gtk.NewCSSProvider()
	css.LoadFromPath("resources/style.css")
	gtk.StyleContextAddProviderForDisplay(display, css, 0)

	window.SetTitlebar(titlebar)
	window.SetTitle("Marmalade")
	window.SetResizable(false)
	window.SetDefaultSize(500, 150)
	window.SetVisible(true)

	about_button := gtk.NewButtonFromIconName("help-about-symbolic")
	titlebar.PackStart(about_button)
	about_button.Connect("clicked", create_about_dialog)

	/* MAIN CONTENT */

	grid := gtk.NewGrid()
	grid.SetRowSpacing(7)
	grid.SetColumnSpacing(0)
	grid.SetMarginStart(30)
	grid.SetMarginEnd(30)
	grid.SetMarginTop(15)
	grid.SetMarginBottom(20)
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
			server.Config.Save()
			go srv.Start(err_channel)
			button.SetLabel("Stop MediaPipe")
		}
	})

	create_webcam_setting(grid, err_channel)
	create_camera_settings(grid, window)
	create_misc_settings(grid, window)

	/* ERROR HANDLING */

	check_venv_folder(window, err_channel)

	error_window, error_label := create_error_window()
	go error_handler(button, error_window, error_label, err_channel)
}

func update_numeric_config(input *gtk.Entry, target *float64) error {
	value := input.Text()
	validator, err := regexp.Compile(`\D`)
	if err != nil {
		return err
	}

	not_numeric := validator.MatchString(value)
	if not_numeric {
		value = validator.ReplaceAllString(value, "")
		input.SetText(value)
	}

	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}

	*target = number
	return nil
}
