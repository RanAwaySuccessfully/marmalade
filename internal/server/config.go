package server

import (
	"encoding/json"
	"io"
	"os"
)

var Config ConfigSchema

const (
	MediaPipeDelegateCPU = iota
	MediaPipeDelegateGPU
	//MediaPipeDelegateNPU_GOOGLE
)

type ConfigSchema struct {
	Camera    int    `json:"camera"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	FPS       int    `json:"fps"`
	Format    string `json:"format"`
	ModelFace string `json:"model_face"`
	ModelHand string `json:"model_hand"`
	ModelPose string `json:"model_pose"`
	HwAccel   struct {
		ForceFFMPEG bool   `json:"force_ffmpeg"`
		DelegateMP  int    `json:"delegate_mp"`
		Decode      bool   `json:"decode"`
		PrimeId     string `json:"prime_id"`
		//DecodeId  string `json:"decode_id"`
	} `json:"hw_accel"` // hardware acceleration
	VTSApi struct {
		Enabled bool `json:"enabled"`
		Port    int  `json:"port"`
	} `json:"vts_api"` // vts3p
	VTSPlugin struct {
		Enabled bool   `json:"enabled"`
		UseFace bool   `json:"use_face"`
		UseHand bool   `json:"use_hand"`
		UsePose bool   `json:"use_pose"`
		Port    int    `json:"port"`
		Token   string `json:"token"`
	} `json:"vts_plugin"` // vtsplugin
	VMCApi struct {
		Enabled bool `json:"enabled"`
		UseFace bool `json:"use_face"`
		UseHand bool `json:"use_hand"`
		UsePose bool `json:"use_pose"`
		Port    int  `json:"port"`
	} `json:"vmc_api"` // vmcap
	VRChatOSC struct {
		Enabled bool `json:"enabled"`
		UseFace bool `json:"use_face"`
		UseHand bool `json:"use_hand"`
		UsePose bool `json:"use_pose"`
		Port    int  `json:"port"`
	} `json:"vrc_osc"` // vrcosc
}

// for compatibility with fields that have been changed / renamed
type oldSchema struct {
	Port   int    `json:"port"`
	Model  string `json:"model"`
	UseGpu bool   `json:"use_gpu"`
	//PrimeId string `json:"prime_id"` // this setting unfortunately cannot be carried over from v0.4.x
}

func (config *ConfigSchema) Read() error {
	/*
		folder := os.Getenv("XDG_CONFIG_HOME")
		if folder == "" {

			folder := os.Getenv("HOME")
			if folder == "" {
				return errors.New("User has no $XDG_CONFIG_HOME and no $HOME environment variables set.")
			}

			folder = folder + "/.config"
		}
	*/

	// TODO
	// check if folder exists, if not, create it...
	// check if file exists, if not, create it...

	//file, err := os.Open(folder + "/MarmaladeVT/config.json")
	file, err := os.Open("config.json")
	if err != nil {
		return err
	}

	var oldConfig oldSchema

	dec := json.NewDecoder(file)
	dec.Decode(&oldConfig)

	file.Seek(0, io.SeekStart)
	dec.Decode(&config)

	if oldConfig.Model != "" {
		config.ModelFace = oldConfig.Model
	}

	if oldConfig.Port != 0 {
		config.VTSApi.Port = oldConfig.Port
	}

	if oldConfig.UseGpu == true {
		config.HwAccel.DelegateMP = 1
	}

	file.Close()

	return nil
}

func (config *ConfigSchema) Save() error {
	file, err := os.OpenFile("config.json", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(file)
	enc.SetIndent("", "    ")
	enc.Encode(&config)

	file.Close()
	return nil
}
