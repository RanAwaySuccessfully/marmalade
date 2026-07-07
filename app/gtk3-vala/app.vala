using Gtk; // TODO: use the command-line version of Marmalade under the hood, do NOT try to interface with Go code
using GLib;

// https://valadoc.org/glib-2.0/GLib.Process.spawn_async.html
// https://valadoc.org/v4l2/index.htm
// https://valadoc.org/json-glib-1.0/index.htm

public class MarmaladeApp : Gtk.Application {
    public MarmaladeApp() {
        Object(application_id: "xyz.randev.marmalade.gtk3vala");
    }

    public static int main(string[] args) {        
        var app = new MarmaladeApp();
        return app.run(args);
    }

    public override void activate() {
        Gtk.Window.set_default_icon_name("xyz.randev.marmalade");
        new Marmalade();
        Gtk.main();
    }
}

[GtkTemplate (ui = "/xyz/randev/marmalade/app.ui")]
public class Marmalade : Gtk.ApplicationWindow {
    public Marmalade() {
        this.show_all();
    }

    [GtkCallback]
    private void about_button_clicked(Gtk.Button source) {
        new AboutDialog();
    }
}

[GtkTemplate (ui = "/xyz/randev/marmalade/dialog_about.ui")]
public class AboutDialog : Gtk.AboutDialog {
    public AboutDialog() {
        this.add_credit_section("Logo by", { "vexamour" });
        this.show_all();
    }
}
