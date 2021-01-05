package knock

import (
	"fmt"
	"os"
	"runtime"

	"github.com/cmj0121/argparse"
	"github.com/cmj0121/logger"
)

type KnockRunner interface {
	Run(*logger.Logger)
}

// the knock interface
type Knock struct {
	argparse.Help

	// the internal logger
	*logger.Logger `-`
	LogLevel       string `name:"log" choices:"warn info debug verbose" help:"log level"`

	// number of worker, default #CPU
	NumWorker int `short:"w" name:"worker" help:"number of worker"`

	*Info `help:"show self net info"`
	*Scan `help:"run raw scan"`
	*Web  `help:"scan the web information"`
}

func New() (knock *Knock) {
	knock = &Knock{
		Logger:    logger.New(PROJ_NAME),
		NumWorker: runtime.NumCPU(),
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

	var runner KnockRunner
	switch {
	case knock.Info != nil:
		runner = knock.Info
	case knock.Web != nil:
		runner = knock.Web
	case knock.Scan != nil:
		runner = knock.Scan
	default:
		parser.HelpMessage(nil)
		return
	}

	runner.Run(knock.Logger)
	return
}

func (knock *Knock) Version(parser *argparse.ArgParse) (exit bool) {
	os.Stdout.WriteString(Version() + "\n")
	exit = true
	return
}
