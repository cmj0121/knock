package producer

import (
	"regexp/syntax"
	"time"

	"github.com/rs/zerolog/log"
)

func NewRegexpProducer(s string) (producer *RegexpProducer, err error) {
	var re *syntax.Regexp

	log.Debug().Str("syntax", s).Msg("compile the regexp")
	if re, err = syntax.Parse(s, syntax.Perl); err == nil {
		producer = &RegexpProducer{
			Regexp: re,
			Closed: make(chan struct{}, 1),
		}
	}
	return
}

// the Regexp-based random string generator
type RegexpProducer struct {
	// the compiled regular-expression syntax
	*syntax.Regexp

	// the signle for close the current connection and the subscriber
	// should close all allocated resources.
	Closed chan struct{}
}

// produce the token via the regexp pattern
func (ctx *RegexpProducer) Produce(wait time.Duration) (ch <-chan string) {
	tmp := make(chan string, 1)

	go func() {
		defer close(tmp)

		for token := range NewState(ctx.Regexp).Next() {
			select {
			case tmp <- token:
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
func (ctx *RegexpProducer) Close() {
	close(ctx.Closed)
}

type State struct {
	*syntax.Regexp
}

// create a new state from regexp pattern
func NewState(re *syntax.Regexp) *State {
	return &State{
		Regexp: re,
	}
}

// generate the possible token list
func (s *State) Next() (ch <-chan string) {
	tmp := make(chan string, 1)

	go func() {
		defer close(tmp)

		switch s.Regexp.Op {
		case syntax.OpEmptyMatch:
			// matches empty string
		case syntax.OpLiteral:
			tmp <- string(s.Regexp.Rune)
		case syntax.OpCharClass:
			// matches Runes interpreted as range pair list
			for idx := 0; idx < len(s.Regexp.Rune); idx += 2 {
				for ch := s.Regexp.Rune[idx]; ch <= s.Regexp.Rune[idx+1]; ch++ {
					tmp <- string(ch)
				}
			}
		default:
			log.Error().Str("Op", s.Regexp.Op.String()).Interface("re", s.Regexp).Msg("not implemented")
			return
		}
	}()

	ch = tmp
	return
}
