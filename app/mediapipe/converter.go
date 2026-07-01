package main

import (
	"encoding/json"
	"errors"
	"marmalade/internal/interfaces"
	"marmalade/internal/server"
	"os"
	"plugin"
)

const AvHwDeviceTypeVAAPI uint32 = 3
const AvHwDeviceTypeQSV uint32 = 5

type Mapping struct {
	FourCC  string
	CodecID int
	PixFmt  int
}

func (mp *MediaPipe) find_ffmpeg_plugin() error {
	ffmpeg, err := plugin.Open("ffmpeg6_plugin.so")
	if err != nil {
		return err
	}

	symbol, err := ffmpeg.Lookup("FFmpeg")
	if err != nil {
		return err
	}

	var ok bool
	mp.ffmpeg, ok = symbol.(interfaces.FFMPEG)
	if !ok {
		return errors.New("FFMPEG interface does not match")
	}

	return nil
}

func (mp *MediaPipe) converter_init(format string) error {
	var codec_id uint32
	var pix_fmt int32

	file, err := os.Open("fourcc.json")
	if err != nil {
		return interfaces.CreateError("opening FourCC file", err)
	}

	var mapping []Mapping

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&mapping)
	if err != nil {
		return interfaces.CreateError("reading FourCC file", err)
	}

	for _, mapItem := range mapping {
		if mapItem.FourCC == format {
			codec_id = uint32(mapItem.CodecID)
			pix_fmt = int32(mapItem.PixFmt)
		}
	}

	if codec_id == 0 {
		err := errors.New("format " + format + " is not mapped on file fourcc.json. You may manually add the mapping, or ask the developer to do so.")
		return interfaces.CreateError("finding codec for FFmpeg", err)
	}

	err = mp.find_ffmpeg_plugin()
	if err != nil {
		return interfaces.CreateError("initializing FFmpeg plugin", err)
	}

	mp.ffmpeg.Init(codec_id, pix_fmt)
	hw := mp.ffmpeg.FindHwAccel()

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
				err = mp.ffmpeg.UseHwAccel(device_str)
			} else {
				err = mp.ffmpeg.UseHwAccel("")
			}

			if err != nil {
				return interfaces.CreateError("creating hardware context", err)
			}

		} /* else if can_qsv {
			quicksync_ctx := C.av_hwdevice_ctx_alloc(C.AV_HWDEVICE_TYPE_QSV)
			if quicksync_ctx == nil {
				return interfaces.CreateError("unable to allocate quicksync context", nil)
			}

			ret := C.av_hwdevice_ctx_init(quicksync_ctx)
			if ret < 0 {
				err := conv.get_error(ret)
				return interfaces.CreateError("initializing quicksync context", err)
			}

			conv.inputCtx.hwaccel_flags = 1
			conv.inputCtx.hw_device_ctx = quicksync_ctx
		} */
	}

	err = mp.ffmpeg.Ready()
	if err != nil {
		return interfaces.CreateError("opening decoder", err)
	}

	return nil
}
