#include "v4l2.h"

int check_real_video_capture_device(char* device_filepath, char* cardname) {
    int dev = v4l2_open(device_filepath, 0);
    if (dev == -1) {
        return -errno;
    }

    struct v4l2_capability capabilities;

    int ret = v4l2_ioctl(dev, VIDIOC_QUERYCAP, &capabilities);
    if (ret == -1) {
        return -errno;
    }

    strcpy(cardname, capabilities.card);

    // V4L2_CAP_META_CAPTURE
    int result = ((capabilities.device_caps & V4L2_CAP_VIDEO_CAPTURE) == V4L2_CAP_VIDEO_CAPTURE);
    v4l2_close(dev);
    return result;
}