package knock

import (
	"context"
)

// list all the word-list as the demo runner
type Demo struct {
}

func (demo *Demo) Open() (err error) {
	err = nil
	return
}

func (demo *Demo) Close() (err error) {
	err = nil
	return
}

func (demo *Demo) Run(receiver chan<- Response, broker <-chan string) {
	for {
		msg, ok := <-broker
		if !ok {
			// finished all message from the broker
			break
		}
		// send the progress
		receiver <- Response{
			Type:    RESP_PROGRESS,
			Message: msg,
		}
	}
}

func (demo *Demo) Broker(ctx context.Context) (ch <-chan string) {
	ch = nil
	return
}
