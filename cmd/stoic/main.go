package main

import (
	"os"
	"path/filepath"

	"github.com/stoic-cli/stoic-cli-core/cmd/stoic/cmd"
)

func translateSymlink() {
	executable, err := os.Executable()
	if err != nil {
		return
	}
	target, err := os.Readlink(executable)
	if err != nil {
		return
	}

	dir, tool := filepath.Split(executable)
	executable = filepath.Join(dir, target)
	os.Args = append([]string{executable, "run", "--", tool}, os.Args[1:]...)
}

func main() {
	translateSymlink()
	cmd.Execute()
}
