//go:build withgtk4

package gtk4

import (
	"fmt"
	"marmalade/server"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type miscWidgets struct {
	model_label *gtk.Label
	model_input *gtk.Entry
	port_label  *gtk.Label
	port_input  *gtk.Entry
	gpu_label   *gtk.Label
	gpu_input   *gtk.Switch
}

func create_misc_settings(grid *gtk.Grid, window *gtk.ApplicationWindow) {
	misc_row := gtk.NewExpander("Misc settings")
	misc_row.AddCSSClass("boldText")
	misc_row.SetMarginTop(5)
	misc_row.SetMarginBottom(5)

	misc_widgets := create_misc_widgets()
	grid.Attach(misc_row, 0, 3, 2, 1)

	misc_row.Connect("notify::expanded", func() {
		expanded := misc_row.Expanded()
		_, row, _, _ := grid.QueryChild(misc_row)
		row++

		if expanded {
			show_misc_widgets(grid, &misc_widgets, row)
		} else {
			hide_misc_widgets(grid, row)
			window.SetDefaultSize(500, 150)
		}
	})
}

func create_misc_widgets() miscWidgets {
	model_label := gtk.NewLabel("Model filename:")
	model_label.SetHAlign(gtk.AlignStart)
	model_input := gtk.NewEntry()
	model := fmt.Sprintf(server.Config.Model)
	model_input.SetText(model)

	model_input.Connect("changed", func() {
		value := model_input.Text()
		server.Config.Model = value
	})

	port_label := gtk.NewLabel("UDP port:")
	port_label.SetHAlign(gtk.AlignStart)
	port_input := gtk.NewEntry()
	port := fmt.Sprintf("%d", int(server.Config.Port))
	port_input.SetText(port)

	port_input.Connect("changed", func() {
		update_numeric_config(port_input, &server.Config.Port)
	})

	gpu_label := gtk.NewLabel("Use GPU?")
	gpu_label.SetHAlign(gtk.AlignStart)
	gpu_input := gtk.NewSwitch()
	gpu_input.SetState(server.Config.UseGpu)

	gpu_input.Connect("state-set", func() {
		state := gpu_input.State()
		server.Config.UseGpu = state
	})

	widgets := miscWidgets{
		model_label,
		model_input,
		port_label,
		port_input,
		gpu_label,
		gpu_input,
	}

	return widgets
}

func show_misc_widgets(grid *gtk.Grid, widgets *miscWidgets, row int) {
	grid.InsertRow(row)
	grid.InsertRow(row + 1)
	grid.InsertRow(row + 2)

	grid.Attach(widgets.model_label, 0, row, 1, 1)
	grid.Attach(widgets.model_input, 1, row, 1, 1)

	grid.Attach(widgets.port_label, 0, row+1, 1, 1)
	grid.Attach(widgets.port_input, 1, row+1, 1, 1)

	grid.Attach(widgets.gpu_label, 0, row+2, 1, 1)
	grid.Attach(widgets.gpu_input, 1, row+2, 1, 1)
}

func hide_misc_widgets(grid *gtk.Grid, row int) {
	grid.RemoveRow(row + 2)
	grid.RemoveRow(row + 1)
	grid.RemoveRow(row)
}
