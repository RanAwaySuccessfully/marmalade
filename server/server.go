package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type udpMessage struct {
	created float64
	source  string
	fresh   bool
	Type    string    `json:"messageType"`
	Time    float64   `json:"time"`
	SentBy  string    `json:"sentBy"`
	Ports   []float64 `json:"ports"`
}

type ServerData struct {
	udpListener net.PacketConn
	mu          sync.Mutex
	exit        bool
	clients     map[string]*udpMessage
}

var Server = ServerData{
	clients: make(map[string]*udpMessage),
	exit:    true,
}

func (server *ServerData) Started() bool {
	return !server.exit
}

func (server *ServerData) Start(err_ch chan error) {
	server.exit = false
	Config.Read()

	port := fmt.Sprintf(":%d", int(Config.Port))

	var err error
	server.udpListener, err = net.ListenPacket("udp", port)
	if err != nil {
		err_ch <- err
		return
	}

	fmt.Println("[MARMALADE] Listening...")

	camera := fmt.Sprintf("--camera=%d", int(Config.Camera))
	width := fmt.Sprintf("--width=%d", int(Config.Width))
	height := fmt.Sprintf("--height=%d", int(Config.Height))
	fps := fmt.Sprintf("--fps=%d", int(Config.FPS))
	model := fmt.Sprintf("--model=%s", Config.Model)
	cam_fmt := fmt.Sprintf("--fmt=%s", Config.Format)
	var cmd *exec.Cmd

	if Config.UseGpu {
		cmd = exec.Command(
			"env",
			"VIRTUAL_ENV=../.venv",
			"../.venv/bin/python3",
			"main.py",
			camera,
			width,
			height,
			fps,
			model,
			cam_fmt,
			"--use-gpu",
		)
	} else {
		cmd = exec.Command(
			"env",
			"VIRTUAL_ENV=../.venv",
			"../.venv/bin/python3",
			"main.py",
			camera,
			width,
			height,
			fps,
			model,
			cam_fmt,
		)
	}

	cmd.Dir = "python"
	stdin, err := cmd.StdinPipe()
	if err != nil {
		err_ch <- err
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		err_ch <- err
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		err_ch <- err
		return
	}

	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	err = cmd.Start()
	if err != nil {
		err_ch <- err
		return
	}

	go server.updateClients(stdin, err_ch)
	go server.Wait(cmd, err_ch)

	for !server.exit {
		buf := make([]byte, 1024)

		n, addr, err := server.udpListener.ReadFrom(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				continue
			}

			err_ch <- err
			continue
		}

		if n >= 1024 {
			continue
		}

		data := buf[:n]
		err = server.handlePacket(data, addr)
		if err != nil {
			err_ch <- err
		}
	}

	fmt.Println("[MARMALADE] Ending...")
	fmt.Fprintln(stdin, "end")
	fmt.Println("[MARMALADE] Ended")
}

func (server *ServerData) Wait(cmd *exec.Cmd, err_ch chan error) {
	err := cmd.Wait()
	if err != nil {
		// perhaps I should also keep a copy of stderr?
		err_ch <- err
	} else {
		err_ch <- os.ErrProcessDone
	}
}

func (server *ServerData) Stop() {
	server.exit = true

	if server.udpListener != nil {
		server.udpListener.Close()
	}
}

func (server *ServerData) handlePacket(buf []byte, addr net.Addr) error {
	var msg udpMessage

	err := json.Unmarshal(buf, &msg)
	if err != nil {
		return err
	}

	if msg.Type != "iOSTrackingDataRequest" {
		return nil
	}

	if msg.Time < 0.5 {
		msg.Time = 0.5
	}

	if msg.Time > 10 {
		msg.Time = 10
	}

	msg.source = addr.String()
	msg.Time *= 1000

	server.mu.Lock()
	if server.clients[msg.SentBy] == nil {
		msg.fresh = true
	} else {
		msg.fresh = server.clients[msg.SentBy].fresh
	}

	server.clients[msg.SentBy] = &msg
	server.mu.Unlock()

	return nil
}

func (server *ServerData) updateClients(stdin io.WriteCloser, err_ch chan error) {

	for !server.exit {
		start := time.Now().UnixMilli()

		min := int64(100)

		server.mu.Lock()

		for clientId, client := range server.clients {

			ip, _, _ := strings.Cut(client.source, ":")

			if client.Time <= 0 {
				delete(server.clients, clientId)
				err := server.sendUpdate(stdin, "-", ip, client.Ports)
				if err != nil {
					err_ch <- err
				}

				continue
			}

			if client.fresh {
				err := server.sendUpdate(stdin, "+", ip, client.Ports)
				if err != nil {
					err_ch <- err
				}

				client.fresh = false
			}

			client.Time -= float64(min)
		}

		server.mu.Unlock()

		end := time.Now().UnixMilli()
		diff := end - start

		if diff < min {
			waitFor := time.Duration(min - diff)
			time.Sleep(waitFor * time.Millisecond)
		}
	}
}

func (server *ServerData) sendUpdate(stdin io.WriteCloser, action string, ip string, ports []float64) error {
	for _, port := range ports {
		data := fmt.Sprintf("%s%s:%d", action, ip, int(port))
		_, err := fmt.Fprintln(stdin, data)
		if err != nil {
			return err
		}
	}

	return nil
}
