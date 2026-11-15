package webcam

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"marmalade/internal/devices"
	"marmalade/internal/errs"
	"marmalade/internal/server"

	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
)

type RunCallback func(data []byte, width uint32, height uint32, timestamp int64) error

type Converter interface {
	Convert(input []byte, output io.Writer) error
	End()
}

type GOV4L struct {
	webcam    *device.Device
	converter Converter
}

func NewGOV4L(fourcc v4l2.FourCCType) (*GOV4L, error) {
	v4l := GOV4L{}

	pix_format := v4l2.PixFormat{
		Width:       uint32(server.Config.Width),
		Height:      uint32(server.Config.Height),
		PixelFormat: uint32(fourcc),
	}

	device_path := fmt.Sprintf("/dev/video%d", int(server.Config.Camera))

	var err error
	v4l.webcam, err = device.Open(
		device_path,
		device.WithBufferSize(1),
		device.WithPixFormat(pix_format),
		device.WithFPS(uint32(server.Config.FPS)),
	)

	if err != nil {
		return nil, errs.CreateError("opening camera device", err)
	}

	//v4l.check_supported_settings()

	v4l.webcam.GetFrames()
	err = v4l.webcam.Start(context.Background())
	if err != nil {
		return nil, errs.CreateError("starting camera feed", err)
	}

	// "RGB3" is the format MediaPipe needs
	if fourcc != v4l2.PixelFmtRGB24 {

		if server.Config.HwAccel.ForceFFMPEG {
			v4l.converter, err = NewFFMPEG(fourcc)
			if err != nil {
				return nil, err
			}

		} else {
			v4l.converter, err = NewV4LConvert(v4l.webcam)
			if err != nil {
				return nil, err
			}
		}
	}

	return &v4l, nil
}

func (v4l *GOV4L) Run(err_ch chan error, detect RunCallback) {
	for frame := range v4l.webcam.GetFrames() {

		format, err := v4l.webcam.GetPixFormat()
		if err != nil {
			err_ch <- errs.CreateError("reading webcam current format", err)
			break
		}

		timestamp := frame.Timestamp.UnixMilli()

		var srgb_data []byte
		if v4l.converter == nil {
			srgb_data = frame.Data
		} else {
			output := bytes.Buffer{}
			err := v4l.converter.Convert(frame.Data, &output)
			if err != nil {
				err_ch <- err
			}

			srgb_data = output.Bytes()
		}

		err = detect(srgb_data, format.Width, format.Height, timestamp)
		if err != nil {
			err_ch <- err
			frame.Release()
			break
		}

		frame.Release()
	}
}

func (v4l *GOV4L) End() {
	if v4l.webcam != nil {
		v4l.webcam.Close()
	}

	if v4l.converter != nil {
		v4l.converter.End()
	}
}

func (v4l *GOV4L) check_supported_settings() error {
	fmt_real, err := v4l.webcam.GetPixFormat()
	if err != nil {
		return errs.CreateError("reading webcam current format", err)
	}

	fps_real, err := v4l.webcam.GetFrameRate()
	if err != nil {
		return errs.CreateError("reading webcam current frame rate", err)
	}

	pixelformat := devices.PixFmtToString(fmt_real.PixelFormat)

	fmt.Printf("[%s] %dx%d@%d\n", pixelformat, fmt_real.Width, fmt_real.Height, fps_real)
	return nil
}
