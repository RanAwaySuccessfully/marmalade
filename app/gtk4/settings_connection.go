package main

import (
	"marmalade/internal/server"
	"strconv"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type connectionWidgets struct {
	mimic_label *gtk.Label
	mimic_box   *gtk.Box
	vts_label   *gtk.Label
	vts_box     *gtk.Box
	vmc_label   *gtk.Label
	vmc_box     *gtk.Box
}

func create_connection_settings(grid *gtk.Grid, window *gtk.ApplicationWindow) {
	misc_widgets := connectionWidgets{}
	create_mimic_widgets(&misc_widgets)
	create_vts_widgets(&misc_widgets)
	create_vmc_widgets(&misc_widgets)

	misc_row := UI.GetObject("misc_expander").(*gtk.Expander)

	misc_row.Connect("notify::expanded", func() {
		expanded := misc_row.Expanded()
		_, row, _, _ := grid.QueryChild(misc_row)
		row++

		if expanded {
			show_connection_widgets(grid, &misc_widgets, row)
		} else {
			hide_connection_widgets(grid, row)
		}
	})
}

func create_mimic_widgets(widgets *connectionWidgets) {
	mimic_label := gtk.NewLabel("VTS iPhone Mimic:")
	mimic_label.SetHAlign(gtk.AlignStart)

	mimic_enable := gtk.NewSwitch()
	mimic_enable.SetSensitive(false)
	mimic_enable.SetActive(true)

	mimic_port_label := gtk.NewLabel("UDP Port:")
	mimic_port_label.SetHAlign(gtk.AlignEnd)
	mimic_port_label.SetMarginStart(25)

	port := strconv.FormatFloat(server.Config.Port, 'f', 0, 64)
	mimic_port := gtk.NewEntry()
	mimic_port.SetText(port)
	mimic_port.SetPlaceholderText("21412")
	mimic_port.Connect("changed", func() {
		update_numeric_config(mimic_port, &server.Config.Port)
	})

	mimic_settings := gtk.NewButtonFromIconName("applications-system-symbolic")
	mimic_settings.AddCSSClass("smallIcon")
	mimic_settings.SetSensitive(false)

	mimic_box := gtk.NewBox(gtk.OrientationHorizontal, 3)
	mimic_box.Append(mimic_enable)
	mimic_box.Append(mimic_port_label)
	mimic_box.Append(mimic_port)
	mimic_box.Append(mimic_settings)

	widgets.mimic_label = mimic_label
	widgets.mimic_box = mimic_box
}

func create_vts_widgets(widgets *connectionWidgets) {
	mimic_label := gtk.NewLabel("VTS Plugin:")
	mimic_label.SetHAlign(gtk.AlignStart)

	mimic_enable := gtk.NewSwitch()
	mimic_enable.SetSensitive(false)

	mimic_port_label := gtk.NewLabel("UDP Port:")
	mimic_port_label.SetHAlign(gtk.AlignEnd)
	mimic_port_label.SetMarginStart(25)

	mimic_port := gtk.NewEntry()
	mimic_port.SetPlaceholderText("8001")
	mimic_port.SetSensitive(false)

	mimic_settings := gtk.NewButtonFromIconName("applications-system-symbolic")
	mimic_settings.AddCSSClass("smallIcon")
	mimic_settings.SetSensitive(false)

	mimic_box := gtk.NewBox(gtk.OrientationHorizontal, 3)
	mimic_box.Append(mimic_enable)
	mimic_box.Append(mimic_port_label)
	mimic_box.Append(mimic_port)
	mimic_box.Append(mimic_settings)

	widgets.vts_label = mimic_label
	widgets.vts_box = mimic_box
}

func create_vmc_widgets(widgets *connectionWidgets) {
	mimic_label := gtk.NewLabel("VMC Protocol:")
	mimic_label.SetHAlign(gtk.AlignStart)

	mimic_enable := gtk.NewSwitch()
	mimic_enable.SetSensitive(false)

	mimic_port_label := gtk.NewLabel("UDP Port:")
	mimic_port_label.SetHAlign(gtk.AlignEnd)
	mimic_port_label.SetMarginStart(25)

	mimic_port := gtk.NewEntry()
	mimic_port.SetPlaceholderText("39540")
	mimic_port.SetSensitive(false)

	mimic_settings := gtk.NewButtonFromIconName("applications-system-symbolic")
	mimic_settings.AddCSSClass("smallIcon")
	mimic_settings.SetSensitive(false)

	mimic_box := gtk.NewBox(gtk.OrientationHorizontal, 3)
	mimic_box.Append(mimic_enable)
	mimic_box.Append(mimic_port_label)
	mimic_box.Append(mimic_port)
	mimic_box.Append(mimic_settings)

	widgets.vmc_label = mimic_label
	widgets.vmc_box = mimic_box
}

func show_connection_widgets(grid *gtk.Grid, widgets *connectionWidgets, row int) {
	grid.InsertRow(row)
	grid.InsertRow(row + 1)
	grid.InsertRow(row + 2)

	grid.Attach(widgets.mimic_label, 0, row, 1, 1)
	grid.Attach(widgets.mimic_box, 1, row, 1, 1)

	grid.Attach(widgets.vts_label, 0, row+1, 1, 1)
	grid.Attach(widgets.vts_box, 1, row+1, 1, 1)

	grid.Attach(widgets.vmc_label, 0, row+2, 1, 1)
	grid.Attach(widgets.vmc_box, 1, row+2, 1, 1)
}

func hide_connection_widgets(grid *gtk.Grid, row int) {
	grid.RemoveRow(row + 2)
	grid.RemoveRow(row + 1)
	grid.RemoveRow(row)
}
