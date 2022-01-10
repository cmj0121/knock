package knock

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cmj0121/knock/internal/task"
	"github.com/cmj0121/stropt"
)

// the knock instance, generate the word list, pass to the task and then
// get the response.
type Knock struct {
	stropt.Model

	// number milliseconds to wait per each task
	Wait time.Duration `shortcut:"w" desc:"number of milliseconds to wait per each task"`
	// number of the Worker
	Worker int `shortcut:"W" desc:"number of worker"`
	// set the progress message disable
	Silence bool `shortcut:"s" desc:"do not show the progress message"`

	Token *string `shortcut:"t" attr:"flag" desc:"test on the specified token"`

	*os.File `shortcut:"f" attr:"flag" desc:"external word-list file"`

	// the pre-defined task
	*task.Debug `desc:"show the tokens only (default action)"`
	*task.DNS   `desc:"try to find all possible DNS record"`
	*task.Web   `desc:"try to find all possible web path"`
	*task.Find  `desc:"try to find the host by given token"`

	// the shared channel to notify workers closed
	closed chan struct{}
	// the shared channel to notify main thread about all workers closed
	finished chan struct{}
	// the shared channel to collect the result from tasks
	ch_collector chan task.Message
}

func New() (knock *Knock) {
	knock = &Knock{
		Wait:   50 * time.Millisecond,
		Worker: runtime.NumCPU(),

		Debug: &task.Debug{},

		closed:       make(chan struct{}, 1),
		finished:     make(chan struct{}, 1),
		ch_collector: make(chan task.Message, 1),
	}
	return
}

// run the knock with provides arguments
func (knock *Knock) Run() (err error) {
	parser := stropt.MustNew(knock)
	parser.Version(Version())
	parser.Run()

	if knock.Wait == 0 {
		// set silence when wait=0
		knock.Silence = true
	}

	var runner task.Task
	switch {
	case knock.DNS != nil:
		runner = knock.DNS
	case knock.Web != nil:
		runner = knock.Web
	case knock.Find != nil:
		runner = knock.Find
	default:
		runner = knock.Debug
	}

	wg := sync.WaitGroup{}
	ctx := task.Context{
		Closed:    knock.closed,
		Wait:      knock.Wait,
		Collector: knock.ch_collector,
	}

	// run the reducer to receive message
	go knock.reducer()

	var mode task.TaskMode
	if mode, err = runner.Prologue(&ctx); err != nil {
		err = fmt.Errorf("%v prologue: %v", runner.Name(), err)
		return
	}

	ctx.Producer = knock.run_producer(&ctx, mode)

	// start all the worker
	for idx := 0; idx < knock.Worker; idx++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := runner.Execute(&ctx); err != nil {
				// catch error, show the message
				knock.ch_collector <- task.Message{
					Status: task.ERROR,
					Msg:    fmt.Sprintf("execute task %#v: %v", runner.Name(), err),
				}
			}
		}()
	}

	// wait all task finished, and notify main thread
	go func() {
		wg.Wait()
		runner.Epilogue(&ctx)
		close(knock.finished)
	}()

	// exactly run the knock, wait finished or catch Ctrl-C
	knock.run()
	return
}

// run the knock main thread and want tasks finished or force stop
// via Ctrl-C
func (knock *Knock) run() {
	sigint := make(chan os.Signal, 1)
	done := make(chan struct{}, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)

	// another go-routine for wait SIGNINT
	go func() {
		// wait signal or knock closed
		select {
		case <-knock.closed:
			// main thread closed
		case <-sigint:
			// catch Ctrl-C
		}

		// notify knock should be closed
		close(done)
	}()

	// wait either timeout or catch Ctrl-C
	select {
	case <-knock.finished:
		// all tasks finished
	case <-done:
		// catch Ctrl-C
	}

	knock.gradeful_shutdown()
}

func (knock *Knock) run_producer(ctx *task.Context, mode task.TaskMode) (producer <-chan string) {
	switch {
	case knock.Token != nil:
		producer = knock.producer(strings.NewReader(*knock.Token))
	case ctx.Producer != nil:
		// already create the producer
		producer = ctx.Producer
	case mode&task.M_NO_PRODUCER != 0:
		producer = knock.producer(strings.NewReader("."))
	case knock.File == nil:
		producer = knock.producer(strings.NewReader(word_lists))
	default:
		producer = knock.producer(knock.File)
	}

	return
}

// generate the tokens to the tasks
func (knock *Knock) producer(r io.Reader) (p <-chan string) {
	ch := make(chan string, 1)

	go func() {
		defer close(ch)

		scanner := bufio.NewScanner(r)
		token_buff := []string{}

		for scanner.Scan() {
			token := scanner.Text()
			token_buff = append(token_buff, token)
		}

		rand.Shuffle(len(token_buff), func(i, j int) {
			token_buff[i], token_buff[j] = token_buff[j], token_buff[i]
		})

		for idx := range token_buff {
			select {
			case <-knock.closed:
				return
			case ch <- token_buff[idx]:
			}

			time.Sleep(knock.Wait)
		}
	}()

	p = ch
	return
}

// show the message
func (knock *Knock) reducer() {
	progress := 0
	progress_bar := []string{"|", "/", "-", "\\"}
	show_progress := false

	defer func() {
		if show_progress {
			fmt.Printf("\x1b[s\x1b[2K\x1b[u")
		}
	}()

	for {
		select {
		case <-knock.closed:
			return
		case message := <-knock.ch_collector:
			show_progress = false

			switch message.Status {
			case task.RESULT:
				fmt.Printf("[%v] ........................ %v\n", "+", message.Msg)
			case task.ERROR:
				fmt.Printf("[%v] ........................ %v\n", "!", message.Msg)
			case task.TRACE:
				// 2K clear entire line and cursor position does not change
				//  s saves the cursor position/state in SCO console mode
				//  u restores the cursor position/state in SCO console mode
				if !knock.Silence {
					fmt.Printf("\x1b[s\x1b[2K[%v] ........................ %v\x1b[u", progress_bar[progress], message.Msg)
					progress = (progress + 1) % len(progress_bar)
				}
				show_progress = true
			default:
				fmt.Printf("[%v] ........................ %v\n", "?", message.Msg)
			}
		}
	}
}

// the post-script for the Knock.
func (knock *Knock) gradeful_shutdown() {
	// notify reducer and all worker stop
	close(knock.closed)
	// wait 1 seconds
	time.Sleep(time.Second)
	// the final stesp need to execute after Knock stop
	close(knock.ch_collector)
	// exit program
	fmt.Println("\n~ Bye ~")
	os.Exit(0)
}
