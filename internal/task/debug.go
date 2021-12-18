package task

import (
	"fmt"
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
	for {
		select {
		case token, running := <-ctx.Producer:
			if !running {
				// no-more token, close the task
				return
			}
			// print the token
			fmt.Println(token)
		case <-ctx.Closed:
			// closed by the main thread
			return
		}
	}
}
