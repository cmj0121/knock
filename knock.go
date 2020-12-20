package knock

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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

	Port string  `short:"s" help:"port (INT), port list  (INT,INT) or port range (INT-INT)"`
	IP   *string `help:"ip address (STR), ip list (STR,STR) or ip with mask (STR/INT)"`

	*Info `help:"show self net info"`
}

func New() (knock *Knock) {
	knock = &Knock{
		Logger: logger.New(PROJ_NAME),
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

	ports := []int{}
	switch {
	case knock.Port == "":
		// not set the port number
	case RE_PORT_LIST.MatchString(knock.Port):
		for _, p := range strings.Split(knock.Port, ",") {
			if port, err := strconv.Atoi(p); err != nil {
				err = fmt.Errorf("invalid port %v: %v", p, err)
				panic(err)
			} else {
				switch {
				case port < 0:
					err = fmt.Errorf("invalid port %v", port)
					panic(err)
				case port > 65535:
					err = fmt.Errorf("invalid port %v", port)
					panic(err)
				}
				// save to the port-list
				ports = append(ports, port)
			}
		}
	case RE_PORT_RANGE.MatchString(knock.Port):
		port_range := strings.Split(knock.Port, "-")
		if port_start, err := strconv.Atoi(port_range[0]); err != nil {
			err = fmt.Errorf("invalid port %v: %v", port_start, err)
			panic(err)
		} else if port_end, err := strconv.Atoi(port_range[1]); err != nil {
			err = fmt.Errorf("invalid port %v: %v", port_end, err)
			panic(err)
		} else {
			switch {
			case port_start < 0:
				err = fmt.Errorf("invalid port start %v", port_start)
				panic(err)
			case port_end > 65535:
				err = fmt.Errorf("invalid port end %v", port_end)
				panic(err)
			case port_start >= port_end:
				err = fmt.Errorf("invalid port range %v-%v", port_start, port_end)
				panic(err)
			}

			for p := port_start; p <= port_end; p++ {
				ports = append(ports, p)
			}
		}
	default:
		err := fmt.Errorf("invalid port: %v", knock.Port)
		panic(err)
	}
	knock.Debug("scan port #%v ports", len(ports))

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
	}

	return
}

func (knock *Knock) Version(parser *argparse.ArgParse) (exit bool) {
	os.Stdout.WriteString(Version() + "\n")
	exit = true
	return
}
