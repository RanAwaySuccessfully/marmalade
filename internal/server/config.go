package server

import (
	"encoding/json"
	"io"
	"os"
)

var Config ConfigSchema

type ConfigSchema struct {
	Camera         float64 `json:"camera"`
	Width          float64 `json:"width"`
	Height         float64 `json:"height"`
	FPS            float64 `json:"fps"`
	Format         string  `json:"format"`
	ModelFace      string  `json:"model_face"`
	ModelHand      string  `json:"model_hand"`
	UseGpu         bool    `json:"use_gpu"`
	PrimeId        string  `json:"prime_id"`
	VTSApiUse      bool    `json:"vts_api_use"`
	VTSApiPort     float64 `json:"vts_api_port"`
	VTSPluginUse   bool    `json:"vts_plugin_use"`
	VTSPluginPort  float64 `json:"vts_plugin_port"`
	VTSPluginToken string  `json:"vts_plugin_token"`
}

// for compatibility with fields that have been changed / renamed
type oldSchema struct {
	Port  float64 `json:"port"`
	Model string  `json:"model"`
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
		config.VTSApiPort = oldConfig.Port
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
