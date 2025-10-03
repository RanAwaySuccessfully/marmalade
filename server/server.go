package server

import (
	"encoding/json"
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
	Type    string    `json:"messageType"`
	Time    float64   `json:"time"`
	SentBy  string    `json:"sentBy"`
	Ports   []float64 `json:"ports"`
}

type ServerData struct {
	mu      sync.Mutex
	exit    bool
	clients map[string]*udpMessage
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

	listener, err := net.ListenPacket("udp", port)
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

	go server.sendToClients(stdin, err_ch)

	for !server.exit {
		buf := make([]byte, 1024)

		deadline := time.Now().Add(time.Second)
		listener.SetDeadline(deadline)

		n, addr, err := listener.ReadFrom(buf)
		if err != nil {
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
	cmd.Process.Signal(os.Interrupt)
	cmd.Wait()
	fmt.Println("[MARMALADE] Ended")
}

func (server *ServerData) Stop() {
	server.exit = true
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
	server.clients[msg.SentBy] = &msg
	server.mu.Unlock()

	return nil
}

func (server *ServerData) sendToClients(stdin io.WriteCloser, err_ch chan error) {

	counter := 0

	for !server.exit {
		start := time.Now().UnixMilli()

		// minimum amount of milliseconds this loop iteration must take to maintain 60FPS
		min := int64(17)

		if counter == 0 {
			min = 16
		}

		server.mu.Lock()

		for clientId, msg := range server.clients {

			if msg.Time <= 0 {
				delete(server.clients, clientId)
				continue
			}

			ip, _, _ := strings.Cut(msg.source, ":")

			for _, port := range msg.Ports {
				target := ip + ":" + fmt.Sprintf("%d", int(port))
				_, err := fmt.Fprintln(stdin, target) // send target address to the Python script
				if err != nil {
					err_ch <- err
				}
			}

			msg.Time -= float64(min)
		}

		server.mu.Unlock()

		end := time.Now().UnixMilli()
		diff := end - start

		if diff < min {
			waitFor := time.Duration(min - diff)
			time.Sleep(waitFor * time.Millisecond)
		}

		counter++
		counter %= 3
	}
}
