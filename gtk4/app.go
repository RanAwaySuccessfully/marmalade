//go:build withgtk4

package gtk4

import (
	_ "embed"
	"marmalade/server"
	"regexp"
	"strconv"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

var savedConfigRevealer *gtk.Revealer

//go:embed resources/style.css
var EmbeddedCSS string

func Activate(app *gtk.Application) {
	server.Config.Read()

	window := gtk.NewApplicationWindow(app)
	titlebar := gtk.NewHeaderBar()

	display := window.Widget.Display()
	css := gtk.NewCSSProvider()
	css.LoadFromString(EmbeddedCSS)
	gtk.StyleContextAddProviderForDisplay(display, css, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	window.SetTitlebar(titlebar)
	window.SetTitle("Marmalade")
	window.SetResizable(false)
	set_window_size(window)
	window.SetVisible(true)

	about_button := gtk.NewButtonFromIconName("help-about-symbolic")
	titlebar.PackStart(about_button)
	about_button.Connect("clicked", create_about_dialog)

	/* MAIN CONTENT */

	main_box := gtk.NewBox(gtk.OrientationVertical, 0)
	window.SetChild(main_box)

	grid := gtk.NewGrid()
	grid.SetRowSpacing(7)
	grid.SetColumnSpacing(0)
	grid.SetMarginStart(30)
	grid.SetMarginEnd(30)
	grid.SetMarginTop(15)
	grid.SetMarginBottom(20)
	main_box.Append(grid)

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

	create_webcam_setting(grid, err_channel)
	create_camera_settings(grid, window)
	create_misc_settings(grid, window)

	savedConfigRevealer = gtk.NewRevealer()
	main_box.Append(savedConfigRevealer)
	create_footer()

	/* ERROR HANDLING */

	check_venv_folder(window, err_channel)
	go error_handler(button, err_channel)
}

func set_window_size(window *gtk.ApplicationWindow) {
	window.SetDefaultSize(450, 150)
}

func create_footer() {
	footer_box := gtk.NewBox(gtk.OrientationVertical, 5)
	savedConfigRevealer.SetChild(footer_box)

	action_bar := gtk.NewActionBar()
	footer_box.Append(action_bar)

	footer_warning := gtk.NewLabel("You have unsaved changes.")
	action_bar.SetCenterWidget(footer_warning)

	save_button := gtk.NewButtonWithLabel("Save")
	save_button.Connect("clicked", func() {
		server.Config.Save()
		update_unsaved_config(false)
	})

	action_bar.PackEnd(save_button)
}

func update_numeric_config(input *gtk.Entry, target *float64) error {
	value := input.Text()
	if value == "" {
		update_unsaved_config(true)
		*target = 0
		return nil
	}

	validator, err := regexp.Compile(`\D`)
	if err != nil {
		return err
	}

	not_numeric := validator.MatchString(value)
	if not_numeric {
		value = validator.ReplaceAllString(value, "")
		pos := input.Position()
		input.SetText(value)
		input.SetPosition(pos - 1)
		return nil
	}

	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}

	update_unsaved_config(true)

	*target = number
	return nil
}

func update_unsaved_config(value bool) {
	if savedConfigRevealer != nil {
		savedConfigRevealer.SetRevealChild(value)
	}
}
