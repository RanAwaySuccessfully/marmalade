//go:build withgtk4

package gtk4

import (
	_ "embed"
	"errors"
	"fmt"
	"marmalade/server"
	"os"
	"os/exec"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

//go:embed resources/icons/marmalade_logo.svg
var EmbeddedAboutLogo []byte

//go:embed resources/version.txt
var EmbeddedVersion string

func create_about_dialog() {
	version := "v" + EmbeddedVersion

	authors := make([]string, 0, 1)
	authors = append(authors, "RanAwaySuccessfully")

	artists := make([]string, 0, 1)
	artists = append(artists, "vexamour")

	dialog := gtk.NewAboutDialog()
	dialog.SetLogoIconName("xyz.randev.marmalade")
	dialog.SetProgramName("Marmalade")
	dialog.SetComments("API server for MediaPipe, mimicking VTube Studio for iPhone")
	dialog.SetWebsite("https://github.com/RanAwaySuccessfully/marmalade")
	dialog.SetWebsiteLabel("GitHub")
	dialog.SetCopyright("© 2025 RanAwaySuccessfully")
	dialog.SetVersion(version)
	dialog.SetAuthors(authors)
	dialog.AddCreditSection("Logo by", artists)

	dialog.SetVisible(true)
}

func create_error_window(err error) {
	window := gtk.NewWindow()
	window.SetTitle("Marmalade - Error")
	window.SetDefaultSize(300, 100)
	window.SetResizable(false)
	window.SetHideOnClose(true)
	window.SetVisible(true)

	box := gtk.NewBox(gtk.OrientationVertical, 5)
	box.SetMarginStart(10)
	box.SetMarginEnd(10)
	box.SetMarginTop(5)
	box.SetMarginBottom(7)
	window.SetChild(box)

	label := gtk.NewLabel(err.Error())
	label.SetWrap(true)
	label.SetVExpand(true)
	box.Append(label)

	button := gtk.NewButton()
	button.SetLabel("Close")
	box.Append(button)

	button.Connect("clicked", func() {
		window.Close()
	})
}

func error_handler(button *gtk.Button, err_channel chan error) {
	for err := range err_channel {
		srv := &server.Server
		srv.Stop()
		button.SetLabel("Start MediaPipe")
		// updating a label's text outside of glib.IdleAdd() could cause a crash...but it seems to be working fine so far

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

			// TODO: .Stderr is empty due to it being collected over on server.Start() at io.Copy(os.Stderr, stderr)
			err = fmt.Errorf("[%d] %s\n%s", exitCode, errTitle, string(exitError.Stderr))
		}

		glib.IdleAdd(func() {
			create_error_window(err)
		})
	}
}
