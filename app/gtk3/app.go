package main

import (
	"fmt"
	"marmalade/app/gtk3/ui"
	"marmalade/internal/resources"
	"marmalade/internal/server"
	"os"
	"regexp"
	"strconv"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
)

var savedConfigRevealer *gtk.Revealer

func main() {
	isIncompatible := gtk.CheckVersion(3, 24, 0)
	if isIncompatible != "" {
		fmt.Fprintf(os.Stderr, "[MARMALADE] Incompatible: %s\n", isIncompatible)
		os.Exit(109)
	}

	theme := gtk.NewIconTheme()
	hasIcon := theme.HasIcon("xyz.randev.marmalade")
	if !hasIcon {
		resources.InstallIcon()
	}

	gtk.WindowSetDefaultIconName("xyz.randev.marmalade")

	app := gtk.NewApplication("xyz.randev.marmalade.gtk3", gio.ApplicationDefaultFlags)
	app.ConnectActivate(func() {
		activate(app)
		gtk.Main()
	})

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func activate2(app *gtk.Application) {
	builder := gtk.NewBuilderFromString(ui.App)

	window := get_object(builder, "main_app").(*gtk.ApplicationWindow)
	app.AddWindow(&window.Window)
	window.SetVisible(true)
}

func get_object(builder *gtk.Builder, id string) glib.Objector {
	return builder.GetObject(id).Cast()
}

func activate(app *gtk.Application) {
	server.Config.Read()

	window := gtk.NewApplicationWindow(app)
	window.ConnectDestroy(gtk.MainQuit)
	titlebar := gtk.NewHeaderBar()
	titlebar.SetShowCloseButton(true)

	window.SetTitlebar(titlebar)
	window.SetTitle("Marmalade")
	window.SetResizable(false)
	set_window_size(window)

	about_button := gtk.NewButtonFromIconName("help-about-symbolic", 4)
	titlebar.PackStart(about_button)
	about_button.Connect("clicked", create_about_dialog)

	/* MAIN CONTENT */

	main_box := gtk.NewBox(gtk.OrientationVertical, 0)
	window.Add(main_box)

	grid := gtk.NewGrid()
	grid.SetRowSpacing(7)
	grid.SetColumnSpacing(0)
	grid.SetMarginStart(30)
	grid.SetMarginEnd(30)
	grid.SetMarginTop(15)
	grid.SetMarginBottom(20)
	main_box.PackStart(grid, true, true, 0)

	button := gtk.NewButtonWithLabel("Start MediaPipe")
	button.SetHExpand(true)
	grid.Attach(button, 0, 0, 2, 1)

	err_channel := make(chan error, 1)

	button.Connect("clicked", func() {
		srv := &server.Server
		started := srv.Started()

		if started {
			srv.Stop()
			button.SetLabel("Stopping MediaPipe...")
			button.SetSensitive(false)
		} else {
			go srv.Start(err_channel)
			button.SetLabel("Stop MediaPipe")
		}
	})

	create_webcam_setting(grid, err_channel)
	create_camera_settings(grid, window)
	create_misc_settings(grid, window)

	savedConfigRevealer = gtk.NewRevealer()
	main_box.Add(savedConfigRevealer)
	create_footer()

	/* ERROR HANDLING */

	go error_handler(button, err_channel)

	window.SetVisible(true)
	window.ShowAll()
}

func set_window_size(window *gtk.ApplicationWindow) {
	window.SetSizeRequest(450, 150)
}

func create_footer() {
	footer_box := gtk.NewBox(gtk.OrientationVertical, 5)
	savedConfigRevealer.Add(footer_box)

	action_bar := gtk.NewActionBar()
	footer_box.Add(action_bar)

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

func query_child_row(grid *gtk.Grid, child gtk.Widgetter) int {
	var row64 int64
	value := glib.NewValue(row64)
	grid.ChildGetProperty(child, "top-attach", value)
	row64 = value.GoValueAsType(glib.TypeInt64).(int64)
	return int(row64)
}
