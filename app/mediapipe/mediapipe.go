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
	handm_lm   unsafe.Pointer
	handm_path *C.char
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

	delegate := 0
	if server.Config.UseGpu {
		delegate = 1
	}

	if server.Config.ModelFace != "" {
		mp.facem_path = C.CString(server.Config.ModelFace)
		mp.facem_lm = C.face_landmarker_start(mp.facem_path, C.int(delegate))
		if mp.facem_lm == nil {
			error_str := C.GoString(C.mediapipe_read_error())
			C.mediapipe_free_error()
			return errors.New(error_str)
		}
	}

	if server.Config.ModelHand != "" {
		mp.handm_path = C.CString(server.Config.ModelHand)
		mp.handm_lm = C.hand_landmarker_start(mp.handm_path, C.int(delegate))
		if mp.handm_lm == nil {
			error_str := C.GoString(C.mediapipe_read_error())
			C.mediapipe_free_error()
			return errors.New(error_str)
		}
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
		ret := C.mediapipe_detect(mp.facem_lm, mp.handm_lm, data_ptr, C.int(data_size), C.int(format.Width), C.int(format.Height), C.long(timestamp))
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
	if mp.webcam != nil {
		mp.webcam.Close()
	}

	if mp.converter != nil {
		mp.converter.end()
	}

	ret := C.mediapipe_stop(mp.facem_lm, mp.handm_lm)
	if ret < 0 {
		error_str := C.GoString(C.mediapipe_read_error())
		C.mediapipe_free_error()
		return errors.New(error_str)
	}

	if mp.facem_path != nil {
		C.free(unsafe.Pointer(mp.facem_path))
	}

	if mp.handm_path != nil {
		C.free(unsafe.Pointer(mp.handm_path))
	}

	return nil
}

func mediapipe_send_result(result any) {
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
