package devices

/*
#cgo LDFLAGS: -lv4l2

#include <linux/videodev2.h>
#include <libv4l2.h>
#include <errno.h>

int m_v4l2_open(char* file, int oflag) {
    return v4l2_open(file, oflag);
}

int m_v4l2_ioctl(int fd, long request, void* obj) {
    return v4l2_ioctl(fd, request, obj);
}
*/
import "C"
import (
	"errors"
	"syscall"
	"unsafe"

	"github.com/vladimirvivien/go4vl/v4l2"
)

func v4l2_open(path string, flags int) (uintptr, error) {
	c_path := C.CString(path)
	device, err := C.m_v4l2_open(c_path, C.int(flags))
	if device == -1 {
		return uintptr(0), err
	}

	return uintptr(device), nil
}

func close_device(device uintptr) {
	C.v4l2_close(C.int(device))
}

func get_formats2(device uintptr, result *VideoCapture) error {
	formats := make([]v4l2.FormatDescription, 0)
	var err error

	for index := 0; ; index++ {
		fmtDesc := C.struct_v4l2_fmtdesc{
			index: C.uint(index),
			_type: C.V4L2_BUF_TYPE_VIDEO_CAPTURE,
		}

		result, err := C.m_v4l2_ioctl(C.int(device), C.VIDIOC_ENUM_FMT, unsafe.Pointer(&fmtDesc))
		if result == -1 {
			if errors.Is(err, syscall.EINVAL) {
				break
			} else {
				return err
			}
		}

		formats = append(formats, v4l2.FormatDescription{
			Index:       uint32(fmtDesc.index),
			StreamType:  uint32(fmtDesc._type),
			Flags:       uint32(fmtDesc.flags),
			Description: C.GoString((*C.char)(unsafe.Pointer(&fmtDesc.description[0]))),
			PixelFormat: uint32(fmtDesc.pixelformat),
			MBusCode:    uint32(fmtDesc.mbus_code),
		})
	}

	if len(formats) <= 0 {
		return err
	}

	result.Formats = make([]VideoFormat, 0, len(formats))

	for _, format_data := range formats {

		isCompressed := ((format_data.Flags & v4l2.FmtDescFlagCompressed) == v4l2.FmtDescFlagCompressed)
		isEmulated := ((format_data.Flags & v4l2.FmtDescFlagEmulated) == v4l2.FmtDescFlagEmulated)

		var pixelformat string
		for i := range 4 {
			pixelformat += string(byte(format_data.PixelFormat >> (i * 8)))
		}

		format := VideoFormat{
			Id:         pixelformat,
			Data:       format_data,
			Compressed: isCompressed,
			Emulated:   isEmulated,
		}

		err = get_resolutions2(uintptr(device), &format)
		if err != nil {
			return err
		}

		result.Formats = append(result.Formats, format)
	}

	return nil
}

func get_resolutions2(device uintptr, format *VideoFormat) error {
	resolutions := make([]v4l2.FrameSizeEnum, 0)
	var err error

	for index := 0; ; index++ {
		fmtDesc := C.struct_v4l2_fmtdesc{
			index: C.uint(index),
			_type: C.V4L2_BUF_TYPE_VIDEO_CAPTURE,
		}

		result, err := C.m_v4l2_ioctl(C.int(device), C.VIDIOC_ENUM_FRAMESIZES, unsafe.Pointer(&fmtDesc))
		if result == -1 {
			if errors.Is(err, syscall.EINVAL) {
				break
			} else {
				return err
			}
		}

		/*
			resolutions = append(resolutions, v4l2.FrameSizeEnum{
				Index:       uint32(fmtDesc.index),
				StreamType:  uint32(fmtDesc._type),
				Flags:       uint32(fmtDesc.flags),
				Description: C.GoString((*C.char)(unsafe.Pointer(&fmtDesc.description[0]))),
				PixelFormat: uint32(fmtDesc.pixelformat),
				MBusCode:    uint32(fmtDesc.mbus_code),
			})
		*/
	}

	if len(resolutions) <= 0 {
		return err
	}

	format.Resolutions = make([]VideoFormatResolution, 0, len(resolutions))

	for _, resolution_data := range resolutions {
		resolution := VideoFormatResolution{
			Data: resolution_data,
		}

		err = get_frame_rates2(device, format, &resolution)
		if err != nil {
			return err
		}

		format.Resolutions = append(format.Resolutions, resolution)
	}

	return nil
}

func get_frame_rates2(device uintptr, format *VideoFormat, resolution *VideoFormatResolution) error {
	resolution.FrameRates = make([]v4l2.FrameIntervalEnum, 0)
	index := 0

	for {
		width := resolution.Data.Size.MaxWidth
		height := resolution.Data.Size.MaxHeight

		frame_interval, err := v4l2.GetFormatFrameInterval(device, uint32(index), format.Data.PixelFormat, width, height)
		if err != nil {
			if len(resolution.FrameRates) <= 0 {
				return err
			} else {
				return nil
			}
		}

		resolution.FrameRates = append(resolution.FrameRates, frame_interval)
		index++
	}
}
