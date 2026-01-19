package main

import (
	"fmt"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func init_ports_actions_plugin(app *gtk.Application) {
	facem_variant := glib.NewVariantBoolean(true)
	facem_action := gio.NewSimpleActionStateful("ports_plugin_facem", nil, facem_variant)
	facem_action.ConnectActivate(func(parameter *glib.Variant) {
		fmt.Println("ports_plugin_facem")
		return
	})

	app.ActionMap.AddAction(facem_action)

	handm_variant := glib.NewVariantBoolean(true)
	handm_action := gio.NewSimpleActionStateful("ports_plugin_handm", nil, handm_variant)
	handm_action.ConnectActivate(func(parameter *glib.Variant) {
		fmt.Println("ports_plugin_handm")
		return
	})

	app.ActionMap.AddAction(handm_action)

	about_action := gio.NewSimpleAction("ports_plugin_about", nil)
	about_action.ConnectActivate(func(parameter *glib.Variant) {
		settings := UI.GetObject("plugin_settings").(*gtk.MenuButton)
		settings.SetMenuModel(nil)

		popover := UI.GetObject("plugin_popover").(*gtk.Popover)
		settings.SetPopover(popover)
		settings.Popup()
	})

	app.ActionMap.AddAction(facem_action)
	app.ActionMap.AddAction(about_action)
}

func init_ports_actions_vmcap(app *gtk.Application) {
	facem_variant := glib.NewVariantBoolean(true)
	facem_action := gio.NewSimpleActionStateful("ports_vmcap_facem", nil, facem_variant)
	facem_action.ConnectActivate(func(parameter *glib.Variant) {
		fmt.Println("ports_vmcap_facem")
		return
	})

	app.ActionMap.AddAction(facem_action)

	about_action := gio.NewSimpleAction("ports_vmcap_about", nil)
	about_action.ConnectActivate(func(parameter *glib.Variant) {
		settings := UI.GetObject("vmcap_settings").(*gtk.MenuButton)
		settings.SetMenuModel(nil)

		popover := UI.GetObject("vmcap_popover").(*gtk.Popover)
		settings.SetPopover(popover)
		settings.Popup()
	})

	app.ActionMap.AddAction(facem_action)
	app.ActionMap.AddAction(about_action)
}
