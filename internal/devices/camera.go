package devices

import (
	"os"
	"regexp"
	"strconv"
	"syscall"

	"github.com/vladimirvivien/go4vl/v4l2"
)

type VideoCapture struct {
	Index   uint8
	Name    string
	Formats []VideoFormat
}

type VideoFormat struct {
	Data        v4l2.FormatDescription
	Resolutions []VideoFormatResolution
	Id          string
	Compressed  bool
	Emulated    bool
}

type VideoFormatResolution struct {
	Data       v4l2.FrameSizeEnum
	FrameRates []v4l2.FrameIntervalEnum
}

func ListVideoCaptures() ([]VideoCapture, error) {
	devFiles, err := os.ReadDir("/dev/")
	if err != nil {
		return nil, err
	}

	var inputs []VideoCapture

	for _, devFile := range devFiles {
		name := devFile.Name()

		index, err := get_video_index(name)
		if err != nil {
			return nil, err
		} else if index == -1 {
			continue
		}

		devicePath := "/dev/" + name
		device, err := v4l2.OpenDevice(devicePath, syscall.O_RDWR, 0)
		if err != nil {
			return nil, err
		}

		capabilities, err := v4l2.GetCapability(device)
		if err != nil {
			v4l2.CloseDevice(device)
			return nil, err
		}

		isVideoCapture := (capabilities.DeviceCapabilities & v4l2.CapVideoCapture) == v4l2.CapVideoCapture

		if isVideoCapture {
			cardname := capabilities.Card

			input_device := VideoCapture{
				Index:   uint8(index),
				Name:    cardname,
				Formats: nil,
			}

			inputs = append(inputs, input_device)
		}

		err = v4l2.CloseDevice(device)
		if err != nil {
			return nil, err
		}
	}

	return inputs, nil
}

func GetVideoCaptureDetails(camera_id uint8) (*VideoCapture, error) {
	devicePath := "/dev/video" + strconv.Itoa(int(camera_id)) // convert int to string

	device, err := v4l2.OpenDevice(devicePath, syscall.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	defer v4l2.CloseDevice(device)

	capabilities, err := v4l2.GetCapability(device)
	if err != nil {
		return nil, err
	}

	cardname := capabilities.Card

	result := VideoCapture{
		Index: camera_id,
		Name:  cardname,
	}

	err = get_formats(device, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
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

	index, err := strconv.Atoi(res[2]) // convert string to int
	if err != nil {
		return -1, err
	}

	return index, nil
}
