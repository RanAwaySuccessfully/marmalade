# Marmalade

Allows MediaPipe to be used on Linux by mimicking VTube Studio's iPhone Raw Tracking data. You can connect it to programs such as VBridger.

## Installing

1. Download the latest release of Marmalade.
2. Download `face_landmarker_v2_with_blendshapes.task` from Google's MediaPipe page and place it inside the `python` folder.
3. Install `python3` and `pip3`.
4. Run `mediapipe-install.sh`. Make sure to change your active folder via `cd scripts`, or run the file by double-clicking it.

And you're done. You can just run the program at any time, and it'll take care of the rest for you.

If you don't have GTK4 installed, you can still run the program via the command line.

## Building, Testing, Debugging

You'll need to install [Go](https://go.dev/).

For building, run: `go build`

For running it without building it, run: `go run`

If you want to debug it, it comes with Visual Studio Code configuration. You can debug the entire thing using `Go: Launch Package` or just the Python code by using `Python Debugger: Current File` while having the `main.py` open and selected.

## License and Credits

Still thinking about what license to use, but it'll definitely be open source.

Relies heavily on code written by lilacGalaxy on this [GitHub Repo](https://github.com/lilac-galaxy/lilacs-mediapipe-forward-vts-plugin).

Uses [gotk4](https://github.com/diamondburned/gotk4) for its GUI.