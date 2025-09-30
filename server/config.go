package server

import (
	"encoding/json"
	"errors"
	"os"
)

var serverConfig ServerConfig

type ServerConfig struct {
	Port   float64 `json:"port"`
	Camera float64 `json:"camera"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	FPS    float64 `json:"fps"`
	Model  string  `json:"model"`
	UseGpu bool    `json:"use_gpu"`
}

func ReadConfig() error {
	folder := os.Getenv("XDG_CONFIG_HOME")
	if folder == "" {

		folder := os.Getenv("HOME")
		if folder == "" {
			return errors.New("User has no $XDG_CONFIG_HOME and no $HOME environment variables set.")
		}

		folder = folder + "/.config"
	}

	// check if folder exists, if not, create it...
	// check if file exists, if not, create it...

	//file, err := os.Open(folder + "/MarmaladeVT/config.json")
	file, err := os.Open("config.json")
	if err != nil {
		return err
	}

	dec := json.NewDecoder(file)
	dec.Decode(&serverConfig)

	file.Close()

	return nil
}
