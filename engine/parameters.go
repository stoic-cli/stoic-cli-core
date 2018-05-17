package engine

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

func (e *engine) Parameters() map[string]interface{} {
	env := map[string]string{}
	for _, envVar := range os.Environ() {
		pos := strings.Index(envVar, "=")
		env[envVar[:pos]] = envVar[pos+1:]
	}

	var goarm string
	if runtime_goarm == 0 {
		goarm = ""
	} else {
		goarm = fmt.Sprintf("%i", runtime_goarm)
	}

	params := map[string]interface{}{
		"Env":           env,
		"Arch":          runtime.GOARCH,
		"OS":            runtime.GOOS,
		"Arm":           goarm,
		"NativeArchive": ".tar.gz",
		"WindowsBat":    "",
		"WindowsExe":    "",
	}

	if runtime.GOOS == "windows" {
		params["NativeArchive"] = ".zip"
		params["WindowsBat"] = ".bat"
		params["WindowsExe"] = ".exe"
	}

	return params
}
