#include <libavcodec/avcodec.h>
#include <libavutil/frame.h>
#include <libswscale/swscale.h>

int ffmpeg_convert_frame(struct SwsContext*, AVFrame*, AVFrame*);
void* ffmpeg_get_frame_data_ptr(AVFrame*, int);