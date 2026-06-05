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
	mutex   sync.Mutex
	socket  net.Conn
	encoder *gob.Encoder
	sender  func(any)
}

var ipc = IPC{}

func main() {
	fmt.Println("[MP +TOAST] Starting...")

	useGob := (len(os.Args) > 1) && (os.Args[1] == "--ipc")

	var err error
	if useGob {
		ipc.socket, err = net.Dial("unix", "marmalade.sock")
		if err != nil {
			err = create_error("creating Unix socket IPC", err)
			fmt.Fprintln(os.Stderr, err)
			return
		}
		//defer ipc.socket.Close()

		ipc.encoder = gob.NewEncoder(ipc.socket)
		ipc.sender = send_result_socket
	} else {
		ipc.sender = send_result_stdout
	}

	mp := MediaPipe{}
	err = mp.start()
	if err != nil {
		mp.stop()
		fmt.Fprintln(os.Stderr, err)
		fmt.Println("[MP +TOAST] Stopping early...")
		os.Exit(111)
		return
	}

	err_channel := make(chan error, 1)
	go mp.detect(err_channel)

	sig_channel := make(chan os.Signal, 1)
	signal.Notify(sig_channel, os.Interrupt, syscall.SIGTERM)

	select {
	case err = <-err_channel:
		fmt.Fprintln(os.Stderr, err)
		os.Exit(110)
	case sig := <-sig_channel:
		fmt.Printf("[MP +TOAST] Terminating: %v\n", sig)
	}

	go mp.stop()
	fmt.Println("[MP +TOAST] Stopping...")
}

func send_result_socket(result any) {
	ipc.mutex.Lock()

	err := ipc.encoder.Encode(result)
	if err != nil {
		err = create_error("sending data to Unix socket IPC", err)
		fmt.Fprintln(os.Stderr, err)
	}

	ipc.mutex.Unlock()
}

func send_result_stdout(result any) {
	text, err := json.Marshal(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	} else {
		fmt.Println(string(text))
	}
}
