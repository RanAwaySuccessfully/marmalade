#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "libtoast.h"
#include "mediapipe/tasks/c/core/common.h"
#include "mediapipe/tasks/c/vision/core/image.h"

void* mediapipe_create_img(int* status_ptr, void* img_ptr, int img_size, int width, int height) {
  MpImagePtr image;
  char* error_msg;

  MpStatus status = MpImageCreateFromUint8Data(kMpImageFormatSrgb, width, height, (const uint8_t*)img_ptr, img_size, &image, &error_msg);
  if (status != 0) {
    save_last_error(error_msg);
    *status_ptr = status;
    return NULL;
  }

  *status_ptr = status;
  return image;
}

void mediapipe_free_img(void* mpimg_ptr) {
  MpImageFree((MpImagePtr)mpimg_ptr);
}

void set_base_options(struct BaseOptions* base_options) {
  base_options->model_asset_buffer = NULL;
  base_options->model_asset_buffer_count = 0;
  base_options->host_environment = HOST_ENVIRONMENT_UNKNOWN;
  base_options->host_system = HOST_SYSTEM_LINUX;
  base_options->host_version = NULL;
  base_options->ca_bundle_path = NULL;
}

void set_confidence(float** target, float** source, int count) {
  for (int i = 0; i < count; i++) {
    if (*source[i] >= 0) {
      *target[i] = *source[i];
    }
  }
}

// ERROR HANDLING

char* mediapipe_last_error = NULL;

void save_last_error(char* error_msg) {
  size_t length = strlen(error_msg);

  mediapipe_last_error = (char*)realloc(mediapipe_last_error, (length + 1) * sizeof(char));
  if (mediapipe_last_error != NULL) {
    strncpy(mediapipe_last_error, error_msg, length);
    mediapipe_last_error[length] = '\0';
  }

  MpErrorFree(error_msg);
}

char* mediapipe_read_error() {
  return mediapipe_last_error;
}

void mediapipe_free_error() {
  free(mediapipe_last_error);
  mediapipe_last_error = NULL;
}
