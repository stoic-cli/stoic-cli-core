// +build !windows

package runner

import (
	"path/filepath"
)

func (ve *virtualEnv) Scripts() string {
	return filepath.Join(ve.root, "bin")
}
