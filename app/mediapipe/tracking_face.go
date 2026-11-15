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

type FaceLandmarker struct {
	landmarker C.MpFaceLandmarkerPtr
	model_path *C.char
}

func (face *FaceLandmarker) Enabled() bool {
	if server.Config.ModelFace == "" {
		return false
	}

	if server.Config.VTSApi.Enabled {
		return true
	}

	if server.Config.VTSPlugin.Enabled && server.Config.VTSPlugin.UseFace {
		return true
	}

	if server.Config.VMCApi.Enabled && server.Config.VMCApi.UseFace {
		return true
	}

	if server.Config.VRChatOSC.Enabled && server.Config.VRChatOSC.UseFace {
		return true
	}

	return false
}

func (face *FaceLandmarker) Create(model_path string, delegate_opt uint32, confidence []float32) error {
	face.model_path = C.CString(model_path)

	options := C.struct_MpFaceLandmarkerOptions{
		base_options: C.struct_MpBaseOptions{
			model_asset_path: face.model_path,
			delegate:         delegate_opt,
		},
		running_mode:                          C.MP_RUNNING_MODE_LIVE_STREAM,
		output_face_blendshapes:               true,
		output_facial_transformation_matrixes: true,
		num_faces:                             1,
		min_face_detection_confidence:         C.float(confidence[0]),
		min_face_presence_confidence:          C.float(confidence[1]),
		min_tracking_confidence:               C.float(confidence[2]),
	}

	set_base_options(&options.base_options)

	var error_msg *C.char
	status := C.MpFaceLandmarkerCreate(&options, &face.landmarker, &error_msg)
	if status != 0 {
		err := C.GoString(error_msg)
		return errors.New(err)
	}

	C.face_landmarker_callback_set(&options)

	return nil
}

func (face *FaceLandmarker) Detect(image C.MpImagePtr, timestamp int64) error {
	options := C.MpImageProcessingOptions{
		has_region_of_interest: 0, // false
		rotation_degrees:       0,
	}

	var error_msg *C.char
	status := C.MpFaceLandmarkerDetectAsync(face.landmarker, image, &options, C.int64_t(timestamp), &error_msg)
	if status != 0 {
		err := C.GoString(error_msg)
		return errors.New(err)
	}

	return nil
}

func (face *FaceLandmarker) Close() error {
	var error_msg *C.char
	status := C.MpFaceLandmarkerClose(face.landmarker, &error_msg)
	if status != 0 {
		err := C.GoString(error_msg)
		return errors.New(err)
	}

	if face.model_path != nil {
		C.free(unsafe.Pointer(face.model_path))
	}

	return nil
}

//export face_landmarker_callback_external
func face_landmarker_callback_external(mp_result *C.struct_MpFaceLandmarkerResult, status C.int, timestamp C.long) {
	result := server.FaceTracking{}

	// Convert between C data types and structs to Go

	if mp_result.face_blendshapes_count != 0 {
		result.Blendshapes = make([]server.Category, 0, int(mp_result.face_blendshapes.categories_count))

		for i := 0; i < int(mp_result.face_blendshapes.categories_count); i++ {
			mp_blendshape := C.face_landmarker_blendshape(mp_result, C.uint(i))
			blendshape := ConvertCategory(mp_blendshape)
			result.Blendshapes = append(result.Blendshapes, blendshape)
		}
	}

	if mp_result.face_landmarks_count != 0 {
		result.Landmarks = make([]server.Landmark, 0, int(mp_result.face_landmarks.landmarks_count))

		for i := 0; i < int(mp_result.face_landmarks.landmarks_count); i++ {
			mp_landmark := C.face_landmarker_landmark(mp_result, C.uint(i))
			landmark := ConvertNormalizedLandmark(mp_landmark)
			result.Landmarks = append(result.Landmarks, landmark)
		}
	}

	result.Matrixes = make([]server.Matrix, 0, int(mp_result.facial_transformation_matrixes_count))

	for i := 0; i < int(mp_result.facial_transformation_matrixes_count); i++ {
		mp_matrix := C.face_landmarker_matrix(mp_result, C.uint(i))

		matrix := server.Matrix{
			Rows: uint32(mp_matrix.rows),
			Cols: uint32(mp_matrix.cols),
		}

		length := matrix.Rows * matrix.Cols
		matrix.Data = make([]float32, 0, length)

		for j := uint32(0); j < length; j++ {
			value := C.face_landmarker_matrix_data(mp_matrix, C.uint(j))
			matrix.Data = append(matrix.Data, float32(*value))
		}

		result.Matrixes = append(result.Matrixes, matrix)
	}

	payload := server.TrackingData{FaceData: result}
	payload.Status = int(status)
	payload.Timestamp = int(timestamp)
	payload.Type = server.FaceTrackingType

	ipc.sender(payload)
}
