package runner

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/stoic-cli/stoic-cli-core"
)

const (
	getPipCacheKey = "run-python/get-pip.py"
	getPipURL      = "https://bootstrap.pypa.io/get-pip.py"

	readyBase        = ".ready"
	requirementsBase = ".requirements"

	pipRequirements = "" +
		"pip==10.0.1 --hash=sha256:717cdffb2833be8409433a93746744b59505f42146e8d37de6c62b430e25d6d7\n" +
		"setuptools==39.1.0 --hash=sha256:0cb8b8625bfdcc2d43ea4b9cdba0b39b2b7befc04f3088897031082aa16ce186\n" +
		"virtualenv==15.2.0 --hash=sha256:e8e05d4714a1c51a2f5921e62f547fcb0f713ebbe959e0a7f585cc8bef71d11f\n" +
		"wheel==0.31.0 --hash=sha256:9cdc8ab2cc9c3c2e2727a4b67c22881dbb0e1c503d592992594c5e131c867107\n" +
		""
)

func getPipScript(cache stoic.Cache) (string, error) {
	var err error

	script, err := ioutil.TempFile("", "get-pip-*.py")
	if err != nil {
		return "", errors.Wrap(err, "unable to set up temp file for get-pip.py")
	}
	defer func() {
		if err != nil {
			os.Remove(script.Name())
		}
		script.Close()
	}()

	if cacheReader := cache.Get(getPipCacheKey); cacheReader != nil {
		defer cacheReader.Close()

		_, err = io.Copy(script, cacheReader)
		if err != nil {
			return "", errors.Wrap(err, "unable to read get-pip.py from cache")
		}
	} else {
		resp, err := http.Get(getPipURL)
		if err != nil {
			return "", errors.Wrapf(err,
				"unable to download get-pip.py script from %v", getPipURL)
		}
		defer resp.Body.Close()

		err = cache.Put(getPipCacheKey, io.TeeReader(resp.Body, script))
		if err != nil {
			return "", errors.Wrapf(err,
				"unable to download get-pip.py script from %v", getPipURL)
		}
	}

	return script.Name(), nil
}

func setupPythonEnv(root string, python string, cache stoic.Cache) (PythonEnv, error) {
	buf := bytes.NewBuffer(nil)
	fmt.Fprintf(buf, "# Python: %s\n%s", python, pipRequirements)

	pipCache := filepath.Join(root, "pip-cache")
	pythonName := filepath.Base(python)

	requirements := buf.Bytes()
	envHash := sha256.Sum256(requirements)
	envRoot := filepath.Join(root, fmt.Sprintf("%s-%.4x", pythonName, envHash))

	pe := newPythonEnv(envRoot, python, pipCache)

	marker := filepath.Join(envRoot, readyBase)
	if fileExists(marker) {
		return pe, nil
	}

	err := os.MkdirAll(pe.Scripts(), os.ModePerm)
	if err != nil {
		return nil, errors.Wrap(err,
			"unable to setup directory for python environmnent")
	}

	defer func() {
		if err != nil {
			os.RemoveAll(envRoot)
		}
	}()

	envRequirements := filepath.Join(envRoot, requirementsBase)
	err = ioutil.WriteFile(envRequirements, requirements, 0644)
	if err != nil {
		return nil, errors.Wrap(err,
			"unable to write requirements for python environment")
	}

	getPip, err := getPipScript(cache)
	if err != nil {
		return nil, err
	}
	defer os.Remove(getPip)

	cmd := exec.Command(python, getPip,
		"--disable-pip-version-check",
		"--no-warn-script-location",
		"--ignore-installed",
		"--isolated",
		"--cache-dir", pipCache,
		pe.(*pythonEnv).InstallModeForSetup(),
		"--require-hashes",
		"--requirement", envRequirements,

		// Disable implicit packages
		"--no-setuptools", "--no-wheel",

		// HACK: Disable implicit pip
		// MUST BE LAST parameter. See https://github.com/pypa/pip/issues/3685
		"--src")

	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	cmd.Env = pe.(*pythonEnv).EnvironForSetup()

	err = cmd.Run()
	if err != nil {
		return nil, errors.Wrap(err,
			"unable to setup pip in python environment")
	}

	if err := ioutil.WriteFile(marker, currentTimestamp(), 0644); err != nil {
		jww.DEBUG.Printf("failed to mark python environment as ready: %v", err)
	}
	return pe, nil
}

func setupVirtualEnv(pe PythonEnv, requirementsFile string) (VirtualEnv, error) {
	requirements, err := ioutil.ReadFile(requirementsFile)
	if err != nil {
		return nil, errors.Wrapf(err,
			"unable to read requirements for virtual environment from %v",
			requirementsFile)
	}

	venvHash := fmt.Sprintf("%x", sha256.Sum256(requirements))
	venvBase := filepath.Join(pe.Root(), "env", venvHash[:2], venvHash[2:])

	ve := newVirtualEnv(pe, venvBase)

	marker := filepath.Join(ve.Root(), readyBase)
	if fileExists(marker) {
		// macOS: homebrew installations of python can be regularly updated,
		// breaking virtual environments, so check for it.
		python, err := os.Readlink(filepath.Join(ve.Root(), ".Python"))
		if os.IsNotExist(err) {
			// Breakage detection does not apply, assume ve is good
			return ve, nil
		}
		if fileExists(python) {
			return ve, nil
		}
	}

	initVenv := exec.Command(pe.Python(),
		"-S", "-m", "virtualenv", "--quiet",

		// Disable implicit packages
		"--no-pip", "--no-setuptools", "--no-wheel",
		ve.Root(),
	)
	initVenv.Stdout = os.Stderr
	initVenv.Stderr = os.Stderr
	initVenv.Env = pe.Environ()

	// virtualenv reported failing on Linux with non UTF-8 locale
	initVenv.Env = append(initVenv.Env, "LANG=en_US.UTF-8")

	err = initVenv.Run()
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize virtual environment")
	}

	defer func() {
		if err != nil {
			_ = os.RemoveAll(ve.Root())
		}
	}()

	venvRequirements := filepath.Join(ve.Root(), requirementsBase)
	err = ioutil.WriteFile(venvRequirements, requirements, 0644)
	if err != nil {
		return nil, errors.Wrap(err,
			"unable to write requirements in virtual environment")
	}

	installRequirements := exec.Command(ve.Python(),
		"-m", "pip", "install",
		"--disable-pip-version-check",
		"--no-warn-script-location",
		"--ignore-installed",
		"--requirement", venvRequirements,
	)
	installRequirements.Stdout = os.Stderr
	installRequirements.Stderr = os.Stderr
	installRequirements.Env = pe.Environ()

	err = installRequirements.Run()
	if err != nil {
		return nil, errors.Wrap(err,
			"unable to setup requirements in virtual environment")
	}

	if err := ioutil.WriteFile(marker, currentTimestamp(), 0644); err != nil {
		jww.DEBUG.Printf("failed to mark virtual environment as ready: %v", err)
	}
	return ve, nil
}
