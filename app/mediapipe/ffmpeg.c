#include "ffmpeg.h"

int ffmpeg_convert_frame(struct SwsContext* outputCtx, AVFrame* inputFrame, AVFrame* outputFrame) {
	return sws_scale(
		outputCtx,
		(const uint8_t* const*)inputFrame->data, inputFrame->linesize,
		0, inputFrame->height,
		outputFrame->data, outputFrame->linesize
	);
}

void* ffmpeg_get_frame_data_ptr(AVFrame* frame, int y) {
    return frame->data[0] + y * frame->linesize[0];
}