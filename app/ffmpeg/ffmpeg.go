package main

/*
#include "ffmpeg.h"
*/
import "C"
import (
	"errors"
	"io"
	"marmalade/internal/errs"
	"unsafe"
)

var FFmpeg = FFMPEG_Plugin{}

type FFMPEG_Plugin struct {
	inputCodec  *C.struct_AVCodec
	inputCtx    *C.struct_AVCodecContext
	inputPacket *C.struct_AVPacket
	inputFrame  *C.struct_AVFrame

	copyHwCtx  *C.AVBufferRef
	copyFrame  *C.struct_AVFrame
	copyDevice *C.char

	outputCtx   *C.struct_SwsContext
	outputFrame *C.struct_AVFrame
}

func (ff *FFMPEG_Plugin) Init(codec_id uint32, pix_fmt int32) {
	ff.inputCodec = C.avcodec_find_decoder(codec_id)
	ff.inputCtx = C.avcodec_alloc_context3(ff.inputCodec)

	if pix_fmt != -1 {
		ff.inputCtx.pix_fmt = pix_fmt
	}
}

func (ff *FFMPEG_Plugin) FindHwAccel() []uint32 {
	supported := []uint32{}

	for i := 0; ; i++ {
		hw := C.avcodec_get_hw_config(ff.inputCodec, C.int(i))
		if hw == nil {
			break
		}

		supported = append(supported, hw.device_type)
	}

	return supported
}

func (ff *FFMPEG_Plugin) UseHwAccel(device string) error {
	if device != "" {
		ff.copyDevice = C.CString(device)
	}

	ret := C.av_hwdevice_ctx_create(&ff.copyHwCtx, C.AV_HWDEVICE_TYPE_VAAPI, ff.copyDevice, nil, 0)
	if ret < 0 {
		err := get_error(ret)
		return err
	}

	ff.inputCtx.hwaccel_flags = 1
	ff.inputCtx.hw_device_ctx = ff.copyHwCtx

	return nil
}

func (ff *FFMPEG_Plugin) Ready() error {
	ret := C.avcodec_open2(ff.inputCtx, ff.inputCodec, nil)
	if ret < 0 {
		err := get_error(ret)
		return err
	}

	ff.inputFrame = C.av_frame_alloc()
	if ff.inputFrame == nil {
		return errors.New("unable to allocate input frame")
	}

	ff.inputPacket = C.av_packet_alloc()
	if ff.inputPacket == nil {
		return errors.New("unable to allocate input packet")
	}

	return nil
}

func (ff *FFMPEG_Plugin) Convert(input []byte, output io.Writer) error {
	data_length := len(input)
	data := C.CBytes(input)
	data = C.av_realloc(data, C.size_t(data_length+C.AV_INPUT_BUFFER_PADDING_SIZE))
	defer C.av_free(data)

	ret := C.av_packet_from_data(ff.inputPacket, (*C.uchar)(data), (C.int)(data_length))
	if ret < 0 {
		err := get_error(ret)
		return errs.CreateError("creating input packet", err)
	}

	length := C.avcodec_send_packet(ff.inputCtx, ff.inputPacket) // packet -> codec
	if length < 0 {
		err := get_error(length)
		return errs.CreateError("reading input packet", err)
	}

	for length >= 0 { // this loop is likely not necessary
		length = C.avcodec_receive_frame(ff.inputCtx, ff.inputFrame) // codec -> frame

		if length == -C.EAGAIN || length == C.AVERROR_EOF {
			break
		} else if length < 0 {
			err := get_error(length)
			return errs.CreateError("decoding packet into frame", err)
		}

		frame := ff.inputFrame

		// uses hardware acceleration
		if ff.copyHwCtx != nil {
			if ff.copyFrame == nil {
				ff.copyFrame = C.av_frame_alloc()
				ff.copyFrame.height = frame.height
				ff.copyFrame.width = frame.width
			}

			ret = C.av_hwframe_transfer_data(ff.copyFrame, frame, 0)
			if ret < 0 {
				err := get_error(ret)
				return errs.CreateError("transferring hardware frame", err)
			}

			frame = ff.copyFrame
		}

		// do pixelformat conversion
		if ff.inputFrame.format != C.AV_PIX_FMT_RGB24 {

			if ff.outputFrame == nil {
				err := ff.init_output_frame(frame)
				if err != nil {
					return err
				}
			}

			C.ffmpeg_convert_frame(ff.outputCtx, frame, ff.outputFrame)

			frame = ff.outputFrame
		}

		for y := C.int(0); y < frame.height; y++ {
			start := C.ffmpeg_get_frame_data_ptr(frame, y)
			bytes := C.GoBytes(start, frame.width*3)
			output.Write(bytes)
		}
	}

	return nil
}

func (ff *FFMPEG_Plugin) init_output_frame(inputFrame *C.AVFrame) error {
	ff.outputFrame = C.av_frame_alloc()
	ff.outputFrame.width = inputFrame.width
	ff.outputFrame.height = inputFrame.height
	ff.outputFrame.format = C.AV_PIX_FMT_RGB24

	ret := C.av_frame_get_buffer(ff.outputFrame, 0)
	if ret < 0 {
		err := get_error(ret)
		return errs.CreateError("allocating output frame pointers", err)
	}

	ff.outputCtx = C.sws_getContext(
		inputFrame.width, inputFrame.height, int32(inputFrame.format),
		ff.outputFrame.width, ff.outputFrame.height, int32(ff.outputFrame.format),
		C.SWS_FAST_BILINEAR, nil, nil, nil,
	)

	if ff.outputCtx == nil {
		return errors.New("unable to allocate the thing that will do the conversion")
	}

	return nil
}

func (ff *FFMPEG_Plugin) End() {
	if ff.copyHwCtx != nil {
		C.av_buffer_unref(&ff.copyHwCtx)
	}

	if ff.outputCtx != nil {
		C.sws_freeContext(ff.outputCtx)
	}

	if ff.outputFrame != nil {
		C.av_frame_free(&ff.outputFrame)
	}

	if ff.inputFrame != nil {
		C.av_frame_free(&ff.inputFrame)
	}

	/*
		if conv.inputPacket != nil {
			C.av_packet_free(&conv.inputPacket)
		}
	*/

	if ff.inputCtx != nil {
		C.avcodec_free_context(&ff.inputCtx)
	}

	if ff.copyDevice != nil {
		C.free(unsafe.Pointer(ff.copyDevice))
	}
}

func get_error(errnum C.int) error {
	error_size := C.sizeof_uchar * 100
	error_c := C.malloc(C.ulong(error_size))
	error_c_char := (*C.char)(error_c)

	result := "unknown error"

	ret := C.av_strerror(errnum, error_c_char, C.ulong(error_size))
	if ret >= 0 {
		result = C.GoString(error_c_char)
	}

	C.free(error_c)
	return errors.New(result)
}
