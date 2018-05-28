// +build !windows

package runner

import (
	"path/filepath"
)

func (pe *pythonEnv) SitePackages() string {
	return filepath.Join(pe.root, "lib", "python", "site-packages")
}

func (pe *pythonEnv) Scripts() string {
	return filepath.Join(pe.root, "bin")
}

func (pe *pythonEnv) EnvironForSetup() []string {
	return []string{"PYTHONUSERBASE=" + pe.Root()}
}

func (pe *pythonEnv) InstallModeForSetup() string {
	return "--user"
}
