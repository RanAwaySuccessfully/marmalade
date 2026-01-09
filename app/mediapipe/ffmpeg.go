package main

/*
#cgo CFLAGS: -I/ffmpeg/include
#cgo LDFLAGS: -L/ffmpeg/lib -lavcodec -lavutil -lswscale

#include <libavcodec/avcodec.h>
#include <libavutil/frame.h>
#include <libswscale/swscale.h>

int do_sws_scale(struct SwsContext* outputCtx, AVFrame* inputFrame, AVFrame* outputFrame) {
	return sws_scale(
		outputCtx,
		(const uint8_t* const*)inputFrame->data, inputFrame->linesize,
		0, inputFrame->height,
		outputFrame->data, outputFrame->linesize
	);
}

void* get_frame_data_ptr(AVFrame* frame, int y) {
    return frame->data[0] + y * frame->linesize[0];
}
*/
import "C"
import (
	"bytes"
	"errors"
	"fmt"
)

type ConverterFFMPEG struct {
	inputCtx    *C.struct_AVCodecContext
	inputPacket *C.struct_AVPacket
	inputFrame  *C.struct_AVFrame

	outputCtx   *C.struct_SwsContext
	outputFrame *C.struct_AVFrame
}

func (conv *ConverterFFMPEG) init(format string) error {
	var codec_id uint32
	var pix_fmt int32

	switch format {
	case "YUYV":
		codec_id = C.AV_CODEC_ID_RAWVIDEO
		pix_fmt = C.AV_PIX_FMT_YUYV422
	case "MJPG":
		codec_id = C.AV_CODEC_ID_MJPEG
	default:
	}

	codec := C.avcodec_find_decoder(codec_id)
	conv.inputCtx = C.avcodec_alloc_context3(codec)

	if codec_id == C.AV_CODEC_ID_RAWVIDEO {
		conv.inputCtx.pix_fmt = pix_fmt
	}

	ret := C.avcodec_open2(conv.inputCtx, codec, nil)
	if ret < 0 {
		return conv.get_error("opening decoder", ret)
	}

	/*
		output_codec := C.avcodec_find_encoder(C.AV_CODEC_ID_RAWVIDEO)
		conv.outputCodec = C.avcodec_alloc_context3(output_codec)
		conv.outputCodec.pix_fmt = C.AV_PIX_FMT_RGB24
		ret = C.avcodec_open2(conv.outputCodec, output_codec, nil)
		if ret < 0 {
			return conv.get_error("opening encoder", ret)
		}
	*/

	conv.inputFrame = C.av_frame_alloc()
	if conv.inputFrame == nil {
		return errors.New("unable to allocate input frame")
	}

	conv.inputPacket = C.av_packet_alloc()
	if conv.inputPacket == nil {
		return errors.New("unable to allocate input packet")
	}

	return nil
}

func (conv *ConverterFFMPEG) init_output_frame() error {
	conv.outputFrame = C.av_frame_alloc()
	conv.outputFrame.width = conv.inputFrame.width
	conv.outputFrame.height = conv.inputFrame.height
	conv.outputFrame.format = C.AV_PIX_FMT_RGB24

	ret := C.av_frame_get_buffer(conv.outputFrame, 0)
	if ret < 0 {
		return conv.get_error("allocating output frame pointers", ret)
	}

	conv.outputCtx = C.sws_getContext(
		conv.inputFrame.width, conv.inputFrame.height, int32(conv.inputFrame.format),
		conv.outputFrame.width, conv.outputFrame.height, int32(conv.outputFrame.format),
		C.SWS_FAST_BILINEAR, nil, nil, nil,
	)

	return nil
}

func (conv *ConverterFFMPEG) convert(input []byte) ([]byte, error) {
	data_length := len(input)
	data := C.CBytes(input)
	data = C.av_realloc(data, C.size_t(data_length+C.AV_INPUT_BUFFER_PADDING_SIZE))
	//defer C.av_free(data)

	ret := C.av_packet_from_data(conv.inputPacket, (*C.uchar)(data), (C.int)(data_length))
	if ret < 0 {
		return nil, conv.get_error("creating input packet", ret)
	}

	length := C.avcodec_send_packet(conv.inputCtx, conv.inputPacket) // packet -> codec
	if length < 0 {
		return nil, conv.get_error("reading input packet", length)
	}

	output := bytes.Buffer{}

	for length >= 0 {
		length = C.avcodec_receive_frame(conv.inputCtx, conv.inputFrame) // codec -> frame

		if length == -C.EAGAIN || length == C.AVERROR_EOF {
			break
		} else if length < 0 {
			return nil, conv.get_error("decoding packet into frame", length)
		}

		frame := conv.inputFrame

		// do pixelformat conversion
		if conv.inputFrame.format != C.AV_PIX_FMT_RGB24 {

			if conv.outputFrame == nil {
				err := conv.init_output_frame()
				if err != nil {
					return nil, err
				}

				if conv.outputCtx == nil {
					return nil, errors.New("what???")
				}
			} else {
				// sorry, nothing
			}

			C.do_sws_scale(conv.outputCtx, conv.inputFrame, conv.outputFrame)

			frame = conv.outputFrame
		}

		for y := C.int(0); y < frame.height; y++ {
			start := C.get_frame_data_ptr(frame, y)
			bytes := C.GoBytes(start, frame.width*3)
			output.Write(bytes)
		}
	}

	output_data := output.Bytes()
	return output_data, nil
}

func (conv *ConverterFFMPEG) end() {
	if conv.outputCtx != nil {
		C.sws_freeContext(conv.outputCtx)
	}

	if conv.outputFrame != nil {
		C.av_frame_free(&conv.outputFrame)
	}

	if conv.inputFrame != nil {
		C.av_frame_free(&conv.inputFrame)
	}

	if conv.inputPacket != nil {
		C.av_packet_free(&conv.inputPacket)
	}

	if conv.inputCtx != nil {
		C.avcodec_free_context(&conv.inputCtx)
	}
}

func (conv *ConverterFFMPEG) get_error(prefix string, errnum C.int) error {
	error_size := C.sizeof_uchar * 100
	error_c := C.malloc(C.ulong(error_size))
	error_c_char := (*C.char)(error_c)

	result := "unknown error"

	ret := C.av_strerror(errnum, error_c_char, C.ulong(error_size))
	if ret >= 0 {
		result = C.GoString(error_c_char)
	}

	C.free(error_c)

	if prefix != "" {
		result = fmt.Sprintf("(%s) %s", prefix, result)
	}

	return errors.New(result)
}
