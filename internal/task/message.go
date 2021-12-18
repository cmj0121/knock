package task

// type message status
type Status int

const (
	// the result of the task
	RESULT Status = iota
	// the error message
	ERROR
	// the debug message
	TRACE
)

type Message struct {
	Status

	Msg string
}
