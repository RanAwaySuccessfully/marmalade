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

var clients = Clients{
	list: make(map[string]*udpMessage),
}

type Clients struct {
	mu   sync.Mutex
	exit bool
	list map[string]*udpMessage
}

type udpMessage struct {
	created float64
	source  string
	Type    string    `json:"messageType"`
	Time    float64   `json:"time"`
	SentBy  string    `json:"sentBy"`
	Ports   []float64 `json:"ports"`
}

func Start(err_ch chan error) {
	ReadConfig()

	port := fmt.Sprintf(":%d", int(serverConfig.Port))

	listener, err := net.ListenPacket("udp", port)
	if err != nil {
		err_ch <- err
		return
	}

	fmt.Println("[MARMALADE] Listening...")

	camera := fmt.Sprintf("--camera=%d", int(serverConfig.Camera))
	width := fmt.Sprintf("--width=%d", int(serverConfig.Width))
	height := fmt.Sprintf("--height=%d", int(serverConfig.Height))
	fps := fmt.Sprintf("--fps=%d", int(serverConfig.FPS))
	model := fmt.Sprintf("--model=%s", serverConfig.Model)
	var cmd *exec.Cmd

	if serverConfig.UseGpu {
		cmd = exec.Command(
			"scripts/mediapipe-run.sh",
			camera,
			width,
			height,
			fps,
			model,
			"--use-gpu",
		)
	} else {
		cmd = exec.Command(
			"scripts/mediapipe-run.sh",
			camera,
			width,
			height,
			fps,
			model,
		)
	}

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

	go sendToClients(stdin, err_ch)

	for !clients.exit {
		buf := make([]byte, 1024)
		n, addr, err := listener.ReadFrom(buf)
		if err != nil {
			err_ch <- err
			continue
		}

		if n >= 1024 {
			continue
		}

		data := buf[:n]
		err = handlePacket(data, addr)
		if err != nil {
			err_ch <- err
		}
	}

	cmd.Process.Signal(os.Interrupt)
	cmd.Wait()
}

func Stop() {
	clients.exit = true
}

func handlePacket(buf []byte, addr net.Addr) error {
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
	clients.mu.Lock()
	clients.list[msg.SentBy] = &msg
	clients.mu.Unlock()

	return nil
}

func sendToClients(stdin io.WriteCloser, err_ch chan error) {

	counter := 0

	for !clients.exit {
		start := time.Now().UnixMilli()

		// minimum amount of milliseconds this loop iteration must take to maintain 60FPS
		min := int64(17)

		if counter == 0 {
			min = 16
		}

		clients.mu.Lock()

		for clientId, msg := range clients.list {

			if msg.Time <= 0 {
				delete(clients.list, clientId)
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

		clients.mu.Unlock()

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
