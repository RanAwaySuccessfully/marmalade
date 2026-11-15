package webcam

import (
	"errors"
	"io"
	"marmalade/internal/errs"
	"marmalade/internal/server"
	"os/exec"
	"plugin"
	"regexp"
	"strings"
)

const AvHwDeviceTypeVAAPI uint32 = 3
const AvHwDeviceTypeQSV uint32 = 5

type FFMPEG_Plugin interface {
	Init(fourcc uint32)
	FindHwAccel() []uint32
	UseHwAccel(device string) error
	Ready() error
	Convert(input []byte, output io.Writer) error
	End()
}

func NewFFMPEG(fourcc uint32) (FFMPEG_Plugin, error) {
	ffmpeg, err := find_ffmpeg_plugin()
	if err != nil {
		return nil, errs.CreateError("initializing FFmpeg plugin", err)
	}

	ffmpeg.Init(fourcc)
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
	cmd := exec.Command("ffmpeg", "-version")

	var out strings.Builder
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	result := out.String()
	re := regexp.MustCompile(`\d`)

	plugin_filepath := "lib/ffmpeg" + re.FindString(result) + "_plugin.so"
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
