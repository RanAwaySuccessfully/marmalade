package main

/*
const int libavcodec_check_version() {
    #include <libavcodec/version.h>
    return LIBAVCODEC_VERSION_MAJOR;
}
*/
import "C"
import (
	"errors"
	"fmt"
	"io"
	"marmalade/internal/errs"
	"marmalade/internal/server"
	"plugin"
)

const AvHwDeviceTypeVAAPI uint32 = 3
const AvHwDeviceTypeQSV uint32 = 5

type FFMPEG_Plugin interface {
	Init(codec_id uint32, pix_fmt int32)
	FindHwAccel() []uint32
	UseHwAccel(device string) error
	Ready() error
	Convert(input []byte, output io.Writer) error
	End()
}

func NewFFMPEG(format *Mapping) (FFMPEG_Plugin, error) {
	codec_id := format.CodecID
	pix_fmt := format.PixFmt

	if codec_id == 0 {
		err := errors.New("format " + format.Format + " is not mapped on file fourcc.json. You may manually add the mapping, or ask the developer to do so.")
		return nil, errs.CreateError("finding codec for FFmpeg", err)
	}

	ffmpeg, err := find_ffmpeg_plugin()
	if err != nil {
		return nil, errs.CreateError("initializing FFmpeg plugin", err)
	}

	ffmpeg.Init(codec_id, pix_fmt)
	hw := ffmpeg.FindHwAccel()

	can_vaapi := false
	//can_qsv := false

	for _, hw_option := range hw {
		switch hw_option {
		case AvHwDeviceTypeVAAPI:
			can_vaapi = true
		}
	}

	if server.Config.HwAccel.Decode {
		if can_vaapi {
			if server.Config.HwAccel.PrimeId != "" {
				device_str := "/dev/dri/by-path/" + server.Config.HwAccel.PrimeId + "-render"
				err = ffmpeg.UseHwAccel(device_str)
			} else {
				err = ffmpeg.UseHwAccel("")
			}

			if err != nil {
				return nil, errs.CreateError("creating hardware context", err)
			}

		} /* else if can_qsv {
			quicksync_ctx := C.av_hwdevice_ctx_alloc(C.AV_HWDEVICE_TYPE_QSV)
			if quicksync_ctx == nil {
				return errs.CreateError("unable to allocate quicksync context", nil)
			}

			ret := C.av_hwdevice_ctx_init(quicksync_ctx)
			if ret < 0 {
				err := conv.get_error(ret)
				return errs.CreateError("initializing quicksync context", err)
			}

			conv.inputCtx.hwaccel_flags = 1
			conv.inputCtx.hw_device_ctx = quicksync_ctx
		} */
	}

	err = ffmpeg.Ready()
	if err != nil {
		return nil, errs.CreateError("opening decoder", err)
	}

	return ffmpeg, nil
}

func find_ffmpeg_plugin() (FFMPEG_Plugin, error) {
	ver := C.libavcodec_check_version()
	plugin_filepath := ""

	switch ver {
	case 62:
		plugin_filepath = "lib/ffmpeg8_plugin.so"
	case 61:
		plugin_filepath = "lib/ffmpeg7_plugin.so"
	case 60:
		plugin_filepath = "lib/ffmpeg6_plugin.so"
	case 59:
		plugin_filepath = "lib/ffmpeg5_plugin.so"
	case 58:
		plugin_filepath = "lib/ffmpeg4_plugin.so"
	default:
		return nil, fmt.Errorf("No plugin available for libavcodec version: %d", ver)
	}

	plugin_file, err := plugin.Open(plugin_filepath)
	if err != nil {
		return nil, err
	}

	symbol, err := plugin_file.Lookup("FFmpeg")
	if err != nil {
		return nil, err
	}

	ffmpeg, ok := symbol.(FFMPEG_Plugin)
	if !ok {
		return nil, errors.New("FFMPEG_Plugin interface does not match")
	}

	return ffmpeg, nil
}
