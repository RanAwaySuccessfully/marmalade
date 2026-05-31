package main

import "C"
import (
	"marmalade/app/gtk4/ui"
	"marmalade/internal/server"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

//export ports_notify_expanded
func ports_notify_expanded() {
	misc_row := UI.GetObject("misc_expander").(*gtk.Expander)
	expanded := misc_row.Expanded()

	vts3p_label := UI.GetObject("vts3p_label").(*gtk.Label)
	vts3p_box := UI.GetObject("vts3p_box").(*gtk.Box)
	plugin_label := UI.GetObject("vtsplugin_label").(*gtk.Label)
	plugin_box := UI.GetObject("vtsplugin_box").(*gtk.Box)
	vmcap_label := UI.GetObject("vmcap_label").(*gtk.Label)
	vmcap_box := UI.GetObject("vmcap_box").(*gtk.Box)
	vrcosc_label := UI.GetObject("vrcosc_label").(*gtk.Label)
	vrcosc_box := UI.GetObject("vrcosc_box").(*gtk.Box)

	if expanded {
		vts3p_label.SetVisible(true)
		vts3p_box.SetVisible(true)
		plugin_label.SetVisible(true)
		plugin_box.SetVisible(true)
		vmcap_label.SetVisible(true)
		vmcap_box.SetVisible(true)
		vrcosc_label.SetVisible(true)
		vrcosc_box.SetVisible(true)
	} else {
		vts3p_label.SetVisible(false)
		vts3p_box.SetVisible(false)
		plugin_label.SetVisible(false)
		plugin_box.SetVisible(false)
		vmcap_label.SetVisible(false)
		vmcap_box.SetVisible(false)
		vrcosc_label.SetVisible(false)
		vrcosc_box.SetVisible(false)
	}
}

//export ports_vtsplugin_popover_closed
func ports_vtsplugin_popover_closed() {
	plugin_settings := UI.GetObject("vtsplugin_settings").(*gtk.MenuButton)
	menu_model := UI.GetObject("vtsplugin_menu").(*gio.Menu)

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

//export ports_vrcosc_popover_closed
func ports_vrcosc_popover_closed() {
	vrcosc_settings := UI.GetObject("vrcosc_settings").(*gtk.MenuButton)
	menu_model := UI.GetObject("vrcosc_menu").(*gio.Menu)

	vrcosc_settings.SetMenuModel(menu_model)
	vrcosc_settings.Popdown()
}

func init_ports_settings() {
	UI.gtkBuilder.AddFromString(ui.SettingsPorts)

	grid := UI.GetObject("main_grid").(*gtk.Grid)
	init_ports_settings_vtsapi(grid)
	init_ports_settings_vtsplugin(grid)
	init_ports_settings_vmcap(grid)
	init_ports_settings_vrcosc(grid)
}

func init_ports_settings_vtsapi(grid *gtk.Grid) {
	vtsapi_switch := UI.GetObject("vts3p_enable").(*gtk.Switch)
	vtsapi_switch.SetActive(server.Config.VTSApi.Enabled)
	vtsapi_switch.ConnectStateSet(func(state bool) bool {
		server.Config.VTSApi.Enabled = state
		update_unsaved_config(true)
		return false
	})

	vtsapi_port_value := ""
	if server.Config.VTSApi.Port != 0 {
		vtsapi_port_value = int_to_string(server.Config.VTSApi.Port)
	}

	vtsapi_port := UI.GetObject("vts3p_port").(*gtk.Entry)
	vtsapi_port.SetText(vtsapi_port_value)
	vtsapi_port.ConnectChanged(func() {
		update_numeric_config(vtsapi_port, &server.Config.VTSApi.Port)
	})

	vts3p_label := UI.GetObject("vts3p_label").(*gtk.Label)
	vts3p_box := UI.GetObject("vts3p_box").(*gtk.Box)
	grid.Attach(vts3p_label, 0, 31, 1, 1)
	grid.Attach(vts3p_box, 1, 31, 1, 1)
}

func init_ports_settings_vtsplugin(grid *gtk.Grid) {
	plugin_switch := UI.GetObject("vtsplugin_enable").(*gtk.Switch)
	plugin_switch.SetActive(server.Config.VTSPlugin.Enabled)
	plugin_switch.ConnectStateSet(func(state bool) bool {
		server.Config.VTSPlugin.Enabled = state
		update_unsaved_config(true)
		return false
	})

	plugin_port_value := ""
	if server.Config.VTSPlugin.Port != 0 {
		plugin_port_value = int_to_string(server.Config.VTSPlugin.Port)
	}

	plugin_port := UI.GetObject("vtsplugin_port").(*gtk.Entry)
	plugin_port.SetText(plugin_port_value)
	plugin_port.ConnectChanged(func() {
		update_numeric_config(plugin_port, &server.Config.VTSPlugin.Port)
	})

	plugin_label := UI.GetObject("vtsplugin_label").(*gtk.Label)
	plugin_box := UI.GetObject("vtsplugin_box").(*gtk.Box)
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

	vmcap_port_value := ""
	if server.Config.VMCApi.Port != 0 {
		vmcap_port_value = int_to_string(server.Config.VMCApi.Port)
	}

	vmcap_port := UI.GetObject("vmcap_port").(*gtk.Entry)
	vmcap_port.SetText(vmcap_port_value)
	vmcap_port.ConnectChanged(func() {
		update_numeric_config(vmcap_port, &server.Config.VMCApi.Port)
	})

	vmcap_label := UI.GetObject("vmcap_label").(*gtk.Label)
	vmcap_box := UI.GetObject("vmcap_box").(*gtk.Box)
	grid.Attach(vmcap_label, 0, 33, 1, 1)
	grid.Attach(vmcap_box, 1, 33, 1, 1)
}

func init_ports_settings_vrcosc(grid *gtk.Grid) {
	vrcosc_switch := UI.GetObject("vrcosc_enable").(*gtk.Switch)
	vrcosc_switch.SetActive(server.Config.VRChatOSC.Enabled)
	vrcosc_switch.ConnectStateSet(func(state bool) bool {
		server.Config.VRChatOSC.Enabled = state
		update_unsaved_config(true)
		return false
	})

	vrcosc_port_value := ""
	if server.Config.VRChatOSC.Port != 0 {
		vrcosc_port_value = int_to_string(server.Config.VRChatOSC.Port)
	}

	vrcosc_port := UI.GetObject("vrcosc_port").(*gtk.Entry)
	vrcosc_port.SetText(vrcosc_port_value)
	vrcosc_port.ConnectChanged(func() {
		update_numeric_config(vrcosc_port, &server.Config.VRChatOSC.Port)
	})

	vrcosc_label := UI.GetObject("vrcosc_label").(*gtk.Label)
	vrcosc_box := UI.GetObject("vrcosc_box").(*gtk.Box)
	grid.Attach(vrcosc_label, 0, 34, 1, 1)
	grid.Attach(vrcosc_box, 1, 34, 1, 1)
}
