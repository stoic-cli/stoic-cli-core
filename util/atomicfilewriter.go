package util

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type AtomicFileWriter interface {
	io.Writer

	Name() string
	Current() io.Reader

	AbortIfPending()
	Commit() error
}

func OpenToChange(filename string) (AtomicFileWriter, error) {
	lockedFile, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	lock, err := TryLockFile(lockedFile)
	if err != nil {
		return nil, errors.Wrapf(err,
			"unable to obtain lock on file %v", lockedFile)
	}
	defer func() {
		if err != nil {
			lock.Unlock()
		}
	}()

	curr, err := os.Open(lockedFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrapf(err,
			"unable to open current state, %v", lockedFile)
	}

	dir, base := filepath.Split(lockedFile)
	next, err := ioutil.TempFile(dir, base)
	if err != nil {
		return nil, errors.Wrap(err,
			"unable to create temporary file for changes")
	}

	return &state{lock, curr, next}, nil
}

type state struct {
	lock FileLock
	curr *os.File
	next *os.File
}

func (s *state) Name() string {
	return s.lock.Name()
}

type offsetFileReader struct {
	file   *os.File
	offset int64
}

func (r *offsetFileReader) Read(p []byte) (int, error) {
	n, err := r.file.ReadAt(p, r.offset)
	r.offset += int64(n)
	return n, err
}

func (s *state) Current() io.Reader {
	if s.curr == nil {
		return nil
	}
	return &offsetFileReader{s.curr, 0}
}

func (s *state) Write(p []byte) (n int, err error) {
	return s.next.Write(p)
}

func (s *state) AbortIfPending() {
	if s.next == nil {
		return
	}

	s.lock.Unlock()

	if s.curr != nil {
		s.curr.Close()
		s.curr = nil
	}

	os.Remove(s.next.Name())
	s.next.Close()

	// Make unusable
	s.next = nil
}

func (s *state) Commit() error {
	if s.curr != nil {
		s.curr.Close()
		s.curr = nil
	}

	s.next.Sync()
	err := os.Rename(s.next.Name(), s.Name())

	s.next.Close()
	s.lock.Unlock()

	// Make unusable
	s.next = nil

	return err
}
