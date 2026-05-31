# cp ./mediapipe/bazel-bin/mediapipe/tasks/c/libmediapipe.so ./
g++ -g -I./mediapipe/ -L./ -lmediapipe -shared -fPIC -o libtoast.so *.cc