package knock

import (
	"os"

	"github.com/cmj0121/argparse"
	"github.com/cmj0121/logger"
)

// the knock interface
type Knock struct {
	argparse.Help

	// the internal logger
	*logger.Logger `-`
	LogLevel       string `name:"log" choices:"warn info debug verbose" help:"log level"`

	Port string  `short:"s" help:"port (INT), port list  (INT,INT) or port range (INT-INT)"`
	IP   *string `help:"ip address (STR), ip list (STR,STR) or ip with mask (STR/INT)"`
}

func New() (knock *Knock) {
	knock = &Knock{
		Logger: logger.New(PROJ_NAME),
	}
	return
}

func (knock *Knock) ParseAndRun() (err error) {
	parser := argparse.MustNew(knock)
	argparse.RegisterCallback(argparse.FN_VERSION, knock.Version)

	if err = parser.Run(); err != nil {
		// cannot parse the command-line
		return
	}

	knock.Logger.SetLevel(knock.LogLevel)
	knock.Info("start run %v", PROJ_NAME)
	return
}

func (knock *Knock) Version(parser *argparse.ArgParse) (exit bool) {
	os.Stdout.WriteString(Version() + "\n")
	exit = true
	return
}
