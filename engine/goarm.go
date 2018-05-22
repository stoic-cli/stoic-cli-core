package engine

import (
	_ "unsafe"
)

// Expose runtime.goarm
// https://gist.github.com/lucab/f7162ca2d95191c692edc12ea8ccaaef

//go:linkname goarm runtime.goarm
var runtime_goarm uint8
