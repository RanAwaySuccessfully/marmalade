package main

import "C"
import (
	"marmalade/app/gtk4/ui"
	"marmalade/internal/server"
	"strconv"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

//export ports_notify_expanded
func ports_notify_expanded() {
	grid := UI.GetObject("main_grid").(*gtk.Grid)
	misc_row := UI.GetObject("misc_expander").(*gtk.Expander)

	expanded := misc_row.Expanded()
	_, row, _, _ := grid.QueryChild(misc_row)
	row++

	if expanded {
		show_connection_widgets(grid, row)
	} else {
		hide_connection_widgets(grid, row)
	}
}

//export ports_plugin_popover_closed
func ports_plugin_popover_closed() {
	plugin_settings := UI.GetObject("plugin_settings").(*gtk.MenuButton)
	menu_model := UI.GetObject("plugin_menu").(*gio.Menu)

	plugin_settings.SetMenuModel(menu_model)
	plugin_settings.Popdown()
}

//export ports_vmcap_popover_closed
func ports_vmcap_popover_closed() {
	vmcap_settings := UI.GetObject("vmcap_settings").(*gtk.MenuButton)
	menu_model := UI.GetObject("vmcap_menu").(*gio.Menu)

	vmcap_settings.SetMenuModel(menu_model)
	vmcap_settings.Popdown()
}

func init_ports_settings() {
	UI.gtkBuilder.AddFromString(ui.SettingsPorts)

	port := strconv.FormatFloat(server.Config.Port, 'f', 0, 64)
	mimic_port := UI.GetObject("mimic_port").(*gtk.Entry)
	mimic_port.SetText(port)
	mimic_port.Connect("changed", func() {
		update_numeric_config(mimic_port, &server.Config.Port)
	})
}

func init_ports_actions(app *gtk.Application) {
	init_ports_actions_generic(app, "plugin")
	init_ports_actions_generic(app, "vmcap")
}

func init_ports_actions_generic(app *gtk.Application, conn_type string) {
	facem_variant := glib.NewVariantBoolean(true)
	facem_action := gio.NewSimpleActionStateful("ports_"+conn_type+"_facem", nil, facem_variant)
	facem_action.ConnectActivate(func(parameter *glib.Variant) {
		println("ports_" + conn_type + "_facem")
		return
	})

	app.ActionMap.AddAction(facem_action)

	about_action := gio.NewSimpleAction("ports_"+conn_type+"_about", nil)
	about_action.ConnectActivate(func(parameter *glib.Variant) {
		settings := UI.GetObject(conn_type + "_settings").(*gtk.MenuButton)
		settings.SetMenuModel(nil)

		popover := UI.GetObject(conn_type + "_popover").(*gtk.Popover)
		settings.SetPopover(popover)
		settings.Popup()
	})

	app.ActionMap.AddAction(facem_action)
	app.ActionMap.AddAction(about_action)
}

func show_connection_widgets(grid *gtk.Grid, row int) {
	grid.InsertRow(row)
	grid.InsertRow(row + 1)
	grid.InsertRow(row + 2)

	mimic_label := UI.GetObject("mimic_label").(*gtk.Label)
	mimic_box := UI.GetObject("mimic_box").(*gtk.Box)
	grid.Attach(mimic_label, 0, row, 1, 1)
	grid.Attach(mimic_box, 1, row, 1, 1)

	plugin_label := UI.GetObject("plugin_label").(*gtk.Label)
	plugin_box := UI.GetObject("plugin_box").(*gtk.Box)
	grid.Attach(plugin_label, 0, row+1, 1, 1)
	grid.Attach(plugin_box, 1, row+1, 1, 1)

	vmcap_label := UI.GetObject("vmcap_label").(*gtk.Label)
	vmcap_box := UI.GetObject("vmcap_box").(*gtk.Box)
	grid.Attach(vmcap_label, 0, row+2, 1, 1)
	grid.Attach(vmcap_box, 1, row+2, 1, 1)
}

func hide_connection_widgets(grid *gtk.Grid, row int) {
	grid.RemoveRow(row + 2)
	grid.RemoveRow(row + 1)
	grid.RemoveRow(row)
}
