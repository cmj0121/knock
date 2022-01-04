package task

// the interface of the task executed by the knock.
type Task interface {
	// the name of the task, should be unique in the task pools
	Name() string

	// run the necessary prepared actions before executed
	Prologue(ctx *Context)

	// run the necessary clean-up actions after task finished
	Epilogue(ctx *Context)

	// execute the task with passed context, return error when fail
	Execute(ctx *Context) error
}
