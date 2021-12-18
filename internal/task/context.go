package task

// the context used in task
type Context struct {
	// signal the task should be closed
	Closed chan struct{}

	// the token producer (generate the token)
	Producer <-chan string

	// the message collector
	Collector chan<- Message
}
