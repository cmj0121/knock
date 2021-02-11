package knock

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/cmj0121/argparse"
	"github.com/cmj0121/logger"
	"github.com/mattn/go-isatty"
)

// the knock interface which provide the global setting
type Knock struct {
	argparse.Model

	// the internal logger
	*logger.Logger `-`
	LogLevel       string `name:"log" choices:"warn info debug verbose" help:"log level"`

	// number of worker, default #CPU
	NumWorker int    `short:"w" name:"worker" help:"number of worker"`
	WordList  string `short:"W" name:"word-list" choices:"wordlists usernames passwords" help:"default word lists"`

	// the global timeout based on seconds
	Timeuot int `short:"t" help:"global timeout based on seconds"`
	Wait    int `help:"wait ms per each task"`

	*Demo `help:"list all the word-list"`
	*Info `help:"show the current system info"`
	*Scan `help:"scan via network protocol"`
	*Git  `help:"fetch the remote Git public repo"`

	/* ---- private fields */
	receiver chan Response
	wg       sync.WaitGroup
}

// new the knock instance with default config
func New() (knock *Knock) {
	knock = &Knock{
		Logger:    logger.New(PROJ_NAME),
		NumWorker: runtime.NumCPU(),
		WordList:  "wordlists",
		Timeuot:   60,

		// NOTE - #Reducer Buffer=16
		receiver: make(chan Response, 16),
	}
	return
}

// show the knock version info
func (knock *Knock) Version(parser *argparse.ArgParse) (exit bool) {
	os.Stdout.WriteString(Version() + "\n")
	exit = true
	return
}

// parse from the command-line and execute the knock
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

	/* ---- runner ---- */
	var runner Runner
	switch {
	case knock.Demo != nil:
		runner = knock.Demo
	case knock.Info != nil:
		runner = knock.Info
	case knock.Scan != nil:
		runner = knock.Scan
	case knock.Git != nil:
		runner = knock.Git
	default:
		knock.Logger.Crit("not specified runner")
		return
	}

	if err := runner.Open(); err != nil {
		knock.Logger.Crit("cannot open runner %T: %v", runner, err)
		return
	}
	defer runner.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(knock.Timeuot)*time.Second)
	defer cancel()

	knock.Logger.Debug("use runner: %T", runner)

	/* ---- runner ---- */
	// fork all the runner
	broker := knock.Broker(ctx, runner.Broker(ctx))
	for i := 0; i < knock.NumWorker; i++ {
		// run on the goroutine
		knock.wg.Add(1)
		go func(idx int) {
			defer knock.wg.Done()
			runner.Run(knock.receiver, broker)
			knock.Logger.Info("stop runner: #%d/%d", idx+1, knock.NumWorker)
		}(i)
	}

	// close the receiver if all runner closed
	go func() {
		knock.wg.Wait()
		knock.Logger.Info("stop the reducer")
		close(knock.receiver)
	}()

	/* ---- reducer ---- */
	knock.Reducer()
	return
}

// load the broker, wrapper the customized broker if exist, and set delay on each message
func (knock *Knock) Broker(ctx context.Context, broker <-chan string) (ch <-chan string) {
	tmp := make(chan string, 1)

	// make sure always only one work scanner
	go func() {
		defer close(tmp)

		switch broker {
		case nil:
			var scanner *bufio.Scanner

			switch knock.WordList {
			case "wordlists":
				reader := strings.NewReader(wordlists)
				scanner = bufio.NewScanner(reader)
			case "usernames":
				reader := strings.NewReader(usernames)
				scanner = bufio.NewScanner(reader)
			case "passwords":
				reader := strings.NewReader(passwords)
				scanner = bufio.NewScanner(reader)
			default:
				knock.Logger.Crit("not supported default word-lists: %#v", knock.WordList)
				return
			}

			// load the str
			for scanner.Scan() {
				select {
				case <-ctx.Done():
					return
				default:
					text := scanner.Text()
					tmp <- text
					// wait ?ms per each word list generated
					time.Sleep(time.Duration(knock.Wait) * time.Millisecond)
				}
			}
		default:
			for {
				msg, ok := <-broker
				if !ok {
					// end-of-message from customized broker
					return
				}

				// re-direct the message to the exposed broker
				tmp <- msg
				// wait ?ms per each word list generated
				time.Sleep(time.Duration(knock.Wait) * time.Millisecond)
			}
		}
	}()

	ch = tmp
	return
}

// receive the response from runner and show the result on STDOUT
func (knock *Knock) Reducer() {
	knock.Logger.Debug("start the reducer")
	cnt := 0
	progress := []string{"|", "/", "-", "\\"}

	newline := false
	isTerm := isatty.IsTerminal(os.Stdout.Fd())
	isStderrTerm := isatty.IsTerminal(os.Stderr.Fd())
	for {
		if resp, ok := <-knock.receiver; !ok {
			knock.Logger.Info("stop reducer")
			break
		} else {
			// show the message to the console
			switch resp.Type {
			case RESP_ERR:
				switch {
				case isTerm:
					os.Stdout.WriteString(fmt.Sprintf("\x1b[2K\x1b[1000D[!] %v\n", resp.Message))
				case isStderrTerm:
					os.Stderr.WriteString(fmt.Sprintf("\x1b[2K\x1b[1000D[!] %v\n", resp.Message))
				}
			case RESP_DEBUG:
				switch {
				case isTerm:
					os.Stdout.WriteString(fmt.Sprintf("\x1b[2K\x1b[1000D[?] %v\n", resp.Message))
				case isStderrTerm:
					os.Stderr.WriteString(fmt.Sprintf("\x1b[2K\x1b[1000D[?] %v\n", resp.Message))
				}
			case RESP_PROGRESS:
				switch {
				case isTerm:
					msg := fmt.Sprintf("\x1b[2K\x1b[1000D................................. %v %v", progress[cnt], resp.Message)
					os.Stdout.WriteString(msg)
					cnt = (cnt + 1) % len(progress)
					newline = true
				case isStderrTerm:
					msg := fmt.Sprintf("\x1b[2K\x1b[1000D................................. %v %v", progress[cnt], resp.Message)
					os.Stderr.WriteString(msg)
					cnt = (cnt + 1) % len(progress)
					newline = true
				}
			case RESP_RESULT:
				switch {
				case isTerm:
					os.Stdout.WriteString(fmt.Sprintf("\x1b[2K\x1b[1000D%v\n", resp.Message))
					newline = false
				default:
					os.Stdout.WriteString(fmt.Sprintf("%v\n", resp.Message))
					newline = false
				}
			default:
				os.Stdout.WriteString(fmt.Sprintf("[Unknown #%d] %v\n", resp.Type, resp.Message))
				newline = false
			}
		}
	}

	// show the extra NEWLINE
	switch {
	case isTerm:
		if newline {
			os.Stdout.WriteString("\x1b[2K\x1b[1000D\n")
		}
	case isStderrTerm:
		os.Stderr.WriteString("\x1b[2K\x1b[1000D\n")
	}
}
