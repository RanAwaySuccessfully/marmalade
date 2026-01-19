#ifndef LIBTOAST_H
#define LIBTOAST_H

#include "mediapipe/tasks/c/vision/face_landmarker/face_landmarker_result.h"
#include "mediapipe/tasks/c/vision/hand_landmarker/hand_landmarker_result.h"

#ifdef __cplusplus
extern "C" {
#endif

extern void face_landmarker_callback_external(struct FaceLandmarkerResult*, int, signed long);
extern void hand_landmarker_callback_external(struct HandLandmarkerResult*, int, signed long);

struct Category face_landmarker_blendshape(struct FaceLandmarkerResult*, uint32_t);
struct NormalizedLandmark face_landmarker_landmark(struct FaceLandmarkerResult*, uint32_t);
struct Matrix face_landmarker_matrix(struct FaceLandmarkerResult*, uint32_t);
float face_landmarker_matrix_data(struct Matrix*, uint32_t);

uint32_t hand_landmarker_handedness_count(struct HandLandmarkerResult*, uint32_t);
uint32_t hand_landmarker_landmark_count(struct HandLandmarkerResult*, uint32_t);
uint32_t hand_landmarker_world_landmark_count(struct HandLandmarkerResult*, uint32_t);
struct Category hand_landmarker_handedness(struct HandLandmarkerResult*, uint32_t, uint32_t);
struct NormalizedLandmark hand_landmarker_landmark(struct HandLandmarkerResult*, uint32_t, uint32_t);
struct Landmark hand_landmarker_world_landmark(struct HandLandmarkerResult*, uint32_t, uint32_t);

void* face_landmarker_start(char*, int);
void* hand_landmarker_start(char*, int);

int mediapipe_detect(void*, void*, void*, int, int, int, int64_t);
int mediapipe_stop(void*, void*);

char* mediapipe_read_error();
void mediapipe_free_error();

#ifdef __cplusplus
}
#endif

#endif // LIBTOAST_H