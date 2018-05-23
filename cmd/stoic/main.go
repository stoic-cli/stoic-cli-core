package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/stoic-cli/stoic-cli-core/cmd/stoic/cmd"
)

func isStoic(name string) bool {
	if strings.HasPrefix(name, "stoic") {
		if len(name) == len("stoic") {
			return true
		}
		if strings.IndexByte("-.0123456789_", name[len("stoic")]) != -1 {
			return true
		}
	}
	return false
}

func translateSymlink() {
	_, tool := filepath.Split(os.Args[0])
	if !isStoic(tool) {
		return
	}

	executable, err := os.Executable()
	if err != nil {
		executable = "stoic"
	}
	os.Args = append([]string{executable, "run", "--", tool}, os.Args[1:]...)
}

func main() {
	translateSymlink()
	cmd.Execute()
}
