package server

import (
	"strings"

	"github.com/hypebeast/go-osc/osc"
)

type VMCApi struct {
	client *osc.Client
	closed bool
}

func (api *VMCApi) Listen(err_ch chan error) {
	port := 39540
	if Config.VMCApi.Port != 0 {
		port = Config.VMCApi.Port
	}

	api.client = osc.NewClient("127.0.0.1", port)
}

func (api *VMCApi) Send(mp_data *TrackingData, err_ch chan error) {
	//api.sendBlendshape("Blink_L", 0.5) // 0.0 to 1.0

	for _, blendshape := range mp_data.FaceData.Blendshapes {
		category_name := blendshape.CategoryName
		category_name = strings.ToUpper(string(category_name[0])) + category_name[1:]

		// switch left and right

		api.sendBlendshape(category_name, blendshape.Score)
	}

	applyMsg := osc.NewMessage("/VMC/Ext/Blend/Apply")
	api.client.Send(applyMsg)
}

func (api *VMCApi) Close() {

}

func (api *VMCApi) sendBlendshape(blendshape string, value float32) {
	msg := osc.NewMessage("/VMC/Ext/Blend/Val")
	msg.Append(blendshape)
	msg.Append(value)
	api.client.Send(msg)
}
