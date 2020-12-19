package knock

import (
	"fmt"
	"regexp"
)

const (
	PROJ_NAME = "knock"

	MAJOR = 0
	MINOR = 0
	MACRO = 0
)

func Version() (ver string) {
	ver = fmt.Sprintf("%v (v%d.%d.%d)", PROJ_NAME, MAJOR, MINOR, MACRO)
	return
}

var (
	RE_PORT_LIST  = regexp.MustCompile(`^\d+(?:,\d+)*$`)
	RE_PORT_RANGE = regexp.MustCompile(`^\d+\-\d+$`)
)
