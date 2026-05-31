#ifndef LIBTOAST_H
#define LIBTOAST_H

#include "mediapipe/tasks/c/core/base_options.h"
#include "mediapipe/tasks/c/vision/face_landmarker/face_landmarker_result.h"
#include "mediapipe/tasks/c/vision/hand_landmarker/hand_landmarker_result.h"

void save_last_error(char*);
void set_base_options(struct BaseOptions*);

#ifdef __cplusplus
extern "C" {
#endif

void* mediapipe_create_img(int*, void*, int, int, int);
void mediapipe_free_img(void*);

char* mediapipe_read_error();
void mediapipe_free_error();

// FACE LANDMARKER

struct Category* face_landmarker_blendshape(struct FaceLandmarkerResult*, uint32_t);
struct NormalizedLandmark* face_landmarker_landmark(struct FaceLandmarkerResult*, uint32_t);
struct Matrix* face_landmarker_matrix(struct FaceLandmarkerResult*, uint32_t);
float* face_landmarker_matrix_data(struct Matrix*, uint32_t);

extern void face_landmarker_callback_external(struct FaceLandmarkerResult*, int, signed long);
void* mediapipe_lm_face_start(char*, int);
int mediapipe_lm_face_detect(void*, void*, int64_t);
int mediapipe_lm_face_stop(void*);

// HAND LANDMARKER

uint32_t* hand_landmarker_handedness_count(struct HandLandmarkerResult*, uint32_t);
uint32_t* hand_landmarker_landmark_count(struct HandLandmarkerResult*, uint32_t);
uint32_t* hand_landmarker_world_landmark_count(struct HandLandmarkerResult*, uint32_t);
struct Category* hand_landmarker_handedness(struct HandLandmarkerResult*, uint32_t, uint32_t);
struct NormalizedLandmark* hand_landmarker_landmark(struct HandLandmarkerResult*, uint32_t, uint32_t);
struct Landmark* hand_landmarker_world_landmark(struct HandLandmarkerResult*, uint32_t, uint32_t);

extern void hand_landmarker_callback_external(struct HandLandmarkerResult*, int, signed long);
void* mediapipe_lm_hand_start(char*, int);
int mediapipe_lm_hand_detect(void*, void*, int64_t);
int mediapipe_lm_hand_stop(void*);

#ifdef __cplusplus
}
#endif

#endif // LIBTOAST_H