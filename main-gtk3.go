//go:build withgtk3

package main

import (
	"marmalade/gtk3"
	"os"
	"os/exec"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
)

func main() {
	theme := gtk.NewIconTheme()
	hasIcon := theme.HasIcon("xyz.randev.marmalade")
	if !hasIcon {
		cmd := exec.Command(
			"xdg-icon-resource", "install",
			"--novendor",
			"--size", "256",
			"gtk4/resources/icons/marmalade_logo_256.png",
			"xyz.randev.marmalade",
		)

		cmd.Run()
	}

	gtk.WindowSetDefaultIconName("xyz.randev.marmalade")

	app := gtk.NewApplication("xyz.randev.marmalade", gio.ApplicationDefaultFlags)
	app.ConnectActivate(func() { gtk3.Activate(app) })

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
