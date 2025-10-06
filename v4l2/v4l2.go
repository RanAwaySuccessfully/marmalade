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

type VideoFormat struct {
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

		c_path := C.CString("/dev/" + name)
		device := C.m_v4l2_open(c_path, C.int(0))
		if device == -1 {
			err := get_errno()
			return nil, err
		}

		capabilities := C.struct_v4l2_capability{}
		result := C.m_v4l2_vidioc_querycap(device, &capabilities)
		if result == -1 {
			err := get_errno()
			return nil, err
		}

		cardname_c := unsafe.Pointer(&capabilities.card[0])
		cardname := C.GoString((*C.char)(cardname_c))

		isVideoCapture := ((capabilities.device_caps & C.V4L2_CAP_VIDEO_CAPTURE) == C.V4L2_CAP_VIDEO_CAPTURE)
		//isVideoCapture2 := ((capabilities.device_caps & C.V4L2_CAP_META_CAPTURE) == C.V4L2_CAP_META_CAPTURE)

		/*
			fmt_index := C.uint(0)

			format := C.struct_v4l2_fmtdesc{
				index:     fmt_index,
				_type:     C.V4L2_BUF_TYPE_VIDEO_CAPTURE,
				mbus_code: 0,
			}

			result = C.m_v4l2_vidioc_enum_fmt(device, &format)
			if result == -1 {
				err := get_errno()
				return nil, err
			}

			fmtdesc_c := unsafe.Pointer(&format.description[0])
			fmtdesc := C.GoString((*C.char)(fmtdesc_c))
			println(fmtdesc)
		*/

		/*
			https://www.kernel.org/doc/html/v6.14/userspace-api/media/v4l/vidioc-enum-fmt.html
			https://www.kernel.org/doc/html/v6.14/userspace-api/media/v4l/vidioc-enum-framesizes.html#c.V4L.VIDIOC_ENUM_FRAMESIZES
			https://www.kernel.org/doc/html/v6.14/userspace-api/media/v4l/vidioc-enum-frameintervals.html#c.V4L.VIDIOC_ENUM_FRAMEINTERVALS
		*/

		result = C.v4l2_close(device)
		if result == -1 {
			err := get_errno()
			return nil, err
		}

		C.free(unsafe.Pointer(c_path))

		if isVideoCapture {
			input_device := VideoCapture{
				Index: index,
				Name:  cardname,
			}

			inputs = append(inputs, input_device)
		}
	}

	return inputs, nil
}

func get_errno() error {
	errno := C.get_errno()
	c_error := C.strerror(errno)
	err := C.GoString(c_error)
	return errors.New(err)
}
