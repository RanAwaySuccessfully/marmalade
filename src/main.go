package main

import (
	"fmt"
	"os"
)

//	"github.com/diamondburned/gotk4/pkg/gio/v2"
//	"github.com/diamondburned/gotk4/pkg/gtk/v4"

func main() {
	err_channel := make(chan error, 1)
	go startServer(err_channel)

	err := <-err_channel
	fmt.Fprintf(os.Stderr, "[MARMALADE] %v\n", err)
}

/*
func main() {
	app := gtk.NewApplication("com.github.ranawaysuccessfully.marmalade", gio.ApplicationDefaultFlags)
	app.ConnectActivate(func() { activate(app) })

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func activate(app *gtk.Application) {
	label := gtk.NewLabel("Hello from Go!")
	button := gtk.NewButtonWithLabel("Start")
	button.Connect("clicked", func() {
		go startServer()
		button.SetLabel("Started")
	})

	window := gtk.NewApplicationWindow(app)
	window.SetTitle("Marmalade")
	window.SetChild(label)
	window.SetChild(button)
	window.SetDefaultSize(400, 300)
	window.SetVisible(true)

}
*/
