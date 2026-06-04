package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hypebeast/go-osc/osc"
	"github.com/ungerik/go3d/mat4"
	"github.com/ungerik/go3d/quaternion"
	"github.com/ungerik/go3d/vec4"
)

type VMCApi struct {
	client        *osc.Client
	asBones       bool
	closed        bool
	frame_counter int
}

type KalidoKitData struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

type KalidoKitHand struct {
	Wrist              KalidoKitData
	RingProximal       KalidoKitData
	RingIntermediate   KalidoKitData
	RingDistal         KalidoKitData
	IndexProximal      KalidoKitData
	IndexIntermediate  KalidoKitData
	IndexDistal        KalidoKitData
	MiddleProximal     KalidoKitData
	MiddleIntermediate KalidoKitData
	MiddleDistal       KalidoKitData
	ThumbProximal      KalidoKitData
	ThumbIntermediate  KalidoKitData
	ThumbDistal        KalidoKitData
	LittleProximal     KalidoKitData
	LittleIntermediate KalidoKitData
	LittleDistal       KalidoKitData
}

type KalidoKitPose struct {
	RightUpperArm KalidoKitData
	RightLowerArm KalidoKitData
	LeftUpperArm  KalidoKitData
	LeftLowerArm  KalidoKitData
	RightHand     KalidoKitData
	LeftHand      KalidoKitData
	RightUpperLeg KalidoKitData
	RightLowerLeg KalidoKitData
	LeftUpperLeg  KalidoKitData
	LeftLowerLeg  KalidoKitData
	Spine         KalidoKitData
	Hips          struct {
		Position      KalidoKitData `json:"position"`
		WorldPosition KalidoKitData `json:"worldPosition"`
		Rotation      KalidoKitData `json:"rotation"`
	}
}

func (api *VMCApi) Listen(err_ch chan error) {
	port := 39540
	if Config.VMCApi.Port != 0 {
		port = Config.VMCApi.Port
	}

	api.client = osc.NewClient("127.0.0.1", port)
}

func (api *VMCApi) Send(mp_data *TrackingData, err_ch chan error) {
	if api.client == nil {
		return
	}

	api.send_face_data(&mp_data.FaceData)
	api.send_hand_data(&mp_data.HandData)
	api.send_pose_data(&mp_data.PoseData)
}

func (api *VMCApi) Close() {
	api.client = nil
}

func (api *VMCApi) send_face_data(face_data *FaceTracking) {
	if len(face_data.Matrixes) > 0 {
		matrix := face_data.Matrixes[0].Data
		rotations := format_rotation_quaternion(matrix)
		api.sendBone(1, "Head", -matrix[12], matrix[13], -matrix[14], -rotations[0], -rotations[1], -rotations[2], rotations[3])
	}

	hasEyeLeft := false
	var eyeLookOutLeft float32
	var eyeLookInLeft float32
	var eyeLookUpLeft float32
	var eyeLookDownLeft float32

	hasEyeRight := false
	var eyeLookOutRight float32
	var eyeLookInRight float32
	var eyeLookUpRight float32
	var eyeLookDownRight float32

	for _, blendshape := range face_data.Blendshapes {
		category_name := blendshape.CategoryName
		category_name = strings.ToUpper(string(category_name[0])) + category_name[1:]

		// switch left/right
		length := len(category_name)
		isLeft := category_name[length-4:] == "Left"
		isRight := category_name[length-5:] == "Right"

		if isLeft {
			category_name = strings.Replace(category_name, "Left", "Right", 1)
		}

		if isRight {
			category_name = strings.Replace(category_name, "Right", "Left", 1)
		}

		score := blendshape.Score

		switch category_name {
		case "EyeLookOutLeft":
			eyeLookOutLeft = score
			hasEyeLeft = true
		case "EyeLookInLeft":
			eyeLookInLeft = score
			hasEyeLeft = true
		case "EyeLookUpLeft":
			eyeLookUpLeft = score
			hasEyeLeft = true
		case "EyeLookDownLeft":
			eyeLookDownLeft = score
			hasEyeLeft = true
		case "EyeLookOutRight":
			eyeLookOutRight = score
			hasEyeRight = true
		case "EyeLookInRight":
			eyeLookInRight = score
			hasEyeRight = true
		case "EyeLookUpRight":
			eyeLookUpRight = score
			hasEyeRight = true
		case "EyeLookDownRight":
			eyeLookDownRight = score
			hasEyeRight = true
		}

		api.sendBlendshape(category_name, score)
	}

	if hasEyeLeft {
		api.sendEye("LeftEye", (eyeLookDownLeft - eyeLookUpLeft), (eyeLookOutLeft - eyeLookInLeft), 0)
	}

	if hasEyeRight {
		api.sendEye("RightEye", (eyeLookDownRight - eyeLookUpRight), (eyeLookInRight - eyeLookOutRight), 0)
	}
}

