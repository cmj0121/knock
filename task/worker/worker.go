package worker

// the pseudo worker to handle word producer
type Worker interface {
	// the unique name of worker
	Name() string

	// open with necessary resources
	Open() error

	// close the allocated resources
	Close() error

	// execute the worker
	Run(producer <-chan string) error
}
