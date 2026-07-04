## Building, Testing, Debugging

**You do not need to do any of this to install Marmalade.**

If you want to develop or tinker with this program, you'll need to install the [Go programming language](https://go.dev/).

Marmalade is divided between the following apps:
- `cmd`
- `ffmpeg`¹
- `gtk3`
- `gtk4`
- `mediapipe`

<sub>¹ not an app, see the "FFmpeg Plugin" section for more information</sub>

For running one of these directly, run: `go run -v ./app/cmd`

For building an executable, run: `go build -v ./app/cmd`

If building the MediaPipe sub-process, see the section "MediaPipe" below.

Note that this will generate an executable in the folder typed out above, which you should copy the main folder of the repository (the top-most folder). You can tell it to build an executable with a different name, but you can also rename it by adding `-o custom_name_here` to the command (you must add it like so: `go build -v -o custom_name_here ./app/cmd`). Note that Marmalade expects the `mediapipe` executable to have that exact name.

Depending on which executable you're building, you'll also need some extra development files installed:
- gtk3: `libv4l-dev gobject-introspection libgtk-3-dev`
- gtk4: `libv4l-dev gobject-introspection libgtk-4-dev`
- mediapipe: `libv4l-dev libavcodec-dev libavutil-dev libswscale-dev`

If you want to debug it, it comes with some Visual Studio Code configuration depending on what you want to debug:

- If you want to debug the command-line version, run `Go: Debug CMD`.
- If you want to debug the GTK 3 or GTK 4 version, run `Go: Debug GTK 3 Build` or `Go: Debug GTK 4 Build`. Note that this one will pre-build a `marmalade-gtk3` or `marmalade-gtk4` executable (so you can see each step of the build process).
- If you want to debug the MediaPipe sub-process, run `Go: Debug MediaPipe`.

### MediaPipe

Before building the mediapipe sub-process, you'll need to a copy of the following libraries:
- `libmediapipe.so`
- `libopencv_core.so.414`²
- `libopencv_features2d.so.414`²
- `libopencv_imgproc.so.414`²

<sub>¹ can be circumvented by using your system's OpenCV during the build</sub>

Downloading them from an appropriate release of Marmalade and adding them to Marmalade's `lib` folder is highly recommended, but you can built them yourself if you so wish (check the "Building libmediapipe.so" section for more details).

The most recent stable release of MediaPipe (`v0.10.35`) contains a C library that can be used directly by programs like Marmalade via `libmediapipe.so`. Unfortunately, the C API still has some trace amounts of C++ in it, which makes it impossible for Go to connect to it directly, so I created a wrapper called **libtoast** written in C but compiled as C++. In the future, I assume the C API will stabilize and **libtoast** will be removed, but for now this is a necessary component of Marmalade.

You can compile **libtoast** by running the command `make` while your working directory (current folder) is the `cc` folder. This will generate `libtoast.a`. This will fail if `libmediapipe.so` and its dependencies are not found. This requires `make` or an equivalent command to be installed on your system. Once you have this file, you can proceed with compiling the MediaPipe subprocess via `go build -v ./app/mediapipe`.

#### Building libmediapipe.so

A slightly customized version of MediaPipe, adapted to work better with Marmalade, is available at `app/mediapipe/cc/mediapipe` as a Git submodule. If it's not already downloaded, you can download it by running `git submodule update --init --recursive app/mediapipe/cc/mediapipe`. This will also download the OpenCV repository. If you don't want that and would rather use your system's OpenCV, then run that command without the `--recursive` option and then edit the `WORKSPACE` and `third_party/opencv_linux.BUILD` to match the ones in [this commit](https://github.com/google-ai-edge/mediapipe/tree/f8ef212d5c962c0e853db7e59d217056b187084b) (and then appropriatelly change `opencv_linux.BUILD` again to match what your system provides).

If you proceed with using the bundled OpenCV, then you'll need to built it first. A [build-opencv.sh](/app/mediapipe/cc/build-opencv.sh) script is available for convenience. This will generate dynamic library files at `opencv_local/build/install/lib/`. You must copy `libopencv_core.so.4.14.0`, `libopencv_features2d.so.4.14.0` and `libopencv_imgproc.so.4.14.0` into Marmalade's `lib` folder, and then afterwards, edit them so the `so.4.14.0` at the end is changed into `so.414`. You might also need to run the following command on the `lib` folder, as the OpenCV libraries depend on each other and might have some trouble finding each other during runtime:

