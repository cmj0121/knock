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
)

// the knock interface which provide the global setting
type Knock struct {
	argparse.Model

	sync.Mutex `-`

	// the internal logger
	*logger.Logger `-`
	LogLevel       string `name:"log" choices:"warn info debug verbose" help:"log level"`

	// the display format
	Format string `short:"f" default:"yaml" choices:"table yaml json" help:"output format"`

	// number of worker, default #CPU
	NumWorker int `short:"w" name:"worker" help:"number of worker"`

	// default word-list
	WordListFile *os.File `short:"W" name:"word-list" args:"option" help:"default system-wide word-list"`

	// the global timeout based on seconds
	Timeuot int `short:"t" help:"global timeout based on seconds"`

	*Demo `help:"list all the word-list"`
	*Info `help:"show the current system info"`

	/* ---- private fields */
	receiver chan Response
	wg       sync.WaitGroup
}

// new the knock instance with default config
func New() (knock *Knock) {
	knock = &Knock{
		Logger:    logger.New(PROJ_NAME),
		NumWorker: runtime.NumCPU(),
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(knock.Timeuot)*time.Second)
	defer cancel()

	var runner Runner

	switch {
	case knock.Demo != nil:
		runner = knock.Demo
	case knock.Info != nil:
		runner = knock.Info
	default:
		return
	}

	/* ---- runner ---- */
	// fork all the runner
	broker := knock.WordList(ctx)
	for i := 0; i < knock.NumWorker; i++ {
		// run on the goroutine
		knock.wg.Add(1)
		go func() {
			defer knock.wg.Done()
			runner.Run(broker, knock.receiver)
		}()
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

// list of the word-list
func (knock *Knock) WordList(ctx context.Context) (ch <-chan string) {
	tmp := make(chan string, 1)

	// make sure always only one work scanner
	go func() {
		knock.Lock()
		defer knock.Unlock()

		var scanner *bufio.Scanner

		switch {
		case knock.WordListFile == nil:
			// load the default word lists
			knock.Logger.Info("load the default word-list")
			scanner = bufio.NewScanner(strings.NewReader(wordlists))
		default:
			knock.Logger.Info("load the customized word-list: %#v", knock.WordListFile.Name())
			scanner = bufio.NewScanner(knock.WordListFile)
		}

		// load the str
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				close(tmp)
				return
			default:
				text := scanner.Text()
				tmp <- text
			}
		}
		close(tmp)
		return
	}()

	ch = tmp
	return
}

func (knock *Knock) Reducer() {
	knock.Logger.Debug("start the reducer")
	cnt := 0
	progress := []string{"|", "/", "-", "\\"}

	newline := false
	for {
		if resp, ok := <-knock.receiver; !ok {
			knock.Logger.Info("stop reducer")
			break
		} else {
			// show the message to the console
			switch resp.Type {
			case RESP_PROGRESS:
				msg := fmt.Sprintf("\x1b[2K\x1b[1000D.................................... %v %v", progress[cnt], resp.Message)
				os.Stdout.WriteString(msg)
				cnt = (cnt + 1) % len(progress)
				newline = true
			case RESP_RESULT:
				os.Stdout.WriteString(fmt.Sprintf("\x1b[2K\x1b[1000D%v\n", resp.Message))
				newline = false
			default:
				os.Stdout.WriteString(fmt.Sprintf("[Unknown #%d] %v\n", resp.Type, resp.Message))
				newline = false
			}
		}
	}

	if newline {
		// show the extra NEWLINE
		os.Stdout.WriteString("\n")
	}
}
