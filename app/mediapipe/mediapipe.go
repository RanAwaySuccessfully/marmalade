package main

/*
#cgo CFLAGS: -I./cc/ -I./cc/mediapipe/
#cgo LDFLAGS: -L./cc/ -ltoast -lmediapipe

#include <libtoast.h>
#include <stdlib.h>
*/
import "C"
import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"marmalade/internal/server"
	"os"
	"unsafe"

	"github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
)

type MediaPipe struct {
	webcam     *device.Device
	converter  *ConverterFFMPEG
	facem_lm   unsafe.Pointer
	facem_path *C.char
}

func (mp *MediaPipe) start() error {
	server.Config.Read()

	buffer := bytes.Buffer{}
	buffer.WriteString(server.Config.Format)

	var fourcc uint32
	err := binary.Read(&buffer, binary.LittleEndian, &fourcc)
	if err != nil {
		return err
	}

	format := v4l2.PixFormat{
		Width:       uint32(server.Config.Width),
		Height:      uint32(server.Config.Height),
		PixelFormat: uint32(fourcc),
	}

	device_path := fmt.Sprintf("/dev/video%d", int(server.Config.Camera))

	mp.webcam, err = device.Open(
		device_path,
		device.WithBufferSize(1),
		device.WithPixFormat(format),
		device.WithFPS(uint32(server.Config.FPS)),
	)

	if err != nil {
		return err
	}

	mp.webcam.GetFrames()
	err = mp.webcam.Start(context.Background())
	if err != nil {
		return err
	}

	mp.converter = &ConverterFFMPEG{}
	err = mp.converter.init(server.Config.Format)
	if err != nil {
		return err
	}

	mp.facem_path = C.CString(server.Config.ModelFace)

	delegate := 0
	if server.Config.UseGpu {
		delegate = 1
	}

	mp.facem_lm = C.mediapipe_start(mp.facem_path, C.int(delegate))
	if mp.facem_lm == nil {
		error_str := C.GoString(C.mediapipe_read_error())
		C.mediapipe_free_error()
		return errors.New(error_str)
	}

	return nil
}

func (mp *MediaPipe) detect(err_channel chan error) {
	for frame := range mp.webcam.GetFrames() {
		//start := time.Now().UnixMilli()

		srgb_frame, err := mp.converter.convert(frame.Data)
		if err != nil {
			err_channel <- err
			break
		}

		format, err := mp.webcam.GetPixFormat()
		if err != nil {
			err_channel <- err
			break
		}

		data_size := len(srgb_frame)
		data_ptr := C.CBytes(srgb_frame)
		timestamp := frame.Timestamp.UnixMilli()
		ret := C.mediapipe_detect(mp.facem_lm, data_ptr, C.int(data_size), C.int(format.Width), C.int(format.Height), C.long(timestamp))
		frame.Release()
		C.free(data_ptr)

		if ret < 0 {
			error_str := C.GoString(C.mediapipe_read_error())
			err_channel <- errors.New(error_str)
			C.mediapipe_free_error()
			break
		}

		/*
			end := time.Now().UnixMilli()
			diff := end - start
			fmt.Printf("%d\n", diff)
		*/
	}

	close(err_channel)
}

func (mp *MediaPipe) stop() error {
	mp.webcam.Close()
	mp.converter.end()

	ret := C.mediapipe_stop(mp.facem_lm)
	if ret < 0 {
		error_str := C.GoString(C.mediapipe_read_error())
		C.mediapipe_free_error()
		return errors.New(error_str)
	}

	C.free(unsafe.Pointer(mp.facem_path))

	return nil
}

//export mediapipe_call_facem_result
func mediapipe_call_facem_result(mp_result *C.struct_FaceLandmarkerResult, status C.int, timestamp C.long) {
	result := server.FaceTracking{}

	// Convert between C data types and structs to Go

	if mp_result.face_blendshapes_count != 0 {
		result.Blendshapes = make([]server.Blendshape, 0, int(mp_result.face_blendshapes.categories_count))

		for i := 0; i < int(mp_result.face_blendshapes.categories_count); i++ {
			mp_blendshape := C.face_landmarker_blendshape(mp_result, C.uint(i))

			blendshape := server.Blendshape{
				Index:        int(mp_blendshape.index),
				Score:        float32(mp_blendshape.score),
				CategoryName: C.GoString(mp_blendshape.category_name),
				DisplayName:  C.GoString(mp_blendshape.display_name),
			}

			result.Blendshapes = append(result.Blendshapes, blendshape)
		}
	}

	if mp_result.face_landmarks_count != 0 {
		result.Landmarks = make([]server.Landmark, 0, int(mp_result.face_landmarks.landmarks_count))

		for i := 0; i < int(mp_result.face_landmarks.landmarks_count); i++ {
			mp_landmark := C.face_landmarker_landmark(mp_result, C.uint(i))

			landmark := server.Landmark{
				X:             float32(mp_landmark.x),
				Y:             float32(mp_landmark.y),
				Z:             float32(mp_landmark.z),
				HasVisibility: bool(mp_landmark.has_visibility),
				Visibility:    float32(mp_landmark.visibility),
				HasPresence:   bool(mp_landmark.has_presence),
				Presence:      float32(mp_landmark.presence),
				Name:          C.GoString(mp_landmark.name),
			}

			result.Landmarks = append(result.Landmarks, landmark)
		}
	}

	result.Matrixes = make([]server.Matrix, 0, int(mp_result.facial_transformation_matrixes_count))

	for i := 0; i < int(mp_result.facial_transformation_matrixes_count); i++ {
		mp_matrix := C.face_landmarker_matrix(mp_result, C.uint(i))

		matrix := server.Matrix{
			Rows: uint32(mp_matrix.rows),
			Cols: uint32(mp_matrix.cols),
		}

		length := matrix.Rows * matrix.Cols
		matrix.Data = make([]float32, 0, length)

		for j := uint32(0); j < length; j++ {
			value := C.face_landmarker_matrix_data(&mp_matrix, C.uint(j))
			matrix.Data = append(matrix.Data, float32(value))
		}

		result.Matrixes = append(result.Matrixes, matrix)
	}

	result.Status = int(status)
	result.Timestamp = int(timestamp)

	// Send that sucker

	result.Type = uint8(server.FaceTrackingType)

	unlocked := ipc.mutex.TryLock()
	if !unlocked {
		fmt.Fprintln(os.Stderr, "[MP +TOAST] output is busy. dropping message...")
		return
	}

	if ipc.enabled {
		encoder := json.NewEncoder(ipc.socket)
		encoder.Encode(&result)

		newline := []byte("\n")
		ipc.socket.Write(newline)

	} else {
		text, err := json.Marshal(result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
		} else {
			fmt.Println(string(text))
		}
	}

	ipc.mutex.Unlock()
}
