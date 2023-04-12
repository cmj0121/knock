package producer

import (
	"time"
)

// the pseudo word producer
type Producer interface {
	// set prefix
	Prefix(prefix string)

	// generate the word-list
	Produce(wait time.Duration) (ch <-chan string)

	// explicitly close the current producer
	Close()
}
