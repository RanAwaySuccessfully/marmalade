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
	signals := make(map[*glib.Object]glib.SignalHandle)

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

		signals[object] = signalId
	})

	// NOTE: if using X11, DO NOT PUT BREAKPOINTS INSIDE THIS FUNCTION
	factory.ConnectUnbind(func(object *glib.Object) {
		listItem := object.Cast().(*gtk.ListItem)

		box := listItem.Child().(*gtk.Box)
		label := box.FirstChild().(*gtk.Label)
		label.SetText("")

		icon := box.LastChild().(*gtk.Image)
		icon.SetVisible(false)

		var signalId glib.SignalHandle

		for objPointer, signal := range signals {
			if object.Eq(objPointer) {
				signalId = signal
				input.HandlerDisconnect(signalId)
				delete(signals, objPointer)
			}
		}
	})

	factory.ConnectTeardown(func(object *glib.Object) {
		listItem := object.Cast().(*gtk.ListItem)
		listItem.SetChild(nil)
	})

	return factory
}
