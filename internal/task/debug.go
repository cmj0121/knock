package task

import (
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

// run the necessary prepared actions before executed
func (debug Debug) Prologue(ctx *Context) {
}

// run the necessary clean-up actions after task finished
func (debug Debug) Epilogue(ctx *Context) {
}

// execute the debug, show the word and wait closed
func (debug Debug) Execute(ctx *Context) (err error) {
	for {
		select {
		case token, running := <-ctx.Producer:
			if !running {
				// no-more token, close the task
				return
			}
			// print the token
			ctx.Collector <- Message{
				Status: TRACE,
				Msg:    token,
			}

			time.Sleep(time.Millisecond * 150)
		case <-ctx.Closed:
			// closed by the main thread
			return
		}
	}
}
