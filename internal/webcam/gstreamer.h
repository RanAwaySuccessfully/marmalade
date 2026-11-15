//go:build gstreamer

#include <glib.h>
#include <gst/app/gstappsink.h>
#include <gst/video/video.h>
#include <gst/gstbin.h>
#include <gst/gstbuffer.h>
#include <gst/gstclock.h>
#include <gst/gstmacros.h>
#include <gst/gstparse.h>

extern GstFlowReturn on_new_sample_external(GstAppSink*, gpointer);
GstFlowReturn on_new_sample(GstAppSink*, gpointer);

void set_signals(GstElement*, void*);
GstBin* gst_bin(GstElement*);
GstClockTime gst_buffer_pts(GstBuffer*);

void list_caps();