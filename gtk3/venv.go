//go:build withgtk3

package gtk3

import (
	"os"
	"os/exec"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
)

func check_venv_folder(app_window *gtk.ApplicationWindow, err_chan chan error) {
	info, err := os.Stat(".venv")
	if err != nil || !info.IsDir() {

		app_window.SetVisible(false)

		window := gtk.NewWindow(WindowToplevel)
		window.SetTitle("Marmalade - .venv folder missing")
		window.SetDefaultSize(400, 100)
		window.SetResizable(false)
		window.SetVisible(true)

		box := gtk.NewBox(OrientationVertical, 5)
		box.SetMarginStart(10)
		box.SetMarginEnd(10)
		box.SetMarginTop(5)
		box.SetMarginBottom(7)
		window.Add(box)

		label := gtk.NewLabel("The folder .venv is missing. This likely indicates that mediapipe-install.sh has not been run yet. Run it now?")
		label.SetLineWrap(true)
		label.SetVExpand(true)
		box.Add(label)

		button_box := gtk.NewBox(OrientationHorizontal, 5)
		box.Add(button_box)

		button := gtk.NewButtonWithLabel("Yes")
		button.SetHExpand(true)
		button_box.Add(button)

		button_no := gtk.NewButtonWithLabel("No")
		button_no.SetHExpand(true)
		button_box.Add(button_no)

		button_no.Connect("clicked", func() {
			app_window.SetVisible(true)
			window.Close()
		})

		button.Connect("clicked", func() {
			button.SetSensitive(false)
			button_no.SetSensitive(false)

			label.SetText("Installing MediaPipe...")

			spinner := gtk.NewSpinner()
			box.Add(spinner)
			spinner.Start()

			go install_mediapipe(app_window, window, err_chan)
		})
	}
}

func install_mediapipe(app_window *gtk.ApplicationWindow, window *gtk.Window, err_chan chan error) {
	cmd := exec.Command("./mediapipe-install.sh")
	cmd.Dir = "scripts"

	err := cmd.Run()
	if err != nil {
		err_chan <- err
	}

	glib.IdleAdd(func() {
		window.Close()
		app_window.SetVisible(true)
	})
}
