#sudo apt install libopencv-core-dev libopencv-highgui-dev libopencv-calib3d-dev libopencv-features2d-dev libopencv-imgproc-dev libopencv-video-dev
LOCALREPOS=$(realpath bazel-local)
cd $(realpath mediapipe)
XDG_CACHE_HOME=/tmp
bazel-7.4.1 clean
bazel-7.4.1 build --compilation_mode dbg --verbose_failures --distdir=$LOCALREPOS -c opt --linkopt -s --strip never --define MEDIAPIPE_DISABLE_GPU=1 //mediapipe/tasks/c:libmediapipe.so
# --define MEDIAPIPE_DISABLE_GPU=1
# --copt -DMESA_EGL_NO_X11_HEADERS --copt -DEGL_NO_X11
#bazel-7.4.1 build --verbose_failures --distdir=$LOCALREPOS -c opt --linkopt -s --strip always --define MEDIAPIPE_DISABLE_GPU=1 //mediapipe/examples/desktop/holistic_tracking:holistic_tracking_cpu
