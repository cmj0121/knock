package producer

import (
	"time"
)

// the pseudo word producer
type Producer interface {
	// set prefix and suffix
	Prefix(prefix string)
	Suffix(suffix string)

	// generate the word-list
	Produce(wait time.Duration) (ch <-chan string)

	// explicitly close the current producer
	Close()
}

type ProducerBase struct {
	prefix string
	suffix string
}

func (ctx *ProducerBase) Prefix(prefix string) {
	ctx.prefix = prefix
}

func (ctx *ProducerBase) Suffix(suffix string) {
	ctx.suffix = suffix
}
