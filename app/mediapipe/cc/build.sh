# cp ./mediapipe/bazel-bin/mediapipe/tasks/c/libmediapipe.so ./
g++ -g -I./mediapipe/ -L./ -lmediapipe -shared -fPIC libtoast.cc -o libtoast.so