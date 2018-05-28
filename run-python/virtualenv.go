package runner

import (
	"os"
	"path/filepath"
)

// VirtualEnv represents a Python virtual environment
type VirtualEnv interface {
	// Root returns the base path of the environment
	Root() string

	// Python returns the path to the environment's Python executable
	Python() string

	// Scripts returns the path containing scripts and binaries for the
	// environment
	Scripts() string

	// EnvPath returns the PATH environment variable setup for the environment
	EnvPath() string

	// Environ returns os.Environ(), adjusted with this environment's settings.
	Environ() []string
}

func newVirtualEnv(pe PythonEnv, root string) VirtualEnv {
	return &virtualEnv{pe, root}
}

type virtualEnv struct {
	pe   PythonEnv
	root string
}

func (ve *virtualEnv) Root() string { return ve.root }

func (ve *virtualEnv) Python() string {
	return filepath.Join(ve.Scripts(), "python")
}

func (ve *virtualEnv) EnvPath() string {
	vePath := ve.Scripts()
	if curPath := os.Getenv("PATH"); curPath != "" {
		vePath = vePath + string(os.PathListSeparator) + curPath
	}
	return vePath
}

func (ve *virtualEnv) Environ() []string {
	// TODO: Should filter out PYTHONHOME from environment, if set
	return append(os.Environ(),
		"PATH="+ve.EnvPath(),
		"PYTHONPATH="+ve.pe.SitePackages(),
		"VIRTUAL_ENV="+ve.root)
}
