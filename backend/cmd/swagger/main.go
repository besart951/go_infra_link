package main

import (
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command(
		"go",
		"run",
		"github.com/swaggo/swag/cmd/swag@v1.16.4",
		"init",
		"-g",
		"cmd/app/main.go",
		"-d",
		".",
		"-o",
		"./docs",
		"--parseDependency",
		"--parseInternal",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}
