package main

/*
#cgo CFLAGS: -I${SRCDIR}/cc/ -I${SRCDIR}/cc/mediapipe/
#cgo LDFLAGS: ${SRCDIR}/cc/libtoast.a -L${SRCDIR}/../../lib/ -lmediapipe -l:libopencv_core.so.414 -l:libopencv_features2d.so.414 -l:libopencv_imgproc.so.414 -Wl,-rpath,./lib

#include <libtoast.h>
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"marmalade/internal/devices"
	"marmalade/internal/errs"
	"marmalade/internal/server"
	"marmalade/internal/webcam"
	"unsafe"
)

type WebCamAPI interface {
	Run(chan error, webcam.RunCallback)
	End()
}

type MediaPipe struct {
	api        WebCamAPI
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

	isFlatpak := true
	if isFlatpak {
		mp.api, err = webcam.NewGStreamer(0)
	} else {
		mp.api, err = webcam.NewGOV4L(fourcc)
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

func (mp *MediaPipe) run(err_ch chan error) {
	mp.api.Run(err_ch, mp.detect)
	close(err_ch)
}

func (mp *MediaPipe) detect(data []byte, width uint32, height uint32, timestamp int64) error {
	//start := time.Now().UnixMilli()

	data_size := len(data)
	data_ptr := C.CBytes(data)
	ret := C.int(0)

	img_ptr := C.mediapipe_create_img(&ret, data_ptr, C.int(data_size), C.int(width), C.int(height))
	C.free(data_ptr)
	if ret < 0 {
		err := mediapipe_get_error()
		return errs.CreateError("creating MediaPipe image from webcam frame", err)
	}

	if mp.facem_lm != nil {
		ret = C.mediapipe_lm_face_detect(mp.facem_lm, img_ptr, C.long(timestamp))
		if ret < 0 {
			C.mediapipe_free_img(img_ptr)
			err := mediapipe_get_error()
			return errs.CreateError("running FaceLandmarker detection", err)
		}
	}

	if mp.handm_lm != nil {
		ret = C.mediapipe_lm_hand_detect(mp.handm_lm, img_ptr, C.long(timestamp))
		if ret < 0 {
			C.mediapipe_free_img(img_ptr)
			err := mediapipe_get_error()
			return errs.CreateError("running HandLandmarker detection", err)
		}
	}

	if mp.posem_lm != nil {
		ret = C.mediapipe_lm_pose_detect(mp.posem_lm, img_ptr, C.long(timestamp))
		if ret < 0 {
			C.mediapipe_free_img(img_ptr)
			err := mediapipe_get_error()
			return errs.CreateError("running PoseLandmarker detection", err)
		}
	}

	/*
		end := time.Now().UnixMilli()
		diff := end - start
		fmt.Printf("%d\n", diff)
	*/

	C.mediapipe_free_img(img_ptr)
	return nil
}

func (mp *MediaPipe) stop() error {
	if mp.api != nil {
		mp.api.End()
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

func mediapipe_get_error() error {
	error_str := C.GoString(C.mediapipe_read_error())
	C.mediapipe_free_error()
	return errors.New(error_str)
}
