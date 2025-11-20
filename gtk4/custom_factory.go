//go:build withgtk4

package gtk4

import (
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

func create_custom_factory() *gtk.SignalListItemFactory {
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
		stringObj := listItem.Item().Cast().(*gtk.StringObject)

		label := listItem.Child().(*gtk.Label)
		label.SetText(stringObj.String())
	})

	factory.ConnectTeardown(func(object *glib.Object) {
		listItem := object.Cast().(*gtk.ListItem)
		listItem.SetChild(nil)
	})

	return factory
}

func create_custom_list_factory(input *gtk.DropDown) *gtk.SignalListItemFactory {
	factory := gtk.NewSignalListItemFactory()
	signals := make(map[uint]glib.SignalHandle)

	factory.ConnectSetup(func(object *glib.Object) {
		listItem := object.Cast().(*gtk.ListItem)

		box := gtk.NewBox(gtk.OrientationHorizontal, 5)

		label := gtk.NewLabel("")
		label.SetEllipsize(pango.EllipsizeEnd)
		label.SetHAlign(gtk.AlignStart)
		label.SetHExpand(true)
		box.Append(label)

		icon := gtk.NewImageFromIconName("object-select-symbolic")
		icon.SetVisible(false)
		box.Append(icon)

		listItem.SetChild(box)
	})

	factory.ConnectBind(func(object *glib.Object) {
		listItem := object.Cast().(*gtk.ListItem)
		stringObj := listItem.Item().Cast().(*gtk.StringObject)

		box := listItem.Child().(*gtk.Box)
		label := box.FirstChild().(*gtk.Label)
		label.SetText(stringObj.String())

		icon := box.LastChild().(*gtk.Image)
		icon.SetVisible(false)

		/*
			listItem.Selected() will be true if the item is being hovered over, rather than if the item is currently selected or "activated"
			there is no property such as listItem.Activated(), so I'm forced to keep a reference to the DropDown element as a hacky workaround
		*/
		if listItem.Position() == input.Selected() {
			icon.SetVisible(true)
		}

		signalId := input.Connect("notify::selected", func() {
			if listItem.Position() == input.Selected() {
				icon.SetVisible(true)
			} else {
				icon.SetVisible(false)
			}
		})

		index := listItem.Position()
		signals[index] = signalId
	})

	// NOTE: if using X11, DO NOT PUT BREAKPOINTS INSIDE THIS FUNCTION
	factory.ConnectUnbind(func(object *glib.Object) {
		listItem := object.Cast().(*gtk.ListItem)

		box := listItem.Child().(*gtk.Box)
		icon := box.LastChild().(*gtk.Image)
		icon.SetVisible(false)

		index := listItem.Position()
		signalId := signals[index]
		input.HandlerDisconnect(signalId)
		delete(signals, index)
	})

	factory.ConnectTeardown(func(object *glib.Object) {
		listItem := object.Cast().(*gtk.ListItem)
		listItem.SetChild(nil)
	})

	return factory
}
