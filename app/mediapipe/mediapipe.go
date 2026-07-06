package main

/*
#cgo CFLAGS: -I${SRCDIR}/cc/ -I${SRCDIR}/cc/mediapipe/
#cgo LDFLAGS: ${SRCDIR}/cc/libtoast.a -L${SRCDIR}/../../lib/ -lmediapipe -l:libopencv_core.so.414 -l:libopencv_features2d.so.414 -l:libopencv_imgproc.so.414 -Wl,-rpath,./lib

#include <libtoast.h>
#include <stdlib.h>
*/
import "C"
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"marmalade/internal/devices"
	"marmalade/internal/errs"
	"marmalade/internal/server"
	"unsafe"

	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
)

type Converter interface {
	Convert(input []byte, output io.Writer) error
	End()
}

type MediaPipe struct {
	webcam     *device.Device
	converter  Converter
	facem_lm   unsafe.Pointer
	facem_path *C.char
	handm_lm   unsafe.Pointer
	handm_path *C.char
	posem_lm   unsafe.Pointer
	posem_path *C.char
}

func (mp *MediaPipe) start() error {
	server.Config.Read()

	fourcc, err := devices.StringToPixFmt(server.Config.Format)
	if err != nil {
		return errs.CreateError("converting FourCC", err)
	}

	pix_format := v4l2.PixFormat{
		Width:       uint32(server.Config.Width),
		Height:      uint32(server.Config.Height),
		PixelFormat: uint32(fourcc),
	}

	device_path := fmt.Sprintf("/dev/video%d", int(server.Config.Camera))

	mp.webcam, err = device.Open(
		device_path,
		device.WithBufferSize(1),
		device.WithPixFormat(pix_format),
		device.WithFPS(uint32(server.Config.FPS)),
	)

	if err != nil {
		return errs.CreateError("opening camera device", err)
	}

	//mp.check_supported_settings()

	mp.webcam.GetFrames()
	err = mp.webcam.Start(context.Background())
	if err != nil {
		return errs.CreateError("starting camera feed", err)
	}

	// "RGB3" is the format MediaPipe needs
	if fourcc != v4l2.PixelFmtRGB24 {
		pixfmt_map, err := find_mapping(fourcc)

		if pixfmt_map == nil && !server.Config.HwAccel.ForceFFMPEG {
			mp.converter, err = NewV4LConvert(mp.webcam)
			if err != nil {
				return err
			}

		} else {
			mp.converter, err = NewFFMPEG(pixfmt_map)
			if err != nil {
				return err
			}
		}
	}

	delegate := 0
	if server.Config.HwAccel.DelegateMP != 0 {
		delegate = server.Config.HwAccel.DelegateMP
	}

	anyFaceApi := server.Config.VTSApi.Enabled ||
		(server.Config.VTSPlugin.Enabled && server.Config.VTSPlugin.UseFace) ||
		(server.Config.VMCApi.Enabled && server.Config.VMCApi.UseFace) ||
		(server.Config.VRChatOSC.Enabled && server.Config.VRChatOSC.UseHand)

	anyHandApi := (server.Config.VMCApi.Enabled && server.Config.VMCApi.UseHand) ||
		(server.Config.VTSPlugin.Enabled && server.Config.VTSPlugin.UseHand) ||
		(server.Config.VRChatOSC.Enabled && server.Config.VRChatOSC.UseHand)

	anyPoseApi := (server.Config.VMCApi.Enabled && server.Config.VMCApi.UsePose) ||
		(server.Config.VTSPlugin.Enabled && server.Config.VTSPlugin.UsePose) ||
		(server.Config.VRChatOSC.Enabled && server.Config.VRChatOSC.UsePose)

	if (server.Config.ModelFace != "") && anyFaceApi {
		confidences := [3]C.float{-1, -1, -1}

		mp.facem_path = C.CString(server.Config.ModelFace)
		mp.facem_lm = C.mediapipe_lm_face_start(mp.facem_path, C.int(delegate), &confidences[0])
		if mp.facem_lm == nil {
			err := mediapipe_get_error()
			return errs.CreateError("creating MediaPipe FaceLandmarker", err)
		}
	}

	if (server.Config.ModelHand != "") && anyHandApi {
		confidences := [3]C.float{-1, -1, -1}

		mp.handm_path = C.CString(server.Config.ModelHand)
		mp.handm_lm = C.mediapipe_lm_hand_start(mp.handm_path, C.int(delegate), &confidences[0])
		if mp.handm_lm == nil {
			err := mediapipe_get_error()
			return errs.CreateError("creating MediaPipe HandLandmarker", err)
		}
	}

	if (server.Config.ModelPose != "") && anyPoseApi {
		confidences := [3]C.float{-1, -1, -1}

		mp.posem_path = C.CString(server.Config.ModelPose)
		mp.posem_lm = C.mediapipe_lm_pose_start(mp.posem_path, C.int(delegate), &confidences[0])
		if mp.posem_lm == nil {
			err := mediapipe_get_error()
			return errs.CreateError("creating MediaPipe PoseLandmarker", err)
		}
	}

	return nil
}

