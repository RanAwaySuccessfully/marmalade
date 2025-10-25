//go:build withgtk4

package gtk4

import (
	_ "embed"
	"errors"
	"fmt"
	"marmalade/server"
	"os"
	"os/exec"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

//go:embed resources/icons/marmalade_logo.svg
var EmbeddedAboutLogo []byte

func create_about_dialog() {
	authors := make([]string, 0, 1)
	authors = append(authors, "RanAwaySuccessfully")

	artists := make([]string, 0, 1)
	artists = append(artists, "vexamour")

	dialog := gtk.NewAboutDialog()

	gbytes := glib.NewBytesWithGo(EmbeddedAboutLogo)
	texture, err := gdk.NewTextureFromBytes(gbytes)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	} else {
		dialog.SetLogo(texture)
	}

	dialog.SetProgramName("Marmalade")
	dialog.SetComments("API server for MediaPipe, mimicking VTube Studio for iPhone")
	dialog.SetWebsite("https://github.com/RanAwaySuccessfully/marmalade")
	dialog.SetWebsiteLabel("GitHub")
	dialog.SetCopyright("© 2025 RanAwaySuccessfully")
	dialog.SetVersion("v0.3.1")
	dialog.SetAuthors(authors)
	dialog.AddCreditSection("Logo by", artists)
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
		srv := &server.Server
		srv.Stop()
		button.SetLabel("Start MediaPipe")

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

			// .Stderr is empty due to it being collected over on server.Start() at io.Copy(os.Stderr, stderr)
			err = fmt.Errorf("[%d] %s\n%s", exitCode, errTitle, string(exitError.Stderr))
		}

		error_label.SetText(err.Error())
		error_window.SetVisible(true)
	}
}
