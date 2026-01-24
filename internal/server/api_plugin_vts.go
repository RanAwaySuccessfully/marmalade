package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"marmalade/internal/resources"
	"math/rand/v2"
	"strconv"
	"sync"

	"github.com/coder/websocket"
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

func (plugin *VTSPlugin) listen(err_ch chan error) {
	plugin.callbacks = make(map[string]*vtsPluginCallback)
	plugin.authenticated = false
	plugin.closed = false

	port := "8001"
	if Config.VTSPlugin.Port != 0 {
		port = strconv.Itoa(Config.VTSPlugin.Port) // convert int to string
	}

	var err error
	plugin.conn, _, err = websocket.Dial(context.Background(), "ws://localhost:"+port, nil)
	if err != nil {
		err_ch <- err
		return
	}
	defer plugin.conn.CloseNow()

	plugin.doAuth(false)

	for !plugin.closed {
		go plugin.handleMessage(err_ch)
		err, ok := <-err_ch
		if !ok {
			plugin.conn.Close(websocket.StatusNormalClosure, "connection closed")
			break
		}

		status := websocket.CloseStatus(err)
		if status == websocket.StatusNormalClosure {
			break
		}

		if err != nil {
			plugin.conn.Close(websocket.StatusInternalError, err.Error())
			break
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

	plugin.conn.Write(context.Background(), websocket.MessageText, data)
	if callback != nil {
		plugin.callbacks[req_id] = callback
	}

	return nil
}

func (plugin *VTSPlugin) handleMessage(err_ch chan error) {
	_, reader, err := plugin.conn.Reader(context.Background())
	if err != nil {
		err_ch <- err // this can throw strange errors when MediaPipe is stopped
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
		// handle messages that don't fit into the req/res architecture
	}
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

func (plugin *VTSPlugin) doAuth(forceFetchToken bool) {
	payload_data := make(map[string]any)
	payload_data["pluginName"] = "Marmalade"
	payload_data["pluginDeveloper"] = "RanAwaySuccessfully"

	payload := make(map[string]any)
	payload["data"] = payload_data

	var callback vtsPluginCallback // function

	if Config.VTSPlugin.Token == "" || forceFetchToken {
		payload_data["pluginIcon"] = resources.EmbeddedIconLogoSmall
		payload["messageType"] = "AuthenticationTokenRequest"

		callback = vtsPluginCallback(plugin.handleAuthToken)
	} else {
		payload_data["authenticationToken"] = Config.VTSPlugin.Token
		payload["messageType"] = "AuthenticationRequest"

		callback = vtsPluginCallback(plugin.handleAuth)
	}

	plugin.sendMessage(payload, &callback)
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

	plugin.doAuth(false)
}

func (plugin *VTSPlugin) handleAuth(msg_map map[string]any, err_ch chan error) {
	var msg vtsPluginMessageAuth

	err := mapToStruct(msg_map, &msg)
	if err != nil {
		err_ch <- err
		return
	}

	if !msg.Data.Authenticated {
		plugin.doAuth(true)
		return
	}

	plugin.authenticated = true
}

// SEND MEDIAPIPE DATA

type vtsPayloadMessageLive2D struct {
	Id     string  `json:"id"`
	Weight float64 `json:"weight"`
	Value  float64 `json:"value"`
}

func (plugin *VTSPlugin) send(mp_data *TrackingData, err_ch chan error) {
	if !plugin.authenticated {
		return
	}

	payload := make(map[string]any)
	payload["messageType"] = "InjectParameterDataRequest"

	// now the magic starts!

	payload_parameters := make([]vtsPayloadMessageLive2D, 0, 100)

	parameter1 := vtsPayloadMessageLive2D{
		Id:     "FaceAngleX",
		Weight: 1,
		Value:  12.31, // between -1000000 and 1000000
	}

	payload_parameters = append(payload_parameters, parameter1)

	payload_data := make(map[string]any)
	payload_data["parameterValues"] = payload_parameters
	payload_data["faceFound"] = len(mp_data.FaceData.Blendshapes) != 0
	payload_data["mode"] = "set" // add?

	payload["data"] = payload_data

	err := plugin.sendMessage(payload, nil)
	if err != nil {
		err_ch <- err
	}
}

func (plugin *VTSPlugin) close() {
	if plugin.conn != nil {
		plugin.conn.Close(websocket.StatusNormalClosure, "Closing plugin connection.")
	}

	plugin.closed = true
}
