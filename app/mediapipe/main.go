package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Inter-process communication
type IPC struct {
	mutex      sync.Mutex
	mutex_type uint8
	socket     net.Conn
	enabled    bool
	encoder    *gob.Encoder
	sender     func(uint8, any)
}

var ipc = IPC{}

func main() {
	fmt.Println("[MP +TOAST] Starting...")

	ipc.enabled = false
	if (len(os.Args) > 1) && (os.Args[1] == "--ipc") {
		ipc.enabled = true
	}

	var err error
	if ipc.enabled {
		ipc.socket, err = net.Dial("unix", "marmalade.sock")
		if err != nil {
			fmt.Fprintf(os.Stderr, "[MP +TOAST] %v\n", err)
			return
		}

		ipc.encoder = gob.NewEncoder(ipc.socket)
		ipc.sender = send_result_socket
	} else {
		ipc.sender = send_result_stdout
	}

	mp := MediaPipe{}
	err = mp.start()
	if err != nil {
		mp.stop()
		fmt.Fprintf(os.Stderr, "[MP +TOAST] %v\n", err)
		return
	}

	err_channel := make(chan error, 1)
	go mp.detect(err_channel)

	sig_channel := make(chan os.Signal, 1)
	signal.Notify(sig_channel, os.Interrupt, syscall.SIGTERM)

	select {
	case err = <-err_channel:
		fmt.Fprintf(os.Stderr, "[MP +TOAST] %v\n", err)
	case sig := <-sig_channel:
		fmt.Printf("[MP +TOAST] Terminating: %v\n", sig)
	}

	go mp.stop()
	fmt.Println("[MP +TOAST] Stopping...")
}

func send_result_socket(msg_type uint8, result any) {
	/*
		if msg_type == ipc.mutex_type {
			// if CPU usage is high, we want to discard some results instead of queueing them
			locked := ipc.mutex.TryLock()
			if !locked {
				return
			}
		} else {
	*/
	ipc.mutex.Lock()
	//ipc.mutex_type = msg_type
	//}

	err := ipc.encoder.Encode(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[MP +TOAST] %v\n", err)
	}

	ipc.mutex.Unlock()
}

func send_result_stdout(msg_type uint8, result any) {
	text, err := json.Marshal(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	} else {
		fmt.Println(string(text))
	}
}
