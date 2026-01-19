package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/diamondburned/gotk4/pkg/core/glib"
)

type ServerData struct {
	ErrPipe    *ServerErrPipe
	started    bool
	exit       bool
	mpListener net.Listener
	mpData     TrackingData
	mpCmd      *exec.Cmd
	VTSApi     *VTSApi
	VTSPlugin  *VTSPlugin
}

var Server = ServerData{
	exit: true,
}

func (server *ServerData) Started() bool {
	return !server.exit
}

func (server *ServerData) Start(err_ch chan error, callback func()) {
	server.started = false
	server.exit = false

	fmt.Println("[MARMALADE] Listening...")

	var err error
	server.mpCmd, err = server.createMediaPipeProcess()
	if err != nil {
		err_ch <- err
		return
	}

	err = server.mpCmd.Start()
	if err != nil {
		err_ch <- err
		return
	}

	err = os.Remove("marmalade.sock")
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			err_ch <- err
			return
		}
	}

	listener, err := net.Listen("unix", "marmalade.sock")
	if err != nil {
		err_ch <- err
		return
	}

	if Config.VTSApiUse {
		server.VTSApi = &VTSApi{}
		go server.VTSApi.listen(err_ch)
	}

	if Config.VTSPluginUse {
		server.VTSPlugin = &VTSPlugin{}
		go server.VTSPlugin.listen(err_ch)
	}

	go server.waitMediaPipeProcess(server.mpCmd, err_ch)

	for !server.exit {
		fmt.Println("[MARMALADE] Waiting for MediaPipe...")
		conn, err := listener.Accept()
		if err != nil {
			err_ch <- err
		}

		fmt.Println("[MARMALADE] MediaPipe connection started")
		data := []byte{}

		for !server.exit {
			reader := bufio.NewReader(conn)
			line, isPrefix, err := reader.ReadLine()
			data = append(data, line...)

			if isPrefix {
				continue
			}

			if !server.started {
				server.started = true
				glib.IdleAdd(callback)
			}

			if err != nil {
				if err != io.EOF {
					err_ch <- err
				}

				break
			}

			if len(data) == 0 {
				continue
			}

			server.sendToClients(data, err_ch)
			data = []byte{}
		}
	}

	listener.Close()
	fmt.Println("[MARMALADE] Ended")
}

func (server *ServerData) Stop() {

	if !server.exit {
		fmt.Println("[MARMALADE] Ending...")

		server.exit = true

		if server.VTSApi != nil {
			server.VTSApi.close()
		}

		if server.VTSPlugin != nil {
			server.VTSPlugin.close()
		}

		if server.mpCmd.Process != nil {
			err := server.mpCmd.Process.Signal(syscall.SIGTERM)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v", err)
			}
		}

	}
}

func (server *ServerData) sendToClients(mp_string []byte, err_ch chan error) {
	var mp_data_small anyTracking // For checking tracking type ahead of time

	if server.exit {
		return
	}

	err := json.Unmarshal(mp_string, &mp_data_small)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[MARMALADE] mp socket data: %v\n", err)
		return
	}

	switch mp_data_small.Type {
	case uint8(FaceTrackingType):
		var mp_data FaceTracking
		err := json.Unmarshal(mp_string, &mp_data)
		if err != nil {
			err_ch <- err
			return
		}

		server.mpData.facem = mp_data
	case uint8(HandTrackingType):
		var mp_data HandTracking
		err := json.Unmarshal(mp_string, &mp_data)
		if err != nil {
			err_ch <- err
			return
		}

		server.mpData.handm = mp_data
	}

	if server.VTSApi != nil {
		server.VTSApi.send(&server.mpData.facem, err_ch)
	}

	if server.VTSPlugin != nil {
		server.VTSPlugin.send(&server.mpData, err_ch)
	}
}

func int_to_string(number int) string {
	return strconv.Itoa(number)
}
