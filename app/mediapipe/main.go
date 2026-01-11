package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
)

// Inter-process communication
type IPC struct {
	mutex   sync.Mutex
	socket  net.Conn
	enabled bool
}

var ipc = IPC{}

func main() {
	println("[MP +TOAST] Starting...")

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
	signal.Notify(sig_channel, os.Interrupt)

	select {
	case err = <-err_channel:
		fmt.Fprintf(os.Stderr, "[MP +TOAST] %v\n", err)
	case sig := <-sig_channel:
		fmt.Printf("[MP +TOAST] Terminating: %v\n", sig)
	}

	go mp.stop()

	/*
		err = mp.stop()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[MP +TOAST] %v\n", err)
		}
	*/

	println("[MP +TOAST] Stopping...")
}
