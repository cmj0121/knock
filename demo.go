package knock

import (
	"io"
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

// does not provide the customized word-list
func (demo *Demo) Reader() (r io.Reader) {
	r = nil
	return
}
