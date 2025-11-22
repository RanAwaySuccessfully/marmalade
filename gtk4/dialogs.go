//go:build withgtk4

package gtk4

import (
	"errors"
	"fmt"
	"marmalade/resources"
	"marmalade/server"
	"os"
	"os/exec"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func create_about_dialog() {
	version := "v" + resources.EmbeddedVersion

	authors := make([]string, 0, 1)
	authors = append(authors, "RanAwaySuccessfully")

	artists := make([]string, 0, 1)
	artists = append(artists, "vexamour")

	dialog := gtk.NewAboutDialog()
	dialog.SetProgramName("Marmalade")
	dialog.SetComments("API server for MediaPipe, mimicking VTube Studio for iPhone")
	dialog.SetWebsite("https://github.com/RanAwaySuccessfully/marmalade")
	dialog.SetWebsiteLabel("GitHub")
	dialog.SetCopyright("© 2025 RanAwaySuccessfully")
	dialog.SetVersion(version)
	dialog.SetAuthors(authors)
	dialog.AddCreditSection("Logo by", artists)

	display := dialog.Widget.Display()
	theme := gtk.IconThemeGetForDisplay(display)
	hasIcon := theme.HasIcon("xyz.randev.marmalade")

	if hasIcon {
		dialog.SetLogoIconName("xyz.randev.marmalade")
	} else {
		gbytes := glib.NewBytesWithGo(resources.EmbeddedAboutLogo)
		texture, err := gdk.NewTextureFromBytes(gbytes)

		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		} else {
			dialog.SetLogo(texture)
		}
	}

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
