package server

import "github.com/hypebeast/go-osc/osc"

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
}

func (api *VRChatOSC) Send(mp_data *TrackingData, err_ch chan error) {
	if len(mp_data.FaceData.Matrixes) > 0 {
		matrix := mp_data.FaceData.Matrixes[0].Data
		y, x, z := format_rotation_angles(matrix)

		msg_rot := osc.NewMessage("/tracking/trackers/head/rotation")
		api.sendPosition("/tracking/trackers/head/position", -matrix[12], matrix[13], -matrix[14])
		api.sendRotation("/tracking/trackers/head/rotation", x, y, z)
		api.client.Send(msg_rot)
	}

	/*
		hip, chest, 2x feet, 2x knees, 2x elbows (upper arms)
		1 to 8
		/tracking/trackers/8/position
		/tracking/trackers/8/rotation
	*/
}

func (api *VRChatOSC) Close() {
	api.client = nil
}

func (api *VRChatOSC) sendPosition(msg_type string, px float32, py float32, pz float32) {
	scale := api.height / api.avatarHeight

	msg := osc.NewMessage(msg_type)
	msg.Append(px * scale)
	msg.Append(py * scale)
	msg.Append(pz * scale)
	api.client.Send(msg)
}

func (api *VRChatOSC) sendRotation(msg_type string, rx float32, ry float32, rz float32) {
	msg := osc.NewMessage(msg_type)
	msg.Append(rx)
	msg.Append(ry)
	msg.Append(rz)
	api.client.Send(msg)
}