func (mp *MediaPipe) detect(err_ch chan error) {

	for frame := range mp.webcam.GetFrames() {
		//start := time.Now().UnixMilli()

		var srgb_frame []byte
		var err error

		if mp.converter == nil {
			srgb_frame = frame.Data
		} else {
			output := bytes.Buffer{}
			err = mp.converter.Convert(frame.Data, &output) // uses about 1% CPU and 40MB of RAM...pretty good!
			if err != nil {
				err_ch <- err // errs.CreateError() is already being called over at ffmpeg.go
				break
			}

			srgb_frame = output.Bytes()
		}

		format, err := mp.webcam.GetPixFormat()
		if err != nil {
			err_ch <- errs.CreateError("reading webcam current format", err)
			break
		}

		data_size := len(srgb_frame)
		data_ptr := C.CBytes(srgb_frame)
		ret := C.int(0)

		img_ptr := C.mediapipe_create_img(&ret, data_ptr, C.int(data_size), C.int(format.Width), C.int(format.Height))
		C.free(data_ptr)
		if ret < 0 {
			err := mediapipe_get_error()
			err_ch <- errs.CreateError("creating MediaPipe image from webcam frame", err)
			break
		}

		timestamp := frame.Timestamp.UnixMilli()

		if mp.facem_lm != nil {
			ret = C.mediapipe_lm_face_detect(mp.facem_lm, img_ptr, C.long(timestamp))
			if ret < 0 {
				C.mediapipe_free_img(img_ptr)
				err := mediapipe_get_error()
				err_ch <- errs.CreateError("running FaceLandmarker detection", err)
				break
			}
		}

		if mp.handm_lm != nil {
			ret = C.mediapipe_lm_hand_detect(mp.handm_lm, img_ptr, C.long(timestamp))
			if ret < 0 {
				C.mediapipe_free_img(img_ptr)
				err := mediapipe_get_error()
				err_ch <- errs.CreateError("running HandLandmarker detection", err)
				break
			}
		}

		if mp.posem_lm != nil {
			ret = C.mediapipe_lm_pose_detect(mp.posem_lm, img_ptr, C.long(timestamp))
			if ret < 0 {
				C.mediapipe_free_img(img_ptr)
				err := mediapipe_get_error()
				err_ch <- errs.CreateError("running PoseLandmarker detection", err)
				break
			}
		}

		frame.Release()
		C.mediapipe_free_img(img_ptr)

		/*
			end := time.Now().UnixMilli()
			diff := end - start
			fmt.Printf("%d\n", diff)
		*/
	}

	close(err_ch)
}

func (mp *MediaPipe) stop() error {
	if mp.webcam != nil {
		mp.webcam.Close()
	}

	if mp.converter != nil {
		mp.converter.End()
	}

	if mp.facem_lm != nil {
		ret := C.mediapipe_lm_face_stop(mp.facem_lm)
		if ret < 0 {
			err := mediapipe_get_error()
			return errs.CreateError("stopping MediaPipe FaceLandmarker", err)
		}
	}

	if mp.handm_lm != nil {
		ret := C.mediapipe_lm_hand_stop(mp.handm_lm)
		if ret < 0 {
			err := mediapipe_get_error()
			return errs.CreateError("stopping MediaPipe HandLandmarker", err)
		}
	}

	if mp.posem_lm != nil {
		ret := C.mediapipe_lm_pose_stop(mp.posem_lm)
		if ret < 0 {
			err := mediapipe_get_error()
			return errs.CreateError("stopping MediaPipe PoseLandmarker", err)
		}
	}

	if mp.facem_path != nil {
		C.free(unsafe.Pointer(mp.facem_path))
	}

	if mp.handm_path != nil {
		C.free(unsafe.Pointer(mp.handm_path))
	}

	if mp.posem_path != nil {
		C.free(unsafe.Pointer(mp.posem_path))
	}

	return nil
}

func (mp *MediaPipe) check_supported_settings() error {
	fmt_real, err := mp.webcam.GetPixFormat()
	if err != nil {
		return errs.CreateError("reading webcam current format", err)
	}

	fps_real, err := mp.webcam.GetFrameRate()
	if err != nil {
		return errs.CreateError("reading webcam current frame rate", err)
	}

	pixelformat := devices.PixFmtToString(fmt_real.PixelFormat)

	fmt.Printf("[%s] %dx%d@%d\n", pixelformat, fmt_real.Width, fmt_real.Height, fps_real)
	return nil
}

func mediapipe_get_error() error {
	error_str := C.GoString(C.mediapipe_read_error())
	C.mediapipe_free_error()
	return errors.New(error_str)
}
