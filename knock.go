package knock

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// the knock instance to run the brute-force task
type Knock struct {
	// the command-line options
	Debug bool `short:"d" help:"Show the debug message (auto apply --pretty-logger, -vvvv)."`

	Quiet        bool `short:"q" group:"logger" xor:"verbose,quiet" help:"Disable all logger."`
	Verbose      int  `short:"v" group:"logger" xor:"verbose,quiet" type:"counter" help:"Show the verbose logger."`
	PrettyLogger bool `group:"logger" help:"Show the pretty logger."`
}

// create the Knock instance with the default settings.
func New() (knock *Knock) {
	knock = &Knock{}
	return
}

// run the knock, parse by the passed arguments from CLI and return
// the result.
func (knock *Knock) Run() (exitcode int) {
	kong.Parse(knock)

	knock.prologue()
	log.Info().Msg("start Knock ...")

	exitcode = 0
	return
}

// setup the necessary before run knock
func (knock *Knock) prologue() {
	knock.setupLogger()
}

// setup the logger sub-system
func (knock *Knock) setupLogger() {
	if knock.PrettyLogger {
		writter := zerolog.ConsoleWriter{Out: os.Stderr}
		log.Logger = zerolog.New(writter).With().Timestamp().Logger()
	}

	switch knock.Verbose {
	case -1:
		zerolog.SetGlobalLevel(zerolog.Disabled)
	case 0:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case 1:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case 2:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case 3:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}

// the callback function after kong parse the CLI arguments
func (knock *Knock) AfterApply() (err error) {
	if knock.Debug {
		knock.PrettyLogger = true
		knock.Verbose = 4
	}

	if knock.Quiet {
		knock.Verbose = -1
	}

	return
}
