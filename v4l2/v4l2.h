#include <linux/videodev2.h>
#include <libv4l2.h>
#include <errno.h>

#include <stdlib.h>
#include <string.h>

int get_errno();

int m_v4l2_open(char* file, int oflag);
int m_v4l2_vidioc_querycap(int fd, struct v4l2_capability* capabilities);
int m_v4l2_vidioc_enum_fmt(int fd, struct v4l2_fmtdesc* capabilities);