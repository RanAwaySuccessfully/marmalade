package main

/*
#cgo CFLAGS: -I${SRCDIR}/cc/mediapipe/
#cgo LDFLAGS: -L${SRCDIR}/../../lib/ -lmediapipe -l:libopencv_core.so.414 -l:libopencv_features2d.so.414 -l:libopencv_imgproc.so.414 -Wl,-rpath,./lib

#include "mediapipe.h"
*/
import "C"
import (
	"errors"
	"marmalade/internal/devices"
	"marmalade/internal/errs"
	"marmalade/internal/server"
	"marmalade/internal/webcam"
	"math"
)

type WebCamAPI interface {
	Run(chan error, webcam.RunCallback)
	End()
}

type Landmarker interface {
	Check() bool
	Create(string, uint32, []float32) error
	Detect(C.MpImagePtr, int64) error
	Close(error)
}

type MediaPipe struct {
	capture WebCamAPI
	face    FaceLandmarker
	hand    HandLandmarker
	pose    PoseLandmarker
}

func (mp *MediaPipe) start() error {
	server.Config.Read()

	fourcc, err := devices.StringToPixFmt(server.Config.Format)
	if err != nil {
		return errs.CreateError("converting FourCC", err)
	}

	isFlatpak := true
	if isFlatpak {
		mp.capture, err = webcam.NewGStreamer(0)
	} else {
		mp.capture, err = webcam.NewGOV4L(fourcc)
	}

	delegate := 0
	if server.Config.HwAccel.DelegateMP != 0 {
		delegate = server.Config.HwAccel.DelegateMP
	}

	if mp.face.Enabled() {
		confidences := []float32{-1, -1, -1}
		err := mp.face.Create(server.Config.ModelFace, uint32(delegate), confidences)
		if err != nil {
			return errs.CreateError("creating MediaPipe FaceLandmarker", err)
		}
	}

	if mp.hand.Enabled() {
		confidences := []float32{-1, -1, -1}
		err := mp.hand.Create(server.Config.ModelHand, uint32(delegate), confidences)
		if err != nil {
			return errs.CreateError("creating MediaPipe HandLandmarker", err)
		}
	}

	if mp.pose.Enabled() {
		confidences := []float32{-1, -1, -1}
		err := mp.pose.Create(server.Config.ModelPose, uint32(delegate), confidences)
		if err != nil {
			return errs.CreateError("creating MediaPipe PoseLandmarker", err)
		}
	}

	return nil
}

func (mp *MediaPipe) run(err_ch chan error) {
	mp.capture.Run(err_ch, mp.detect)
	close(err_ch)
}

func (mp *MediaPipe) detect(data []byte, width uint32, height uint32, timestamp int64) error {
	//start := time.Now().UnixMilli()

	data_size := len(data)
	data_ptr := C.CBytes(data)

	var image C.MpImagePtr
	var error_msg *C.char
	status := C.MpImageCreateFromUint8Data(C.kMpImageFormatSrgb, C.int(width), C.int(height), (*C.uint8_t)(data_ptr), C.int(data_size), &image, &error_msg)
	C.free(data_ptr)

	if status < 0 {
		err := errors.New(C.GoString(error_msg))
		return errs.CreateError("creating MediaPipe image from webcam frame", err)
	}

	if mp.face.landmarker != nil {
		err := mp.face.Detect(image, timestamp)
		if err != nil {
			C.MpImageFree(image)
			return errs.CreateError("running FaceLandmarker detection", err)
		}
	}

	if mp.hand.landmarker != nil {
		err := mp.hand.Detect(image, timestamp)
		if err != nil {
			C.MpImageFree(image)
			return errs.CreateError("running HandLandmarker detection", err)
		}
	}

	if mp.pose.landmarker != nil {
		err := mp.pose.Detect(image, timestamp)
		if err != nil {
			C.MpImageFree(image)
			return errs.CreateError("running PoseLandmarker detection", err)
		}
	}

	/*
		end := time.Now().UnixMilli()
		diff := end - start
		fmt.Printf("%d\n", diff)
	*/

	C.MpImageFree(image)
	return nil
}

func (mp *MediaPipe) stop() error {
	if mp.capture != nil {
		mp.capture.End()
	}

	if mp.face.landmarker != nil {
		err := mp.face.Close()
		if err != nil {
			return errs.CreateError("stopping MediaPipe FaceLandmarker", err)
		}
	}

	if mp.hand.landmarker != nil {
		err := mp.hand.Close()
		if err != nil {
			return errs.CreateError("stopping MediaPipe HandLandmarker", err)
		}
	}

	if mp.pose.landmarker != nil {
		err := mp.pose.Close()
		if err != nil {
			return errs.CreateError("stopping MediaPipe PoseLandmarker", err)
		}
	}

	return nil
}

func set_base_options(options *C.struct_MpBaseOptions) {
	options.model_asset_buffer = nil
	options.model_asset_buffer_count = 0
	options.host_environment = C.MP_HOST_ENVIRONMENT_UNKNOWN
	options.host_system = C.MP_HOST_SYSTEM_LINUX
	options.host_version = nil
	options.ca_bundle_path = nil
}

func removeNaN(value float32) float32 {
	if math.IsNaN(float64(value)) {
		return 0 // should this be -1?
	} else {
		return value
	}
}

func ConvertCategory(mp_category *C.struct_MpCategory) server.Category {
	return server.Category{
		Index:        int(mp_category.index),
		Score:        removeNaN(float32(mp_category.score)),
		CategoryName: C.GoString(mp_category.category_name),
		DisplayName:  C.GoString(mp_category.display_name),
	}
}

func ConvertNormalizedLandmark(mp_landmark *C.struct_MpNormalizedLandmark) server.Landmark {
	return server.Landmark{
		X:             removeNaN(float32(mp_landmark.x)),
		Y:             removeNaN(float32(mp_landmark.y)),
		Z:             removeNaN(float32(mp_landmark.z)),
		HasVisibility: bool(mp_landmark.has_visibility),
		Visibility:    removeNaN(float32(mp_landmark.visibility)),
		HasPresence:   bool(mp_landmark.has_presence),
		Presence:      removeNaN(float32(mp_landmark.presence)),
		Name:          C.GoString(mp_landmark.name),
	}
}

func ConvertLandmark(mp_landmark *C.struct_MpLandmark) server.Landmark {
	return server.Landmark{
		X:             removeNaN(float32(mp_landmark.x)),
		Y:             removeNaN(float32(mp_landmark.y)),
		Z:             removeNaN(float32(mp_landmark.z)),
		HasVisibility: bool(mp_landmark.has_visibility),
		Visibility:    removeNaN(float32(mp_landmark.visibility)),
		HasPresence:   bool(mp_landmark.has_presence),
		Presence:      removeNaN(float32(mp_landmark.presence)),
		Name:          C.GoString(mp_landmark.name),
	}
}
