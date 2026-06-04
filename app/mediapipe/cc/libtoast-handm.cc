#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "libtoast.h"
#include "mediapipe/tasks/c/core/base_options.h"
#include "mediapipe/tasks/c/vision/core/image_processing_options.h"
#include "mediapipe/tasks/c/vision/hand_landmarker/hand_landmarker.h"

// ARRAY HELPERS

uint32_t* hand_landmarker_handedness_count(struct HandLandmarkerResult* result, uint32_t hand) {
  return &result->handedness[hand].categories_count;
}

uint32_t* hand_landmarker_landmark_count(struct HandLandmarkerResult* result, uint32_t hand) {
  return &result->hand_landmarks[hand].landmarks_count;
}

uint32_t* hand_landmarker_world_landmark_count(struct HandLandmarkerResult* result, uint32_t hand) {
  return &result->hand_world_landmarks[hand].landmarks_count;
}

struct Category* hand_landmarker_handedness(struct HandLandmarkerResult* result, uint32_t hand, uint32_t index) {
  return &result->handedness[hand].categories[index];
}

struct NormalizedLandmark* hand_landmarker_landmark(struct HandLandmarkerResult* result, uint32_t hand, uint32_t index) {
  return &result->hand_landmarks[hand].landmarks[index];
}

struct Landmark* hand_landmarker_world_landmark(struct HandLandmarkerResult* result, uint32_t hand, uint32_t index) {
  return &result->hand_world_landmarks[hand].landmarks[index];
}

// MAIN FUNCTIONS

void mediapipe_lm_hand_callback(MpStatus status, const struct HandLandmarkerResult* result, MpImagePtr image, int64_t timestamp_ms) {
  hand_landmarker_callback_external((struct HandLandmarkerResult*)result, (int)status, timestamp_ms);
  MpHandLandmarkerCloseResult((struct HandLandmarkerResult*)result);
}

void* mediapipe_lm_hand_start(char* hand_model_path, int delegate_opt, float confidence[3]) {
	struct HandLandmarkerOptions hand_landmarker_options;
  hand_landmarker_options.base_options.model_asset_path = hand_model_path;
  set_base_options(&hand_landmarker_options.base_options);

  if (delegate_opt) {
    hand_landmarker_options.base_options.delegate = (Delegate)delegate_opt;
  } else {
    hand_landmarker_options.base_options.delegate = (Delegate)CPU;
  }

  hand_landmarker_options.running_mode = LIVE_STREAM;
  hand_landmarker_options.num_hands = 2;
  hand_landmarker_options.result_callback = mediapipe_lm_hand_callback;

  if (confidence[0] >= 0) {
    hand_landmarker_options.min_hand_detection_confidence = confidence[0];
  }

  if (confidence[1] >= 0) {
    hand_landmarker_options.min_hand_presence_confidence = confidence[1];
  }

  if (confidence[2] >= 0) {
    hand_landmarker_options.min_tracking_confidence = confidence[2];
  }

  MpHandLandmarkerPtr hand_landmarker = NULL;
  char* error_msg = NULL;

	MpStatus status = MpHandLandmarkerCreate(&hand_landmarker_options, &hand_landmarker, &error_msg);
  if (status != 0) {
    save_last_error(error_msg);
    return NULL;
  }

  return hand_landmarker;
}

int mediapipe_lm_hand_detect(void* hand_landmarker, void* mp_image, int64_t timestamp) {
  struct ImageProcessingOptions options;
  options.has_region_of_interest = false;
  options.rotation_degrees = 0;

  char* error_msg;

  MpStatus status = MpHandLandmarkerDetectAsync((MpHandLandmarkerPtr)hand_landmarker, (MpImagePtr)mp_image, &options, timestamp, &error_msg);
  if (status != 0) {
    save_last_error(error_msg);
  }

  return (int)status;
}

int mediapipe_lm_hand_stop(void* hand_landmarker) {
  char* error_msg;

  MpStatus status = MpHandLandmarkerClose((MpHandLandmarkerPtr)hand_landmarker, &error_msg);
  if (status != 0) {
    save_last_error(error_msg);
  }

  return (int)status;
}