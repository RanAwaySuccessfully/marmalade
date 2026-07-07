#!/usr/bin/env -S vala --vapidir . --pkg gtk+-3.0

using Gtk; // TODO: use the command-line version of Marmalade under the hood, do NOT try to interface with Go code

public class MarmaladeApp : Gtk.Application {
    public static Gtk.Builder builder;

    public MarmaladeApp() {
        Object(application_id: "xyz.randev.marmalade.gtk3v");
    }

    public static int main(string[] args) {
        var app = new MarmaladeApp();
        return app.run(args);
    }

    public override void activate() {
        Gtk.Window.set_default_icon_name("xyz.randev.marmalade");

        builder = new Gtk.Builder();

        try {
            builder.add_from_file("app/gtk3-vala/ui/app.ui");
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
