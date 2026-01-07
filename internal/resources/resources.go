package resources

import (
	_ "embed"
	"os"
	"os/exec"
)

//go:embed icons/marmalade_logo_256.png
var EmbeddedIconLogo []byte

//go:embed version.txt
var EmbeddedVersion string

func InstallIcon() error {
	err := os.WriteFile("/tmp/marmalade_logo_256.png", EmbeddedIconLogo, 0o600)
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"xdg-icon-resource", "install",
		"--novendor",
		"--size", "256",
		"/tmp/marmalade_logo_256.png",
		"xyz.randev.marmalade",
	)

	return cmd.Run()
}
