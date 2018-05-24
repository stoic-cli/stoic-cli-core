package runner

import (
	"path/filepath"
)

func (pe pythonEnvironment) SitePackages() string {
	return filepath.Join(pe.root, "Lib", "site-packages")
}

func (pe pythonEnvironment) Scripts() string {
	return filepath.Join(pe.root, "Scripts")
}

func (pe pythonEnvironment) installModeForSetup() string {
	return "--prefix=" + pe.root
}

func (pe pythonEnvironment) environForSetup() []string {
	return nil
}

func (ve virtualEnvironment) Scripts() string {
	return filepath.Join(ve.root, "Scripts")
}
