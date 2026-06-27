package main

/*
#cgo CFLAGS: -I/ffmpeg/include
#cgo LDFLAGS: -L/ffmpeg/lib -lavcodec -lavutil -lswscale

#include "ffmpeg.h"
*/
import "C"
import (
	"bytes"
	"encoding/json"
	"errors"
	"marmalade/internal/server"
	"os"
	"unsafe"
)

type ConverterFFMPEG struct {
	inputCtx    *C.struct_AVCodecContext
	inputPacket *C.struct_AVPacket
	inputFrame  *C.struct_AVFrame

	copyFrame  *C.struct_AVFrame
	copyDevice *C.char
	copyFromHw bool

	outputCtx   *C.struct_SwsContext
	outputFrame *C.struct_AVFrame
}

type Mapping struct {
	FourCC  string
	CodecID int
	PixFmt  int
}

func (conv *ConverterFFMPEG) init(format string) error {
	var codec_id uint32
	var pix_fmt int32

	file, err := os.Open("fourcc.json")
	if err != nil {
		return create_error("opening FourCC file", err)
	}

	var mapping []Mapping

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&mapping)
	if err != nil {
		return create_error("reading FourCC file", err)
	}

	for _, mapItem := range mapping {
		if mapItem.FourCC == format {
			codec_id = uint32(mapItem.CodecID)
			pix_fmt = int32(mapItem.PixFmt)
		}
	}

	if codec_id == 0 {
		err := errors.New("format " + format + " is not mapped on file fourcc.json. You may manually add the mapping, or ask the developer to do so.")
		return create_error("finding codec for FFmpeg", err)
	}

	codec := C.avcodec_find_decoder(codec_id)
	conv.inputCtx = C.avcodec_alloc_context3(codec)

	if pix_fmt != -1 {
		conv.inputCtx.pix_fmt = pix_fmt
	}

	can_vaapi := false
	//can_qsv := false

	for i := 0; ; i++ {
		hw := C.avcodec_get_hw_config(codec, C.int(i))
		if hw == nil {
			break
		}

		switch hw.device_type {
		case C.AV_HWDEVICE_TYPE_VAAPI:
			can_vaapi = true
			continue
			/*
				case C.AV_HWDEVICE_TYPE_QSV:
					can_qsv = true
					continue
			*/
			/*
				default:
					println(C.GoString(C.av_hwdevice_get_type_name(hw.device_type)))
			*/
		}
	}

	if server.Config.HwAccel.Decode {
		if can_vaapi {
			var hwCtx *C.AVBufferRef

			if server.Config.HwAccel.PrimeId != "" {
				device_str := "/dev/dri/by-path/" + server.Config.HwAccel.PrimeId + "-render"
				conv.copyDevice = C.CString(device_str)
			}

			ret := C.av_hwdevice_ctx_create(&hwCtx, C.AV_HWDEVICE_TYPE_VAAPI, conv.copyDevice, nil, 0)
			if ret < 0 {
				err := conv.get_error(ret)
				return create_error("creating hardware context", err)
			}

			conv.inputCtx.hwaccel_flags = 1
			conv.inputCtx.hw_device_ctx = hwCtx
			conv.copyFromHw = true
		} /* else if can_qsv {
			quicksync_ctx := C.av_hwdevice_ctx_alloc(C.AV_HWDEVICE_TYPE_QSV)
			if quicksync_ctx == nil {
				return create_error("unable to allocate quicksync context", nil)
			}

			ret := C.av_hwdevice_ctx_init(quicksync_ctx)
			if ret < 0 {
				err := conv.get_error(ret)
				return create_error("initializing quicksync context", err)
			}

			conv.inputCtx.hwaccel_flags = 1
			conv.inputCtx.hw_device_ctx = quicksync_ctx
		} */
	}

	ret := C.avcodec_open2(conv.inputCtx, codec, nil)
	if ret < 0 {
		err := conv.get_error(ret)
		return create_error("opening decoder", err)
	}

	conv.inputFrame = C.av_frame_alloc()
	if conv.inputFrame == nil {
		return create_error("unable to allocate input frame", nil)
	}

	conv.inputPacket = C.av_packet_alloc()
	if conv.inputPacket == nil {
		return create_error("unable to allocate input packet", nil)
	}

	return nil
}

