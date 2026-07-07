package main

import "C"
import (
	"errors"
	"fmt"
	"marmalade/app/gtk3-vala/ui"
	"marmalade/internal/server"
	"os"
	"os/exec"
)

// export ui_getembed
func ui_getembed(fileNo C.int) *C.char {
	return C.CString(ui.App) // caller is responsible for freeing the error
}

var status int
var last_error string

// will be done on the C side:
// - gpu listing
// - camerainfo (just print out a list using code block)
// - client list (just print out a list using code block)

// export srv_config
func srv_config() {
	// I'll need to return a C struct here but also, I need to write a function to free that struct from memory
}

// export srv_status
func srv_status() C.int {
	return C.int(status)
	// 4 -> stopped
	// 3 -> stopping
	// 2 -> started
	// 1 -> starting
}

// export srv_error
func srv_error() *C.char {
	if last_error == "" {
		return nil
	}

	c_error := C.CString(last_error) // caller is responsible for freeing the error
	last_error = ""
	return c_error
}

// export srv_start
func srv_start() {
	status = 1
	err_chan := make(chan error)
	server.Server.Start(err_chan, func() {
		status = 2
	})

	go srv_error_handler(err_chan)
}

// export srv_stop
func srv_stop() {
	server.Server.Stop()
	status = 3
}

func srv_error_handler(err_chan chan error) {
	for err := range err_chan {
		server.Server.Stop()
		status = 4

		if errors.Is(err, os.ErrProcessDone) {
			continue
		}

		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			exitError = err.(*exec.ExitError)
			exitCode := exitError.ExitCode()

			// exitError.Stderr is empty, so we use our own copy of Stderr instead
			err = fmt.Errorf("[%d] %s\n%s", exitCode, server.Server.ErrPipe.Log)
		}

		last_error = err.Error()
	}
}

func main() {}
