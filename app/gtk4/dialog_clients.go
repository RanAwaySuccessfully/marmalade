package main

import (
	"marmalade/app/gtk4/ui"
	"marmalade/internal/server"
	"reflect"
	"time"

	"github.com/diamondburned/gotk4/pkg/core/gioutil"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

var clientUtil = gioutil.NewListModelType[server.Client]()

type ClientList = gioutil.ListModel[server.Client]

func listclients_show_dialog() {
	window_ptr := UI.GetObject("listclient_dialog")

	var window *gtk.Window
	var selection *gtk.NoSelection

	model := clientUtil.New()

	if window_ptr == nil {
		// initialize dialog

		UI.gtkBuilder.AddFromString(ui.DialogClients)
		window = UI.GetObject("listclient_dialog").(*gtk.Window)
		column_view := UI.GetObject("listclient_columns").(*gtk.ColumnView)

		listclients_create_factories()

		selection = gtk.NewNoSelection(model)
		column_view.SetModel(selection)

		button := UI.GetObject("listclient_close_button").(*gtk.Button)
		button.ConnectClicked(func() {
			selection.SetModel(nil)
			window.Close()
		})

	} else {
		window = window_ptr.(*gtk.Window)
		column_view := UI.GetObject("listclient_columns").(*gtk.ColumnView)

		selection_model := column_view.Model()
		selection = selection_model.ListModel.Cast().(*gtk.NoSelection)
	}

	window.SetVisible(true)

	go listclients_update_model(window, model)
}

func listclients_update_model(window *gtk.Window, model *ClientList) {
	for {
		if !window.Visible() {
			break
		}

		time.Sleep(1 * time.Second)

		glib.IdleAdd(func() {
			// Remove all
			for i := 0; i < model.Len(); i++ {
				model.Remove(i)
			}

			clients := server.Server.GetClientList()

			for _, client := range clients {
				model.Append(client)
			}
		})
	}
}

func listclients_create_factories() {
	name_factory := columnview_factory_create("Name")
	type_factory := columnview_factory_create("Type")
	source_factory := columnview_factory_create("Source")
	target_factory := columnview_factory_create("Target")

	name_column := UI.GetObject("listclient_column_name").(*gtk.ColumnViewColumn)
	name_column.SetFactory(&name_factory.ListItemFactory)

	type_column := UI.GetObject("listclient_column_type").(*gtk.ColumnViewColumn)
	type_column.SetFactory(&type_factory.ListItemFactory)

	source_column := UI.GetObject("listclient_column_source").(*gtk.ColumnViewColumn)
	source_column.SetFactory(&source_factory.ListItemFactory)

	target_column := UI.GetObject("listclient_column_target").(*gtk.ColumnViewColumn)
	target_column.SetFactory(&target_factory.ListItemFactory)
}

func columnview_factory_create(fieldName string) *gtk.SignalListItemFactory {
	factory := gtk.NewSignalListItemFactory()

	factory.ConnectSetup(func(object *glib.Object) {
		listItem := object.Cast().(*gtk.ListItem)

		label := gtk.NewLabel("")
		label.SetEllipsize(pango.EllipsizeEnd)
		label.SetHAlign(gtk.AlignStart)
		label.SetHExpand(true)

		listItem.SetChild(label)
	})

	factory.ConnectBind(func(object *glib.Object) {
		listItem := object.Cast().(*gtk.ListItem)
		clientObj := clientUtil.ObjectValue(listItem.Item())

		reflectedObj := reflect.ValueOf(clientObj)
		field := reflect.Indirect(reflectedObj).FieldByName(fieldName)

		label := listItem.Child().(*gtk.Label)
		label.SetText(field.String())
	})

	factory.ConnectTeardown(func(object *glib.Object) {
		listItem := object.Cast().(*gtk.ListItem)
		listItem.SetChild(nil)
	})

	return factory
}
