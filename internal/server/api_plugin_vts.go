package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"marmalade/internal/resources"
	"math/rand/v2"
	"sync"

	"github.com/coder/websocket"
	"github.com/ungerik/go3d/vec3"
)

type VTSPlugin struct {
	mutex         sync.Mutex
	conn          *websocket.Conn
	callbacks     map[string]*vtsPluginCallback
	closed        bool
	authenticated bool
}

type vtsPluginCallback func(map[string]any, chan error)

type vtsPluginMessage struct {
	ApiName     string `json:"apiName"`
	ApiVersion  string `json:"apiVersion"`
	Timestamp   int64  `json:"timestamp"`
	RequestID   string `json:"requestID"`
	MessageType string `json:"messageType"`
}

func (plugin *VTSPlugin) Listen(err_ch chan error) {
	plugin.callbacks = make(map[string]*vtsPluginCallback)
	plugin.authenticated = false
	plugin.closed = false

	port := "8001"
	if Config.VTSPlugin.Port != 0 {
		port = int_to_string(Config.VTSPlugin.Port)
	}

	var err error
	plugin.conn, _, err = websocket.Dial(context.Background(), "ws://localhost:"+port, nil)
	if err != nil {
		err_ch <- err
		return
	}
	defer plugin.conn.CloseNow()

	err = plugin.doAuth(false)
	if err != nil {
		err_ch <- err
		return
	}

	for !plugin.closed {
		_, reader, err := plugin.conn.Reader(context.Background())
		if err != nil {
			var ws_error websocket.CloseError
			if errors.As(err, &ws_error) {
				if ws_error.Code == websocket.StatusNormalClosure {
					return
				}
			}

			err_ch <- err
			return
		}

		var msg_map map[string]any

		dec := json.NewDecoder(reader)
		err = dec.Decode(&msg_map)
		if err != nil {
			err_ch <- err
			return
		}

		var msg vtsPluginMessage
		err = mapToStruct(msg_map, &msg)
		if err != nil {
			err_ch <- err
			return
		}

		req_id := msg.RequestID
		ptr := plugin.callbacks[req_id]
		if ptr != nil {
			callback := *ptr
			callback(msg_map, err_ch)
		} else {
			//fmt.Printf("[%d] %s\n", msg.Timestamp, msg.MessageType)
			// handle messages that don't fit into the req/res architecture
		}
	}
}

func (plugin *VTSPlugin) sendMessage(payload map[string]any, callback *vtsPluginCallback) error {
	req_id := fmt.Sprintf("%x", rand.Int())

	payload["apiName"] = "VTubeStudioPublicAPI"
	payload["apiVersion"] = "1.0"
	payload["requestID"] = req_id

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	err = plugin.conn.Write(context.Background(), websocket.MessageText, data)
	if err != nil {
		return err
	}

	if callback != nil {
		plugin.callbacks[req_id] = callback
	}

	return nil
}

func mapToStruct(input map[string]any, output any) error {
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonBytes, output)
	if err != nil {
		return err
	}

	return nil
}

// AUTH

type vtsPluginMessageAuthToken struct {
	Data struct {
		Token   string `json:"authenticationToken"`
		ErrorID int    `json:"errorID"`
		Message string `json:"message"`
	} `json:"data"`
}

type vtsPluginMessageAuth struct {
	Data struct {
		Authenticated bool   `json:"authenticated"`
		Reason        string `json:"reason"`
	} `json:"data"`
}

func (plugin *VTSPlugin) doAuth(forceFetchToken bool) error {
	payload_data := make(map[string]any)
	payload_data["pluginName"] = "Marmalade"
	payload_data["pluginDeveloper"] = "RanAwaySuccessfully"

	payload := make(map[string]any)
	payload["data"] = payload_data

	var callback vtsPluginCallback // function

	if Config.VTSPlugin.Token == "" || forceFetchToken {
		payload_data["pluginIcon"] = resources.EmbeddedIconLogoSmall
		payload["messageType"] = "AuthenticationTokenRequest"

		callback = plugin.handleAuthToken
	} else {
		payload_data["authenticationToken"] = Config.VTSPlugin.Token
		payload["messageType"] = "AuthenticationRequest"

		callback = plugin.handleAuth
	}

	return plugin.sendMessage(payload, &callback)
}

func (plugin *VTSPlugin) handleAuthToken(msg_map map[string]any, err_ch chan error) {
	var msg vtsPluginMessageAuthToken

	err := mapToStruct(msg_map, &msg)
	if err != nil {
		err_ch <- err
		return
	}

	if msg.Data.ErrorID != 0 {
		if msg.Data.ErrorID == 50 {
			Config.VTSPlugin.Token = ""
			Config.Save()
		}

		err_ch <- errors.New("VTube Studio Plugin: " + msg.Data.Message)
		return
	}

	Config.VTSPlugin.Token = msg.Data.Token
	Config.Save()

	err = plugin.doAuth(false)
	if err != nil {
		err_ch <- err
		return
	}
}

