package server

import (
	"encoding/json"
	"os"
)

var Config ServerConfig

type ServerConfig struct {
	Port    float64 `json:"port"`
	Camera  float64 `json:"camera"`
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
	FPS     float64 `json:"fps"`
	Format  string  `json:"format"`
	Model   string  `json:"model"`
	UseGpu  bool    `json:"use_gpu"`
	PrimeId string  `json:"prime_id"`
}

func (config *ServerConfig) Read() error {
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

	// TODO: check if folder exists, if not, create it...
	// TODO: check if file exists, if not, create it...

	//file, err := os.Open(folder + "/MarmaladeVT/config.json")
	file, err := os.Open("config.json")
	if err != nil {
		return err
	}

	dec := json.NewDecoder(file)
	dec.Decode(&config)

	file.Close()

	return nil
}

func (config *ServerConfig) Save() error {
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
