package v4l2

/*
#cgo LDFLAGS: -lv4l2

#include "v4l2.h"
*/
import "C"
import (
	"errors"
	"os"
	"regexp"
	"strconv"
	"unsafe"
)

type VideoCapture struct {
	Index int
	Name  string
}

func GetInputDevices() ([]VideoCapture, error) {
	devices, err := os.ReadDir("/dev/")
	if err != nil {
		return nil, err
	}

	var inputs []VideoCapture

	for _, device := range devices {
		name := device.Name()

		regex, err := regexp.Compile(`^(video)(\d{1,2})$`) // device name must be like: videoX where X is a number between 0 and 99
		if err != nil {
			return nil, err
		}

		res := regex.FindStringSubmatch(name)
		if res == nil {
			continue
		}

		index, err := strconv.Atoi(res[2]) // convert string to integer
		if err != nil {
			return nil, err
		}

		cardname_size := C.sizeof_uchar * 33
		cardname_c := C.malloc(C.ulong(cardname_size))
		cardname_c_char := (*C.char)(cardname_c)

		c_path := C.CString("/dev/" + name)
		result := C.check_real_video_capture_device(c_path, cardname_c_char)
		C.free(unsafe.Pointer(c_path))

		cardname := C.GoString(cardname_c_char)
		C.free(cardname_c)

		if result < 0 {
			c_error := C.strerror(-result)
			err := C.GoString(c_error)
			return nil, errors.New(err)

		} else if result == 1 {
			input_device := VideoCapture{
				Index: index,
				Name:  cardname,
			}

			inputs = append(inputs, input_device)
		}
	}

	return inputs, nil
}
