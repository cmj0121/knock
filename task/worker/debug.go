package worker

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

// the debugger worker and just show the word in STDOUT
type Debug struct {
}


// the unique name of worker
func (Debug) Name() string {
	return "Debug"
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

func (Debug) Run(producer <-chan string) (err error) {
	for word := range producer {
		log.Debug().Str("word", word).Msg("handle producer")
		fmt.Println(word)
	}

	return
}
