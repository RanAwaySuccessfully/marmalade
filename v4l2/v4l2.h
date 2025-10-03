#include <linux/videodev2.h>
#include <libv4l2.h>
#include <errno.h>

#include <stdlib.h>
#include <string.h>

int check_real_video_capture_device(char* device_filepath, char* cardname);