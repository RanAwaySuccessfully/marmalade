package main

import "C"
import (
	"fmt"
	"marmalade/app/gtk3/ui"
	"marmalade/internal/resources"
	"marmalade/internal/server"
	"os"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
)

var UI Builder

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
		// GTK3 app is ready to go, start building UI from here

		server.Config.Read()

		UI = NewBuilder(ui.App)
		UI.errChannel = make(chan error, 1)

		window := UI.GetObject("main_app").(*gtk.ApplicationWindow)
		app.AddWindow(&window.Window)

		init_webcam_setting()
		init_camera_widgets()
		init_mediapipe_widgets()
		init_ports_settings()

		/* ERROR HANDLING */
		button := UI.GetObject("main_button").(*gtk.Button)
		go error_handler(button, UI.errChannel)

		window.SetVisible(true)
		window.ShowAll()

		listclients_button := UI.GetObject("list_clients_button").(*gtk.Button)
		listclients_button.SetVisible(false)

		camera_notify_expanded()
		mediapipe_notify_expanded()
		ports_notify_expanded()

		UI.gtkBuilder.ConnectSignals(nil)

		go gtk.Main()
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

	dialog.SetVisible(true)

	dialog.ConnectResponse(func(response int) {
		if response == int(gtk.ResponseDeleteEvent) {
			dialog.Close()
		}
	})
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
