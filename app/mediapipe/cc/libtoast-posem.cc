#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "libtoast.h"
#include "mediapipe/tasks/c/core/base_options.h"
#include "mediapipe/tasks/c/vision/core/image_processing_options.h"
#include "mediapipe/tasks/c/vision/pose_landmarker/pose_landmarker.h"

// ARRAY HELPERS

uint32_t* pose_landmarker_landmark_count(struct PoseLandmarkerResult* result) {
  return &result->pose_landmarks[0].landmarks_count;
}

uint32_t* pose_landmarker_world_landmark_count(struct PoseLandmarkerResult* result) {
  return &result->pose_landmarks[0].landmarks_count;
}

struct NormalizedLandmark* pose_landmarker_landmark(struct PoseLandmarkerResult* result, uint32_t index) {
  return &result->pose_landmarks[0].landmarks[index];
}

struct Landmark* pose_landmarker_world_landmark(struct PoseLandmarkerResult* result, uint32_t index) {
  return &result->pose_world_landmarks[0].landmarks[index];
}

// MAIN FUNCTIONS

void mediapipe_lm_pose_callback(MpStatus status, const struct PoseLandmarkerResult* result, MpImagePtr image, int64_t timestamp_ms) {
  pose_landmarker_callback_external((struct PoseLandmarkerResult*)result, (int)status, timestamp_ms);
  MpPoseLandmarkerCloseResult((struct PoseLandmarkerResult*)result);
}

void* mediapipe_lm_pose_start(char* pose_model_path, int delegate_opt, float confidence[3]) {
	struct PoseLandmarkerOptions pose_landmarker_options;
  pose_landmarker_options.base_options.model_asset_path = pose_model_path;
  set_base_options(&pose_landmarker_options.base_options);

  if (delegate_opt) {
    pose_landmarker_options.base_options.delegate = (Delegate)delegate_opt;
  } else {
    pose_landmarker_options.base_options.delegate = (Delegate)CPU;
  }

  pose_landmarker_options.running_mode = LIVE_STREAM;
  pose_landmarker_options.num_poses = 1;
  //pose_landmarker_options.output_segmentation_masks = true;
  pose_landmarker_options.result_callback = mediapipe_lm_pose_callback;

  float* targets[3] = {
    &pose_landmarker_options.min_pose_detection_confidence,
    &pose_landmarker_options.min_pose_presence_confidence,
    &pose_landmarker_options.min_tracking_confidence,
  };

  //set_confidence(targets, &confidence, 3);

  MpPoseLandmarkerPtr pose_landmarker = NULL;
  char* error_msg = NULL;

	MpStatus status = MpPoseLandmarkerCreate(&pose_landmarker_options, &pose_landmarker, &error_msg);
  if (status != 0) {
    save_last_error(error_msg);
    return NULL;
  }

  return pose_landmarker;
}

int mediapipe_lm_pose_detect(void* pose_landmarker, void* mp_image, int64_t timestamp) {
  struct ImageProcessingOptions options;
  options.has_region_of_interest = false;
  options.rotation_degrees = 0;

  char* error_msg;

  MpStatus status = MpPoseLandmarkerDetectAsync((MpPoseLandmarkerPtr)pose_landmarker, (MpImagePtr)mp_image, &options, timestamp, &error_msg);
  if (status != 0) {
    save_last_error(error_msg);
  }

  return (int)status;
}

int mediapipe_lm_pose_stop(void* face_landmarker) {
  char* error_msg;

  MpStatus status = MpPoseLandmarkerClose((MpPoseLandmarkerPtr)face_landmarker, &error_msg);
  if (status != 0) {
    save_last_error(error_msg);
  }

  return (int)status;
}