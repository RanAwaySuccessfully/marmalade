//go:build withgtk4

package gtk4

import (
	"os"
	"os/exec"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func check_venv_folder(app_window *gtk.ApplicationWindow, err_chan chan error) bool {
	info, err := os.Stat(".venv")
	if err != nil || !info.IsDir() {

		window := gtk.NewWindow()
		window.SetTitle("Marmalade - .venv folder missing")
		window.SetDefaultSize(400, 100)
		window.SetResizable(false)
		window.SetVisible(true)

		box := gtk.NewBox(gtk.OrientationVertical, 5)
		box.SetMarginStart(10)
		box.SetMarginEnd(10)
		box.SetMarginTop(5)
		box.SetMarginBottom(7)
		window.SetChild(box)

		label := gtk.NewLabel("The folder .venv is missing. This likely indicates that mediapipe-install.sh has not been run yet. Run it now?")
		label.SetWrap(true)
		label.SetVExpand(true)
		box.Append(label)

		button_box := gtk.NewBox(gtk.OrientationHorizontal, 5)
		box.Append(button_box)

		button := gtk.NewButtonWithLabel("Yes")
		button.SetHExpand(true)
		button_box.Append(button)

		button_no := gtk.NewButtonWithLabel("No")
		button_no.SetHExpand(true)
		button_box.Append(button_no)

		close_dialog := func() {
			app_window.SetVisible(true)
			window.Close()
		}

		button_no.Connect("clicked", close_dialog)

		button.Connect("clicked", func() {
			button.SetSensitive(false)
			button_no.SetSensitive(false)

			label.SetText("Installing MediaPipe...")

			spinner := gtk.NewSpinner()
			spinner.InsertBefore(box, label)
			spinner.Start()

			go install_mediapipe(close_dialog, err_chan)
		})

		return false
	}

	return true
}

func install_mediapipe(close_dialog func(), err_chan chan error) {
	cmd := exec.Command("./mediapipe-install.sh")
	cmd.Dir = "scripts"

	err := cmd.Run()
	if err != nil {
		err_chan <- err
	}

	glib.IdleAdd(close_dialog)
}
