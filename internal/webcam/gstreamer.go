//go:build gstreamer

package webcam

/*
#cgo CFLAGS: -I/usr/include/gstreamer-1.0 -I/usr/include/glib-2.0 -I/usr/lib/x86_64-linux-gnu/glib-2.0/include
#cgo LDFLAGS: -lgstvideo-1.0 -lgstapp-1.0 -lgstbase-1.0 -lgstreamer-1.0 -lgobject-2.0 -lglib-2.0

#include "gstreamer.h"
*/
import "C"
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"marmalade/internal/server"
	"os"
	"os/exec"
	"runtime/cgo"
	"strconv"
	"unsafe"
)

type PipeWireNode struct {
	Info struct {
		Props struct {
			ObjectId    float32 `json:"object.id"`
			APIv4l2Path string  `json:"api.v4l2.path"`
			MediaClass  string  `json:"media.class"`
		} `json:"props"`
	} `json:"info"`
}

type GStreamer struct {
	pipeline     *C.struct__GstElement
	pipeline_cmd *C.char
	sink_name    *C.char
	handler      cgo.Handle
	data_ch      chan GFrame
	done_ch      chan bool
}

type GFrame struct {
	data      []byte
	info      C.struct__GstVideoInfo
	timestamp C.GstClockTime
}

func GetCaps() {
	C.list_caps()
}

func NewGStreamer(fd int) (*GStreamer, error) {
	gst := GStreamer{}
	C.gst_init(nil, nil)

	extra := ""

	switch server.Config.Format {
	case "MJPG":
		extra = "jpegdec ! " // try "decodebin3 ! "
		/*
			case "HEVC":
				fallthrough
			case "H264":
				extra = "decodebin ! "
		*/
	}

	command := extra + "videoconvert ! video/x-raw, format=RGB, width=%d, height=%d, framerate=%d/1 ! appsink name=receiver"
	if fd == 0 {
		object_id, err := get_pipewire_obj_id(server.Config.Camera)
		if err != nil {
			return nil, err
		}

		command = fmt.Sprintf("pipewiresrc target-object=%d ! "+command,
			object_id,
			int(server.Config.Width),
			int(server.Config.Height),
			int(server.Config.FPS),
		)
	} else {
		command = fmt.Sprintf("pipewiresrc fd=%d ! "+command,
			fd,
			int(server.Config.Width),
			int(server.Config.Height),
			int(server.Config.FPS),
		)
	}

	gst.pipeline_cmd = C.CString(command)

	var gerror *C.struct__GError
	gst.pipeline = C.gst_parse_launch(gst.pipeline_cmd, &gerror)

	var err error
	if gerror != nil {
		err = errors.New(C.GoString(gerror.message))
		C.g_clear_error(&gerror)
	}

	if gst.pipeline == nil {
		return nil, err
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	gst.sink_name = C.CString("receiver")
	sink := C.gst_bin_get_by_name(C.gst_bin(gst.pipeline), gst.sink_name)

	gst.data_ch = make(chan GFrame)
	gst.done_ch = make(chan bool)
	gst.handler = cgo.NewHandle(&gst)
	handle_ptr := unsafe.Pointer(gst.handler)
	C.set_signals(sink, handle_ptr)

	return &gst, nil
}

func (gst *GStreamer) Run(err_ch chan error, detect RunCallback) {
	ret := C.gst_element_set_state(gst.pipeline, C.GST_STATE_PLAYING)
	if ret == 0 {
		err_ch <- errors.New("GstElement state change failure")
		return
	}

	for frame := range gst.data_ch {
		fmt.Printf("%dx%d@%d\n", uint32(frame.info.width), uint32(frame.info.height), int64(frame.info.fps_n)) // just checking...
		err := detect(frame.data, uint32(frame.info.width), uint32(frame.info.height), int64(frame.timestamp))
		gst.done_ch <- true // send this so on_new_sample_external() can know we're done
		if err != nil {
			err_ch <- err
			break
		}
	}
}

func (gst *GStreamer) End() {
	if gst.pipeline_cmd != nil {
		C.free(unsafe.Pointer(gst.pipeline_cmd))
	}

	if gst.sink_name != nil {
		C.free(unsafe.Pointer(gst.sink_name))
	}

	gst.handler.Delete()
	C.gst_deinit()
}

//export on_new_sample_external
func on_new_sample_external(sink *C.struct__GstAppSink, handle_ptr unsafe.Pointer) C.GstFlowReturn {
	sample := C.gst_app_sink_pull_sample(sink)

	handle := (cgo.Handle)(handle_ptr)
	gst := handle.Value().(*GStreamer)

	if sample != nil {
		buffer := C.gst_sample_get_buffer(sample)
		var gst_map C.GstMapInfo

		ret := C.gst_buffer_map(buffer, &gst_map, C.GST_MAP_READ)
		if ret == C.TRUE {

			frame := GFrame{}
			frame.data = C.GoBytes(unsafe.Pointer(gst_map.data), C.int(gst_map.size))
			frame.timestamp = C.gst_buffer_pts(buffer)

			caps := C.gst_sample_get_caps(sample)
			C.gst_video_info_from_caps(&frame.info, caps)

			gst.data_ch <- frame
			<-gst.done_ch // wait until we're done using the frame

			C.gst_buffer_unmap(buffer, &gst_map)
		}

		C.gst_sample_unref(sample)
	}

	return C.GST_FLOW_OK
}

func get_pipewire_obj_id(device_id int) (int, error) {
	cmd := exec.Command("pw-dump")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return -1, err
	}

	stdout_result := out.Bytes()

	objs := make([]PipeWireNode, 0)
	err = json.Unmarshal(stdout_result, &objs)
	if err != nil {
		return -1, err
	}

	for _, obj := range objs {
		if obj.Info.Props.MediaClass == "Video/Source" {
			dev_name := "/dev/video" + strconv.Itoa(device_id)
			if obj.Info.Props.APIv4l2Path == dev_name {
				return int(obj.Info.Props.ObjectId), nil
			}
		}
	}

	return -1, nil
}
