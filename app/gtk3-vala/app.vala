#!/usr/bin/env -S vala --vapidir . --pkg libshared --pkg gtk+-3.0

using Gtk;

public class MarmaladeApp : Gtk.Application {
    public static Gtk.Builder builder;

    public MarmaladeApp() {
        Object(application_id: "xyz.randev.marmalade.gtk3");
    }

    public static int main(string[] args) {
        var app = new MarmaladeApp();
        return app.run(args);
    }

    public override void activate() {
        Gtk.Window.set_default_icon_name("xyz.randev.marmalade");

        builder = new Gtk.Builder();

        try {
            string ui_string = (string)LibShared.ui_getembed();
            builder.add_from_file(ui_string);
        } catch (Error e) {
            stderr.printf("Could not load UI: %s\n", e.message);
        }

        builder.connect_signals(null);

        var window = builder.get_object("main_app") as Gtk.ApplicationWindow;
        window.show_all();

        Gtk.main();
    }
}

public class AboutDialog {
    public static void open() {
        var builder = MarmaladeApp.builder;

        try {
            builder.add_from_file("ui/dialog_about.ui");
        } catch (Error e) {
            stderr.printf("Could not load UI: %s\n", e.message);
        }

        var dialog = MarmaladeApp.builder.get_object("about_dialog") as Gtk.AboutDialog;
        dialog.add_credit_section("Logo by", { "vexamour" });
    }
}
