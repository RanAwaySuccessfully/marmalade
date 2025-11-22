//go:build !withgtk4 && !withgtk3

package main

import (
	"marmalade/resources"
	"marmalade/server"

	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-u":
			fmt.Println("Uninstalling Marmalade icons...")
			cmd := exec.Command(
				"xdg-icon-resource", "uninstall",
				"--size", "256",
				"xyz.randev.marmalade",
			)

			cmd.Run()
		case "-v":
			fmt.Println("[MARMALADE] v" + resources.EmbeddedVersion)
		default:
			fmt.Println("Unknown argument. Use -v for version. Use -u for uninstalling icons. Do not provide any command line argument for normal usage.")
		}

		return
	}

	info, err := os.Stat(".venv")
	if err != nil || !info.IsDir() {
		fmt.Println("[MARMALADE] .venv folder is missing. This likely indicates that mediapipe-install.sh has not been run yet.")
		fmt.Println("[MARMALADE] Run it now? [y/N]")

		var response string
		fmt.Scanln(&response)

		if response == "y" || response == "Y" {
			fmt.Println("[MARMALADE] Installing MediaPipe...")
			cmd := exec.Command("./mediapipe-install.sh")
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

	err = server.Config.Read()
	if err != nil {
		log.Fatalln(err)
	}

	srv := &server.Server
	err_channel := make(chan error, 1)
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
