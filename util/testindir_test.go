package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
)

func TestTestInDir(t *testing.T) {
	wd, err := os.Getwd()
	assert.Nil(t, err)

	tid := SetupTestInDir(t)
	testDir := tid.TestDir()

	assert.Equal(t, wd, tid.OriginalWorkDir())

	assert.True(t, filepath.IsAbs(testDir))
	assert.NotEqual(t, wd, testDir)

	wdInTest, err := os.Getwd()
	assert.Nil(t, err)
	assert.Equal(t, testDir, wdInTest)

	_, err = os.Stat(testDir)
	assert.Nil(t, err)

	_, name := tid.SetupTest(t)
	assert.Equal(t, "TestTestInDir", name)

	t.Run("SubTest", func(t *testing.T) {
		assert, name := tid.SetupTest(t)
		assert.Equal("SubTest", name)
	})
	t.Run("", func(t *testing.T) {
		assert, name := tid.SetupTest(t)
		assert.Equal("#00", name)
	})

	tid.Close()

	finalWd, err := os.Getwd()
	assert.Nil(t, err)
	assert.Equal(t, wd, finalWd)

	_, err = os.Stat(testDir)
	assert.True(t, os.IsNotExist(err))
}

func TestSwitchHome(t *testing.T) {
	originalHome, err := homedir.Dir()
	assert.Nil(t, err)

	sh := SwitchHome(t, "new-home")

	inTestHome, err := homedir.Dir()
	assert.Nil(t, err)
	assert.NotEqual(t, originalHome, inTestHome)

	sh.Close()

	finalHome, err := homedir.Dir()
	assert.Nil(t, err)
	assert.Equal(t, originalHome, finalHome)

	t.Run("SubTest", func(t *testing.T) {
		first, err := homedir.Dir()
		assert.Nil(t, err)
		assert.Equal(t, originalHome, first)

		defer SwitchHome(t, "new-home").Close()

		second, err := homedir.Dir()
		assert.Nil(t, err)
		assert.NotEqual(t, originalHome, second)
	})

	afterSubTest, err := homedir.Dir()
	assert.Nil(t, err)
	assert.Equal(t, originalHome, afterSubTest)

}
