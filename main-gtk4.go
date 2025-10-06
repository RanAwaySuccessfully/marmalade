//go:build withgtk4

package main

import (
	"marmalade/gtk4"
	"os"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func main() {
	app := gtk.NewApplication("xyz.randev.marmalade", gio.ApplicationDefaultFlags)
	app.ConnectActivate(func() { gtk4.Activate(app) })

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
