package knock

import (
	"fmt"

	_ "embed"
)

// the meta of this projecet
const (
	// project name
	PROJ_NAME = "knock"
	// project version
	MAJOR = 0 // the API version, bump when change interface
	MINOR = 0 // bump when the new feature implemented
	MACRO = 0 // bump when bug-fixed only
)

// return the version info
func Version() (ver string) {
	ver = fmt.Sprintf("%s (%d.%d.%d)", PROJ_NAME, MACRO, MINOR, MACRO)
	return
}

//go:embed assets/word-list
var word_lists string
