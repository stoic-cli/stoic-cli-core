package runner

import (
	"path/filepath"
)

func (pe *pythonEnv) SitePackages() string {
	return filepath.Join(pe.root, "Lib", "site-packages")
}

func (pe *pythonEnv) Scripts() string {
	return filepath.Join(pe.root, "Scripts")
}

func (pe *pythonEnv) EnvironForSetup() []string {
	return nil
}

func (pe *pythonEnv) InstallModeForSetup() string {
	return "--prefix=" + pe.root
}
