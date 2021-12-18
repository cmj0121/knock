package task

import (
	"context"
	"fmt"
	"time"
)

func init() {
	// register debug task
	RegisterTask(Debug{})
}

// task used for debug, only show the work knock passed
type Debug struct {
}

// show the unique name of the task
func (debug Debug) Name() (name string) {
	name = "debug"
	return
}

// execute the debug, show the word and wait closed
func (debug Debug) Execute(ctx *Context) (err error) {
	fmt.Println("start debug task")

	timeout := 2 * time.Second
	timeout_ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case <-timeout_ctx.Done():
		// closed for task finished
	case <-ctx.Closed:
		// closed by the main thread
	}

	fmt.Println("task closed")
	return
}
