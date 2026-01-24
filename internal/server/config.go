package server

import (
	"encoding/json"
	"io"
	"os"
)

var Config ConfigSchema

type ConfigSchema struct {
	Camera    int    `json:"camera"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	FPS       int    `json:"fps"`
	Format    string `json:"format"`
	ModelFace string `json:"model_face"`
	ModelHand string `json:"model_hand"`
	UseGpu    bool   `json:"use_gpu"`
	PrimeId   string `json:"prime_id"`
	VTSApi    struct {
		Enabled bool `json:"enabled"`
		Port    int  `json:"port"`
	} `json:"vts_api"`
	VTSPlugin struct {
		Enabled bool   `json:"enabled"`
		UseFace bool   `json:"use_face"`
		UseHand bool   `json:"use_hand"`
		Port    int    `json:"port"`
		Token   string `json:"token"`
	} `json:"vts_plugin"`
	VMCApi struct {
		Enabled bool `json:"enabled"`
		UseFace bool `json:"use_face"`
		UseHand bool `json:"use_hand"`
		Port    int  `json:"port"`
	} `json:"vmc_api"`
}

// for compatibility with fields that have been changed / renamed
type oldSchema struct {
	Port  int    `json:"port"`
	Model string `json:"model"`
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
