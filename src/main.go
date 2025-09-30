//go:build !withgtk4

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
)

func main() {
	err_channel := make(chan error, 1)
	go StartServer(err_channel)

	sig_channel := make(chan os.Signal, 1)
	signal.Notify(sig_channel, os.Interrupt)

	select {
	case err := <-err_channel:
		fmt.Fprintf(os.Stderr, "[MARMALADE] %v\n", err)
	case sig := <-sig_channel:
		log.Printf("[MARMALADE] Terminating: %v", sig)
	}

	StopServer()
}
