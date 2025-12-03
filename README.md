# Marmalade

[![build status](https://github.com/ranawaysuccessfully/marmalade/actions/workflows/ubuntu.yml/badge.svg)](https://github.com/ranawaysuccessfully/marmalade/actions) [![latest release](https://img.shields.io/github/v/release/ranawaysuccessfully/marmalade)](https://github.com/RanAwaySuccessfully/marmalade/releases/latest)

Allows MediaPipe to be used on Linux by mimicking VTube Studio's iPhone Raw Tracking data. You can connect it to programs such as VBridger.

| Command-line | GTK 4 (GUI) |
| ---- | ---- |
| ![Command-line](docs/readme_cmd.png) | ![GTK 4](docs/readme_gtk4.png) |

Also available under GTK 3 (GUI).

## Installing

1. Download the [latest release](https://github.com/RanAwaySuccessfully/marmalade/releases/latest) of Marmalade.
2. Download the latest [`face_landmarker.task`](https://ai.google.dev/edge/mediapipe/solutions/vision/face_landmarker) file from Google's MediaPipe page and place it inside the `python` folder.
3. If using any of the GUI versions, you'll also need to have the following installed, although they probably already are installed by default:
    - `libgtk-3`¹ or `gtk3`¹ (>=3.24, only if using GTK 3)
    - `libgtk-4`¹ or `gtk4`¹ (>=4.8, only if using GTK 4)
    - `libv4l`¹
    - `xdg-utils`
    - `pciutils`

<sub>¹ May be suffixed by another version number, for example: `libgtk-3-0t64`, `libgtk-4-1`, `libv4l-0`.</sub>

And you're done. You can just run the program at any time, and it should take care of the rest for you.

### First-time setup

- If using a GUI version, and it notices its icon is not installed, it will install a local copy to distinguish it between just a random executable. If you wish to uninstall the icon, run the command-line version of Marmalade like so: `./marmalade -u`.

## Connecting

If you're running Marmalade on the same PC as the program you want to connect to, then you can use the IP address `127.0.0.1`, which is the loopback IP (always points to your own PC).

If you're running it on another machine over LAN, you'll need to figure out its IP address and to make sure it is reachable via UDP port that Marmalade is configured to use (see "Config file" section below).

### VBridger

Do not select the "MediaPipe" option, instead, select "VTube Studio" and type in the relevant IP address. Even though it says "Connect to iPhone", clicking on that button will connect to Marmalade instead.

### VNyan¹

During the first time setup, select Skip when it asks you which options will be used for tracking.

Go to "Settings", then on the "Tracking" tab, select "Phone / ARKit". Choose "VTube Studio" and type in the relevant IP address.

### VSeeFace¹

Select any option during the first time setup (this will be overriden later).

Go to "General Settings" then scroll down to "iFacialMocap/FaceMocap3D/VTube Studio" and set the tracking app to "VTube Studio", then type in the relevant IP address.

### Warudo

Marmalade does not support VMC (Virtual Motion Capture Protocol). You may try to connect Warudo to another program such as VSeeFace, and have said program connect to Marmalade.

### Notes

I haven't tested with other programs yet, but in case it doesn't work or works weirdly, feel free to open an issue and/or feature request.

¹ During testing, these programs worked best when running using Proton 10. You should also install the font arial.ttf by copying it from a Windows installation to the folder `[wine prefix]/drive_c/windows/Fonts/`. Your wine prefix folder will vary (the default one is at `~/.wine`).

## Config file

**If using a GUI version, you do not need to worry about this file** unless it becomes corrupted somehow, as the UI allows you to edit it seamlessly. If using the command line version, you'll need to edit it manually to use the settings that you want. It is located right beside the app's executable as `config.json`.

Here's what each field in this file is responsible for:

* port: The UDP port that Marmalade will be listening to. If you don't know what to do with this, keep the default value of `21412`.
* camera: Camera ID (index). Starts at `0` and goes up from there.
* width: Camera horizontal resolution (number of pixels).
* height: Camera vertical resolution (number of pixels).
* fps: Camera frames per second.
* format: Camera format. Examples: `"YUYV"`, `"MJPG"`, etc...
* model: Filename of the model file that MediaPipe will use for face tracking.
* use_gpu: Set to `true` to attempt to use the GPU for processing MediaPipe, and leave it at `false` otherwise.
* prime_id: PCIe bus (slot/address) of the GPU that should be used by MediaPipe.* An empty string is valid, in which case, the default GPU will be used. Has no effect if `use_gpu` is `false`.

The fields `model` and `prime_id` are string values, and as such they're surrounded by `"` (double quotes) unlike other fields.

<sub>* This is the same as the `DRI_PRIME` environment variable, and any valid value for it is also valid for this field, although the GUI versions only expects PCIe bus IDs and may glitch otherwise.</sub>

## Building, Testing, Debugging

**You do not need to do any of this to install Marmalade. See the "Installing" section above instead.**

If you want to develop or tinker with this program, you'll need to install the [Go programming language](https://go.dev/). You'll also need to setup the Python dependencies (see section below).

For building, run: `go build -v`

For running it without building it, run: `go run -v ./`

For building or running the GTK 4 version, just add `-tags withgtk4` to the commands above. Do note that in this case, you'll also need to install the `libgtk-4-dev` and `libv4l-dev` packages. For GTK 3 it's `-tags withgtk3` and you'll need `libgtk-3-dev` instead. You *may* also need `gobject-introspection`.

If you want to debug it, it comes with some Visual Studio Code configuration depending on what you want to debug:

- If you want to debug the Go code, specifically the command-line version, run `Go: Launch Package`.
- If you want to debug the Python code, run `Python Debugger: Current File` while having the `main.py` file open and selected. Once it's running, type in `+127.0.0.1:21499` for example, to start sending data to a specific IP address and port.
- If you want to debug the GTK 4 version, run `Go: Debug GTK 4 Build`. Note that this one will pre-build a `marmalade-gtk4` executable to make it start faster. The same applies for the GTK 3 version.

### Python dependencies

Since v0.4, Marmalade comes bundled with Python 3.12, OpenCV and MediaPipe. PEX is used to make a portable and compressed executable.

In order to test Marmalade locally, you'll need to set up a virtual environment (`.venv`) folder that contains MediaPipe, which will use around 850MB of disk space. You can run `scripts/mediapipe-install.sh` while making sure your working directory (current folder) is `scripts`. This will require you to install `python3`, `python3-venv` and `python3-pip`.

If you want to install run PEX manually, you'll also need to make sure you have PEX installed (such as by running `pip3 install pex`). Change your current directory to be the `python` folder, and run the following command:

```sh
pex -v -r requirements.txt --scie eager -o dist/mediapipe
```

MediaPipe will not install if your version of Python is 3.13.

### Build times

The GUI version of this project takes about 10 minutes to compile when building via GitHub Actions (probably faster on your PC), most of this time is taken up by building GTK and its dependencies. This will happen when building the program for the first time, but if you're using VSCode with the Go extension, it will also happen the first time you open a .go file in this project as `.vscode/settings.json` is, by default, configured to the GTK4 version, and so it will get busy generating all the IntelliSense data it needs.

Go has a caching mechanism that makes it so you don't have to go through this every time, but the cache does not last forever, so don't be surprised if you see it recompiling the GTK dependencies again. If you compile the GTK4 version, the GTK3 version will take slightly less time and vice-versa.

## License and Credits

Licensed under the [MIT License](LICENSE). The code under the `python` folder is edited based on code written by lilacGalaxy on this [GitHub Repo](https://github.com/lilac-galaxy/lilacs-mediapipe-forward-vts-plugin), and as such it uses the same license, but has separate copyright, check [its license file](python/LICENSE) for more info.

This project uses [gotk4](https://github.com/diamondburned/gotk4), which are [GTK4](https://docs.gtk.org/gtk4/) language bindings for Go. This project does **not** use libadwaita, although I'm wondering if I should add [libadapta](https://github.com/xapp-project/libadapta) support.

Somewhat inspired by [Facetracker](https://codeberg.org/ZRayEntertainment/Facetracker) which uses OpenSeeFace instead.

Many thanks to Kylo-Neko's [Linux Guide to Vtubing](https://codeberg.org/KyloNeko/Linux-Guide-to-Vtubing) which is what kickstarted my adventuring into seeing if/how I can make this work.
