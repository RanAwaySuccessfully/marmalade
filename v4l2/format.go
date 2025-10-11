package v4l2

import (
	"unsafe"
)

/*
#cgo LDFLAGS: -lv4l2

#include "v4l2.h"
*/
import "C"

func get_formats_for_device(device C.int) ([]VideoFormat, error) {
	index := C.uint(0)

	var formats []VideoFormat

	for {
		format_data := C.struct_v4l2_fmtdesc{
			index:     index,
			_type:     C.V4L2_BUF_TYPE_VIDEO_CAPTURE,
			mbus_code: 0,
		}

		result := C.m_v4l2_vidioc_enum_fmt(device, &format_data)
		if result == -1 {
			errno, err := get_errno()
			if errno == C.EINVAL {
				break
			} else {
				return nil, err
			}
		}

		var pixelformat string
		for i := range 4 {
			pixelformat += string(byte(format_data.pixelformat >> (i * 8)))
		}

		fmtdesc_c := unsafe.Pointer(&format_data.description[0])
		fmtdesc := C.GoString((*C.char)(fmtdesc_c))

		resolutions, res_type, err := get_resolutions_for_format(device, format_data.pixelformat)
		if err != nil {
			return nil, err
		}

		isCompressed := ((format_data.flags & C.V4L2_FMT_FLAG_COMPRESSED) == C.V4L2_FMT_FLAG_COMPRESSED)
		isEmulated := ((format_data.flags & C.V4L2_FMT_FLAG_EMULATED) == C.V4L2_FMT_FLAG_EMULATED)

		format := VideoFormat{
			Id:             pixelformat,
			Name:           fmtdesc,
			Resolutions:    resolutions,
			ResolutionType: res_type,
			Compressed:     isCompressed,
			Emulated:       isEmulated,
		}

		formats = append(formats, format)
		index++
	}

	return formats, nil
}

func get_resolutions_for_format(device C.int, pixelformat C.uint) ([]VideoFormatResolution, uint32, error) {
	index := C.uint(0)

	var resolutions []VideoFormatResolution
	var res_type uint32

	for {
		res_data := C.struct_v4l2_frmsizeenum{
			index:        index,
			pixel_format: pixelformat,
		}

		result := C.m_v4l2_vidioc_enum_framesizes(device, &res_data)
		if result == -1 {
			errno, err := get_errno()
			if errno == C.EINVAL {
				break
			} else {
				return nil, 0, err
			}
		}

		res_type = uint32(res_data._type)

		resolution := VideoFormatResolution{}

		res_inner_data := unsafe.Pointer(&res_data.anon0[0])

		if res_data._type == C.V4L2_FRMSIZE_TYPE_DISCRETE {
			discrete_data := (*C.struct_v4l2_frmsize_discrete)(res_inner_data)

			resolution.Width = uint32(discrete_data.width)
			resolution.Height = uint32(discrete_data.height)

			framerates, fps_type, err := get_framerates_for_resolution(device, pixelformat, discrete_data.width, discrete_data.height)
			if err != nil {
				return nil, 0, err
			}

			resolution.FrameRates = framerates
			resolution.FrameRateType = fps_type
			resolutions = append(resolutions, resolution)

		} else {
			stepwise_data := (*C.struct_v4l2_frmsize_stepwise)(res_inner_data)

			resolution.RangeWidth[0] = uint32(stepwise_data.min_width)
			resolution.RangeWidth[1] = uint32(stepwise_data.max_width)
			resolution.RangeHeight[0] = uint32(stepwise_data.min_height)
			resolution.RangeHeight[1] = uint32(stepwise_data.max_height)

			if res_data._type == C.V4L2_FRMSIZE_TYPE_STEPWISE {
				resolution.Width = uint32(stepwise_data.step_width)
				resolution.Height = uint32(stepwise_data.step_height)
			}

			framerates, fps_type, err := get_framerates_for_resolution(device, pixelformat, stepwise_data.max_width, stepwise_data.max_height)
			if err != nil {
				return nil, 0, err
			}

			resolution.FrameRates = framerates
			resolution.FrameRateType = fps_type
			resolutions = append(resolutions, resolution)

			break
		}

		index++
	}

	return resolutions, uint32(res_type), nil
}

func get_framerates_for_resolution(device C.int, pixelformat C.uint, width C.uint, height C.uint) ([]uint32, uint32, error) {
	index := C.uint(0)

	var framerates []uint32
	var fps_type uint32

	for {
		frame_data := C.struct_v4l2_frmivalenum{
			index:        index,
			pixel_format: pixelformat,
			width:        width,
			height:       height,
		}

		result := C.m_v4l2_vidioc_enum_frameintervals(device, &frame_data)
		if result == -1 {
			errno, err := get_errno()
			if errno == C.EINVAL {
				break
			} else {
				return nil, 0, err
			}
		}

		fps_type = uint32(frame_data._type)

		res_inner_data := unsafe.Pointer(&frame_data.anon0[0])

		if frame_data._type == C.V4L2_FRMIVAL_TYPE_DISCRETE {
			discrete_data := (*C.struct_v4l2_fract)(res_inner_data)
			framerates = append(framerates, frac_to_int(*discrete_data))
		} else {
			stepwise_data := (*C.struct_v4l2_frmival_stepwise)(res_inner_data)
			fps_min := stepwise_data.min
			fps_max := stepwise_data.max

			framerates = append(framerates, frac_to_int(fps_min))
			framerates = append(framerates, frac_to_int(fps_max))

			if frame_data._type == C.V4L2_FRMSIZE_TYPE_STEPWISE {
				fps_step := stepwise_data.step
				framerates = append(framerates, frac_to_int(fps_step))
			}

			break
		}

		index++
	}

	return framerates, fps_type, nil
}

func frac_to_int(frac C.struct_v4l2_fract) uint32 {
	result := uint32(frac.denominator) / uint32(frac.numerator)
	return result
}