func (api *VMCApi) send_hand_data(hand_data *HandTracking) {
	for _, hand := range hand_data.Hand {
		if len(hand.Handedness) <= 0 {
			continue
		}

		handedness := hand.Handedness[0].CategoryName

		if len(hand.WorldLandmarks) <= 0 {
			return
		}

		payload := map[string]any{}
		payload["handedness"] = handedness
		payload["landmarks"] = hand.WorldLandmarks // hand.WorldLandmarks

		jsonData, err := json.Marshal(payload)
		if err != nil {
			println("oops!")
			return
		}

		response := api.requestKalidokitSolve("hand", bytes.NewReader(jsonData))
		defer response.Close()

		var hand_rot KalidoKitHand
		decoder := json.NewDecoder(response)
		decoder.Decode(&hand_rot)

		api.sendKalidokitBone(2, handedness+"Wrist", &hand_rot.Wrist, nil)
		api.sendKalidokitBone(2, handedness+"RingProximal", &hand_rot.RingProximal, nil)
		api.sendKalidokitBone(2, handedness+"RingIntermediate", &hand_rot.RingIntermediate, nil)
		api.sendKalidokitBone(2, handedness+"RingDistal", &hand_rot.RingDistal, nil)
		api.sendKalidokitBone(2, handedness+"IndexProximal", &hand_rot.IndexProximal, nil)
		api.sendKalidokitBone(2, handedness+"IndexIntermediate", &hand_rot.IndexIntermediate, nil)
		api.sendKalidokitBone(2, handedness+"IndexDistal", &hand_rot.IndexDistal, nil)
		api.sendKalidokitBone(2, handedness+"MiddleProximal", &hand_rot.MiddleProximal, nil)
		api.sendKalidokitBone(2, handedness+"MiddleIntermediate", &hand_rot.MiddleIntermediate, nil)
		api.sendKalidokitBone(2, handedness+"MiddleDistal", &hand_rot.MiddleDistal, nil)
		api.sendKalidokitBone(2, handedness+"ThumbProximal", &hand_rot.ThumbProximal, nil)
		api.sendKalidokitBone(2, handedness+"ThumbIntermediate", &hand_rot.ThumbIntermediate, nil)
		api.sendKalidokitBone(2, handedness+"ThumbDistal", &hand_rot.ThumbDistal, nil)
		api.sendKalidokitBone(2, handedness+"LittleProximal", &hand_rot.LittleProximal, nil)
		api.sendKalidokitBone(2, handedness+"LittleIntermediate", &hand_rot.LittleIntermediate, nil)
		api.sendKalidokitBone(2, handedness+"LittleDistal", &hand_rot.LittleDistal, nil)
	}
}

