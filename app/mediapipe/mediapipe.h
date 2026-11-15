#ifndef MARMALADE_MEDIAPIPE_H
#define MARMALADE_MEDIAPIPE_H

#include <stdlib.h>

#include "mediapipe/tasks/c/core/base_options.h"
#include "mediapipe/tasks/c/vision/core/image.h"
#include "mediapipe/tasks/c/vision/core/image_processing_options.h"

#include "mediapipe/tasks/c/vision/face_landmarker/face_landmarker.h"
#include "mediapipe/tasks/c/vision/hand_landmarker/hand_landmarker.h"
#include "mediapipe/tasks/c/vision/pose_landmarker/pose_landmarker.h"

#include "mediapipe/tasks/c/vision/face_landmarker/face_landmarker_result.h"
#include "mediapipe/tasks/c/vision/hand_landmarker/hand_landmarker_result.h"
#include "mediapipe/tasks/c/vision/pose_landmarker/pose_landmarker_result.h"

// FACE LANDMARKER

struct MpCategory* face_landmarker_blendshape(struct MpFaceLandmarkerResult*, uint32_t);
struct MpNormalizedLandmark* face_landmarker_landmark(struct MpFaceLandmarkerResult*, uint32_t);
struct MpMatrix* face_landmarker_matrix(struct MpFaceLandmarkerResult*, uint32_t);
float* face_landmarker_matrix_data(struct MpMatrix*, uint32_t);

extern void face_landmarker_callback_external(struct MpFaceLandmarkerResult*, int, signed long);
void face_landmarker_callback_set(struct MpFaceLandmarkerOptions* options);

// HAND LANDMARKER

uint32_t* hand_landmarker_handedness_count(struct MpHandLandmarkerResult*, uint32_t);
uint32_t* hand_landmarker_landmark_count(struct MpHandLandmarkerResult*, uint32_t);
uint32_t* hand_landmarker_world_landmark_count(struct MpHandLandmarkerResult*, uint32_t);
struct MpCategory* hand_landmarker_handedness(struct MpHandLandmarkerResult*, uint32_t, uint32_t);
struct MpNormalizedLandmark* hand_landmarker_landmark(struct MpHandLandmarkerResult*, uint32_t, uint32_t);
struct MpLandmark* hand_landmarker_world_landmark(struct MpHandLandmarkerResult*, uint32_t, uint32_t);

extern void hand_landmarker_callback_external(struct MpHandLandmarkerResult*, int, signed long);
void hand_landmarker_callback_set(struct MpHandLandmarkerOptions* options);

// POSE LANDMARKER

uint32_t* pose_landmarker_landmark_count(struct MpPoseLandmarkerResult*);
uint32_t* pose_landmarker_world_landmark_count(struct MpPoseLandmarkerResult*);
struct MpNormalizedLandmark* pose_landmarker_landmark(struct MpPoseLandmarkerResult*, uint32_t);
struct MpLandmark* pose_landmarker_world_landmark(struct MpPoseLandmarkerResult*, uint32_t);

extern void pose_landmarker_callback_external(struct MpPoseLandmarkerResult*, int, signed long);
void pose_landmarker_callback_set(struct MpPoseLandmarkerOptions* options);

#endif // MARMALADE_MEDIAPIPE_H