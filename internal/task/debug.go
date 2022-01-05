package task

import (
	"github.com/cmj0121/stropt"
)

// task used for debug, only show the work knock passed
type Debug struct {
	stropt.Model
}

// show the unique name of the task
func (debug Debug) Name() (name string) {
	name = "debug"
	return
}

// run the necessary prepared actions before executed
func (debug Debug) Prologue(ctx *Context) (mode TaskMode, err error) {
	return
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
		case <-ctx.Closed:
			// closed by the main thread
			return
		}
	}
}
