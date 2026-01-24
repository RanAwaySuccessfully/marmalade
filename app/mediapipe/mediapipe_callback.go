package main

/*
#cgo CFLAGS: -I./cc/ -I./cc/mediapipe/
#cgo LDFLAGS: -L./cc/ -ltoast -lmediapipe

#include <libtoast.h>
*/
import "C"
import "marmalade/internal/server"

func ConvertCategory(mp_category *C.struct_Category) server.Category {
	return server.Category{
		Index:        int(mp_category.index),
		Score:        float32(mp_category.score),
		CategoryName: C.GoString(mp_category.category_name),
		DisplayName:  C.GoString(mp_category.display_name),
	}
}

func ConvertNormalizedLandmark(mp_landmark *C.struct_NormalizedLandmark) server.Landmark {
	return server.Landmark{
		X:             float32(mp_landmark.x),
		Y:             float32(mp_landmark.y),
		Z:             float32(mp_landmark.z),
		HasVisibility: bool(mp_landmark.has_visibility),
		Visibility:    float32(mp_landmark.visibility),
		HasPresence:   bool(mp_landmark.has_presence),
		Presence:      float32(mp_landmark.presence),
		Name:          C.GoString(mp_landmark.name),
	}
}

func ConvertLandmark(mp_landmark *C.struct_Landmark) server.Landmark {
	return server.Landmark{
		X:             float32(mp_landmark.x),
		Y:             float32(mp_landmark.y),
		Z:             float32(mp_landmark.z),
		HasVisibility: bool(mp_landmark.has_visibility),
		Visibility:    float32(mp_landmark.visibility),
		HasPresence:   bool(mp_landmark.has_presence),
		Presence:      float32(mp_landmark.presence),
		Name:          C.GoString(mp_landmark.name),
	}
}

//export face_landmarker_callback_external
func face_landmarker_callback_external(mp_result *C.struct_FaceLandmarkerResult, status C.int, timestamp C.long) {
	result := server.FaceTracking{}

	// Convert between C data types and structs to Go

	if mp_result.face_blendshapes_count != 0 {
		result.Blendshapes = make([]server.Category, 0, int(mp_result.face_blendshapes.categories_count))

		for i := 0; i < int(mp_result.face_blendshapes.categories_count); i++ {
			mp_blendshape := C.face_landmarker_blendshape(mp_result, C.uint(i))
			blendshape := ConvertCategory(&mp_blendshape)
			result.Blendshapes = append(result.Blendshapes, blendshape)
		}
	}

	if mp_result.face_landmarks_count != 0 {
		result.Landmarks = make([]server.Landmark, 0, int(mp_result.face_landmarks.landmarks_count))

		for i := 0; i < int(mp_result.face_landmarks.landmarks_count); i++ {
			mp_landmark := C.face_landmarker_landmark(mp_result, C.uint(i))
			landmark := ConvertNormalizedLandmark(&mp_landmark)
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
			value := C.face_landmarker_matrix_data(&mp_matrix, C.uint(j))
			matrix.Data = append(matrix.Data, float32(value))
		}

		result.Matrixes = append(result.Matrixes, matrix)
	}

	payload := server.TrackingData{FaceData: result}
	payload.Status = int(status)
	payload.Timestamp = int(timestamp)
	payload.Type = server.FaceTrackingType

	ipc.sender(server.FaceTrackingType, payload)
}

//export hand_landmarker_callback_external
func hand_landmarker_callback_external(mp_result *C.struct_HandLandmarkerResult, status C.int, timestamp C.long) {
	result := server.HandTracking{}
	result.Hand = make([]server.Hand, 2)

	// Convert between C data types and structs to Go

	if mp_result.handedness_count != 0 {

		for i := 0; i < int(mp_result.handedness_count); i++ {
			count := C.hand_landmarker_handedness_count(mp_result, C.uint(i))

			handedess_slice := make([]server.Category, 0, int(count))

			for j := 0; j < int(count); j++ {
				mp_handedness := C.hand_landmarker_handedness(mp_result, C.uint(i), C.uint(j))
				handedness := ConvertCategory(&mp_handedness)
				handedess_slice = append(handedess_slice, handedness)
			}

			result.Hand[i].Handedness = handedess_slice
		}
	}

	if mp_result.hand_landmarks_count != 0 {

		for i := 0; i < int(mp_result.hand_landmarks_count); i++ {
			count := C.hand_landmarker_landmark_count(mp_result, C.uint(i))

			landmark_slice := make([]server.Landmark, 0, int(count))

			for j := 0; j < int(count); j++ {
				mp_landmark := C.hand_landmarker_landmark(mp_result, C.uint(i), C.uint(j))
				landmark := ConvertNormalizedLandmark(&mp_landmark)
				landmark_slice = append(landmark_slice, landmark)
			}

			result.Hand[i].Landmarks = landmark_slice
		}
	}

	if mp_result.hand_world_landmarks_count != 0 {

		for i := 0; i < int(mp_result.hand_world_landmarks_count); i++ {
			count := C.hand_landmarker_world_landmark_count(mp_result, C.uint(i))

			landmark_slice := make([]server.Landmark, 0, int(count))

			for j := 0; j < int(count); j++ {
				mp_landmark := C.hand_landmarker_world_landmark(mp_result, C.uint(i), C.uint(j))
				landmark := ConvertLandmark(&mp_landmark)
				landmark_slice = append(landmark_slice, landmark)
			}

			result.Hand[i].WorldLandmarks = landmark_slice
		}
	}

	payload := server.TrackingData{HandData: result}
	payload.Status = int(status)
	payload.Timestamp = int(timestamp)
	payload.Type = server.HandTrackingType

	ipc.sender(server.HandTrackingType, payload)
}
