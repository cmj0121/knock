package knock

import (
	"os"
	"runtime"
	"time"

	"github.com/alecthomas/kong"
	"github.com/cmj0121/knock/task/producer"
	"github.com/cmj0121/knock/task/worker"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// the knock instance to run the brute-force task
type Knock struct {
	// the command-line options
	Debug bool `short:"d" help:"Show the debug message (auto apply --pretty-logger, -vvvv)."`

	// number of workers would be generated and running
	Workers int    `short:"w" help:"Number of workers [default: runtime.NumCPU()]"`
	Name    string `arg:"" default:"list" help:"The worker name [default: list]"`

	// the external wordlist
	Wait time.Duration `short:"W" help:"The duration per generate word"`
	File *os.File      `xor:"file,ip" group:"producer" short:"f" help:"The external word-list file."`
	IP   string        `xor:"file,ip" group:"producer" short:"i" help:"The valid IP/mask"`

	// the logger options
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
func (knock *Knock) Run() int {
	kong.Parse(knock)

	knock.prologue()
	return knock.run()
}

func (knock *Knock) run() (exitcode int) {
	log.Info().Msg("start Knock ...")

	var p producer.Producer

	switch knock.IP {
	case "":
		p = producer.NewReaderProducer(knock.File)
	default:
		var err error

		if p, err = producer.NewCIDRProducer(knock.IP); err != nil {
			log.Error().Err(err).Msg("invalid IP")
			return 1
		}
	}

	worker, ok := worker.GetWorker(knock.Name)
	if !ok {
		log.Error().Str("name", knock.Name).Msg("cannot get worker")
		return 1
	}

	if err := worker.Run(p.Produce(knock.Wait)); err != nil {
		log.Error().Err(err).Msg("cannot run worker")
		return 1
	}
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

	if knock.Workers == 0 {
		knock.Workers = runtime.NumCPU()
	}

	return
}
