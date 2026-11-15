#include "mediapipe.h"

// FACE LANDMARKER

struct MpCategory* face_landmarker_blendshape(struct MpFaceLandmarkerResult* result, uint32_t index) {
  return &result->face_blendshapes[0].categories[index];
}

struct MpNormalizedLandmark* face_landmarker_landmark(struct MpFaceLandmarkerResult* result, uint32_t index) {
  return &result->face_landmarks[0].landmarks[index];
}

struct MpMatrix* face_landmarker_matrix(struct MpFaceLandmarkerResult* result, uint32_t index) {
  return &result->facial_transformation_matrixes[index];
}

float* face_landmarker_matrix_data(struct MpMatrix* matrix, uint32_t index) {
  return &matrix->data[index];
}

void face_landmarker_callback(MpStatus status, const struct MpFaceLandmarkerResult* result, MpImagePtr image, int64_t timestamp_ms) {
  face_landmarker_callback_external((struct MpFaceLandmarkerResult*)result, (int)status, timestamp_ms);
  MpFaceLandmarkerCloseResult((struct MpFaceLandmarkerResult*)result);
}

void face_landmarker_callback_set(struct MpFaceLandmarkerOptions* options) {
  options->result_callback = face_landmarker_callback;
}

// HAND LANDMARKER

uint32_t* hand_landmarker_handedness_count(struct MpHandLandmarkerResult* result, uint32_t hand) {
  return &result->handedness[hand].categories_count;
}

uint32_t* hand_landmarker_landmark_count(struct MpHandLandmarkerResult* result, uint32_t hand) {
  return &result->hand_landmarks[hand].landmarks_count;
}

uint32_t* hand_landmarker_world_landmark_count(struct MpHandLandmarkerResult* result, uint32_t hand) {
  return &result->hand_world_landmarks[hand].landmarks_count;
}

struct MpCategory* hand_landmarker_handedness(struct MpHandLandmarkerResult* result, uint32_t hand, uint32_t index) {
  return &result->handedness[hand].categories[index];
}

struct MpNormalizedLandmark* hand_landmarker_landmark(struct MpHandLandmarkerResult* result, uint32_t hand, uint32_t index) {
  return &result->hand_landmarks[hand].landmarks[index];
}

struct MpLandmark* hand_landmarker_world_landmark(struct MpHandLandmarkerResult* result, uint32_t hand, uint32_t index) {
  return &result->hand_world_landmarks[hand].landmarks[index];
}

void hand_landmarker_callback(MpStatus status, const struct MpHandLandmarkerResult* result, MpImagePtr image, int64_t timestamp_ms) {
  hand_landmarker_callback_external((struct MpHandLandmarkerResult*)result, (int)status, timestamp_ms);
  MpHandLandmarkerCloseResult((struct MpHandLandmarkerResult*)result);
}

void hand_landmarker_callback_set(struct MpHandLandmarkerOptions* options) {
  options->result_callback = hand_landmarker_callback;
}

// POSE LANDMARKER

uint32_t* pose_landmarker_landmark_count(struct MpPoseLandmarkerResult* result) {
  return &result->pose_landmarks[0].landmarks_count;
}

uint32_t* pose_landmarker_world_landmark_count(struct MpPoseLandmarkerResult* result) {
  return &result->pose_landmarks[0].landmarks_count;
}

struct MpNormalizedLandmark* pose_landmarker_landmark(struct MpPoseLandmarkerResult* result, uint32_t index) {
  return &result->pose_landmarks[0].landmarks[index];
}

struct MpLandmark* pose_landmarker_world_landmark(struct MpPoseLandmarkerResult* result, uint32_t index) {
  return &result->pose_world_landmarks[0].landmarks[index];
}

void pose_landmarker_callback(MpStatus status, const struct MpPoseLandmarkerResult* result, MpImagePtr image, int64_t timestamp_ms) {
  pose_landmarker_callback_external((struct MpPoseLandmarkerResult*)result, (int)status, timestamp_ms);
  MpPoseLandmarkerCloseResult((struct MpPoseLandmarkerResult*)result);
}

void pose_landmarker_callback_set(struct MpPoseLandmarkerOptions* options) {
  options->result_callback = pose_landmarker_callback;
}