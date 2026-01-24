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
	update := true

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

		window.ConnectCloseRequest(func() bool {
			update = false
			return false
		})

	} else {
		window = window_ptr.(*gtk.Window)
		column_view := UI.GetObject("listclient_columns").(*gtk.ColumnView)

		selection_model := column_view.Model()
		selection = selection_model.ListModel.Cast().(*gtk.NoSelection)
		selection.SetModel(model)
	}

	window.SetVisible(true)

	go listclients_update_model(window, model, &update)
}

func listclients_update_model(window *gtk.Window, model *ClientList, update *bool) {

	for *update {
		glib.IdleAdd(func() {
			// Remove all
			for i := model.Len(); i > 0; i-- {
				model.Remove(i - 1)
			}

			clients := server.Server.GetClientList()

			for _, client := range clients {
				model.Append(client)
			}
		})

		time.Sleep(1 * time.Second)
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
		field := get_field_of_obj(clientObj, fieldName)
		fieldValue := field.String()

		label := listItem.Child().(*gtk.Label)
		label.SetText(fieldValue)
	})

	factory.ConnectTeardown(func(object *glib.Object) {
		listItem := object.Cast().(*gtk.ListItem)

		listItem.SetChild(nil)
	})

	return factory
}

func get_field_of_obj(obj any, fieldName string) reflect.Value {
	reflected := reflect.ValueOf(obj)
	field := reflect.Indirect(reflected).FieldByName(fieldName)
	return field
}
