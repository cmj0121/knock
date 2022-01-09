package task

import (
	"time"
)

// the context used in task
type Context struct {
	// signal the task should be closed
	Closed chan struct{}

	// the token producer (generate the token)
	Producer <-chan string
	Wait     time.Duration

	// the message collector
	Collector chan<- Message
}
