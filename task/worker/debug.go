package worker

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

func init() {
	worker := Debug{}
	Register(worker)
}

// the debugger worker and just show the word in STDOUT
type Debug struct {
}

// the unique name of worker
func (Debug) Name() string {
	return "debug"
}

// show the help message
func (Debug) Help() string {
	return "show the words from producer"
}

// the dummy open method
func (Debug) Open() (err error) {
	log.Debug().Msg("dummy open")
	return
}

// the dummy close method
func (Debug) Close() (err error) {
	log.Debug().Msg("dummy close")
	return
}

// execute the worker
func (Debug) Run(producer <-chan string) (err error) {
	for word := range producer {
		log.Debug().Str("word", word).Msg("handle producer")
		fmt.Println(word)
	}

	return
}

// copy the current worker settings and generate a new instance
func (Debug) Dup() (worker Worker) {
	worker = Debug{}
	return
}
