package server

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// MEDIAPIPE TRACKING DATA

const (
	NullTrackingType = iota
	FaceTrackingType
	HandTrackingType
	PoseTrackingType
	HolisticTrackingType
)

type TrackingData struct {
	Type      uint8 `json:"type"`
	Status    int   `json:"status"`
	Timestamp int   `json:"timestamp"`
	FaceData  FaceTracking
	HandData  HandTracking
	PoseData  PoseTracking
}

type Category struct {
	Index        int     `json:"index"`
	Score        float32 `json:"score"`
	CategoryName string  `json:"category_name"`
	DisplayName  string  `json:"display_name"`
}

type Landmark struct {
	X             float32 `json:"x"`
	Y             float32 `json:"y"`
	Z             float32 `json:"z"`
	HasVisibility bool    `json:"has_visibility"`
	Visibility    float32 `json:"visibility"`
	HasPresence   bool    `json:"has_presence"`
	Presence      float32 `json:"presence"`
	Name          string  `json:"name"`
}

// FACE TRACKING

type FaceTracking struct {
	Blendshapes []Category `json:"blendshapes"`
	Landmarks   []Landmark `json:"landmarks"`
	Matrixes    []Matrix   `json:"matrixes"`
}

type Matrix struct {
	Rows uint32    `json:"rows"`
	Cols uint32    `json:"cols"`
	Data []float32 `json:"data"`
}

// HAND TRACKING

type HandTracking struct {
	Hand []Hand `json:"hands"`
}

type Hand struct {
	Handedness     []Category `json:"handedness"`
	Landmarks      []Landmark `json:"landmarks"`
	WorldLandmarks []Landmark `json:"world_landmarks"`
}

// POSE TRACKING

type PoseTracking struct {
	Landmarks      []Landmark `json:"landmarks"`
	WorldLandmarks []Landmark `json:"world_landmarks"`
}

// PROCESS

type MediaPipeProcess struct {
	cmd    *exec.Cmd
	socket net.Listener
}

func (mp *MediaPipeProcess) create(err_pipe io.Writer) error {
	_, err := os.Stat("./mediapipe")
	if errors.Is(err, os.ErrNotExist) { // local testing
		build_cmd := exec.Command("go", "build")
		build_cmd.Dir = "./app/mediapipe"

		err := build_cmd.Run()
		if err != nil {
			return err
		}

		mp.cmd = exec.Command("./app/mediapipe/mediapipe", "--ipc")
		env := mp.cmd.Environ()

		library_path := "LD_LIBRARY_PATH=./app/mediapipe/cc"

		for i := 0; i < len(env); i++ {
			env_var := env[i]
			isLibraryPath := strings.HasPrefix(env_var, "LD_LIBRARY_PATH=")

			if isLibraryPath {
				library_path += ":" + env_var[16:]
			}
		}

		mp.cmd.Env = append(env, library_path)

	} else {
		mp.cmd = exec.Command("./mediapipe", "--ipc")
	}

	if Config.HwAccel.PrimeId != "" {
		replacer := strings.NewReplacer(":", "_", ".", "_")
		prime_id := replacer.Replace(Config.HwAccel.PrimeId)
		prime_env := "DRI_PRIME=" + prime_id
		mp.cmd.Env = append(mp.cmd.Environ(), prime_env)
	}

	stdout, err := mp.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := mp.cmd.StderrPipe()
	if err != nil {
		return err
	}

	go io.Copy(err_pipe, stderr)
	go io.Copy(os.Stdout, stdout)

	err = mp.cmd.Start()
	if err != nil {
		return err
	}

	return nil
}

func (mp *MediaPipeProcess) createSocket() error {
	err := os.Remove("marmalade.sock")
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	mp.socket, err = net.Listen("unix", "marmalade.sock")
	if err != nil {
		return err
	}

	return nil
}

func (mp *MediaPipeProcess) wait(err_ch chan error) {
	err := mp.cmd.Wait()
	if err != nil {
		err_ch <- err
	} else {
		err_ch <- os.ErrProcessDone
	}
}

func (mp *MediaPipeProcess) listen(exit *bool, err_ch chan error, result chan TrackingData) {
	conn, err := mp.socket.Accept()
	if err != nil {
		err_ch <- err
		return
	}

	fmt.Println("[MARMALADE] MediaPipe connection started")
	decoder := gob.NewDecoder(conn)

	for !*exit {
		var mp_data TrackingData

		err := decoder.Decode(&mp_data)
		if err != nil {
			if err != io.EOF {
				err_ch <- err
			}

			break
		}

		result <- mp_data
	}

	close(result)
}

func (mp *MediaPipeProcess) close() {
	if mp.cmd != nil && mp.cmd.Process != nil {
		err := mp.cmd.Process.Signal(syscall.SIGTERM)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
		}

		mp.cmd = nil
	}

	if mp.socket != nil {
		mp.socket.Close()
	}

	os.Remove("marmalade.sock")
}
