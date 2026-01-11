package server

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
)

type ServerErrPipe struct {
	Log string
}

func (err_pipe *ServerErrPipe) Write(data []byte) (n int, err error) {
	n, err = os.Stderr.Write(data)
	err_pipe.Log += string(data[:n])
	return
	// this is interesting...
	// since i specified the variable names on the function definition line, i don't need to specify them on the return statement!
}

func (server *ServerData) createMediaPipeProcess() (*exec.Cmd, error) {
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

	server.ErrPipe = &ServerErrPipe{}
	go io.Copy(server.ErrPipe, stderr)
	go io.Copy(os.Stdout, stdout)

	return cmd, nil
}

func (server *ServerData) waitMediaPipeProcess(cmd *exec.Cmd, err_ch chan error) {
	err := cmd.Wait()
	if err != nil {
		err_ch <- err
	} else {
		err_ch <- os.ErrProcessDone
	}
}
