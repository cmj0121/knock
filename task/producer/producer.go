package producer

// the pseudo word producer
type Producer interface {
	// generate the word-list
	Produce() (ch <-chan string)

	// explicitly close the current producer
	Close()
}
