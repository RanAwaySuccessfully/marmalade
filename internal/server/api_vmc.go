package server

import (
	"strings"

	"github.com/hypebeast/go-osc/osc"
	"github.com/ungerik/go3d/mat4"
	"github.com/ungerik/go3d/quaternion"
	"github.com/ungerik/go3d/vec3"
	"github.com/ungerik/go3d/vec4"
)

type VMCApi struct {
	client  *osc.Client
	asBones bool
	closed  bool
}

type VMCHandTarget struct {
	name   string
	parent int
	invert bool
}

var VMCHandMapping = map[int]VMCHandTarget{}

func init() {
	VMCHandMapping[0] = VMCHandTarget{name: "Hand", parent: 9}

	// might need to be 2-3-4?
	VMCHandMapping[1] = VMCHandTarget{name: "ThumbProximal"}
	VMCHandMapping[2] = VMCHandTarget{name: "ThumbIntermediate"}
	VMCHandMapping[3] = VMCHandTarget{name: "ThumbDistal"}

	VMCHandMapping[5] = VMCHandTarget{name: "IndexProximal"}
	VMCHandMapping[6] = VMCHandTarget{name: "IndexIntermediate"}
	VMCHandMapping[7] = VMCHandTarget{name: "IndexDistal"}

	VMCHandMapping[9] = VMCHandTarget{name: "MiddleProximal"}
	VMCHandMapping[10] = VMCHandTarget{name: "MiddleIntermediate"}
	VMCHandMapping[11] = VMCHandTarget{name: "MiddleDistal"}

	VMCHandMapping[13] = VMCHandTarget{name: "RingProximal"}
	VMCHandMapping[14] = VMCHandTarget{name: "RingIntermediate"}
	VMCHandMapping[15] = VMCHandTarget{name: "RingDistal"}

	VMCHandMapping[17] = VMCHandTarget{name: "LittleProximal"}
	VMCHandMapping[18] = VMCHandTarget{name: "LittleIntermediate"}
	VMCHandMapping[19] = VMCHandTarget{name: "LittleDistal"}
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

	//"/VMC/Ext/Root/Pos"
	//"root"

	if len(mp_data.FaceData.Matrixes) > 0 {
		matrix := mp_data.FaceData.Matrixes[0].Data
		rotations := format_rotation_quaternion(matrix)
		api.sendBone(1, "Head", -matrix[12], matrix[13], -matrix[14], -rotations[0], -rotations[1], -rotations[2], rotations[3])
	}

	var eyeLookOutLeft float32
	var eyeLookInLeft float32
	var eyeLookUpLeft float32
	var eyeLookDownLeft float32

	var eyeLookOutRight float32
	var eyeLookInRight float32
	var eyeLookUpRight float32
	var eyeLookDownRight float32

	for _, blendshape := range mp_data.FaceData.Blendshapes {
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

		api.sendBlendshape(category_name, blendshape.Score)
	}

	api.sendEye("LeftEye", (eyeLookDownLeft - eyeLookUpLeft), (eyeLookOutLeft - eyeLookInLeft), 0)
	api.sendEye("RightEye", (eyeLookDownRight - eyeLookUpRight), (eyeLookInRight - eyeLookOutRight), 0)

	for _, hand := range mp_data.HandData.Hand {
		if len(hand.Handedness) > 0 {
			handedness := hand.Handedness[0].CategoryName

			for i, landmark := range hand.WorldLandmarks {
				mapping, ok := VMCHandMapping[i]
				if !ok {
					continue
				}

				bone_name := handedness + mapping.name

				parent := hand.Landmarks[mapping.parent] // defaults to 0
				parent_bone := vec3.T{
					parent.X,
					parent.Y,
					parent.Z,
				}

				bone := vec3.T{
					landmark.X,
					landmark.Y,
					landmark.Z,
				}

				do_invert := mapping.invert

				var quat quaternion.T
				if do_invert {
					quat = quaternion.Vec3Diff(&bone, &parent_bone)
				} else {
					quat = quaternion.Vec3Diff(&parent_bone, &bone)
				}

				rotation_vector := quat.Vec4()
				rotation := rotation_vector.Slice()
				api.sendBone(2, bone_name, landmark.X, landmark.Y, landmark.Z, rotation[0], rotation[1], rotation[2], rotation[3])
			}

			println(handedness)
		}
	}

	applyMsg := osc.NewMessage("/VMC/Ext/Blend/Apply")
	api.client.Send(applyMsg)
}

func (api *VMCApi) Close() {
	api.client = nil
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
	/*
		msg := osc.NewMessage("/VMC/Ext/Set/Eye")
		//msg.Append(name)
		msg.Append(1)
		msg.Append(x)
		msg.Append(y)
		msg.Append(z)

		api.client.Send(msg)
	*/
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
