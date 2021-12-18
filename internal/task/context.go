package task

// the context used in task
type Context struct {
	// signal the task should be closed
	Closed chan struct{}
}