```sh
patchelf --set-rpath ./lib libopencv*
```

Once you have done so, you may proceed with building **libmediapipe**. Please take a look at the [MediaPipe docs](https://ai.google.dev/edge/mediapipe/framework/getting_started/install) as well as the [Bazel command-line arguments](https://bazel.build/reference/command-line-reference) for more information on how to build MediaPipe. The only target you need to build is `//mediapipe/tasks/c:libmediapipe.so`.

I have provided [build-mediapipe.sh](/app/mediapipe/cc/build-mediapipe.sh) as an example but I provide no guarantees that it will work for you. Once you have compiled MediaPipe, the file `libmediapipe.so` will have been created in the folder `bazel-bin/mediapipe/tasks/c/`. Copy that file to Marmalade's `lib` folder.

### KalidoKit

KalidoKit is necessary for proper hand and pose tracking, on top of MediaPipe. It's a library written in JavaScript, so to build it, you'll need [Bun](https://bun.com/) installed in your system.

Run the `build.sh` script that's located on `app/kalidokit`. Once done, the file `kalidokit-bin` will be created, which is the executable containing all of the code and a copy of the Bun runtime. Copy this file (or create a link to it) on Marmalade's main folder (the one where the `LICENSE` file exists).

### FourCC

The file `fourcc.json` contains a mapping file in order to bridge V4L2's encoding types with FFmpeg's. You can generate this file by running `go run -v ./app/fourcc`. You can alternatively run `go run -v ./app/fourcc -a` in order to add a few additional types that are normally passed along to v4lconvert instead, but can be used with FFmpeg.

Having this as a JSON file instead of being hardcoded makes it easy for anyone to edit it. As long as you know that both V4L2 and FFmpeg support a format, and you know the relevant IDs, you can manually add it to the file and even re-distribute it.

### FFmpeg Plugin

MediaPipe only accepts images in a select few formats like RGB3. As such, if you use any other format then Marmalade will need to use an external library (such as FFmpeg) to convert it to RGB3. Because major versions of FFmpeg can introduce breaking changes, the mediapipe sub-process described above will import a Go plugin (shared library) that interacts with the FFmpeg version installed in your system. **This is not a statically built version of FFmpeg and will not work on its own.**

If you're compiling your own version of the mediapipe sub-process, you'll also have to re-compile the FFmpeg plugin. If so, use the exact same environment to compile both the sub-process and the plugin, as even tiny changes can cause the two to conflict and crash.

To build this plugin, specifically use the command `go build -buildmode=plugin -v -o ffmpegX_plugin.so ./app/ffmpeg`, replacing the X with the version of `ffmpeg` installed in your system (example: `ffmpeg 6.1` -> `ffmpeg6_plugin.so`).

#### Custom FFmpeg version

If you wish to use a specific version of FFmpeg instead, there is a Git submodule that can be downloaded by running `git submodule update --init app/ffmpeg/ffmpeg` which points to FFmpeg's GitHub mirror. With this, you can choose the branch that contains the release you want to compile against (example: `release/6.1`). In order to use this, specifically add the flag `-tags ffmpeg_git` right after the `go build` command, and before all the other flags.

Note that in this case, you'll also need to build FFmpeg itself. A minimal install is enough, and can be done by running `build.sh` located on the `app/ffmpeg` folder. To clean any files created by the build (such as for a subsequent build), you can run `make clean` on the `app/ffmpeg/ffmpeg` folder.

## Build times

The GUI version of this project takes about 10 minutes to compile when building via GitHub Actions (probably faster on your PC), most of this time is taken up by building GTK and its dependencies. This will happen when building the program for the first time, but if you're using VSCode with the Go extension, it will also happen the first time you open a .go file in this project as VSCode will get busy generating all the IntelliSense data it needs.

Go has a caching mechanism that makes it so you don't have to go through this every time, but the cache does not last forever, so don't be surprised if you see it recompiling the GTK dependencies again. If you compile the GTK4 version, the GTK3 version will take slightly less time and vice-versa.