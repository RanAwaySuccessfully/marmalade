package main

import (
	"errors"
	"fmt"
	"marmalade/app/gtk4/ui"
	"marmalade/internal/server"
	"os"
	"os/exec"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func create_error_window(err error) {
	builder := NewBuilder(ui.DialogError)

	label := builder.GetObject("error_dialog_label").(*gtk.Label)
	label.SetText(err.Error())

	button := builder.GetObject("error_dialog_close_button").(*gtk.Button)
	button.ConnectClicked(func() {
		window := builder.GetObject("error_dialog").(*gtk.Window)
		window.Close()
	})
}

func error_handler(button *gtk.Button, err_channel chan error) {
	for err := range err_channel {
		srv := &server.Server
		srv.Stop()

		glib.IdleAdd(func() {
			button.SetLabel("Start MediaPipe")
			button.SetSensitive(true)
		})

		if errors.Is(err, os.ErrProcessDone) {
			continue
		}

		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			exitError = err.(*exec.ExitError)
			exitCode := exitError.ExitCode()

			errTitle := "Unknown error while running python process."

			switch exitCode {
			case 110:
				errTitle = "Unable to connect to camera."
			case 111:
				errTitle = "Unable to start MediaPipe. Is the model (.task) file configured correctly?"
			case 112:
				errTitle = "A client appears to have disconnected."
			case 113:
				errTitle = "Too many failed attempts at reading an image from the camera."
			}

			// exitError.Stderr is empty, so we use our own copy of Stderr instead
			err = fmt.Errorf("[%d] %s\n%s", exitCode, errTitle, srv.ErrPipe.Log)
		}

		glib.IdleAdd(func() {
			create_error_window(err)
		})
	}
}
