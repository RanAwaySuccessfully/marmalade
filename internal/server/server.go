package server

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"syscall"

	"github.com/diamondburned/gotk4/pkg/core/glib"
)

type ServerInstance struct {
	ErrPipe   *ErrPipeProxy
	started   bool
	exit      bool
	mpData    TrackingData
	mpCmd     *exec.Cmd
	VMCApi    *VMCApi
	VTSApi    *VTSApi
	VTSPlugin *VTSPlugin
	VRChatOSC *VRChatOSC
}

var Server = ServerInstance{
	exit: true,
}

func (server *ServerInstance) Started() bool {
	return !server.exit
}

func (server *ServerInstance) Start(err_ch chan error, callback func()) {
	server.started = false
	server.exit = false

	fmt.Println("[MARMALADE] Listening...")

	if server.ErrPipe == nil {
		server.ErrPipe = &ErrPipeProxy{}
	} else {
		server.ErrPipe.Log = ""
	}

	var err error
	server.mpCmd, err = createMediaPipeProcess(server)
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

	socket, err := net.Listen("unix", "marmalade.sock")
	if err != nil {
		err_ch <- err
		return
	}
	defer os.Remove("marmalade.sock")

	if Config.VMCApi.Enabled {
		server.VMCApi = &VMCApi{}
		go server.VMCApi.Listen(err_ch)
	}

	if Config.VTSApi.Enabled {
		server.VTSApi = &VTSApi{}
		go server.VTSApi.Listen(err_ch)
	}

	if Config.VTSPlugin.Enabled {
		server.VTSPlugin = &VTSPlugin{}
		go server.VTSPlugin.Listen(err_ch)
	}

	if Config.VRChatOSC.Enabled {
		server.VRChatOSC = &VRChatOSC{}
		go server.VRChatOSC.Listen(err_ch)
	}

	go waitMediaPipeProcess(server.mpCmd, err_ch)

	conn, err := socket.Accept()
	if err != nil {
		err_ch <- err
		return
	}

	fmt.Println("[MARMALADE] MediaPipe connection started")
	decoder := gob.NewDecoder(conn)

	for !server.exit {
		var mp_data TrackingData

		err := decoder.Decode(&mp_data)
		if err != nil {
			if err != io.EOF {
				err_ch <- err
			}

			break
		}

		if !server.started {
			server.started = true
			glib.IdleAdd(callback)
		}

		server.sendToClients(mp_data, err_ch)
	}

	socket.Close()
	fmt.Println("[MARMALADE] Ended")
}

func (server *ServerInstance) Stop() {

	if !server.exit {
		fmt.Println("[MARMALADE] Ending...")

		server.exit = true

		if server.VMCApi != nil {
			server.VMCApi.Close()
		}

		if server.VTSApi != nil {
			server.VTSApi.Close()
		}

		if server.VTSPlugin != nil {
			server.VTSPlugin.Close()
		}

		if server.mpCmd.Process != nil {
			err := server.mpCmd.Process.Signal(syscall.SIGTERM)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v", err)
			}
		}

	}
}

func (server *ServerInstance) sendToClients(mp_data TrackingData, err_ch chan error) {
	if server.exit {
		return
	}

	switch mp_data.Type {
	case NullTrackingType:
		return
	case FaceTrackingType:
		server.mpData.FaceData = mp_data.FaceData
	case HandTrackingType:
		server.mpData.HandData = mp_data.HandData
	case PoseTrackingType:
		server.mpData.PoseData = mp_data.PoseData
	}

	server.mpData.Status = mp_data.Status
	server.mpData.Timestamp = mp_data.Timestamp

	if server.VMCApi != nil {
		server.VMCApi.Send(&server.mpData, err_ch)
	}

	if server.VTSApi != nil {
		server.VTSApi.Send(&server.mpData, err_ch)
	}

	if server.VTSPlugin != nil {
		server.VTSPlugin.Send(&server.mpData, err_ch)
	}
}
