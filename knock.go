package knock

import (
	"fmt"
	"os"
	"runtime"

	"github.com/cmj0121/argparse"
	"github.com/cmj0121/logger"
	"gopkg.in/yaml.v3"
)

// the knock interface
type Knock struct {
	argparse.Help

	// the internal logger
	*logger.Logger `-`
	LogLevel       string `name:"log" choices:"warn info debug verbose" help:"log level"`

	// number of worker, default #CPU
	NumWorker int `short:"w" name:"worker" help:"number of worker"`

	*Info `help:"show self net info"`
	*Scan `help:"run scan"`
}

func New() (knock *Knock) {
	knock = &Knock{
		Logger:    logger.New(PROJ_NAME),
		NumWorker: runtime.NumCPU(),
		Scan: &Scan{
			MaxPkgSize: MAX_PACKAGE_SIZE,
			Timeout:    60,
			Format:     "yaml",
			Targets:    []*Target{},
		},
	}
	return
}

func (knock *Knock) ParseAndRun() {
	parser := argparse.MustNew(knock)
	argparse.RegisterCallback(argparse.FN_VERSION, knock.Version)
	defer func() {
		if r := recover(); r != nil {
			parser.HelpMessage(fmt.Errorf("%v", r))
			return
		}
	}()

	if err := parser.Run(); err != nil {
		// cannot parse the command-line
		return
	}

	knock.Logger.SetLevel(knock.LogLevel)
	knock.Logger.Info("start run %v", PROJ_NAME)

	switch {
	case knock.Info != nil:
		knock.Info.Load()
		if data, err := yaml.Marshal(knock.Info); err != nil {
			knock.Logger.Warn("cannot marshal info")
			return
		} else {
			// show on the STDOUT
			os.Stdout.Write(data)
		}
	case knock.Scan != nil:
		// run scan
		knock.Scan.Run(knock.Logger)
	}

	return
}

func (knock *Knock) Version(parser *argparse.ArgParse) (exit bool) {
	os.Stdout.WriteString(Version() + "\n")
	exit = true
	return
}