func (plugin *VTSPlugin) handleAuth(msg_map map[string]any, err_ch chan error) {
	var msg vtsPluginMessageAuth

	err := mapToStruct(msg_map, &msg)
	if err != nil {
		err_ch <- err
		return
	}

	if !msg.Data.Authenticated {
		err = plugin.doAuth(true)
		if err != nil {
			err_ch <- err
		}

		return
	}

	plugin.authenticated = true
}

// SEND MEDIAPIPE DATA

type vtsParameter struct {
	Id     string  `json:"id"`
	Weight float64 `json:"weight"`
	Value  float32 `json:"value"`
}

func (plugin *VTSPlugin) Send(mp_data *TrackingData, err_ch chan error) {
	if !plugin.authenticated {
		return
	}

	payload := make(map[string]any)
	payload["messageType"] = "InjectParameterDataRequest"

	// now the magic starts!

	payload_parameters := make([]vtsParameter, 0, 50)

	switch mp_data.Type {
	case FaceTrackingType:
		plugin.format_facem(mp_data.FaceData, &payload_parameters)
	case HandTrackingType:
		plugin.format_handm(mp_data.HandData, &payload_parameters)
	}

	/*
		BodyAngleX
		BodyAngleY
		BodyAngleZ
		BodyPositionX
		BodyPositionY
		BodyPositionZ
	*/

	payload_data := make(map[string]any)
	payload_data["parameterValues"] = payload_parameters
	payload_data["faceFound"] = len(mp_data.FaceData.Blendshapes) != 0
	payload_data["mode"] = "add"

	payload["data"] = payload_data
	err := plugin.sendMessage(payload, nil)
	if err != nil {
		err_ch <- err
	}
}

func (plugin *VTSPlugin) Close() {
	if plugin.conn != nil {
		plugin.conn.Close(websocket.StatusNormalClosure, "Closing plugin connection.")
	}

	plugin.closed = true
}

func (plugin *VTSPlugin) format_facem(mp_data FaceTracking, payload_parameters *[]vtsParameter) {
	if len(mp_data.Matrixes) > 0 {
		matrix := mp_data.Matrixes[0].Data
		y, x, z := format_rotation_angles(matrix)

		add_parameter(payload_parameters, "FaceAngleX", y)
		add_parameter(payload_parameters, "FaceAngleY", x)
		add_parameter(payload_parameters, "FaceAngleZ", -z)

		add_parameter(payload_parameters, "FacePositionX", -matrix[12])
		add_parameter(payload_parameters, "FacePositionY", matrix[13])
		add_parameter(payload_parameters, "FacePositionZ", -matrix[14])
	}

	var mouth_open float32

	var mouth_smile float32
	var mouth_x float32
	var mouth_shrug float32
	var mouth_press float32

	var brow_left float32
	var brow_right float32

	var eye_left_x float32
	var eye_right_x float32
	var eye_left_y float32
	var eye_right_y float32

	for _, blendshape := range mp_data.Blendshapes {
		switch blendshape.CategoryName {

		case "mouthClose":
			mouth_open -= blendshape.Score
		case "mouthPucker":
			add_parameter(payload_parameters, "MouthPucker", (blendshape.Score*2)-1)
		case "mouthSmileLeft":
			mouth_smile += blendshape.Score
		case "mouthSmileRight":
			mouth_smile += blendshape.Score
		case "mouthLeft":
			mouth_x -= blendshape.Score
		case "mouthRight":
			mouth_x += blendshape.Score
		case "mouthShrugLower":
			mouth_shrug -= blendshape.Score
		case "mouthShrugUpper":
			mouth_shrug += blendshape.Score
		case "mouthPressLeft":
			mouth_press -= blendshape.Score
		case "mouthPressRight":
			mouth_press += blendshape.Score
		case "mouthFunnel":
			add_parameter(payload_parameters, "MouthFunnel", blendshape.Score)

		case "browInnerUp":
			add_parameter(payload_parameters, "BrowInnerUp", blendshape.Score)
		case "browDownLeft":
			brow_left -= blendshape.Score
		case "browOuterUpLeft":
			brow_left += blendshape.Score
		case "browDownRight":
			brow_right -= blendshape.Score
		case "browOuterUpRight":
			brow_right += blendshape.Score

		case "eyeSquintLeft":
			add_parameter(payload_parameters, "EyeSquintL", blendshape.Score)
		case "eyeSquintRight":
			add_parameter(payload_parameters, "EyeSquintR", blendshape.Score)
		case "eyeBlinkLeft":
			add_parameter(payload_parameters, "EyeOpenLeft", -blendshape.Score+1)
		case "eyeBlinkRight":
			add_parameter(payload_parameters, "EyeOpenRight", -blendshape.Score+1)
		case "eyeLookDownLeft":
			eye_left_y += blendshape.Score
		case "eyeLookUpLeft":
			eye_left_y -= blendshape.Score
		case "eyeLookDownRight":
			eye_right_y += blendshape.Score
		case "eyeLookUpRight":
			eye_right_y -= blendshape.Score
		case "eyeLookInLeft":
			eye_left_x -= blendshape.Score
		case "eyeLookOutLeft":
			eye_left_x += blendshape.Score
		case "eyeLookInRight":
			eye_right_x += blendshape.Score
		case "eyeLookOutRight":
			eye_right_x -= blendshape.Score

		case "cheekPuff":
			add_parameter(payload_parameters, "CheekPuff", blendshape.Score)
		case "jawOpen":
			mouth_open += blendshape.Score
			add_parameter(payload_parameters, "JawOpen", blendshape.Score)
			//case "tongueOut":
			//	add_parameter(payload_parameters, "TongueOut", blendshape.Score)
		}
	}

	add_parameter(payload_parameters, "MouthOpen", mouth_open)
	add_parameter(payload_parameters, "MouthSmile", mouth_smile/2)

	add_parameter(payload_parameters, "BrowLeftY", brow_left)
	add_parameter(payload_parameters, "BrowRightY", brow_right)
	//add_parameter(payload_parameters, "Brows", 12.31)
	add_parameter(payload_parameters, "MouthX", mouth_x)
	add_parameter(payload_parameters, "MouthShrug", mouth_shrug/2)
	add_parameter(payload_parameters, "MouthPressLipOpen", mouth_press)

	add_parameter(payload_parameters, "EyeLeftX", eye_left_x)
	add_parameter(payload_parameters, "EyeLeftY", eye_left_y)
	add_parameter(payload_parameters, "EyeRightX", eye_right_x)
	add_parameter(payload_parameters, "EyeRightY", eye_right_y)
}