func (api *VMCApi) send_pose_data(pose_data *PoseTracking) {
	if len(pose_data.WorldLandmarks) <= 0 {
		return
	}

	jsonData, err := json.Marshal(pose_data)
	if err != nil {
		println("oops!")
		return
	}

	response := api.requestKalidokitSolve("pose", bytes.NewReader(jsonData))
	defer response.Close()

	var pose KalidoKitPose
	decoder := json.NewDecoder(response)
	decoder.Decode(&pose)

	api.sendKalidokitBone(3, "RightUpperArm", &pose.RightUpperArm, nil)
	api.sendKalidokitBone(3, "RightLowerArm", &pose.RightLowerArm, nil)
	api.sendKalidokitBone(3, "LeftUpperArm", &pose.LeftUpperArm, nil)
	api.sendKalidokitBone(3, "LeftLowerArm", &pose.LeftLowerArm, nil)
	api.sendKalidokitBone(3, "RightHand", &pose.RightHand, nil)
	api.sendKalidokitBone(3, "LeftHand", &pose.LeftHand, nil)
	api.sendKalidokitBone(3, "RightUpperLeg", &pose.RightUpperLeg, nil)
	api.sendKalidokitBone(3, "RightLowerLeg", &pose.RightLowerLeg, nil)
	api.sendKalidokitBone(3, "LeftUpperLeg", &pose.LeftUpperLeg, nil)
	api.sendKalidokitBone(3, "LeftLowerLeg", &pose.LeftLowerLeg, nil)
	api.sendKalidokitBone(3, "Spine", &pose.Spine, nil)
	api.sendKalidokitBone(3, "Hips", &pose.Hips.Rotation, &pose.Hips.WorldPosition)
}

func (api *VMCApi) sendBlendshape(blendshape string, value float32) {
	if api.client == nil {
		return
	}

	msg := osc.NewMessage("/VMC/Ext/Blend/Val")
	msg.Append(blendshape)
	msg.Append(value)
	api.client.Send(msg)
}

func (api *VMCApi) sendEye(name string, x float32, y float32, z float32) {
	msg := osc.NewMessage("/VMC/Ext/Set/Eye")
	//msg.Append(name)
	msg.Append(1)
	msg.Append(x)
	msg.Append(y)
	msg.Append(z)

	api.client.Send(msg)
}

func (api *VMCApi) requestKalidokitSolve(path string, data io.Reader) io.ReadCloser {
	req, err := http.NewRequest("POST", "http://localhost:4242/"+path, data)
	if err != nil {
		println("oops!2")
		return nil
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		println("oops!3")
		return nil
	}

	return resp.Body
}

func (api *VMCApi) sendKalidokitBone(msg_type uint8, bone_name string, rotation *KalidoKitData, position *KalidoKitData) {
	rot := quaternion.FromEulerAngles(-rotation.Y, rotation.X, rotation.Z)

	if position == nil {
		api.sendBone(msg_type, bone_name, 0, 0, 0, rot[0], rot[1], rot[2], rot[3])
	} else {
		api.sendBone(msg_type, bone_name, position.X, position.Y, position.Z, rot[0], rot[1], rot[2], rot[3])
	}
}

func (api *VMCApi) sendBone(msg_type uint8, bone string, px float32, py float32, pz float32, qx float32, qy float32, qz float32, qw float32) {
	if api.client == nil {
		return
	}

	var msg *osc.Message

	api.asBones = true
	if api.asBones {
		msg = osc.NewMessage("/VMC/Ext/Bone/Pos")
	} else {
		switch msg_type {
		case 1: // head-mounted display
			msg = osc.NewMessage("/VMC/Ext/Hmd/Pos")
		case 2: // controllers
			msg = osc.NewMessage("/VMC/Ext/Con/Pos")
		case 3: // trackers
			msg = osc.NewMessage("/VMC/Ext/Tra/Pos")
		}
	}

	msg.Append(bone)
	msg.Append(px)
	msg.Append(py)
	msg.Append(pz)
	msg.Append(qx)
	msg.Append(qy)
	msg.Append(qz)
	msg.Append(qw)

	api.client.Send(msg)
}

func format_rotation_quaternion(matrix []float32) []float32 {
	rotationMatrix := mat4.T{
		vec4.T{matrix[0], matrix[4], matrix[8], matrix[12]},
		vec4.T{matrix[1], matrix[5], matrix[9], matrix[13]},
		vec4.T{matrix[2], matrix[6], matrix[10], matrix[14]},
		vec4.T{matrix[3], matrix[7], matrix[11], matrix[15]},
	}

	quaternion := rotationMatrix.Quaternion()
	vector := quaternion.Vec4()
	return vector.Slice()
}
