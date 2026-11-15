#include <libavcodec/avcodec.h>
#include <libavutil/frame.h>
#include <libswscale/swscale.h>

int ffmpeg_convert_frame(struct SwsContext*, AVFrame*, AVFrame*);
void* ffmpeg_get_frame_data_ptr(AVFrame*, int);

enum AVPixelFormat ff_fmt_v4l2ff(uint32_t v4l2_fmt, enum AVCodecID codec_id);
enum AVCodecID ff_fmt_v4l2codec(uint32_t v4l2_fmt);