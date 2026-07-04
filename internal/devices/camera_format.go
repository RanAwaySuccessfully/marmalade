package devices

import (
	"bytes"
	"encoding/binary"

	"github.com/vladimirvivien/go4vl/v4l2"
)

func PixFmtToString(pixfmt v4l2.FourCCType) string {
	var result string
	for i := range 4 {
		result += string(byte(pixfmt >> (i * 8)))
	}

	return result
}

func StringToPixFmt(pixfmt string) (v4l2.FourCCType, error) {
	buffer := bytes.Buffer{}
	buffer.WriteString(pixfmt)

	var fourcc v4l2.FourCCType
	err := binary.Read(&buffer, binary.LittleEndian, &fourcc)
	return fourcc, err
}

func get_formats(device uintptr, result *VideoCapture) error {
	formats, err := v4l2.GetAllFormatDescriptions(device)
	if len(formats) <= 0 {
		return err
	}

	result.Formats = make([]VideoFormat, 0, len(formats))

	for _, format_data := range formats {

		isCompressed := ((format_data.Flags & v4l2.FmtDescFlagCompressed) == v4l2.FmtDescFlagCompressed)
		isEmulated := ((format_data.Flags & v4l2.FmtDescFlagEmulated) == v4l2.FmtDescFlagEmulated)

		pixelformat := PixFmtToString(format_data.PixelFormat)
		format := VideoFormat{
			Id:         pixelformat,
			Data:       format_data,
			Compressed: isCompressed,
			Emulated:   isEmulated,
		}

		err = get_resolutions(device, &format)
		if err != nil {
			return err
		}

		result.Formats = append(result.Formats, format)
	}

	return nil
}

func get_resolutions(device uintptr, format *VideoFormat) error {
	resolutions, err := v4l2.GetFormatFrameSizes(device, format.Data.PixelFormat)
	if len(resolutions) <= 0 {
		return err
	}

	format.Resolutions = make([]VideoFormatResolution, 0, len(resolutions))

	for _, resolution_data := range resolutions {
		resolution := VideoFormatResolution{
			Data: resolution_data,
		}

		err = get_frame_rates(device, format, &resolution)
		if err != nil {
			return err
		}

		format.Resolutions = append(format.Resolutions, resolution)
	}

	return nil
}

func get_frame_rates(device uintptr, format *VideoFormat, resolution *VideoFormatResolution) error {
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
