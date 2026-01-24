package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ungerik/go3d/mat4"
	"github.com/ungerik/go3d/vec4"
)

type VTSApi struct {
	mutex       sync.Mutex
	udpListener net.PacketConn
	clients     map[string]*VTSClient
	closed      bool
}

type VTSClient struct {
	source    string
	udpSender net.Conn
	message   vtsApiMessage
}

type vtsApiMessage struct {
	Type    string    `json:"messageType"`
	Time    float64   `json:"time"`
	SendFor float64   `json:"sendForSeconds"`
	SentBy  string    `json:"sentBy"`
	Ports   []float64 `json:"ports"`
}

func (api *VTSApi) listen(err_ch chan error) {
	api.clients = make(map[string]*VTSClient)
	api.closed = false

	port := ":21412"
	if Config.VTSApi.Port != 0 {
		port = ":" + int_to_string(int(Config.VTSApi.Port))
	}

	var err error
	api.udpListener, err = net.ListenPacket("udp", port)
	if err != nil {
		err_ch <- err
		return
	}

	go api.updateClients(err_ch)

	for !api.closed {
		buf := make([]byte, 1024)

		n, addr, err := api.udpListener.ReadFrom(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				continue
			}

			err_ch <- err
			continue
		}

		if n >= 1024 {
			continue
		}

		data := buf[:n]
		err = api.handleMessage(data, addr)
		if err != nil {
			err_ch <- err
		}
	}
}

func (api *VTSApi) handleMessage(buf []byte, addr net.Addr) error {
	var msg vtsApiMessage

	err := json.Unmarshal(buf, &msg)
	if err != nil {
		return err
	}

	if msg.Type != "iOSTrackingDataRequest" {
		return nil
	}

	if msg.Time == 0 {
		msg.Time = msg.SendFor
	}

	if msg.Time < 0.5 {
		msg.Time = 0.5
	}

	if msg.Time > 10 {
		msg.Time = 10
	}

	msg.Time *= 1000

	port := int_to_string(int(msg.Ports[0]))

	api.mutex.Lock()

	client := &VTSClient{}
	client.source = addr.String()
	client.udpSender, err = net.Dial("udp", ":"+port)
	client.message = msg

	api.clients[msg.SentBy] = client

	api.mutex.Unlock()

	return err
}

func (api *VTSApi) send(mp_data *TrackingData, err_ch chan error) {
	api_data, err := format_vts_api_data(&mp_data.FaceData, mp_data.Timestamp)
	if err != nil {
		err_ch <- err
		return
	}

	api.mutex.Lock()

	for _, client := range api.clients {
		_, err = client.udpSender.Write(api_data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[MARMALADE] unable to send packet %v\n", err)
		}
	}

	api.mutex.Unlock()
}

func (api *VTSApi) close() {
	if api.udpListener != nil {
		api.udpListener.Close()
	}

	api.closed = true
}

func (api *VTSApi) updateClients(err_ch chan error) {
	for !api.closed {

		start := time.Now().UnixMilli()

		min := int64(100)

		api.mutex.Lock()

		for clientId, client := range api.clients {
			if client.message.Time <= 0 {
				delete(api.clients, clientId)
				err := client.udpSender.Close()
				if err != nil {
					err_ch <- err
				}

				continue
			}

			client.message.Time -= float64(min)
		}

		api.mutex.Unlock()

		end := time.Now().UnixMilli()
		diff := end - start

		if diff < min {
			waitFor := time.Duration(min - diff)
			time.Sleep(waitFor * time.Millisecond)
		}
	}

	api.mutex.Lock()

	for clientId := range api.clients {
		delete(api.clients, clientId)
	}

	api.mutex.Unlock()
}

func format_vts_api_data(mp_data *FaceTracking, timestamp int) ([]byte, error) {
	blendshape_count := len(mp_data.Blendshapes)

	var eyeLookOutLeft float32
	var eyeLookInLeft float32
	var eyeLookUpLeft float32
	var eyeLookDownLeft float32

	var eyeLookOutRight float32
	var eyeLookInRight float32
	var eyeLookUpRight float32
	var eyeLookDownRight float32

	blendshapes := make([]any, 0, blendshape_count)
	for i := 0; i < blendshape_count; i++ {

		blendshape := make(map[string]any)
		category_name := mp_data.Blendshapes[i].CategoryName
		category_name = strings.ToUpper(string(category_name[0])) + category_name[1:]

		// left/right is switched between MediaPipe and VTube Studio parameters
		length := len(category_name)
		isLeft := category_name[length-4:] == "Left"
		isRight := category_name[length-5:] == "Right"

		if isLeft {
			category_name = strings.Replace(category_name, "Left", "Right", 1)
		}

		if isRight {
			category_name = strings.Replace(category_name, "Right", "Left", 1)
		}

		score := mp_data.Blendshapes[i].Score

		switch category_name {
		case "EyeLookOutLeft":
			eyeLookOutLeft = score
		case "EyeLookInLeft":
			eyeLookInLeft = score
		case "EyeLookUpLeft":
			eyeLookUpLeft = score
		case "EyeLookDownLeft":
			eyeLookDownLeft = score
		case "EyeLookOutRight":
			eyeLookOutRight = score
		case "EyeLookInRight":
			eyeLookInRight = score
		case "EyeLookUpRight":
			eyeLookUpRight = score
		case "EyeLookDownRight":
			eyeLookDownRight = score
		}

		blendshape["k"] = category_name
		blendshape["v"] = score
		blendshapes = append(blendshapes, blendshape)
	}

	// VTube Studio considers X=vertical and Y=horizontal
	rotation := make(map[string]any)
	position := make(map[string]any)

	if len(mp_data.Matrixes) > 0 {
		matrix := mp_data.Matrixes[0].Data

		rotationMatrix := mat4.T{
			vec4.T{matrix[0], matrix[4], matrix[8], matrix[12]},
			vec4.T{matrix[1], matrix[5], matrix[9], matrix[13]},
			vec4.T{matrix[2], matrix[6], matrix[10], matrix[14]},
			vec4.T{matrix[3], matrix[7], matrix[11], matrix[15]},
		}

		y, x, z := rotationMatrix.ExtractEulerAngles()

		rotation["x"] = y * (180 / math.Pi)
		rotation["y"] = -x * (180 / math.Pi)
		rotation["z"] = z * (180 / math.Pi)

		position["x"] = matrix[12]
		position["y"] = matrix[13]
		position["z"] = -matrix[14]
	}

	eye_left := make(map[string]any)
	eye_left["x"] = (eyeLookDownLeft - eyeLookUpLeft) * 20
	eye_left["y"] = (eyeLookOutLeft - eyeLookInLeft) * 20
	eye_left["z"] = 0

	eye_right := make(map[string]any)
	eye_right["x"] = (eyeLookDownRight - eyeLookUpRight) * 20
	eye_right["y"] = (eyeLookInRight - eyeLookOutRight) * 20
	eye_right["z"] = 0

	payload := make(map[string]any)
	payload["Timestamp"] = timestamp
	payload["Hotkey"] = -1
	payload["FaceFound"] = blendshape_count != 0
	payload["BlendShapes"] = blendshapes
	payload["Rotation"] = rotation
	payload["Position"] = position
	payload["EyeLeft"] = eye_left
	payload["EyeRight"] = eye_right

	return json.Marshal(payload)
}
