package knock

import (
	"fmt"

	"github.com/alecthomas/kong"
)

// the meta of this projecet
const (
	// project name
	PROJ_NAME = "knock"

	// project version
	MAJOR = 0 // the API version, bump when change interface
	MINOR = 4 // bump when the new feature implemented
	MACRO = 0 // bump when bug-fixed only
)

// return the version info
func Version() (ver string) {
	ver = fmt.Sprintf("%s (v%d.%d.%d)", PROJ_NAME, MACRO, MINOR, MACRO)
	return
}

type VersionFlag bool

func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(Version())
	app.Exit(0)
	return nil
}
