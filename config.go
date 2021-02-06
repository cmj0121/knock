package knock

import (
	"fmt"

	_ "embed"
)

const (
	PROJ_NAME = "knock"

	MAJOR = 1
	MINOR = 0
	MACRO = 0
)

func Version() (ver string) {
	ver = fmt.Sprintf("%v (v%d.%d.%d)", PROJ_NAME, MAJOR, MINOR, MACRO)
	return
}

var (
	//go:embed assets/wordlists
	wordlists string
)

type ResponseType int

const (
	RESP_ERR = iota
	RESP_PROGRESS
	RESP_RESULT
)

// the response when send the request from the Knock
type Response struct {
	// message type
	Type ResponseType

	// the message of the response
	Message string
}

// the runner that receive the word-list from broker and reply the response to receiver
type Runner interface {
	Run(broker <-chan string, receiver chan<- Response)
}
