#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "libtoast.h"
#include "mediapipe/tasks/c/core/base_options.h"
#include "mediapipe/tasks/c/vision/core/image_processing_options.h"
#include "mediapipe/tasks/c/vision/holistic_landmarker/holistic_landmarker.h"

// MAIN FUNCTIONS

void mediapipe_lm_body_callback(MpStatus status, const struct HolisticLandmarkerResult* result, MpImagePtr image, int64_t timestamp_ms) {
  //body_landmarker_callback_external((struct HolisticLandmarkerResult*)result, (int)status, timestamp_ms);
  MpHolisticLandmarkerCloseResult((struct HolisticLandmarkerResult*)result);
}

void* mediapipe_lm_body_start(char* body_model_path, int delegate_opt, float confidence[7]) {
	struct HolisticLandmarkerOptions body_landmarker_options;
  body_landmarker_options.base_options.model_asset_path = body_model_path;
  set_base_options(&body_landmarker_options.base_options);

  if (delegate_opt) {
    body_landmarker_options.base_options.delegate = (Delegate)delegate_opt;
  } else {
    body_landmarker_options.base_options.delegate = (Delegate)CPU;
  }

  body_landmarker_options.running_mode = LIVE_STREAM;
  body_landmarker_options.output_face_blendshapes = true;
  //body_landmarker_options.output_pose_segmentation_masks = true;
  body_landmarker_options.result_callback = mediapipe_lm_body_callback;

  float* targets[7] = {
    &body_landmarker_options.min_face_detection_confidence,
    &body_landmarker_options.min_face_suppression_threshold,
    &body_landmarker_options.min_face_presence_confidence,
    &body_landmarker_options.min_hand_landmarks_confidence,
    &body_landmarker_options.min_pose_detection_confidence,
    &body_landmarker_options.min_pose_suppression_threshold,
    &body_landmarker_options.min_pose_presence_confidence,
  };

  //set_confidence(targets, &confidence, 7);

  MpHolisticLandmarkerPtr body_landmarker = NULL;
  char* error_msg = NULL;

	MpStatus status = MpHolisticLandmarkerCreate(&body_landmarker_options, &body_landmarker, &error_msg);
  if (status != 0) {
    save_last_error(error_msg);
    return NULL;
  }

  return body_landmarker;
}

int mediapipe_lm_body_detect(void* body_landmarker, void* mp_image, int64_t timestamp) {
  struct ImageProcessingOptions options;
  options.has_region_of_interest = false;
  options.rotation_degrees = 0;

  char* error_msg;

  MpStatus status = MpHolisticLandmarkerDetectAsync((MpHolisticLandmarkerPtr)body_landmarker, (MpImagePtr)mp_image, &options, timestamp, &error_msg);
  if (status != 0) {
    save_last_error(error_msg);
  }

  return (int)status;
}

int mediapipe_lm_body_stop(void* face_landmarker) {
  char* error_msg;

  MpStatus status = MpHolisticLandmarkerClose((MpHolisticLandmarkerPtr)face_landmarker, &error_msg);
  if (status != 0) {
    save_last_error(error_msg);
  }

  return (int)status;
}