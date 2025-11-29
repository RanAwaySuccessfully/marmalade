//go:build withgtk3

package main

import (
	"fmt"
	"marmalade/gtk3"
	"marmalade/resources"
	"os"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
)

func main() {
	isIncompatible := gtk.CheckVersion(3, 24, 0)
	if isIncompatible != "" {
		fmt.Fprintf(os.Stderr, "[MARMALADE] Incompatible: %s\n", isIncompatible)
		os.Exit(109)
	}

	theme := gtk.NewIconTheme()
	hasIcon := theme.HasIcon("xyz.randev.marmalade")
	if !hasIcon {
		resources.InstallIcon()
	}

	gtk.WindowSetDefaultIconName("xyz.randev.marmalade")

	app := gtk.NewApplication("xyz.randev.marmalade", gio.ApplicationDefaultFlags)
	app.ConnectActivate(func() {
		gtk3.Activate(app)
		gtk.Main()
	})

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}
