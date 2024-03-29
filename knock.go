package knock

import (
	"os"
	"runtime"
	"time"

	"github.com/alecthomas/kong"
	"github.com/cmj0121/knock/task"
	"github.com/cmj0121/knock/task/producer"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// the knock instance to run the brute-force task
type Knock struct {
	// show version and exit
	Version VersionFlag `short:"V" name:"version" help:"Print version info and quit"`

	// the command-line options
	Debug bool `short:"d" help:"Show the debug message (auto apply --pretty-logger, -vvvv)."`

	// number of workers would be generated and running
	Workers int      `short:"w" help:"Number of workers [default: runtime.NumCPU()]"`
	Name    string   `required:"" arg:"" default:"list" help:"The worker name [default: list]"`
	Args    []string `optional:"" arg:"" help:"The extra arguments to the worker"`

	// the external wordlist
	Wait   time.Duration `default:"25ms" short:"W" help:"The duration per generate word"`
	File   *os.File      `xor:"file,ip,regexp" group:"producer" short:"f" help:"The external word-list file."`
	IP     string        `xor:"file,ip,regexp" group:"producer" short:"i" help:"The valid IP/mask"`
	Regexp string        `xor:"file,ip,regexp" group:"producer" short:"r" help:"The regexp pattern"`
	Prefix string        `group:"producer" help:"The prefix of the token"`
	Suffix string        `group:"producer" help:"The suffix of the token"`

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

	switch {
	case knock.File != nil:
		p = producer.NewReaderProducer(knock.File)
	case knock.Regexp != "":
		var err error

		if p, err = producer.NewRegexpProducer(knock.Regexp); err != nil {
			log.Error().Err(err).Msg("invalid regexp")
			return 1
		}
	case knock.IP != "":
		var err error

		if p, err = producer.NewCIDRProducer(knock.IP); err != nil {
			log.Error().Err(err).Msg("invalid IP")
			return 1
		}
	default:
		p = producer.NewDefaultProducer()
	}
	// set the prefix and suffix of the word list
	p.Prefix(knock.Prefix)
	p.Suffix(knock.Suffix)

	if manager, err := task.New(knock.Name); err != nil {
		log.Error().Err(err).Str("name", knock.Name)
		return 1
	} else if err := manager.NumWorkers(knock.Workers); err != nil {
		log.Error().Err(err)
		return 1
	} else if err := manager.Wait(knock.Wait); err != nil {
		log.Error().Err(err)
		return 1
	} else if err := manager.Run(p, knock.Args...); err != nil {
		log.Error().Err(err)
		return 1
	}

	log.Info().Msg("finish knock ...")
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

	if knock.Name == "list" {
		knock.Workers = 1
	}

	return
}
