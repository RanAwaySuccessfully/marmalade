package server

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
)

func createMediaPipeProcess(server *ServerInstance) (*exec.Cmd, error) {
	var cmd *exec.Cmd

	_, err := os.Stat("./mediapipe")
	if errors.Is(err, os.ErrNotExist) { // local testing
		build_cmd := exec.Command("go", "build")
		build_cmd.Dir = "./app/mediapipe"

		err := build_cmd.Run()
		if err != nil {
			return nil, err
		}

		cmd = exec.Command("./app/mediapipe/mediapipe", "--ipc")
		env := cmd.Environ()

		library_path := "LD_LIBRARY_PATH=./app/mediapipe/cc"

		for i := 0; i < len(env); i++ {
			env_var := env[i]
			isLibraryPath := strings.HasPrefix(env_var, "LD_LIBRARY_PATH=")

			if isLibraryPath {
				library_path += ":" + env_var[16:]
			}
		}

		cmd.Env = append(env, library_path)

	} else {
		cmd = exec.Command("./mediapipe", "--ipc")
	}

	if Config.PrimeId != "" {
		prime_env := "DRI_PRIME=" + Config.PrimeId
		cmd.Env = append(cmd.Environ(), prime_env)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	go io.Copy(server.ErrPipe, stderr)
	go io.Copy(os.Stdout, stdout)

	return cmd, nil
}

func waitMediaPipeProcess(cmd *exec.Cmd, err_ch chan error) {
	err := cmd.Wait()
	if err != nil {
		err_ch <- err
	} else {
		err_ch <- os.ErrProcessDone
	}
}
