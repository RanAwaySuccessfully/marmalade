#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "libtoast.h"
#include "mediapipe/tasks/c/core/base_options.h"
#include "mediapipe/tasks/c/vision/core/image_processing_options.h"
#include "mediapipe/tasks/c/vision/face_landmarker/face_landmarker.h"

// ARRAY HELPERS

struct Category* face_landmarker_blendshape(struct FaceLandmarkerResult* result, uint32_t index) {
  return &result->face_blendshapes[0].categories[index];
}

struct NormalizedLandmark* face_landmarker_landmark(struct FaceLandmarkerResult* result, uint32_t index) {
  return &result->face_landmarks[0].landmarks[index];
}

struct Matrix* face_landmarker_matrix(struct FaceLandmarkerResult* result, uint32_t index) {
  return &result->facial_transformation_matrixes[index];
}

float* face_landmarker_matrix_data(struct Matrix* matrix, uint32_t index) {
  return &matrix->data[index];
}

// MAIN FUNCTIONS

void mediapipe_lm_face_callback(MpStatus status, const struct FaceLandmarkerResult* result, MpImagePtr image, int64_t timestamp_ms) {
  face_landmarker_callback_external((struct FaceLandmarkerResult*)result, (int)status, timestamp_ms);
  MpFaceLandmarkerCloseResult((struct FaceLandmarkerResult*)result);
}

void* mediapipe_lm_face_start(char* face_model_path, int delegate_opt) {
	struct FaceLandmarkerOptions face_landmarker_options;
  face_landmarker_options.base_options.model_asset_path = face_model_path;
  set_base_options(&face_landmarker_options.base_options);

  if (delegate_opt) {
    face_landmarker_options.base_options.delegate = (Delegate)delegate_opt;
  } else {
    face_landmarker_options.base_options.delegate = (Delegate)CPU;
  }

  face_landmarker_options.running_mode = LIVE_STREAM;
  face_landmarker_options.output_face_blendshapes = true;
  face_landmarker_options.output_facial_transformation_matrixes = true;
  face_landmarker_options.num_faces = 1;
  face_landmarker_options.result_callback = mediapipe_lm_face_callback;

  MpFaceLandmarkerPtr face_landmarker = NULL;
  char* error_msg = NULL;

	MpStatus status = MpFaceLandmarkerCreate(&face_landmarker_options, &face_landmarker, &error_msg);
  if (status != 0) {
    save_last_error(error_msg);
    return NULL;
  }

  return face_landmarker;
}

int mediapipe_lm_face_detect(void* face_landmarker, void* mp_image, int64_t timestamp) {
  struct ImageProcessingOptions options;
  options.has_region_of_interest = false;
  options.rotation_degrees = 0;

  char* error_msg;

  MpStatus status = MpFaceLandmarkerDetectAsync((MpFaceLandmarkerPtr)face_landmarker, (MpImagePtr)mp_image, &options, timestamp, &error_msg);
  if (status != 0) {
    save_last_error(error_msg);
  }

  return (int)status;
}

int mediapipe_lm_face_stop(void* face_landmarker) {
  char* error_msg;

  MpStatus status = MpFaceLandmarkerClose((MpFaceLandmarkerPtr)face_landmarker, &error_msg);
  if (status != 0) {
    save_last_error(error_msg);
  }

  return (int)status;
}