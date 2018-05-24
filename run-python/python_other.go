// +build !windows

package runner

import (
	"path/filepath"
)

func (pe pythonEnvironment) SitePackages() string {
	return filepath.Join(pe.root, "lib", "python", "site-packages")
}

func (pe pythonEnvironment) Scripts() string {
	return filepath.Join(pe.root, "bin")
}

func (pe pythonEnvironment) installModeForSetup() string {
	return "--user"
}

func (pe pythonEnvironment) environForSetup() []string {
	return []string{"PYTHONUSERBASE=" + pe.root}
}

func (ve virtualEnvironment) Scripts() string {
	return filepath.Join(ve.root, "bin")
}
