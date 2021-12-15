package knock

import (
	"runtime"
	"fmt"
	"time"
	"os"
	"context"
    "os/signal"
    "syscall"
)

// the knock instance, generate the word list, pass to the task and then
// get the response.
type Knock struct {
	// number of the Worker
	Worker int

	// the shared channel to notify workers closed
	closed chan struct{}
}

func New() (knock *Knock) {
	knock = &Knock{
		Worker: runtime.NumCPU(),
		closed: make(chan struct{}, 1),
	}
	return
}

// run the knock with provides arguments
func (knock *Knock) Run() {
	// start all the worker
	for idx := 0; idx < knock.Worker; idx++ {
		go knock.DummyTask(knock.closed)
	}

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
	case <-done:
	}

	// notify all worker stop
	close(knock.closed)
	fmt.Println("all workers should be closed")

	// graceful shutdown
	fmt.Println("start graceful shutdown...")
	time.Sleep(timeout)
}

func (knock Knock) DummyTask(closed <-chan struct{}) {
	fmt.Println("start dummy task")
	<-closed
	fmt.Println("task closed")
}
