package main

/*
#include "mediapipe.h"
*/
import "C"
import (
	"errors"
	"marmalade/internal/server"
	"unsafe"
)

type HandLandmarker struct {
	landmarker C.MpHandLandmarkerPtr
	model_path *C.char
}

func (hand *HandLandmarker) Enabled() bool {
	if server.Config.ModelHand == "" {
		return false
	}

	if server.Config.VTSPlugin.Enabled && server.Config.VTSPlugin.UseHand {
		return true
	}

	if server.Config.VMCApi.Enabled && server.Config.VMCApi.UseHand {
		return true
	}

	if server.Config.VRChatOSC.Enabled && server.Config.VRChatOSC.UseHand {
		return true
	}

	return false
}

func (hand *HandLandmarker) Create(model_path string, delegate_opt uint32, confidence []float32) error {
	hand.model_path = C.CString(model_path)

	options := C.struct_MpHandLandmarkerOptions{
		base_options: C.struct_MpBaseOptions{
			model_asset_path: hand.model_path,
			delegate:         delegate_opt,
		},
		running_mode:                  C.MP_RUNNING_MODE_LIVE_STREAM,
		num_hands:                     2,
		min_hand_detection_confidence: C.float(confidence[0]),
		min_hand_presence_confidence:  C.float(confidence[1]),
		min_tracking_confidence:       C.float(confidence[2]),
	}

	set_base_options(&options.base_options)

	var error_msg *C.char
	status := C.MpHandLandmarkerCreate(&options, &hand.landmarker, &error_msg)
	if status != 0 {
		err := C.GoString(error_msg)
		return errors.New(err)
	}

	C.hand_landmarker_callback_set(&options)

	return nil
}

func (hand *HandLandmarker) Detect(image C.MpImagePtr, timestamp int64) error {
	options := C.MpImageProcessingOptions{
		has_region_of_interest: 0, // false
		rotation_degrees:       0,
	}

	var error_msg *C.char
	status := C.MpHandLandmarkerDetectAsync(hand.landmarker, image, &options, C.int64_t(timestamp), &error_msg)
	if status != 0 {
		err := C.GoString(error_msg)
		return errors.New(err)
	}

	return nil
}

func (hand *HandLandmarker) Close() error {
	var error_msg *C.char
	status := C.MpHandLandmarkerClose(hand.landmarker, &error_msg)
	if status != 0 {
		err := C.GoString(error_msg)
		return errors.New(err)
	}

	if hand.model_path != nil {
		C.free(unsafe.Pointer(hand.model_path))
	}

	return nil
}

//export hand_landmarker_callback_external
func hand_landmarker_callback_external(mp_result *C.struct_MpHandLandmarkerResult, status C.int, timestamp C.long) {
	result := server.HandTracking{}
	result.Hand = make([]server.Hand, 2)

	// Convert between C data types and structs to Go

	if mp_result.handedness_count != 0 {

		for i := 0; i < int(mp_result.handedness_count); i++ {
			count := C.hand_landmarker_handedness_count(mp_result, C.uint(i))
			int_count := int(*count)

			handedess_slice := make([]server.Category, 0, int_count)

			for j := 0; j < int_count; j++ {
				mp_handedness := C.hand_landmarker_handedness(mp_result, C.uint(i), C.uint(j))
				handedness := ConvertCategory(mp_handedness)
				handedess_slice = append(handedess_slice, handedness)
			}

			result.Hand[i].Handedness = handedess_slice
		}
	}

	if mp_result.hand_landmarks_count != 0 {

		for i := 0; i < int(mp_result.hand_landmarks_count); i++ {
			count := C.hand_landmarker_landmark_count(mp_result, C.uint(i))
			int_count := int(*count)

			landmark_slice := make([]server.Landmark, 0, int_count)

			for j := 0; j < int_count; j++ {
				mp_landmark := C.hand_landmarker_landmark(mp_result, C.uint(i), C.uint(j))
				landmark := ConvertNormalizedLandmark(mp_landmark)
				landmark_slice = append(landmark_slice, landmark)
			}

			result.Hand[i].Landmarks = landmark_slice
		}
	}

	if mp_result.hand_world_landmarks_count != 0 {

		for i := 0; i < int(mp_result.hand_world_landmarks_count); i++ {
			count := C.hand_landmarker_world_landmark_count(mp_result, C.uint(i))
			int_count := int(*count)

			landmark_slice := make([]server.Landmark, 0, int(*count))

			for j := 0; j < int_count; j++ {
				mp_landmark := C.hand_landmarker_world_landmark(mp_result, C.uint(i), C.uint(j))
				landmark := ConvertLandmark(mp_landmark)
				landmark_slice = append(landmark_slice, landmark)
			}

			result.Hand[i].WorldLandmarks = landmark_slice
		}
	}

	payload := server.TrackingData{HandData: result}
	payload.Status = int(status)
	payload.Timestamp = int(timestamp)
	payload.Type = server.HandTrackingType

	ipc.sender(payload)
}
