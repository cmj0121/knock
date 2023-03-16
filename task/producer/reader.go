package producer

import (
	"bufio"
	"io"
	"time"

	"github.com/rs/zerolog/log"
)

func NewReaderProducer(r io.Reader) *ReaderProducer {
	return &ReaderProducer{
		Reader: r,
		Closed: make(chan struct{}, 1),
	}
}

// the word producer by the known io.Reader, separate by the newline
type ReaderProducer struct {
	// the data reader
	io.Reader

	// the signle for close the current connection and the subscriber
	// should close all allocated resources.
	Closed chan struct{}
}

// produce the word list from the current context
func (ctx *ReaderProducer) Produce(wait time.Duration) (ch <-chan string) {
	tmp := make(chan string, 1)

	go func() {
		defer close(tmp)

		scanner := bufio.NewScanner(ctx.Reader)
		scanner.Split(bufio.ScanWords)

		for {
			if !scanner.Scan() {
				log.Debug().Msg("end-of-word")
				return
			}

			select {
			case tmp <- scanner.Text():
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
func (ctx *ReaderProducer) Close() {
	close(ctx.Closed)
}
