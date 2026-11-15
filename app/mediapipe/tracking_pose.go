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

type PoseLandmarker struct {
	landmarker C.MpPoseLandmarkerPtr
	model_path *C.char
}

func (pose *PoseLandmarker) Enabled() bool {
	if server.Config.ModelPose == "" {
		return false
	}

	if server.Config.VTSPlugin.Enabled && server.Config.VTSPlugin.UsePose {
		return true
	}

	if server.Config.VMCApi.Enabled && server.Config.VMCApi.UsePose {
		return true
	}

	if server.Config.VRChatOSC.Enabled && server.Config.VRChatOSC.UsePose {
		return true
	}

	return false
}

func (pose *PoseLandmarker) Create(model_path string, delegate_opt uint32, confidence []float32) error {
	pose.model_path = C.CString(model_path)

	options := C.struct_MpPoseLandmarkerOptions{
		base_options: C.struct_MpBaseOptions{
			model_asset_path: pose.model_path,
			delegate:         delegate_opt,
		},
		running_mode: C.MP_RUNNING_MODE_LIVE_STREAM,
		//output_segmentation_masks: true,
		num_poses:                     1,
		min_pose_detection_confidence: C.float(confidence[0]),
		min_pose_presence_confidence:  C.float(confidence[1]),
		min_tracking_confidence:       C.float(confidence[2]),
	}

	set_base_options(&options.base_options)

	var error_msg *C.char
	status := C.MpPoseLandmarkerCreate(&options, &pose.landmarker, &error_msg)
	if status != 0 {
		err := C.GoString(error_msg)
		return errors.New(err)
	}

	C.pose_landmarker_callback_set(&options)

	return nil
}

func (pose *PoseLandmarker) Detect(image C.MpImagePtr, timestamp int64) error {
	options := C.MpImageProcessingOptions{
		has_region_of_interest: 0, // false
		rotation_degrees:       0,
	}

	var error_msg *C.char
	status := C.MpPoseLandmarkerDetectAsync(pose.landmarker, image, &options, C.int64_t(timestamp), &error_msg)
	if status != 0 {
		err := C.GoString(error_msg)
		return errors.New(err)
	}

	return nil
}

func (pose *PoseLandmarker) Close() error {
	var error_msg *C.char
	status := C.MpPoseLandmarkerClose(pose.landmarker, &error_msg)
	if status != 0 {
		err := C.GoString(error_msg)
		return errors.New(err)
	}

	if pose.model_path != nil {
		C.free(unsafe.Pointer(pose.model_path))
	}

	return nil
}

//export pose_landmarker_callback_external
func pose_landmarker_callback_external(mp_result *C.struct_MpPoseLandmarkerResult, status C.int, timestamp C.long) {
	result := server.PoseTracking{}

	// Convert between C data types and structs to Go

	if mp_result.pose_landmarks_count != 0 {
		count := C.pose_landmarker_landmark_count(mp_result)
		int_count := int(*count)

		landmark_slice := make([]server.Landmark, 0, int_count)

		for j := 0; j < int_count; j++ {
			mp_landmark := C.pose_landmarker_landmark(mp_result, C.uint(j))
			landmark := ConvertNormalizedLandmark(mp_landmark)
			landmark_slice = append(landmark_slice, landmark)
		}

		result.Landmarks = landmark_slice
	}

	if mp_result.pose_world_landmarks_count != 0 {
		count := C.pose_landmarker_world_landmark_count(mp_result)
		int_count := int(*count)

		landmark_slice := make([]server.Landmark, 0, int(*count))

		for j := 0; j < int_count; j++ {
			mp_landmark := C.pose_landmarker_world_landmark(mp_result, C.uint(j))
			landmark := ConvertLandmark(mp_landmark)
			landmark_slice = append(landmark_slice, landmark)
		}

		result.WorldLandmarks = landmark_slice
	}

	payload := server.TrackingData{PoseData: result}
	payload.Status = int(status)
	payload.Timestamp = int(timestamp)
	payload.Type = server.PoseTrackingType

	ipc.sender(payload)
}
