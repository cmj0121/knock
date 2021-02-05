package knock

// list all the word-list as the demo runner
type Demo struct {
}

func (demo *Demo) Run(broker <-chan string, receiver chan<- Response) {
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

		// process the data
	}
}
