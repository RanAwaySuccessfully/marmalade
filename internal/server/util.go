package server

import (
	"os"
	"strconv"
)

type ErrPipeProxy struct {
	Log string
}

func (err_pipe *ErrPipeProxy) Write(data []byte) (n int, err error) {
	// ^[WIE]0000

	n, err = os.Stderr.Write(data)
	err_pipe.Log += string(data[:n])
	return
	// this is interesting...
	// since i specified the variable names on the function definition line, i don't need to specify them on the return statement!
}

func int_to_string(number int) string {
	return strconv.Itoa(number) // convert int to string
}
