package main

/*
#cgo LDFLAGS: -lv4lconvert

#include <stdint.h>
#include <stdlib.h>
#include <linux/videodev2.h>
#include <libv4lconvert.h>

static void set_v4l2_fmtpix(struct v4l2_format* format, struct v4l2_pix_format pix) {
	format->fmt.pix = pix;
}
*/
import "C"
import (
	"errors"
	"io"

	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
)

type V4LConvert struct {
	format  v4l2.PixFormat
	data    *C.struct_v4lconvert_data
	src_fmt C.struct_v4l2_format
	dst_fmt C.struct_v4l2_format
}

func NewV4LConvert(dev *device.Device) (*V4LConvert, error) {
	converter := &V4LConvert{}

	var err error
	converter.format, err = dev.GetPixFormat()
	if err != nil {
		return nil, err
	}

	fd := dev.Fd()
	converter.data = C.v4lconvert_create(C.int(fd))

	// TODO: check if input pixel format is supported by V4LConvert

	// INPUT

	converter.src_fmt = C.struct_v4l2_format{
		_type: C.V4L2_BUF_TYPE_VIDEO_CAPTURE,
	}

	src_pixfmt := C.struct_v4l2_pix_format{
		width:       C.uint32_t(converter.format.Width),
		height:      C.uint32_t(converter.format.Height),
		pixelformat: C.uint32_t(converter.format.PixelFormat),
	}

	C.set_v4l2_fmtpix(&converter.src_fmt, src_pixfmt)

	// OUTPUT

	converter.dst_fmt = C.struct_v4l2_format{
		_type: C.V4L2_BUF_TYPE_VIDEO_CAPTURE,
	}

	dst_pixfmt := C.struct_v4l2_pix_format{
		width:       C.uint32_t(converter.format.Width),
		height:      C.uint32_t(converter.format.Height),
		pixelformat: C.uint32_t(v4l2.PixelFmtRGB24),
	}

	C.set_v4l2_fmtpix(&converter.dst_fmt, dst_pixfmt)

	return converter, nil
}

func (converter *V4LConvert) Convert(input []byte, output io.Writer) error {
	src_length := len(input)
	src_data := C.CBytes(input)
	defer C.free(src_data)

	dst_length := converter.format.Width * converter.format.Height * 3
	dst_data := C.malloc(C.size_t(dst_length))
	defer C.free(dst_data)

	ret := C.v4lconvert_convert(
		converter.data,
		&converter.src_fmt,
		&converter.dst_fmt,
		(*C.uchar)(src_data),
		C.int(src_length),
		(*C.uchar)(dst_data),
		C.int(dst_length),
	)

	if ret < 0 {
		err_msg_c := C.v4lconvert_get_error_message(converter.data)
		err_msg := C.GoString(err_msg_c)
		return errors.New("v4lconvert_convert failed: " + err_msg)
	}

	output_bytes := C.GoBytes(dst_data, C.int(dst_length))
	output.Write(output_bytes)
	return nil
}

func (converter *V4LConvert) End() {
	if converter.data != nil {
		C.v4lconvert_destroy(converter.data)
	}
}
