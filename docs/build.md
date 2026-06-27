## Building, Testing, Debugging

**You do not need to do any of this to install Marmalade.**

If you want to develop or tinker with this program, you'll need to install the [Go programming language](https://go.dev/).

Marmalade is divided between the following apps:
- `cmd`
- `gtk3`
- `gtk4`
- `mediapipe`

For running one of these directly, run: `go run -v ./app/cmd`

For building an executable, run: `go build -v ./app/cmd`

If building the MediaPipe sub-process, see the section "MediaPipe" below.

Note that this will generate an executable in the folder typed out above, which you should copy the main folder of the repository (the top-most folder). You can tell it to build an executable with a different name, but you can also rename it by adding `-o custom_name_here` to the command. Note that Marmalade expects the `mediapipe` executable to have that exact name.

Depending on which executable you're building, you'll also need some extra development files installed:
- gtk3: `libv4l-dev gobject-introspection libgtk-3-dev`
- gtk4: `libv4l-dev gobject-introspection libgtk-4-dev`
- mediapipe: `libv4l-dev libavcodec-dev libavutil-dev libswscale-dev`

If you want to debug it, it comes with some Visual Studio Code configuration depending on what you want to debug:

- If you want to debug the command-line version, run `Go: Debug CMD`.
- If you want to debug the GTK 3 or GTK 4 version, run `Go: Debug GTK 3 Build` or `Go: Debug GTK 4 Build`. Note that this one will pre-build a `marmalade-gtk3` or `marmalade-gtk4` executable (so you can see each step of the build process).
- If you want to debug the MediaPipe sub-process, run `Go: Debug MediaPipe`.

### MediaPipe

Before building the mediapipe sub-process, you'll need to build or download a copy of `libmediapipe.so` and `libtoast.so`.

The most recent stable release of MediaPipe (`v0.10.35`) contains a C library that can be used directly by programs like Marmalade via **libmediapipe**. Unfortunately, the C API still has some trace amounts of C++ in it, which makes it impossible for Go to connect to it directly, so I created a wrapper called **libtoast** written in C but compiled as C++. In the future, I assume the C API will stabilize and **libtoast** will be removed, but for now this is a necessary component of Marmalade.

For compiling **libmediapipe**, a Git submodule is available at `app/mediapipe/cc/mediapipe` containing a fork of the MediaPipe version currently used by Marmalade, alongside a few extra patches I made for compatibility. If it's not already downloaded, you can download it by running `git submodule update --init --recursive`. Once you have downloaded the repo, please take a look at the [MediaPipe docs](https://ai.google.dev/edge/mediapipe/framework/getting_started/install) as well as the [Bazel command-line arguments](https://bazel.build/reference/command-line-reference) for more information on how to build MediaPipe. The only target you need to build is `//mediapipe/tasks/c:libmediapipe.so`.

I have provided [bazel-build.sh](/app/mediapipe/cc/bazel-build.sh) as an example but I provide no guarantees that it will work for you. Once you have compiled MediaPipe, the file `libmediapipe.so` will have been created inside the MediaPipe submodule in the folder `bazel-bin/mediapipe/tasks/c/`. Copy that file to Marmalade's `cc` folder (the same folder that contains the file `libtoast.cc`).

You can compile **libtoast** by running the command `make` while your working directory (current folder) is the `cc` folder. This will generate `libtoast.a`. This will fail if `libmediapipe.so` is not found. This requires `make` or an equivalent command to be installed on your system. Once you have this file, you can proceed with compiling the MediaPipe subprocess via `go build -v ./app/mediapipe`.

### FourCC

The file `fourcc.json` contains a mapping file in order to bridge V4L2's encoding types with FFMPEG's. You can generate this file by running `go run -v ./app/fourcc`.

### Build times

The GUI version of this project takes about 10 minutes to compile when building via GitHub Actions (probably faster on your PC), most of this time is taken up by building GTK and its dependencies. This will happen when building the program for the first time, but if you're using VSCode with the Go extension, it will also happen the first time you open a .go file in this project as VSCode will get busy generating all the IntelliSense data it needs.

Go has a caching mechanism that makes it so you don't have to go through this every time, but the cache does not last forever, so don't be surprised if you see it recompiling the GTK dependencies again. If you compile the GTK4 version, the GTK3 version will take slightly less time and vice-versa.