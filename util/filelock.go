package util

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type FileLock interface {
	// Name returns an absolute path to the locked file
	Name() string

	// Unlock releases the file lock
	Unlock()
}

type fileLock struct {
	filename string
	lock     *os.File
}

// TryLockFile and the returned FileLock implements Cooperative File Locking
// using a ".lock"-suffixed file that is atomically created to acquire the lock.
//
// Cooperative file locking, such as implemented here, is not enforced at any
// level. Instead, it relies on participating threads/processes following the
// same procedure before changing protected files.
func TryLockFile(filename string) (FileLock, error) {
	filename, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	lockFile := filename + ".lock"

	lock, err := os.OpenFile(lockFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0444)
	if err != nil {
		return nil, err
	}

	executable, _ := os.Executable()
	pid := os.Getpid()
	timestamp := time.Now()

	fmt.Fprintf(lock,
		"Cooperative lock on file '%v', created on %v, by %v (PID %v)\n"+
			"It's safe to delete this file if the process that produced it is "+
			"no longer running (e.g., the process may have crashed).\n",
		filename, timestamp, executable, pid)

	return &fileLock{
		filename: filename,
		lock:     lock,
	}, nil
}

func (fl *fileLock) Name() string {
	return fl.filename
}

func (fl *fileLock) Unlock() {
	if fl.lock == nil {
		panic("invariant violation with attempt to unlock already unlocked file")
	}

	os.Remove(fl.lock.Name())

	fl.lock.Close()
	fl.lock = nil
}
