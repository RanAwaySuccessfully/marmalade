#ifndef LIBTOAST_H
#define LIBTOAST_H

#include "mediapipe/tasks/c/vision/face_landmarker/face_landmarker_result.h"

#ifdef __cplusplus
extern "C" {
#endif

extern void mediapipe_call_facem_result(struct FaceLandmarkerResult*, int, signed long);

struct Category face_landmarker_blendshape(struct FaceLandmarkerResult*, uint32_t);
struct NormalizedLandmark face_landmarker_landmark(struct FaceLandmarkerResult*, uint32_t);
struct Matrix face_landmarker_matrix(struct FaceLandmarkerResult*, uint32_t);
float face_landmarker_matrix_data(struct Matrix*, uint32_t);

void* mediapipe_start(char*, int);
int mediapipe_detect(void*, void*, int, int, int, int64_t);
int mediapipe_stop(void*);
char* mediapipe_read_error();
void mediapipe_free_error();

#ifdef __cplusplus
}
#endif

#endif // LIBTOAST_H