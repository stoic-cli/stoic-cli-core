package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileLock(t *testing.T) {
	tid := SetupTestInDir(t)
	defer tid.Close()

	t.Run("TryLockFileSucceedsWithRelativePath", func(t *testing.T) {
		assert, testFile := tid.SetupTest(t)

		lock, err := TryLockFile(testFile)
		assert.Nil(err)
		assert.NotNil(lock)

		absTestFile, err := filepath.Abs(testFile)
		assert.Nil(err)
		assert.False(filepath.IsAbs(testFile))
		assert.Equal(absTestFile, lock.Name())
	})
	t.Run("TryLockFileSucceedsWithAbsolutePath", func(t *testing.T) {
		assert, testFile := tid.SetupTest(t)
		absTestFile, err := filepath.Abs(testFile)
		assert.Nil(err)

		lock, err := TryLockFile(absTestFile)
		assert.Nil(err)
		assert.NotNil(lock)

		assert.Equal(absTestFile, lock.Name())
	})
	t.Run("TryLockFileKeepsLockFileWhileLocked", func(t *testing.T) {
		assert, testFile := tid.SetupTest(t)
		lockFile := testFile + ".lock"

		_, err := os.Stat(lockFile)
		assert.True(os.IsNotExist(err))

		lock, err := TryLockFile(testFile)
		assert.Nil(err)
		assert.NotNil(lock)

		_, err = os.Stat(lockFile)
		assert.Nil(err)

		lock.Unlock()

		_, err = os.Stat(lockFile)
		assert.True(os.IsNotExist(err))
	})
	t.Run("FileLockIsResistantToChdir", func(t *testing.T) {
		assert, testFile := tid.SetupTest(t)

		absLockFile, err := filepath.Abs(testFile + ".lock")
		assert.Nil(err)

		assert.False(filepath.IsAbs(testFile))
		assert.True(filepath.IsAbs(absLockFile))

		_, err = os.Stat(absLockFile)
		assert.True(os.IsNotExist(err))

		lock, err := TryLockFile(testFile)
		assert.Nil(err)
		assert.NotNil(lock)

		_, err = os.Stat(absLockFile)
		assert.Nil(err)

		err = os.Chdir(tid.OriginalWorkDir())
		assert.Nil(err)

		defer os.Chdir(tid.TestDir())

		_, err = os.Stat(absLockFile)
		assert.Nil(err)

		lock.Unlock()

		_, err = os.Stat(absLockFile)
		assert.True(os.IsNotExist(err))
	})
	t.Run("TryLockFileOnLockedFileFailsWithoutSideEffects", func(t *testing.T) {
		assert, testFile := tid.SetupTest(t)
		lockFile := testFile + ".lock"

		lock, err := TryLockFile(testFile)
		assert.Nil(err)
		assert.NotNil(lock)

		secondLock, err := TryLockFile(testFile)
		assert.NotNil(err)
		assert.Nil(secondLock)

		_, err = os.Stat(lockFile)
		assert.Nil(err)

		lock.Unlock()

		_, err = os.Stat(lockFile)
		assert.True(os.IsNotExist(err))
	})
	t.Run("UnlockPanicsOnUnlockedFile", func(t *testing.T) {
		assert, testFile := tid.SetupTest(t)

		lock, err := TryLockFile(testFile)
		assert.Nil(err)
		assert.NotNil(lock)

		lock.Unlock()

		var havePanicked bool
		defer func() {
			assert.True(havePanicked)
		}()
		defer func() {
			if r := recover(); r != nil {
				havePanicked = true
			}
		}()

		assert.False(havePanicked)
		lock.Unlock()
	})
}
