package knock

import (
	"fmt"
	"regexp"
	"time"
)

const (
	PROJ_NAME = "knock"

	MAJOR = 0
	MINOR = 1
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

const (
	// the global wait seconds when task finished, for receive the pending response
	TASK_WAIT_SECONDS = time.Second * 4
)
