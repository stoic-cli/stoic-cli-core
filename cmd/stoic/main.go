package main

import (
	"os"
	"path/filepath"

	"github.com/stoic-cli/stoic-cli-core/cmd/stoic/cmd"
)

func translateSymlink() {
	_, tool := filepath.Split(os.Args[0])
	if tool != "stoic" {
		executable, err := os.Executable()
		if err != nil {
			executable = "stoic"
		}
		os.Args = append([]string{executable, "run", "--", tool}, os.Args[1:]...)
	}
}

func main() {
	translateSymlink()
	cmd.Execute()
}
