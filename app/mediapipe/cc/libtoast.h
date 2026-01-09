#ifndef LIBTOAST_H
#define LIBTOAST_H

#include "mediapipe/tasks/c/vision/face_landmarker/face_landmarker_result.h"

#ifdef __cplusplus
extern "C" {
#endif

extern void mediapipe_call_facem_result(void*, signed long);
extern void mediapipe_call_HELP(int);

struct Category face_landmarker_blendshape(void*, uint32_t);
struct NormalizedLandmark face_landmarker_landmark(void*, uint32_t);
struct Matrix face_landmarker_matrix(void*, uint32_t);

void* mediapipe_start(char*);
int mediapipe_detect(void*, void*, int, int, int, int64_t);
int mediapipe_stop(void*);
char* mediapipe_read_error();
void mediapipe_free_error();

#ifdef __cplusplus
}
#endif

#endif // LIBTOAST_H