#ifndef LIBTOAST_H
#define LIBTOAST_H

#include "mediapipe/tasks/c/core/base_options.h"
#include "mediapipe/tasks/c/vision/face_landmarker/face_landmarker_result.h"
#include "mediapipe/tasks/c/vision/hand_landmarker/hand_landmarker_result.h"
#include "mediapipe/tasks/c/vision/pose_landmarker/pose_landmarker_result.h"
//#include "mediapipe/tasks/c/vision/holistic_landmarker/holistic_landmarker_result.h"

void save_last_error(char*);
void set_base_options(struct MpBaseOptions*);
void set_confidence(float**, float**, int);

#ifdef __cplusplus
extern "C" {
#endif

void* mediapipe_create_img(int*, void*, int, int, int);
void mediapipe_free_img(void*);

char* mediapipe_read_error();
void mediapipe_free_error();

// FACE LANDMARKER

struct MpCategory* face_landmarker_blendshape(struct MpFaceLandmarkerResult*, uint32_t);
struct MpNormalizedLandmark* face_landmarker_landmark(struct MpFaceLandmarkerResult*, uint32_t);
struct MpMatrix* face_landmarker_matrix(struct MpFaceLandmarkerResult*, uint32_t);
float* face_landmarker_matrix_data(struct MpMatrix*, uint32_t);

extern void face_landmarker_callback_external(struct MpFaceLandmarkerResult*, int, signed long);
void* mediapipe_lm_face_start(char*, int, float[3]);
int mediapipe_lm_face_detect(void*, void*, int64_t);
int mediapipe_lm_face_stop(void*);

// HAND LANDMARKER

uint32_t* hand_landmarker_handedness_count(struct MpHandLandmarkerResult*, uint32_t);
uint32_t* hand_landmarker_landmark_count(struct MpHandLandmarkerResult*, uint32_t);
uint32_t* hand_landmarker_world_landmark_count(struct MpHandLandmarkerResult*, uint32_t);
struct MpCategory* hand_landmarker_handedness(struct MpHandLandmarkerResult*, uint32_t, uint32_t);
struct MpNormalizedLandmark* hand_landmarker_landmark(struct MpHandLandmarkerResult*, uint32_t, uint32_t);
struct MpLandmark* hand_landmarker_world_landmark(struct MpHandLandmarkerResult*, uint32_t, uint32_t);

extern void hand_landmarker_callback_external(struct MpHandLandmarkerResult*, int, signed long);
void* mediapipe_lm_hand_start(char*, int, float[3]);
int mediapipe_lm_hand_detect(void*, void*, int64_t);
int mediapipe_lm_hand_stop(void*);

// POSE LANDMARKER

uint32_t* pose_landmarker_landmark_count(struct MpPoseLandmarkerResult*);
uint32_t* pose_landmarker_world_landmark_count(struct MpPoseLandmarkerResult*);
struct MpNormalizedLandmark* pose_landmarker_landmark(struct MpPoseLandmarkerResult*, uint32_t);
struct MpLandmark* pose_landmarker_world_landmark(struct MpPoseLandmarkerResult*, uint32_t);

extern void pose_landmarker_callback_external(struct MpPoseLandmarkerResult*, int, signed long);
void* mediapipe_lm_pose_start(char*, int, float[3]);
int mediapipe_lm_pose_detect(void*, void*, int64_t);
int mediapipe_lm_pose_stop(void*);

// HOLISTIC LANDMARKER

/*
extern void body_landmarker_callback_external(struct MpHolisticLandmarkerResult*, int, signed long);
void* mediapipe_lm_body_start(char*, int, float[7]);
int mediapipe_lm_body_detect(void*, void*, int64_t);
int mediapipe_lm_body_stop(void*);
*/

#ifdef __cplusplus
}
#endif

#endif // LIBTOAST_H