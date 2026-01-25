package main

import (
	"marmalade/app/gtk3/ui"
	"marmalade/internal/server"
	"time"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v3"
)

var dialog_client_isopen bool

func listclients_show_dialog() {
	if dialog_client_isopen {
		return
	} else {
		dialog_client_isopen = true
	}

	columns := []glib.Type{
		glib.TypeString,
		glib.TypeString,
		glib.TypeString,
		glib.TypeString,
	}

	model := gtk.NewTreeStore(columns)
	update := true

	_, err := UI.gtkBuilder.AddFromString(ui.DialogClients)
	if err != nil {
		UI.errChannel <- err
		return
	}

	window := UI.GetObject("listclient_dialog").(*gtk.Window)
	column_view := UI.GetObject("listclient_columns").(*gtk.TreeView)
	column_view.SetModel(model)

	button := UI.GetObject("listclient_close_button").(*gtk.Button)
	button.ConnectClicked(func() {
		column_view.SetModel(nil)
		window.Close()
	})

	window.ConnectDestroy(func() {
		update = false
		dialog_client_isopen = false
	})

	window.SetVisible(true)
	window.ShowAll()

	go listclients_update_model(window, model, &update)
}

func listclients_update_model(window *gtk.Window, model *gtk.TreeStore, update *bool) {

	for *update {
		glib.IdleAdd(func() {
			iter, ok := model.IterFirst()
			for ok {
				ok = model.Remove(iter)
			}

			clients := server.Server.GetClientList()

			for _, client := range clients {
				iter := model.Insert(nil, -1)

				name := glib.NewValue(client.Name)
				typev := glib.NewValue(client.Type)
				source := glib.NewValue(client.Source)
				target := glib.NewValue(client.Target)

				columns := []int{0, 1, 2, 3}
				values := []glib.Value{*name, *typev, *source, *target}

				model.Set(iter, columns, values)
			}
		})

		time.Sleep(1 * time.Second)
	}
}
