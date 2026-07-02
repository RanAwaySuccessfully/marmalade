cd ffmpeg
./configure \
  --disable-everything \
  --disable-doc \
  --disable-ffplay \
  --disable-ffprobe \
  --enable-shared \
  --extra-cflags="-O3"
make