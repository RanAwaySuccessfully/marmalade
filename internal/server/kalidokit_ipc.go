package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"syscall"
)

// KALIDOKIT DATA

const (
	NullKalidoKitType = iota
	HandKalidoKitType
	PoseKalidoKitType
)

type KalidoKitData struct {
	Type          uint8
	LeftHandData  KalidoKitHand
	RightHandData KalidoKitHand
	PoseData      KalidoKitPose
}

type KalidoKitCoords struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

type KalidoKitHand struct {
	Wrist              KalidoKitCoords
	RingProximal       KalidoKitCoords
	RingIntermediate   KalidoKitCoords
	RingDistal         KalidoKitCoords
	IndexProximal      KalidoKitCoords
	IndexIntermediate  KalidoKitCoords
	IndexDistal        KalidoKitCoords
	MiddleProximal     KalidoKitCoords
	MiddleIntermediate KalidoKitCoords
	MiddleDistal       KalidoKitCoords
	ThumbProximal      KalidoKitCoords
	ThumbIntermediate  KalidoKitCoords
	ThumbDistal        KalidoKitCoords
	LittleProximal     KalidoKitCoords
	LittleIntermediate KalidoKitCoords
	LittleDistal       KalidoKitCoords
}

type KalidoKitPose struct {
	RightUpperArm KalidoKitCoords
	RightLowerArm KalidoKitCoords
	LeftUpperArm  KalidoKitCoords
	LeftLowerArm  KalidoKitCoords
	RightHand     KalidoKitCoords
	LeftHand      KalidoKitCoords
	RightUpperLeg KalidoKitCoords
	RightLowerLeg KalidoKitCoords
	LeftUpperLeg  KalidoKitCoords
	LeftLowerLeg  KalidoKitCoords
	Spine         KalidoKitCoords
	Hips          struct {
		Position      KalidoKitCoords `json:"position"`
		WorldPosition KalidoKitCoords `json:"worldPosition"`
		Rotation      KalidoKitCoords `json:"rotation"`
	}
}

type KalidoKitProcess struct {
	cmd     *exec.Cmd
	socket  net.Listener
	encoder *json.Encoder
}

func (ka *KalidoKitProcess) create() error {
	ka.cmd = exec.Command("./kalidokit-bin")

	stdout, err := ka.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := ka.cmd.StderrPipe()
	if err != nil {
		return err
	}

	go io.Copy(os.Stderr, stderr)
	go io.Copy(os.Stdout, stdout)

	err = ka.cmd.Start()
	if err != nil {
		return err
	}

	return nil
}

func (ka *KalidoKitProcess) createSocket() error {
	err := os.Remove("kalidokit.sock")
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	ka.socket, err = net.Listen("unix", "kalidokit.sock")
	if err != nil {
		return err
	}

	return nil
}

func (ka *KalidoKitProcess) wait(err_ch chan error) {
	err := ka.cmd.Wait()
	if err != nil {
		err_ch <- err
	} else {
		err_ch <- os.ErrProcessDone
	}
}

func (ka *KalidoKitProcess) listen(exit *bool, err_ch chan error, result chan KalidoKitData) {
	conn, err := ka.socket.Accept()
	if err != nil {
		err_ch <- err
		return
	}

	fmt.Println("[MARMALADE] KalidoKit connection started")
	ka.encoder = json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	for !*exit {
		var ka_data KalidoKitData

		err := decoder.Decode(&ka_data)
		if err != nil {
			if err != io.EOF {
				err_ch <- err
			}

			break
		}

		result <- ka_data
	}

	close(result)
}

func (ka *KalidoKitProcess) send(value any) error {
	// TODO: should probably receive a JSON with a type field, so that verification on the JS side is easier
	return ka.encoder.Encode(value)
}

func (ka *KalidoKitProcess) close() {
	if ka.cmd.Process != nil {
		err := ka.cmd.Process.Signal(syscall.SIGTERM)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
		}

		ka.cmd = nil
	}

	if ka.socket != nil {
		ka.socket.Close()
	}

	os.Remove("kalidokit.sock")
}
