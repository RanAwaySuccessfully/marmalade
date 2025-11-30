package server

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strconv"
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

func (server *ServerData) create_python_process() (*exec.Cmd, error) {
	args := make([]string, 0, 10)
	args = append(args, "main.py")

	if Config.Camera != 0 {
		camera := "--camera=" + int_to_string(int(Config.Camera))
		args = append(args, camera)
	}

	if Config.Width != 0 {
		width := "--width=" + int_to_string(int(Config.Width))
		args = append(args, width)
	}

	if Config.Height != 0 {
		height := "--height=" + int_to_string(int(Config.Height))
		args = append(args, height)
	}

	if Config.FPS != 0 {
		fps := "--fps=" + int_to_string(int(Config.FPS))
		args = append(args, fps)
	}

	if Config.Model != "" {
		model := "--model=" + Config.Model
		args = append(args, model)
	}

	if Config.Format != "" {
		cam_fmt := "--fmt=" + Config.Format
		args = append(args, cam_fmt)
	}

	if Config.UseGpu {
		args = append(args, "--use-gpu")
	}

	// python executable built by PEX
	filepath := "./mediapipe"

	_, err := os.Stat("python/mediapipe")
	if errors.Is(err, os.ErrNotExist) {
		// python interpreter (.venv)
		filepath = "../.venv/bin/python3"
	}

	cmd := exec.Command(filepath, args...)
	cmd.Dir = "python"

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

func int_to_string(number int) string {
	return strconv.Itoa(number)
}
