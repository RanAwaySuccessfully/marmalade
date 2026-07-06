package server

import (
	"github.com/hypebeast/go-osc/osc"
	"github.com/ungerik/go3d/vec3"
)

type VRChatOSC struct {
	client       *osc.Client
	closed       bool
	height       float32
	avatarHeight float32
}

func (api *VRChatOSC) Listen(err_ch chan error) {
	port := 9000
	if Config.VRChatOSC.Port != 0 {
		port = Config.VRChatOSC.Port
	}

	api.client = osc.NewClient("127.0.0.1", port)

	api.height = 1.7
	api.avatarHeight = 1.87
	// TODO: these parameters need to be configurable...
}

func (api *VRChatOSC) Send(mp_data *TrackingData, err_ch chan error) {
	if len(mp_data.FaceData.Matrixes) > 0 {
		matrix := mp_data.FaceData.Matrixes[0].Data
		y, x, z := format_rotation_angles(matrix)

		api.sendPosition("/tracking/trackers/head/position", matrix[12], matrix[13], matrix[14])
		api.sendRotation("/tracking/trackers/head/rotation", x, y, z)
	}

	if len(mp_data.PoseData.WorldLandmarks) > 0 {
		left_hip := mp_data.PoseData.WorldLandmarks[23]
		right_hip := mp_data.PoseData.WorldLandmarks[24]
		arr_hip := vec3.T{right_hip.X, right_hip.Y, right_hip.Z}

		arr_hip.Add(&vec3.T{left_hip.X, left_hip.Y, left_hip.Z})
		arr_hip.Scale(0.5)
		api.sendPosition("/tracking/trackers/1/position", arr_hip[0], arr_hip[1], arr_hip[2])
		//api.sendRotation("/tracking/trackers/1/rotation", 0.5, 0.5, 0.5)

		left_shoulder := mp_data.PoseData.WorldLandmarks[23]
		right_shoulder := mp_data.PoseData.WorldLandmarks[24]
		arr_chest := vec3.T{right_shoulder.X, right_shoulder.Y, right_shoulder.Z}
		arr_chest.Add(&vec3.T{left_shoulder.X, left_shoulder.Y, left_shoulder.Z})
		arr_chest.Scale(0.5)

		arr_chest.Add(&arr_hip)
		arr_chest.Scale(0.5)
		api.sendPosition("/tracking/trackers/2/position", arr_chest[0], arr_chest[1], arr_chest[2])
		//api.sendRotation("/tracking/trackers/2/rotation", 0.5, 0.5, 0.5)

		left_feet := mp_data.PoseData.WorldLandmarks[27]
		right_feet := mp_data.PoseData.WorldLandmarks[28]
		api.sendPosition("/tracking/trackers/3/position", left_feet.X, left_feet.Y, left_feet.Z)
		api.sendPosition("/tracking/trackers/4/position", right_feet.X, right_feet.Y, right_feet.Z)
		//api.sendRotation("/tracking/trackers/3/rotation", 0.5, 0.5, 0.5)
		//api.sendRotation("/tracking/trackers/4/rotation", 0.5, 0.5, 0.5)

		left_knee := mp_data.PoseData.WorldLandmarks[25]
		right_knee := mp_data.PoseData.WorldLandmarks[26]
		api.sendPosition("/tracking/trackers/5/position", left_knee.X, left_knee.Y, left_knee.Z)
		api.sendPosition("/tracking/trackers/6/position", right_knee.X, right_knee.Y, right_knee.Z)
		//api.sendRotation("/tracking/trackers/5/rotation", 0.5, 0.5, 0.5)
		//api.sendRotation("/tracking/trackers/6/rotation", 0.5, 0.5, 0.5)

		left_elbow := mp_data.PoseData.WorldLandmarks[13]
		right_elbow := mp_data.PoseData.WorldLandmarks[14]
		api.sendPosition("/tracking/trackers/7/position", left_elbow.X, left_elbow.Y, left_elbow.Z)
		api.sendPosition("/tracking/trackers/8/position", right_elbow.X, right_elbow.Y, right_elbow.Z)
		//api.sendRotation("/tracking/trackers/7/rotation", 0.5, 0.5, 0.5)
		//api.sendRotation("/tracking/trackers/8/rotation", 0.5, 0.5, 0.5)
	}
}

func (api *VRChatOSC) Close() {
	api.client = nil
}

func (api *VRChatOSC) sendPosition(msg_type string, px float32, py float32, pz float32) {
	scale := api.height / api.avatarHeight

	msg := osc.NewMessage(msg_type)
	msg.Append(-px * scale)
	msg.Append(py * scale)
	msg.Append(-pz * scale)
	api.client.Send(msg)
}

func (api *VRChatOSC) sendRotation(msg_type string, rx float32, ry float32, rz float32) {
	msg := osc.NewMessage(msg_type)
	msg.Append(rx)
	msg.Append(ry)
	msg.Append(rz)
	api.client.Send(msg)
}
