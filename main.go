//go:build !withgtk4

package main

import (
	"marmalade/server"

	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
)

func main() {
	info, err := os.Stat(".venv")
	if err != nil || !info.IsDir() {
		fmt.Println("[MARMALADE] .venv folder is missing. This likely indicates that mediapipe-install.sh has not been run yet.")
		fmt.Println("[MARMALADE] Run it now? [y/N]")

		var response string
		fmt.Scanln(&response)

		if response == "y" || response == "Y" {
			fmt.Println("[MARMALADE] Installing MediaPipe...")
			cmd := exec.Command("scripts/mediapipe-install.sh")
			cmd.Dir = "scripts"

			err := cmd.Run()
			if err != nil {
				fmt.Println("[MARMALADE] Unable to install MediaPipe. Error details below:")
				log.Fatalln(err)
			}

			fmt.Println("[MARMALADE] Installed!")
		} else {
			fmt.Println("[MARMALADE] Skipping...")
		}
	}

	err_channel := make(chan error, 1)
	srv := &server.Server
	go srv.Start(err_channel)

	sig_channel := make(chan os.Signal, 1)
	signal.Notify(sig_channel, os.Interrupt)

	select {
	case err := <-err_channel:
		fmt.Fprintf(os.Stderr, "[MARMALADE] %v\n", err)
	case sig := <-sig_channel:
		log.Printf("[MARMALADE] Terminating: %v", sig)
	}

	srv.Stop()
}
