package main

import "C"
import (
	"marmalade/app/gtk3/ui"
	"marmalade/internal/server"
	"strconv"

	"github.com/diamondburned/gotk4/pkg/gtk/v3"
)

//export create_dialog_vtsapi
func create_dialog_vtsapi() {
	title := "Marmalade - VTS 3rd Party API"
	label := "Marmalade can mimic VTube Studio's \"3rd Party PC Apps\" API, and will connect to any application that is expecting it (usually worded like \"VTube Studio / iPhone\"). Note that there is no support for hand tracking using this type of connection."
	create_dialog_port(title, label, "21412", &server.Config.VTSApi.Port, nil, nil)
}

//export create_dialog_vtsplugin
func create_dialog_vtsplugin() {
	title := "Marmalade - VTS Plugin"
	label := "Marmalade can connect directly to VTube Studio as a plugin. Make sure VTube Studio's Plugin API is enabled and that you have authorized Marmalade to connect."
	create_dialog_port(title, label, "8001", &server.Config.VTSPlugin.Port, &server.Config.VTSPlugin.UseFace, &server.Config.VTSPlugin.UseHand)
}

//export create_dialog_vmcapi
func create_dialog_vmcapi() {
	title := "Marmalade - VMC Protocol"
	label := "Marmalade supports the Virtual Motion Capture (VMC) protocol."
	create_dialog_port(title, label, "39540", &server.Config.VMCApi.Port, &server.Config.VMCApi.UseFace, &server.Config.VMCApi.UseHand)
}

func create_dialog_port(title string, label string, placeholder string, port *int, facem *bool, handm *bool) {
	builder := NewBuilder(ui.DialogPorts)

	window := builder.GetObject("ports_dialog").(*gtk.Window)
	window.SetTitle(title)

	button := builder.GetObject("ports_dialog_close_button").(*gtk.Button)
	button.ConnectClicked(func() {
		window.Close()
	})

	label_element := builder.GetObject("ports_label").(*gtk.Label)
	label_element.SetText(label)

	port_value := ""
	if *port != 0 {
		port_value = strconv.Itoa(*port) // convert int to string
	}

	port_input := builder.GetObject("ports_input").(*gtk.Entry)
	port_input.SetPlaceholderText(placeholder)
	port_input.SetText(port_value)
	port_input.ConnectChanged(func() {
		update_numeric_config(port_input, port)
	})

	if facem != nil {
		facem_input := builder.GetObject("ports_facem").(*gtk.Switch)

		facem_input.SetState(*facem)
		facem_input.ConnectStateSet(func(state bool) bool {
			*facem = state
			update_unsaved_config(true)
			return false
		})
	}

	if handm != nil {
		handm_input := builder.GetObject("ports_handm").(*gtk.Switch)

		handm_input.SetState(*handm)
		handm_input.ConnectStateSet(func(state bool) bool {
			*handm = state
			update_unsaved_config(true)
			return false
		})
	}

	window.SetVisible(true)
	window.ShowAll()

	if facem == nil {
		facem_row := builder.GetObject("ports_facem_row").(*gtk.ListBoxRow)
		facem_row.SetVisible(false)
	}

	if handm == nil {
		handm_row := builder.GetObject("ports_handm_row").(*gtk.ListBoxRow)
		handm_row.SetVisible(false)
	}
}
