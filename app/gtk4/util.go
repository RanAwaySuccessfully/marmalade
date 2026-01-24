package main

import (
	"regexp"
	"strconv"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// BUILDER

type Builder struct {
	gtkBuilder *gtk.Builder
	errChannel chan error
}

func NewBuilder(xml string) Builder {
	b := Builder{}

	if xml != "" {
		b.gtkBuilder = gtk.NewBuilderFromString(xml)
	}

	return b
}

func (b *Builder) Add(xml string) {
	if b.gtkBuilder == nil {
		b.gtkBuilder = gtk.NewBuilderFromString(xml)
	} else {
		err := b.gtkBuilder.AddFromString(xml)
		if err != nil {
			panic(err)
		}
	}
}

func (b *Builder) GetObject(id string) glib.Objector {
	return b.gtkBuilder.GetObject(id).Cast()
}

// SIGNAL MAP

type SignalMap map[uint]glib.SignalHandle

func (signalPtr *SignalMap) Add(index uint, signal glib.SignalHandle) {
	signals := *signalPtr
	signals[index] = signal
}

func (signalPtr *SignalMap) Remove(index uint) glib.SignalHandle {
	signals := *signalPtr
	signalId := signals[index]
	delete(signals, index)

	return signalId
}

// MISCELLANEOUS

func update_numeric_config(input *gtk.Entry, target *int) error {
	value := input.Text()
	if value == "" {
		update_unsaved_config(true)
		*target = 0
		return nil
	}

	validator, err := regexp.Compile(`\D`)
	if err != nil {
		return err
	}

	not_numeric := validator.MatchString(value)
	if not_numeric {
		value = validator.ReplaceAllString(value, "")
		pos := input.Position()
		input.SetText(value)
		input.SetPosition(pos - 1)
		return nil
	}

	number, err := strconv.Atoi(value) // convert string to int
	if err != nil {
		return err
	}

	update_unsaved_config(true)

	*target = int(number)
	return nil
}

func update_unsaved_config(value bool) {
	revealer := UI.GetObject("unsaved_revealer").(*gtk.Revealer)
	if revealer != nil {
		revealer.SetRevealChild(value)
	}
}
