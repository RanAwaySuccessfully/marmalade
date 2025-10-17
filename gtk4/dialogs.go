//go:build withgtk4

package gtk4

import (
	"errors"
	"marmalade/server"
	"os"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func create_about_dialog() {
	authors := make([]string, 0, 1)
	authors = append(authors, "RanAwaySuccessfully")

	artists := make([]string, 0, 1)
	artists = append(artists, "vexamour")

	logo_file := gtk.NewPictureForFilename("resources/icons/marmalade_logo.svg")
	logo := logo_file.Paintable()

	dialog := gtk.NewAboutDialog()
	dialog.SetProgramName("Marmalade")
	dialog.SetComments("API server for MediaPipe, mimicking VTube Studio for iPhone")
	dialog.SetLogo(logo)
	dialog.SetWebsite("https://github.com/RanAwaySuccessfully/marmalade")
	dialog.SetWebsiteLabel("GitHub")
	dialog.SetCopyright("© 2025 RanAwaySuccessfully")
	dialog.SetVersion("v0.2.0")
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

		error_label.SetText(err.Error())
		error_window.SetVisible(true)
	}
}
