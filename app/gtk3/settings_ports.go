package main

import "C"
import (
	"marmalade/app/gtk3/ui"
	"marmalade/internal/server"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
)

//export ports_notify_expanded
func ports_notify_expanded() {
	misc_row := UI.GetObject("misc_expander").(*gtk.Expander)
	expanded := misc_row.Expanded()

	mimic_label := UI.GetObject("mimic_label").(*gtk.Label)
	mimic_box := UI.GetObject("mimic_box").(*gtk.Box)
	plugin_label := UI.GetObject("plugin_label").(*gtk.Label)
	plugin_box := UI.GetObject("plugin_box").(*gtk.Box)
	vmcap_label := UI.GetObject("vmcap_label").(*gtk.Label)
	vmcap_box := UI.GetObject("vmcap_box").(*gtk.Box)

	if expanded {
		mimic_label.SetVisible(true)
		mimic_box.SetVisible(true)
		plugin_label.SetVisible(true)
		plugin_box.SetVisible(true)
		vmcap_label.SetVisible(true)
		vmcap_box.SetVisible(true)
	} else {
		mimic_label.SetVisible(false)
		mimic_box.SetVisible(false)
		plugin_label.SetVisible(false)
		plugin_box.SetVisible(false)
		vmcap_label.SetVisible(false)
		vmcap_box.SetVisible(false)
	}

	set_window_size()
}

//export ports_plugin_popover_closed
func ports_plugin_popover_closed() {
	plugin_settings := UI.GetObject("plugin_settings").(*gtk.MenuButton)
	menu_model := UI.GetObject("plugin_menu").(*gio.Menu)

	plugin_settings.SetMenuModel(menu_model)
	//plugin_settings.Popdown()
}

//export ports_vmcap_popover_closed
func ports_vmcap_popover_closed() {
	vmcap_settings := UI.GetObject("vmcap_settings").(*gtk.MenuButton)
	menu_model := UI.GetObject("vmcap_menu").(*gio.Menu)

	vmcap_settings.SetMenuModel(menu_model)
	//vmcap_settings.Popdown()
}

func init_ports_settings() {
	UI.gtkBuilder.AddFromString(ui.SettingsPorts)

	grid := UI.GetObject("main_grid").(*gtk.Grid)
	init_ports_settings_vtsapi(grid)
	init_ports_settings_vtsplugin(grid)
	init_ports_settings_vmcap(grid)
}

func init_ports_settings_vtsapi(grid *gtk.Grid) {
	vtsapi_switch := UI.GetObject("mimic_enable").(*gtk.Switch)
	vtsapi_switch.SetActive(server.Config.VTSApi.Enabled)
	vtsapi_switch.ConnectStateSet(func(state bool) bool {
		server.Config.VTSApi.Enabled = state
		update_unsaved_config(true)
		return false
	})

	mimic_label := UI.GetObject("mimic_label").(*gtk.Label)
	mimic_box := UI.GetObject("mimic_box").(*gtk.Box)
	grid.Attach(mimic_label, 0, 31, 1, 1)
	grid.Attach(mimic_box, 1, 31, 1, 1)
}

func init_ports_settings_vtsplugin(grid *gtk.Grid) {
	plugin_switch := UI.GetObject("plugin_enable").(*gtk.Switch)
	plugin_switch.SetActive(server.Config.VTSPlugin.Enabled)
	plugin_switch.ConnectStateSet(func(state bool) bool {
		server.Config.VTSPlugin.Enabled = state
		update_unsaved_config(true)
		return false
	})

	plugin_label := UI.GetObject("plugin_label").(*gtk.Label)
	plugin_box := UI.GetObject("plugin_box").(*gtk.Box)
	grid.Attach(plugin_label, 0, 32, 1, 1)
	grid.Attach(plugin_box, 1, 32, 1, 1)
}

func init_ports_settings_vmcap(grid *gtk.Grid) {
	vmcap_switch := UI.GetObject("vmcap_enable").(*gtk.Switch)
	vmcap_switch.SetActive(server.Config.VMCApi.Enabled)
	vmcap_switch.ConnectStateSet(func(state bool) bool {
		server.Config.VMCApi.Enabled = state
		update_unsaved_config(true)
		return false
	})

	vmcap_label := UI.GetObject("vmcap_label").(*gtk.Label)
	vmcap_box := UI.GetObject("vmcap_box").(*gtk.Box)
	grid.Attach(vmcap_label, 0, 33, 1, 1)
	grid.Attach(vmcap_box, 1, 33, 1, 1)
}
