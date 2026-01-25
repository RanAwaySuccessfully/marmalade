package main

import "C"
import (
	"fmt"
	"marmalade/app/gtk4/ui"
	"marmalade/internal/resources"
	"marmalade/internal/server"
	"os"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

var UI Builder

func main() {
	isIncompatible := gtk.CheckVersion(4, 8, 0)
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

	app := gtk.NewApplication("xyz.randev.marmalade", gio.ApplicationDefaultFlags)
	app.ConnectActivate(func() {
		// GTK4 app is ready to go, start building UI from here

		server.Config.Read()

		UI = NewBuilder(ui.App)
		UI.errChannel = make(chan error, 1)

		window := UI.GetObject("main_app").(*gtk.ApplicationWindow)
		app.AddWindow(&window.Window)

		display := window.Widget.Display()
		css := gtk.NewCSSProvider()
		css.LoadFromData(ui.EmbeddedCSS)
		gtk.StyleContextAddProviderForDisplay(display, css, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

		init_webcam_setting()
		init_camera_widgets()
		init_mediapipe_widgets()
		init_ports_settings()

		init_ports_actions_vmcap(app)
		init_ports_actions_plugin(app)

		/* ERROR HANDLING */
		button := UI.GetObject("main_button").(*gtk.Button)
		go error_handler(button, UI.errChannel)

		window.SetVisible(true)
	})

	defer server.Server.Stop()

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

//export about_button_clicked
func about_button_clicked() {
	version := "v" + resources.EmbeddedVersion

	builder := NewBuilder(ui.DialogAbout)
	dialog := builder.GetObject("about_dialog").(*gtk.AboutDialog)

	artists := make([]string, 0, 1)
	artists = append(artists, "vexamour")

	dialog.AddCreditSection("Logo by", artists)
	dialog.SetVersion(version)
}

//export main_button_clicked
func main_button_clicked() {
	button := UI.GetObject("main_button").(*gtk.Button)
	srv := &server.Server
	started := srv.Started()

	listclients_button := UI.GetObject("list_clients_button").(*gtk.Button)

	if started {
		srv.Stop()
		button.SetLabel("Stopping MediaPipe...")
		button.SetSensitive(false)

		listclients_button.SetVisible(false)
	} else {
		go srv.Start(UI.errChannel, func() {
			button.SetLabel("Stop MediaPipe")
			button.SetSensitive(true)
		})

		button.SetLabel("Starting MediaPipe...")
		button.SetSensitive(false)

		listclients_button.SetVisible(true)
	}
}

//export save_button_clicked
func save_button_clicked() {
	server.Config.Save()
	update_unsaved_config(false)
}

//export listclients_button_clicked
func listclients_button_clicked() {
	listclients_show_dialog()
}
