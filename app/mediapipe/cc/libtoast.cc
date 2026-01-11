#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "libtoast.h"
#include "./mediapipe/mediapipe/tasks/c/vision/core/image.h"
#include "./mediapipe/mediapipe/tasks/c/vision/core/image_processing_options.h"
#include "./mediapipe/mediapipe/tasks/c/vision/face_landmarker/face_landmarker.h"

void face_landmarker_callback(MpStatus status, const struct FaceLandmarkerResult* result, MpImagePtr image, int64_t timestamp_ms) {
  mediapipe_call_facem_result((struct FaceLandmarkerResult*)result, (int)status, timestamp_ms);
  MpFaceLandmarkerCloseResult((struct FaceLandmarkerResult*)result);
}

struct Category face_landmarker_blendshape(struct FaceLandmarkerResult* result, uint32_t index) {
  return result->face_blendshapes[0].categories[index];
}

struct NormalizedLandmark face_landmarker_landmark(struct FaceLandmarkerResult* result, uint32_t index) {
  return result->face_landmarks[0].landmarks[index];
}

struct Matrix face_landmarker_matrix(struct FaceLandmarkerResult* result, uint32_t index) {
  return result->facial_transformation_matrixes[index];
}

float face_landmarker_matrix_data(struct Matrix* matrix, uint32_t index) {
  return matrix->data[index];
}

// ERROR HANDLING

char* mediapipe_last_error = NULL;

void save_last_error(char* error_msg) {
  size_t length = strlen(error_msg);

  mediapipe_last_error = (char*)realloc(mediapipe_last_error, (length + 1) * sizeof(char));
  if (mediapipe_last_error != NULL) {
    strncpy(mediapipe_last_error, error_msg, length);
  }

  MpErrorFree(error_msg);
}

char* mediapipe_read_error() {
  return mediapipe_last_error;
}

void mediapipe_free_error() {
  free(mediapipe_last_error);
  mediapipe_last_error = NULL;
}

// MEDIAPIPE MAIN FUNCTIONS

void* mediapipe_start(char* face_model_path) {
	struct FaceLandmarkerOptions face_landmarker_options;
  face_landmarker_options.base_options.model_asset_buffer = NULL;
  face_landmarker_options.base_options.model_asset_buffer_count = 0;
  face_landmarker_options.base_options.model_asset_path = face_model_path;
  face_landmarker_options.running_mode = LIVE_STREAM;
  face_landmarker_options.output_face_blendshapes = true;
  face_landmarker_options.output_facial_transformation_matrixes = true;
  face_landmarker_options.num_faces = 1;
  face_landmarker_options.result_callback = face_landmarker_callback;

  MpFaceLandmarkerPtr face_landmarker = NULL;
  char* error_msg = NULL;

	MpStatus status = MpFaceLandmarkerCreate(&face_landmarker_options, &face_landmarker, &error_msg);
  if (status != 0) {
    save_last_error(error_msg);
    return NULL;
  }

  return face_landmarker;
}

int mediapipe_detect(void* ctx, void* data_ptr, int data_size, int width, int height, int64_t timestamp) {
  MpFaceLandmarkerPtr face_landmarker = (MpFaceLandmarkerPtr)ctx;
  struct ImageProcessingOptions options;
  options.has_region_of_interest = false;
  options.rotation_degrees = 0;

  MpImagePtr image;
  char* error_msg;

  MpStatus status = MpImageCreateFromUint8Data(kMpImageFormatSrgb, width, height, (const uint8_t*)data_ptr, data_size, &image, &error_msg);
  if (status != 0) {
    save_last_error(error_msg);
    return -1;
  }
  
  status = MpFaceLandmarkerDetectAsync(face_landmarker, image, &options, timestamp, &error_msg);
  if (status != 0) {
    save_last_error(error_msg);
    return -1;
  }

  MpImageFree(image);
  return 0;
}

int mediapipe_stop(void* ctx) {
  if (ctx == NULL) {
    return 0;
  }

  MpFaceLandmarkerPtr face_landmarker = (MpFaceLandmarkerPtr)ctx;

  char* error_msg;
  MpStatus status = MpFaceLandmarkerClose(face_landmarker, &error_msg);
  if (status != 0) {
    save_last_error(error_msg);
    return -1;
  }

  return (int)status;
}