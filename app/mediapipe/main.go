package main

/*
#cgo CFLAGS: -I./cc/ -I./cc/mediapipe/
#cgo LDFLAGS: -L./cc/ -ltoast -lmediapipe

#include <libtoast.h>
#include <stdlib.h>
*/
import "C"
import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"unsafe"

	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
)

func main() {
	facem_path := "face_landmarker.task"
	c_facem_path := C.CString(facem_path)

	ctx := C.mediapipe_start(c_facem_path)
	if ctx == nil {
		error_str := C.GoString(C.mediapipe_read_error())
		log.Fatal(error_str)
	}

	format := v4l2.PixFormat{
		Width:       1920,
		Height:      1080,
		PixelFormat: v4l2.PixelFmtMJPEG,
	}

	dev, err := device.Open(
		"/dev/video2",
		device.WithBufferSize(1),
		device.WithPixFormat(format),
		device.WithVideoCaptureEnabled(),
		device.WithFPS(30),
	)

	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	dev.GetFrames()
	if err := dev.Start(context.Background()); err != nil {
		log.Fatal(err)
	}

	converter := ConverterFFMPEG{}
	err = converter.init("MJPG")
	if err != nil {
		log.Fatal(err)
	}

	err_channel := make(chan error, 1)
	go read(dev, ctx, converter, err_channel)

	select {
	case err := <-err_channel:
		fmt.Fprintf(os.Stderr, "[MEDIAPIPE + LIBTOAST] %v\n", err)
	}

	println("it has ended")

	converter.end()
	ret := C.mediapipe_stop(ctx)
	if ret < 0 {
		error_str := C.GoString(C.mediapipe_read_error())
		log.Fatal(error_str)
	}

	C.free(unsafe.Pointer(c_facem_path))
}

func read(dev *device.Device, ctx unsafe.Pointer, conv ConverterFFMPEG, err_channel chan error) {
	for frame := range dev.GetFrames() {
		srgb_frame, err := conv.convert(frame.Data)
		if err != nil {
			err_channel <- err
			break
		}

		fmt, err := dev.GetPixFormat()
		if err != nil {
			err_channel <- err
			break
		}

		data_size := len(srgb_frame)
		data_ptr := C.CBytes(srgb_frame)
		timestamp := frame.Timestamp.UnixMilli()
		ret := C.mediapipe_detect(ctx, data_ptr, C.int(data_size), C.int(fmt.Width), C.int(fmt.Height), C.long(timestamp))
		frame.Release()

		if ret < 0 {
			error_str := C.GoString(C.mediapipe_read_error())
			err_channel <- errors.New(error_str)
			C.mediapipe_free_error()
			break
		}
	}

	dev.Stop()
	close(err_channel)
}

//export mediapipe_call_HELP
func mediapipe_call_HELP(value C.int) {
	log.Printf("%d", int(value))
}

//export mediapipe_call_facem_result
func mediapipe_call_facem_result(facem_ptr unsafe.Pointer, timestamp C.long) {

	facem_result := (*C.struct_FaceLandmarkerResult)(facem_ptr)
	if facem_result.face_blendshapes != nil {
		for i := 0; i < int(facem_result.face_blendshapes.categories_count); i++ {
			blendshape := C.face_landmarker_blendshape(facem_ptr, C.uint(i))
			println(C.GoString(blendshape.category_name))
		}
	}

	if facem_result.face_landmarks != nil {
		for i := 0; i < int(facem_result.face_landmarks.landmarks_count); i++ {
			_ = C.face_landmarker_landmark(facem_ptr, C.uint(i))
			//println(i)
		}
	}

	for i := 0; i < int(facem_result.facial_transformation_matrixes_count); i++ {
		matrix := C.face_landmarker_matrix(facem_ptr, C.uint(i))
		println(matrix.rows)
	}
}
