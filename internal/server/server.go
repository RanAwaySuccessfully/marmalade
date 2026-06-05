package server

import (
	"fmt"

	"github.com/diamondburned/gotk4/pkg/core/glib"
)

type ServerInstance struct {
	ErrPipe   *ErrPipeProxy
	started   bool
	exit      bool
	mpData    TrackingData
	mpCmd     *MediaPipeProcess
	kaData    KalidoKitData
	kaCmd     *KalidoKitProcess
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

	server.mpCmd = &MediaPipeProcess{}
	err := server.mpCmd.createSocket()
	if err != nil {
		err_ch <- err
		return
	}

	err = server.mpCmd.create(server.ErrPipe)
	if err != nil {
		err_ch <- err
		return
	}

	needs_kalidokit := false

	if Config.VMCApi.Enabled {
		needs_kalidokit = true
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

	go server.mpCmd.wait(err_ch)

	var ka_ch chan KalidoKitData
	if needs_kalidokit {
		server.kaCmd = &KalidoKitProcess{}
		err = server.kaCmd.createSocket()
		if err != nil {
			err_ch <- err
			return
		}

		err = server.kaCmd.create()
		if err != nil {
			err_ch <- err
			return
		}

		go server.kaCmd.wait(err_ch)

		ka_ch = make(chan KalidoKitData)
		go server.kaCmd.listen(&server.exit, err_ch, ka_ch)
	}

	result := make(chan TrackingData)
	go server.mpCmd.listen(&server.exit, err_ch, result)

	for mp_data := range result {
		if !server.started {
			server.started = true
			glib.IdleAdd(callback)
		}

		server.sendToClients(mp_data, ka_ch, err_ch)
	}

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

		if server.VRChatOSC != nil {
			server.VRChatOSC.Close()
		}

		if server.kaCmd != nil {
			server.kaCmd.close()
		}

		if server.mpCmd != nil {
			server.mpCmd.close()
		}
	}
}

func (server *ServerInstance) sendToClients(mp_data TrackingData, ka_ch chan KalidoKitData, err_ch chan error) {
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

	if server.kaCmd != nil {
		switch mp_data.Type {
		case HandTrackingType:
			err := server.kaCmd.send(mp_data.HandData)
			if err != nil {
				err_ch <- err
			}

			ka_data, ok := <-ka_ch
			if !ok {
				return
			}

			server.kaData.LeftHandData = ka_data.LeftHandData
			server.kaData.RightHandData = ka_data.RightHandData
		case PoseTrackingType:
			err := server.kaCmd.send(mp_data.PoseData)
			if err != nil {
				err_ch <- err
			}

			ka_data, ok := <-ka_ch
			if !ok {
				return
			}

			server.kaData.PoseData = ka_data.PoseData
		}
	}

	if server.VMCApi != nil {
		server.VMCApi.Send(&server.mpData, mp_data.Type, &server.kaData, err_ch)
	}

	if server.VTSApi != nil {
		server.VTSApi.Send(&server.mpData, err_ch)
	}

	if server.VTSPlugin != nil {
		server.VTSPlugin.Send(&server.mpData, err_ch)
	}

	if server.VRChatOSC != nil {
		server.VRChatOSC.Send(&server.mpData, err_ch)
	}
}
