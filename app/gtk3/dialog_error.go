package main

import (
	"errors"
	"fmt"
	"marmalade/app/gtk3/ui"
	"marmalade/internal/server"
	"os"
	"os/exec"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
)

var error_window_count int // I managed to almost crash my computer from too many error messages opening at once...never again

func create_error_window(err error) {
	if error_window_count >= 5 {
		return
	}

	error_window_count++
	builder := NewBuilder(ui.DialogError)

	label := builder.GetObject("error_dialog_label").(*gtk.Label)
	label.SetText(err.Error())

	window := builder.GetObject("error_dialog").(*gtk.Window)

	button := builder.GetObject("error_dialog_close_button").(*gtk.Button)
	button.ConnectClicked(func() {
		window.Close()
	})

	window.ConnectDestroy(func() {
		error_window_count--
	})

	window.SetVisible(true)
	window.ShowAll()
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

			errTitle := "Unknown error while running sub-process."

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
