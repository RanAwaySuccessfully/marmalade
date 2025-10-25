//go:build withgtk4

package camera

import "github.com/vladimirvivien/go4vl/v4l2"

func get_formats(device uintptr, result *VideoCapture) error {
	formats, err := v4l2.GetAllFormatDescriptions(device)
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

		result.Formats = append(result.Formats, format)

		err = get_resolutions(device, &format)
		if err != nil {
			return err
		}
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

		format.Resolutions = append(format.Resolutions, resolution)

		err = get_frame_rates(device, format, &resolution)
		if err != nil {
			return err
		}
	}

	return nil
}

func get_frame_rates(device uintptr, format *VideoFormat, resolution *VideoFormatResolution) error {
	resolution.FrameRates = make([]v4l2.FrameIntervalEnum, 0)

	for {
		frame_interval, err := v4l2.GetFormatFrameInterval(device, 0, format.Data.PixelFormat, 0, 0)
		if err != nil {
			if len(resolution.FrameRates) <= 0 {
				return err
			} else {
				return nil
			}
		}

		resolution.FrameRates = append(resolution.FrameRates, frame_interval)
	}
}
