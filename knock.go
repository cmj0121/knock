package knock

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/cmj0121/knock/internal/task"
)

// the knock instance, generate the word list, pass to the task and then
// get the response.
type Knock struct {
	// number of the Worker
	Worker int

	// the shared channel to notify workers closed
	closed chan struct{}
	// the shared channel to notify main thread about all workers closed
	finished chan struct{}
}

func New() (knock *Knock) {
	knock = &Knock{
		Worker:   runtime.NumCPU(),
		closed:   make(chan struct{}, 1),
		finished: make(chan struct{}, 1),
	}
	return
}

// run the knock with provides arguments
func (knock *Knock) Run() {
	wg := sync.WaitGroup{}
	task_name := "debug"

	switch runner, ok := task.GetTask(task_name); ok {
	case false:
		fmt.Printf("cannot find task: %v\n", task_name)
		return
	default:
		// start all the worker
		for idx := 0; idx < knock.Worker; idx++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				ctx := task.Context{
					Closed: knock.closed,
				}

				if err := runner.Execute(&ctx); err != nil {
					// catch error, show the message
					fmt.Println(err)
				}
			}()
		}
	}

	// wait all task finished, and notify main thread
	go func() {
		wg.Wait()
		close(knock.finished)
	}()

	// exactly run the knock, wait finished or catch Ctrl-C
	knock.run()
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
			fmt.Println("main thread closed")
		case <-sigint:
			fmt.Println("catch Ctrl-C")
		}

		// notify knock should be closed
		close(done)
	}()

	// set timeout for the main process
	timeout := 4 * time.Second
	timeout_ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// wait either timeout or catch Ctrl-C
	select {
	case <-timeout_ctx.Done():
		// timeout
	case <-knock.finished:
		// all tasks finished
	case <-done:
		// catch Ctrl-C
	}

	knock.gradeful_shutdown()
}

// the post-script for the Knock.
func (knock *Knock) gradeful_shutdown() {
	timeout := 4 * time.Second

	// notify all worker stop
	close(knock.closed)
	fmt.Println("all workers should be closed")

	// graceful shutdown
	fmt.Println("start graceful shutdown...")
	time.Sleep(timeout)
}
