package main

import (
	"marmalade/internal/server"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func init_ports_actions_plugin(app *gtk.Application) {
	facem_variant := glib.NewVariantBoolean(server.Config.VTSPlugin.UseFace)
	facem_action := gio.NewSimpleActionStateful("ports_plugin_facem", nil, facem_variant)
	facem_action.ConnectChangeState(func(parameter *glib.Variant) {
		server.Config.VTSPlugin.UseFace = parameter.Boolean()
		facem_action.SetState(parameter)
		update_unsaved_config(true)
		return
	})

	app.ActionMap.AddAction(facem_action)

	handm_variant := glib.NewVariantBoolean(server.Config.VTSPlugin.UseHand)
	handm_action := gio.NewSimpleActionStateful("ports_plugin_handm", nil, handm_variant)
	handm_action.ConnectChangeState(func(parameter *glib.Variant) {
		server.Config.VTSPlugin.UseHand = parameter.Boolean()
		handm_action.SetState(parameter)
		update_unsaved_config(true)
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

	app.ActionMap.AddAction(about_action)
}

func init_ports_actions_vmcap(app *gtk.Application) {
	facem_variant := glib.NewVariantBoolean(server.Config.VMCApi.UseFace)
	facem_action := gio.NewSimpleActionStateful("ports_vmcap_facem", nil, facem_variant)
	facem_action.ConnectChangeState(func(parameter *glib.Variant) {
		server.Config.VMCApi.UseHand = parameter.Boolean()
		facem_action.SetState(parameter)
		update_unsaved_config(true)
		return
	})

	app.ActionMap.AddAction(facem_action)

	handm_variant := glib.NewVariantBoolean(server.Config.VMCApi.UseHand)
	handm_action := gio.NewSimpleActionStateful("ports_vmcap_handm", nil, handm_variant)
	handm_action.ConnectChangeState(func(parameter *glib.Variant) {
		server.Config.VMCApi.UseHand = parameter.Boolean()
		handm_action.SetState(parameter)
		update_unsaved_config(true)
		return
	})

	app.ActionMap.AddAction(handm_action)

	about_action := gio.NewSimpleAction("ports_vmcap_about", nil)
	about_action.ConnectActivate(func(parameter *glib.Variant) {
		settings := UI.GetObject("vmcap_settings").(*gtk.MenuButton)
		settings.SetMenuModel(nil)

		popover := UI.GetObject("vmcap_popover").(*gtk.Popover)
		settings.SetPopover(popover)
		settings.Popup()
	})

	app.ActionMap.AddAction(about_action)
}
