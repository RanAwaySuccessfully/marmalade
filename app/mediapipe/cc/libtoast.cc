#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "libtoast.h"
#include "./mediapipe/mediapipe/tasks/c/core/base_options.h"
#include "./mediapipe/mediapipe/tasks/c/vision/core/image.h"
#include "./mediapipe/mediapipe/tasks/c/vision/core/image_processing_options.h"
#include "./mediapipe/mediapipe/tasks/c/vision/face_landmarker/face_landmarker.h"
#include "./mediapipe/mediapipe/tasks/c/vision/hand_landmarker/hand_landmarker.h"

void face_landmarker_callback(MpStatus status, const struct FaceLandmarkerResult* result, MpImagePtr image, int64_t timestamp_ms) {
  face_landmarker_callback_external((struct FaceLandmarkerResult*)result, (int)status, timestamp_ms);
  MpFaceLandmarkerCloseResult((struct FaceLandmarkerResult*)result);
}

void hand_landmarker_callback(MpStatus status, const struct HandLandmarkerResult* result, MpImagePtr image, int64_t timestamp_ms) {
  hand_landmarker_callback_external((struct HandLandmarkerResult*)result, (int)status, timestamp_ms);
  MpHandLandmarkerCloseResult((struct HandLandmarkerResult*)result);
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

uint32_t hand_landmarker_handedness_count(struct HandLandmarkerResult* result, uint32_t hand) {
  return result->handedness[hand].categories_count;
}

uint32_t hand_landmarker_landmark_count(struct HandLandmarkerResult* result, uint32_t hand) {
  return result->hand_landmarks[hand].landmarks_count;
}

uint32_t hand_landmarker_world_landmark_count(struct HandLandmarkerResult* result, uint32_t hand) {
  return result->hand_world_landmarks[hand].landmarks_count;
}

struct Category hand_landmarker_handedness(struct HandLandmarkerResult* result, uint32_t hand, uint32_t index) {
  return result->handedness[hand].categories[index];
}

struct NormalizedLandmark hand_landmarker_landmark(struct HandLandmarkerResult* result, uint32_t hand, uint32_t index) {
  return result->hand_landmarks[hand].landmarks[index];
}

struct Landmark hand_landmarker_world_landmark(struct HandLandmarkerResult* result, uint32_t hand, uint32_t index) {
  return result->hand_world_landmarks[hand].landmarks[index];
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

void* face_landmarker_start(char* face_model_path, int delegate_opt) {
	struct FaceLandmarkerOptions face_landmarker_options;
  face_landmarker_options.base_options.model_asset_buffer = NULL;
  face_landmarker_options.base_options.model_asset_buffer_count = 0;
  face_landmarker_options.base_options.model_asset_path = face_model_path;

  if (delegate_opt) {
    face_landmarker_options.base_options.delegate = (Delegate)delegate_opt;
  } else {
    face_landmarker_options.base_options.delegate = (Delegate)CPU;
  }

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

void* hand_landmarker_start(char* hand_model_path, int delegate_opt) {
	struct HandLandmarkerOptions hand_landmarker_options;
  hand_landmarker_options.base_options.model_asset_buffer = NULL;
  hand_landmarker_options.base_options.model_asset_buffer_count = 0;
  hand_landmarker_options.base_options.model_asset_path = hand_model_path;

  if (delegate_opt) {
    hand_landmarker_options.base_options.delegate = (Delegate)delegate_opt;
  } else {
    hand_landmarker_options.base_options.delegate = (Delegate)CPU;
  }

  hand_landmarker_options.running_mode = LIVE_STREAM;
  hand_landmarker_options.num_hands = 2;
  hand_landmarker_options.result_callback = hand_landmarker_callback;

  MpHandLandmarkerPtr hand_landmarker = NULL;
  char* error_msg = NULL;

	MpStatus status = MpHandLandmarkerCreate(&hand_landmarker_options, &hand_landmarker, &error_msg);
  if (status != 0) {
    save_last_error(error_msg);
    return NULL;
  }

  return hand_landmarker;
}

int mediapipe_detect(void* face_ptr, void* hand_ptr, void* imgdata_ptr, int imgdata_size, int width, int height, int64_t timestamp) {
  struct ImageProcessingOptions options;
  options.has_region_of_interest = false;
  options.rotation_degrees = 0;

  MpImagePtr image;
  char* error_msg;

  MpStatus status = MpImageCreateFromUint8Data(kMpImageFormatSrgb, width, height, (const uint8_t*)imgdata_ptr, imgdata_size, &image, &error_msg);
  if (status != 0) {
    save_last_error(error_msg);
    return -1;
  }

  if (face_ptr != NULL) {
    MpFaceLandmarkerPtr face_landmarker = (MpFaceLandmarkerPtr)face_ptr;
    status = MpFaceLandmarkerDetectAsync(face_landmarker, image, &options, timestamp, &error_msg);
    if (status != 0) {
      save_last_error(error_msg);
      return -1;
    }
  }

  if (hand_ptr != NULL) {
    MpHandLandmarkerPtr hand_landmarker = (MpHandLandmarkerPtr)hand_ptr;
    status = MpHandLandmarkerDetectAsync(hand_landmarker, image, &options, timestamp, &error_msg);
    if (status != 0) {
      save_last_error(error_msg);
      return -1;
    }
  }

  MpImageFree(image);
  return 0;
}

int mediapipe_stop(void* face_ptr, void* hand_ptr) {
  MpStatus status;
  char* error_msg;

  if (face_ptr != NULL) {
    MpFaceLandmarkerPtr face_landmarker = (MpFaceLandmarkerPtr)face_ptr;

    status = MpFaceLandmarkerClose(face_landmarker, &error_msg);
    if (status != 0) {
      save_last_error(error_msg);
      return -1;
    }
  }

  if (hand_ptr != NULL) {
    MpHandLandmarkerPtr hand_landmarker = (MpHandLandmarkerPtr)hand_ptr;

    status = MpHandLandmarkerClose(hand_landmarker, &error_msg);
    if (status != 0) {
      save_last_error(error_msg);
      return -1;
    }
  }

  return (int)status;
}