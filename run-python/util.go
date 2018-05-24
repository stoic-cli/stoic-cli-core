package runner

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// lookupAbsolutePath looks for an executable in PATH, and returns an absolute
// path to the found executable.
func lookupAbsolutePath(file string) (string, error) {
	executable, err := exec.LookPath(file)
	if err != nil {
		return "", err
	}

	executable, err = filepath.Abs(executable)
	if err != nil {
		return "", err
	}

	return executable, nil
}

// fileExists checks if a file exists and is readable by performing an os.Stat()
// operation on the file.
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || !os.IsNotExist(err)
}

func currentTimestamp() []byte {
	return []byte(time.Now().Format(time.RFC3339))
}
