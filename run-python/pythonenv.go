package runner

import (
	"os"
)

// PythonEnv represents the base Python setup used by the Runner. It hosts an
// isolated setup of pip, wheel, and virtualenv, and can be used to setup
// virtual environments.
type PythonEnv interface {
	// Root returns the base path of the environment.
	Root() string

	// PipCache returns the directory for `pip`'s cache.
	PipCache() string

	// Python returns the path to the Python executable the environment is setup
	// for.
	Python() string

	// Scripts returns the path containing scripts and binaries for the
	// environment.
	Scripts() string

	// SitePackages returns the path where python packages are installed. This
	// should be added to the PYTHONPATH environment variable , in order to use
	// the environment.
	SitePackages() string

	// Environ returns os.Environ(), adjusted with this environment's settings.
	Environ() []string
}

func newPythonEnv(envRoot string, python string, pipCache string) PythonEnv {
	return &pythonEnv{envRoot, python, pipCache}
}

type pythonEnv struct {
	root     string
	python   string
	pipCache string
}

func (pe *pythonEnv) PipCache() string {
	return pe.pipCache
}

func (pe *pythonEnv) Python() string {
	return pe.python
}

func (pe *pythonEnv) Root() string {
	return pe.root
}

func (pe *pythonEnv) Environ() []string {
	return append(os.Environ(),
		"PIP_CACHE_DIR="+pe.pipCache,
		"PYTHONPATH="+pe.SitePackages(),
	)
}
