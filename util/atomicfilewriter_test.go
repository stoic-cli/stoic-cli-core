package util

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicFileWriter(t *testing.T) {
	tid := SetupTestInDir(t)
	defer tid.Close()

	t.Run("OpenToChangeCanBeUsedToBootstrapState", func(t *testing.T) {
		assert, testFile := tid.SetupTest(t)

		_, err := os.Stat(testFile)
		assert.True(os.IsNotExist(err))

		state, err := OpenToChange(testFile)
		assert.Nil(err)
		assert.NotNil(state)

		assert.Equal(filepath.Join(tid.TestDir(), testFile), state.Name())
		assert.Nil(state.Current())

		n, err := state.Write([]byte("It was never a thing"))
		assert.Nil(err)
		assert.Equal(20, n)

		_, err = os.Stat(testFile)
		assert.True(os.IsNotExist(err))

		state.Commit()

		fi, err := os.Stat(testFile)
		assert.Nil(err)
		assert.Equal(int64(20), fi.Size())
	})
	t.Run("OpenToChangeLocksTheStateFile", func(t *testing.T) {
		assert, testFile := tid.SetupTest(t)

		state, err := OpenToChange(testFile)
		assert.Nil(err)
		assert.NotNil(state)

		secondState, err := OpenToChange(testFile)
		assert.NotNil(err)
		assert.Nil(secondState)
	})
	t.Run("OpenToChangeCanBeUsedSequentially", func(t *testing.T) {
		assert, testFile := tid.SetupTest(t)

		{
			// Read none, write "Version 1"

			state, err := OpenToChange(testFile)
			assert.Nil(err)
			assert.NotNil(state)

			assert.Nil(state.Current())

			n, err := state.Write([]byte("Version 1"))
			assert.Nil(err)
			assert.Equal(9, n)

			err = state.Commit()
			assert.Nil(err)
		}
		{
			// Read "Version 1", write "Version 41"

			state, err := OpenToChange(testFile)
			assert.Nil(err)
			assert.NotNil(state)

			n, err := state.Write([]byte("Version 41"))
			assert.Nil(err)
			assert.Equal(10, n)

			current := state.Current()
			assert.NotNil(current)

			buffer := make([]byte, 20)

			n, err = current.Read(buffer)
			assert.Equal(io.EOF, err)
			assert.Equal(9, n)
			assert.Equal([]byte("Version 1"), buffer[:9])

			n, err = current.Read(buffer)
			assert.Equal(io.EOF, err)
			assert.Equal(0, n)

			err = state.Commit()
			assert.Nil(err)
		}
		{
			// Read "Version 41", write "Version 103", then Abort

			state, err := OpenToChange(testFile)
			assert.Nil(err)
			assert.NotNil(state)

			n, err := state.Write([]byte("Version 103"))
			assert.Nil(err)
			assert.Equal(11, n)

			current := state.Current()
			assert.NotNil(current)

			buffer := make([]byte, 20)

			n, err = current.Read(buffer)
			assert.Equal(io.EOF, err)
			assert.Equal(10, n)
			assert.Equal([]byte("Version 41"), buffer[:10])

			state.AbortIfPending()
		}
		{
			// Read "Version 41"

			state, err := OpenToChange(testFile)
			assert.Nil(err)
			assert.NotNil(state)

			current := state.Current()
			assert.NotNil(current)

			buffer := make([]byte, 20)

			n, err := current.Read(buffer)
			assert.Equal(io.EOF, err)
			assert.Equal(10, n)
			assert.Equal([]byte("Version 41"), buffer[:10])
		}
	})
	t.Run("AbortIfPendingCanBeCalledAfterCommit", func(t *testing.T) {
		assert, testFile := tid.SetupTest(t)

		_, err := os.Stat(testFile)
		assert.True(os.IsNotExist(err))

		defer func() {
			fi, err := os.Stat(testFile)
			assert.Nil(err)
			assert.Equal(int64(20), fi.Size())
		}()

		state, err := OpenToChange(testFile)
		assert.Nil(err)
		assert.NotNil(state)

		defer state.AbortIfPending()

		assert.Equal(filepath.Join(tid.TestDir(), testFile), state.Name())
		assert.Nil(state.Current())

		n, err := state.Write([]byte("It was never a thing"))
		assert.Nil(err)
		assert.Equal(20, n)

		_, err = os.Stat(testFile)
		assert.True(os.IsNotExist(err))

		state.Commit()
	})
}