func (conv *ConverterFFMPEG) init_output_frame(inputFrame *C.AVFrame) error {
	conv.outputFrame = C.av_frame_alloc()
	conv.outputFrame.width = inputFrame.width
	conv.outputFrame.height = inputFrame.height
	conv.outputFrame.format = C.AV_PIX_FMT_RGB24

	ret := C.av_frame_get_buffer(conv.outputFrame, 0)
	if ret < 0 {
		err := conv.get_error(ret)
		return create_error("allocating output frame pointers", err)
	}

	conv.outputCtx = C.sws_getContext(
		inputFrame.width, inputFrame.height, int32(inputFrame.format),
		conv.outputFrame.width, conv.outputFrame.height, int32(conv.outputFrame.format),
		C.SWS_FAST_BILINEAR, nil, nil, nil,
	)

	if conv.outputCtx == nil {
		return create_error("unable to allocate the thing that will do the conversion", nil)
	}

	return nil
}

func (conv *ConverterFFMPEG) convert(input []byte) ([]byte, error) {
	data_length := len(input)
	data := C.CBytes(input)
	data = C.av_realloc(data, C.size_t(data_length+C.AV_INPUT_BUFFER_PADDING_SIZE))
	defer C.av_free(data)

	ret := C.av_packet_from_data(conv.inputPacket, (*C.uchar)(data), (C.int)(data_length))
	if ret < 0 {
		err := conv.get_error(ret)
		return nil, create_error("creating input packet", err)
	}

	length := C.avcodec_send_packet(conv.inputCtx, conv.inputPacket) // packet -> codec
	if length < 0 {
		err := conv.get_error(length)
		return nil, create_error("reading input packet", err)
	}

	output := bytes.Buffer{}

	for length >= 0 { // this loop is likely not necessary
		length = C.avcodec_receive_frame(conv.inputCtx, conv.inputFrame) // codec -> frame

		if length == -C.EAGAIN || length == C.AVERROR_EOF {
			break
		} else if length < 0 {
			err := conv.get_error(length)
			return nil, create_error("decoding packet into frame", err)
		}

		frame := conv.inputFrame

		// uses hardware acceleration
		if conv.copyFromHw {
			if conv.copyFrame == nil {
				conv.copyFrame = C.av_frame_alloc()
				conv.copyFrame.height = frame.height
				conv.copyFrame.width = frame.width
			}

			ret = C.av_hwframe_transfer_data(conv.copyFrame, frame, 0)
			if ret < 0 {
				err := conv.get_error(ret)
				return nil, create_error("transferring hardware frame", err)
			}

			frame = conv.copyFrame
		}

		// do pixelformat conversion
		if conv.inputFrame.format != C.AV_PIX_FMT_RGB24 {

			if conv.outputFrame == nil {
				err := conv.init_output_frame(frame)
				if err != nil {
					return nil, err
				}
			}

			C.ffmpeg_convert_frame(conv.outputCtx, frame, conv.outputFrame)

			frame = conv.outputFrame
		}

		for y := C.int(0); y < frame.height; y++ {
			start := C.ffmpeg_get_frame_data_ptr(frame, y)
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

	/*
		if conv.inputPacket != nil {
			C.av_packet_free(&conv.inputPacket)
		}
	*/

	if conv.inputCtx != nil {
		C.avcodec_free_context(&conv.inputCtx)
	}

	if conv.copyDevice != nil {
		C.free(unsafe.Pointer(conv.copyDevice))
	}
}

func (conv *ConverterFFMPEG) get_error(errnum C.int) error {
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
