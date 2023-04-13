package producer

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	_ "embed"

	"github.com/rs/zerolog/log"
)

//go:embed assets/default.txt
var words string

func NewDefaultProducer() *DefaultProducer {
	return &DefaultProducer{
		Closed: make(chan struct{}, 1),
	}
}

// the word producer by the known io.Reader, separate by the newline
type DefaultProducer struct {
	ProducerBase

	// the signle for close the current connection and the subscriber
	// should close all allocated resources.
	Closed chan struct{}
}

// produce the word list from the current context
func (ctx *DefaultProducer) Produce(wait time.Duration) (ch <-chan string) {
	tmp := make(chan string, 1)

	go func() {
		defer close(tmp)

		reader := strings.NewReader(words)
		scanner := bufio.NewScanner(reader)
		scanner.Split(bufio.ScanWords)

		for {
			if !scanner.Scan() {
				log.Debug().Msg("end-of-word")
				return
			}

			text := fmt.Sprintf("%v%v", ctx.prefix, scanner.Text())
			select {
			case tmp <- text:
			case <-ctx.Closed:
				log.Debug().Msg("explicitly stop the word producer")
				return
			}

			time.Sleep(wait)
		}
	}()

	ch = tmp
	return
}

// explicitly close the current producer
func (ctx *DefaultProducer) Close() {
	close(ctx.Closed)
}
