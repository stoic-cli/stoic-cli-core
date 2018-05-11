package util

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
)

type TestInDir interface {
	io.Closer

	OriginalWorkDir() string
	TestDir() string

	SetupTest(t *testing.T) (*assert.Assertions, string)
}

func SetupTestInDir(t *testing.T) TestInDir {
	testDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		assert.FailNow(t, "unable to create test directory", err)
	}

	assert.True(t, filepath.IsAbs(testDir))

	originalWorkDir, _ := os.Getwd()
	err = os.Chdir(testDir)
	if err != nil {
		os.RemoveAll(testDir)
		assert.FailNow(t, "unabled to change to test directory", err)
	}

	return &testInDir{originalWorkDir, testDir}
}

type testInDir struct {
	originalWorkDir string
	testDir         string
}

func (d *testInDir) Close() error {
	os.Chdir(d.originalWorkDir)
	os.RemoveAll(d.testDir)
	return nil
}

func (d *testInDir) OriginalWorkDir() string { return d.originalWorkDir }
func (d *testInDir) TestDir() string         { return d.testDir }

func (d *testInDir) SetupTest(t *testing.T) (*assert.Assertions, string) {
	assert := assert.New(t)
	testName := t.Name()

	if i := strings.LastIndexByte(testName, '/'); i != -1 {
		return assert, testName[i+1:]
	}
	return assert, testName
}

func SwitchHome(t *testing.T, newHome string) io.Closer {
	newHome, err := filepath.Abs(newHome)
	if err != nil {
		t.Fatal("failed to convert newHome to an absolute path", err)
	}

	originalHomedirDisableCache := homedir.DisableCache
	originalHome, homeEnvWasSet := os.LookupEnv("HOME")

	homedir.DisableCache = true
	os.Setenv("HOME", newHome)

	return switchHomeCloser{
		originalHome,
		homeEnvWasSet,
		originalHomedirDisableCache,
	}
}

type switchHomeCloser struct {
	originalHome                string
	homeEnvWasSet               bool
	originalHomedirDisableCache bool
}

func (sh switchHomeCloser) Close() error {
	if sh.homeEnvWasSet {
		os.Setenv("HOME", sh.originalHome)
	} else {
		os.Unsetenv("HOME")
	}

	// Refresh cache, before disabling
	homedir.Dir()

	if !sh.originalHomedirDisableCache {
		homedir.DisableCache = false
	}

	return nil
}
