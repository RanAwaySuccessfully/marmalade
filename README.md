# Marmalade

Allows MediaPipe to be used on Linux by mimicking VTube Studio's iPhone Raw Tracking data. You can connect it to programs such as VBridger.

## Installing

1. Download the latest release of Marmalade.
2. Download [`face_landmarker.task`](https://ai.google.dev/edge/mediapipe/solutions/vision/face_landmarker) from Google's MediaPipe page and place it inside the `python` folder.
3. Install `python3`, `python3-venv` and `pip3`.
4. (optional)¹ Run `mediapipe-install.sh`. Make sure to change your active folder via `cd scripts`, or run the file by double-clicking it.
5. (optional)² Make sure you have `gtk4` and `libv4l-0` installed.

<small>¹ If Marmalade does not find a .venv folder when starting up, it will ask you if it should run this step for you. This will fail if you haven't done Step 3 yet.

² Unless you want to use the command line version. Do note that, in this case, you'll have to edit `config.json` manually.</small>

And you're done. You can just run the program at any time, and it'll take care of the rest for you.

## Config File

Here's what each field is responsible for:

* port: The UDP port that Marmalade will be listening to. If you don't know what to do with this, keep the default value of `21412`.
* camera: Camera ID (index). Starts at `0` and goes up from there.
* width: Camera horizontal resolution.
* height: Camera vertical resolution.
* fps: Camera frames per second.
* model: Filename of the model file that MediaPipe will use for face tracking. Since this is a string value, it is surrounded by `"` (double quotes) unlike the numeric fields above.
* use_gpu: Set to `true` to attempt to use the GPU for processing MediaPipe, and leave it at `false` otherwise.

## Building, Testing, Debugging

You'll need to install [Go](https://go.dev/).

For building, run: `go build`

For running it without building it, run: `go run -v ./` or `go run -tags withgtk4 -v ./`

If you want to debug it, it comes with some Visual Studio Code configuration. You can debug the entire thing using `Go: Launch Package` or just the Python code by using `Python Debugger: Current File` while having the `main.py` file open and selected.

## License and Credits

Still thinking about what license to use, but it'll definitely be open source.

Relies heavily on code written by lilacGalaxy on this [GitHub Repo](https://github.com/lilac-galaxy/lilacs-mediapipe-forward-vts-plugin).

Uses [gotk4](https://github.com/diamondburned/gotk4) for its GUI.