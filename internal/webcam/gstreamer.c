//go:build gstreamer

#include "gstreamer.h"
#include "glib-object.h"
#include "glibconfig.h"
#include "gst/gstvalue.h"
#include <stdio.h>

GstFlowReturn on_new_sample(GstAppSink* sink, gpointer handle_ptr) {
	return on_new_sample_external(sink, handle_ptr);
}

void set_signals(GstElement* obj, void* handle_ptr) {
	g_object_set(G_OBJECT(obj), "emit-signals", TRUE, "sync", FALSE, NULL);
	g_signal_connect(obj, "new-sample", G_CALLBACK(on_new_sample), handle_ptr);
}

GstBin* gst_bin(GstElement* obj) {
	return GST_BIN(obj);
}

GstClockTime gst_buffer_pts(GstBuffer* buffer) {
	return GST_BUFFER_PTS(buffer);
}

float divide_fraction(const GValue* input) {
    return (float)gst_value_get_fraction_numerator(input) / (float)gst_value_get_fraction_denominator(input);
}

void check_type(const GValue* input) {
    GType type = G_VALUE_TYPE(input);

    if (type == G_TYPE_INT) {
        printf("%d ", g_value_get_int(input));

    } else if (type == GST_TYPE_FRACTION) {
        printf("%f ", divide_fraction(input));

    } else if (type == GST_TYPE_INT_RANGE) {
        printf("%d-%d ", gst_value_get_int_range_min(input), gst_value_get_int_range_max(input));
        
    } else if (type == GST_TYPE_FRACTION_RANGE) {
        const GValue* min = gst_value_get_fraction_range_min(input);
        const GValue* max = gst_value_get_fraction_range_max(input);
        printf("%f-%f ", divide_fraction(min), divide_fraction(max));
        
    } else if (type == GST_TYPE_LIST) {
        gint size = gst_value_list_get_size(input);
        for (int i = 0; i < size; i++) {
            const GValue *inner_value = gst_value_list_get_value(input, i);
            check_type(inner_value);
        }
    }
}

// gst-device-monitor-1.0 Video/Source
void list_caps() {
    gst_init(NULL, NULL);

    GstDeviceMonitor* monitor = gst_device_monitor_new();
	gst_device_monitor_add_filter(monitor, "Video/Source", NULL);
	gst_device_monitor_start(monitor);

	GList* devices = gst_device_monitor_get_devices(monitor);
    for (GList* l = devices; l; l = l->next) {
        GstDevice* device = GST_DEVICE(l->data);
        printf("Device: %s\n", gst_device_get_display_name(device));

        GstCaps* caps = gst_device_get_caps(device);
        guint size = gst_caps_get_size(caps);
        
        for (int i = 0; i < size; i++) {
            GstStructure *structure = gst_caps_get_structure(caps, i);

            printf("---------\n");

            const GValue* width = gst_structure_get_value(structure, "width");
            printf("Width: ");
            check_type(width);
            printf("\n");

            const GValue* height = gst_structure_get_value(structure, "height");
            printf("Height: ");
            check_type(height);
            printf("\n");

            const GValue* fps = gst_structure_get_value(structure, "framerate");
            printf("Framerate: ");
            check_type(fps);
            printf("\n");

            const gchar* format = gst_structure_get_string(structure, "format");
            const gchar* name = gst_structure_get_name(structure);
            printf("Format: %s %s\n", name, format);
        }

        gst_caps_unref(caps);
    }
    
    fflush(stdout);

    g_list_free_full(devices, (GDestroyNotify)gst_object_unref);
    gst_device_monitor_stop(monitor);
    g_object_unref(monitor);

    gst_deinit();
}