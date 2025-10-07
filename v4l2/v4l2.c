#include "v4l2.h"

int get_errno() {
    return errno;
}

int m_v4l2_open(char* file, int oflag) {
    return v4l2_open(file, oflag);
}

int m_v4l2_vidioc_querycap(int fd, struct v4l2_capability* capabilities) {
    return v4l2_ioctl(fd, VIDIOC_QUERYCAP, capabilities);
}

int m_v4l2_vidioc_enum_fmt(int fd, struct v4l2_fmtdesc* capabilities) {
    return v4l2_ioctl(fd, VIDIOC_ENUM_FMT, capabilities);
}

int m_v4l2_vidioc_enum_framesizes(int fd, struct v4l2_frmsizeenum* capabilities) {
    return v4l2_ioctl(fd, VIDIOC_ENUM_FRAMESIZES, capabilities);
}

int m_v4l2_vidioc_enum_frameintervals(int fd, struct v4l2_frmivalenum* capabilities) {
    return v4l2_ioctl(fd, VIDIOC_ENUM_FRAMEINTERVALS, capabilities);
}