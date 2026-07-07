VALAC = valac
PKGS = --pkg gtk+-3.0 --pkg gio-2.0 --pkg glib-2.0
SOURCES = app.vala
RESOURCE_XML = resources.gresource.xml
RESOURCE_C = resources.c
TARGET = marmalade-gtk3-vala

fresh: clean all

all: $(TARGET)

$(RESOURCE_C): $(RESOURCE_XML)
	glib-compile-resources --target=$@ --generate-source $<

$(TARGET): $(SOURCES) $(RESOURCE_C)
	$(VALAC) -o $@ $(PARAMS) --gresources=$(RESOURCE_XML) $(SOURCES) $(RESOURCE_C) $(PKGS)

clean:
	rm -f $(TARGET) $(RESOURCE_C)