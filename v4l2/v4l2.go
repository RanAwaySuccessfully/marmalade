package v4l2

import (
	"errors"
	"os"
	"regexp"
	"strconv"
	"unsafe"
)

/*
#cgo LDFLAGS: -lv4l2

#include "v4l2.h"
*/
import "C"

type VideoCapture struct {
	Index   int
	Name    string
	Formats []VideoFormat
}

type VideoFormat struct {
	Id             string
	Name           string
	Resolutions    []VideoFormatResolution
	ResolutionType uint32
}

type VideoFormatResolution struct {
	Width         uint32
	Height        uint32
	RangeWidth    [2]uint32
	RangeHeight   [2]uint32
	FrameRateType uint32
	FrameRates    []uint32
}

func GetInputDevices() ([]VideoCapture, error) {
	devices, err := os.ReadDir("/dev/")
	if err != nil {
		return nil, err
	}

	var inputs []VideoCapture

	for _, device := range devices {
		name := device.Name()

		index, err := get_video_index(name)
		if err != nil {
			return nil, err
		} else if index == -1 {
			continue
		}

		c_path := C.CString("/dev/" + name)
		device := C.m_v4l2_open(c_path, C.int(0))
		if device == -1 {
			_, err := get_errno()
			return nil, err
		}

		capabilities := C.struct_v4l2_capability{}
		result := C.m_v4l2_vidioc_querycap(device, &capabilities)
		if result == -1 {
			_, err := get_errno()
			return nil, err
		}

		isVideoCapture := ((capabilities.device_caps & C.V4L2_CAP_VIDEO_CAPTURE) == C.V4L2_CAP_VIDEO_CAPTURE)
		//isVideoCapture2 := ((capabilities.device_caps & C.V4L2_CAP_META_CAPTURE) == C.V4L2_CAP_META_CAPTURE)

		if isVideoCapture {
			cardname_c := unsafe.Pointer(&capabilities.card[0])
			cardname := C.GoString((*C.char)(cardname_c))

			formats, err := get_formats_for_device(device)
			if err != nil {
				return nil, err
			}

			input_device := VideoCapture{
				Index:   index,
				Name:    cardname,
				Formats: formats,
			}

			inputs = append(inputs, input_device)
		}

		result = C.v4l2_close(device)
		if result == -1 {
			_, err := get_errno()
			return nil, err
		}

		C.free(unsafe.Pointer(c_path))
	}

	return inputs, nil
}

func get_video_index(name string) (int, error) {
	regex, err := regexp.Compile(`^(video)(\d{1,2})$`) // device name must be like: videoX where X is a number between 0 and 99
	if err != nil {
		return -1, err
	}

	res := regex.FindStringSubmatch(name)
	if res == nil {
		return -1, nil
	}

	index, err := strconv.Atoi(res[2]) // convert string to integer
	if err != nil {
		return -1, err
	}

	return index, nil
}

func get_errno() (int, error) {
	errno := C.get_errno()
	c_error := C.strerror(errno)
	err := C.GoString(c_error)
	return int(errno), errors.New(err)
}