func (plugin *VTSPlugin) format_handm(mp_data HandTracking, payload_parameters *[]vtsParameter) {
	// ratio = distance(finger tip, hand wrist) / distance(finger base, hand wrist)

	left_hand_found := false
	right_hand_found := false

	for _, hand := range mp_data.Hand {
		if len(hand.Handedness) <= 0 {
			continue
		}

		handedness := hand.Handedness[0].CategoryName

		if handedness == "Left" {
			left_hand_found = true
		} else if handedness == "Right" {
			right_hand_found = true
		}

		wrist := hand.Landmarks[0]

		add_parameter(payload_parameters, "Hand"+handedness+"Found", 1)
		add_parameter(payload_parameters, "Hand"+handedness+"PositionX", wrist.X)
		add_parameter(payload_parameters, "Hand"+handedness+"PositionY", wrist.Y)
		add_parameter(payload_parameters, "Hand"+handedness+"PositionZ", wrist.Z)

		entire_hand := float32(0.5)
		thumb := distance_ratio(&wrist, &hand.Landmarks[1], &hand.Landmarks[4])
		index := distance_ratio(&wrist, &hand.Landmarks[5], &hand.Landmarks[8])
		middle := distance_ratio(&wrist, &hand.Landmarks[9], &hand.Landmarks[12])
		ring := distance_ratio(&wrist, &hand.Landmarks[13], &hand.Landmarks[16])
		pinky := distance_ratio(&wrist, &hand.Landmarks[17], &hand.Landmarks[20])

		add_parameter(payload_parameters, "Hand"+handedness+"Open", entire_hand)
		add_parameter(payload_parameters, "Hand"+handedness+"Finger_1_Thumb", thumb)
		add_parameter(payload_parameters, "Hand"+handedness+"Finger_2_Index", index)
		add_parameter(payload_parameters, "Hand"+handedness+"Finger_3_Middle", middle)
		add_parameter(payload_parameters, "Hand"+handedness+"Finger_4_Ring", ring)
		add_parameter(payload_parameters, "Hand"+handedness+"Finger_5_Pinky", pinky)
	}

	if left_hand_found && right_hand_found {
		add_parameter(payload_parameters, "BothHandsFound", 1)
	} else {
		add_parameter(payload_parameters, "BothHandsFound", 0)
	}

	if !left_hand_found {
		add_parameter(payload_parameters, "HandLeftFound", 0)
	}

	if !right_hand_found {
		add_parameter(payload_parameters, "HandRightFound", 0)
	}

	/*
		add_parameter(payload_parameters, "HandLeftAngleX", 12.31)
		add_parameter(payload_parameters, "HandLeftAngleZ", 12.31)

		add_parameter(payload_parameters, "HandRightAngleX", 12.31)
		add_parameter(payload_parameters, "HandRightAngleZ", 12.31)
	*/

	// add_parameter(payload_parameters, "HandDistance", 12.31)
}

func distance_ratio(wrist *Landmark, base *Landmark, tip *Landmark) float32 {
	wrist_vec := vec3.T{wrist.X, wrist.Y, wrist.Z}
	base_vec := vec3.T{base.X, base.Y, base.Z}
	tip_vec := vec3.T{tip.X, tip.Y, tip.Z}

	distance_tip := vec3.Distance(&tip_vec, &wrist_vec)
	distance_base := vec3.Distance(&base_vec, &wrist_vec)
	return distance_tip / distance_base
}

func add_parameter(slice *[]vtsParameter, id string, value float32) {
	*slice = append(*slice, vtsParameter{
		Id:     id,
		Weight: 1,
		Value:  value, // between -1000000 and 1000000
	})
}
