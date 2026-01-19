package main

import (
	"regexp"
	"strconv"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

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

func update_numeric_config(input *gtk.Entry, target *float64) error {
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

	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}

	update_unsaved_config(true)

	*target = number
	return nil
}

func update_unsaved_config(value bool) {
	revealer := UI.GetObject("unsaved_revealer").(*gtk.Revealer)
	if revealer != nil {
		revealer.SetRevealChild(value)
	}
}
