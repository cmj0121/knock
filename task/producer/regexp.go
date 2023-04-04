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

// produce the word via the regexp pattern
func (ctx *RegexpProducer) Produce(wait time.Duration) (ch <-chan string) {
	tmp := make(chan string, 1)

	go func() {
		defer close(tmp)

		for word := range NewState(ctx.Regexp).Next() {
			select {
			case tmp <- word:
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

// generate the possible word list
func (s *State) Next() (ch <-chan string) {
	ch = s.next(s.Regexp)
	return
}

func (s *State) next(re *syntax.Regexp, others ...*syntax.Regexp) (ch chan string) {
	ch = make(chan string, 1)

	go func() {
		defer close(ch)

		switch re.Op {
		case syntax.OpEmptyMatch:
			// matches empty string
		case syntax.OpLiteral:
			switch len(others) {
			case 0:
				ch <- string(re.Rune)
			default:
				for next := range s.next(others[0], others[1:]...) {
					ch <- string(re.Rune) + next
				}
			}
		case syntax.OpCharClass:
			// matches Runes interpreted as range pair list
			for idx := 0; idx < len(re.Rune); idx += 2 {
				for word := re.Rune[idx]; word <= re.Rune[idx+1]; word++ {
					switch len(others) {
					case 0:
						ch <- string(word)
					default:
						for next := range s.next(others[0], others[1:]...) {
							ch <- string(word) + next
						}
					}
				}
			}
		case syntax.OpConcat:
			// matches concatenation of Subs
			for word := range s.next(re.Sub[0]) {
				for next := range s.next(re.Sub[1], re.Sub[2:]...) {
					ch <- word + next
				}
			}
		case syntax.OpRepeat:
			// matches Sub[0] at least Min times, at most Max (Max == -1 is no limit)
			for repeat := re.Min; repeat <= re.Max; repeat++ {
				for word := range s.repeat(repeat, re.Sub[0], re.Sub[1:]...) {
					ch <- word
				}
			}
		default:
			log.Error().Str("Op", re.Op.String()).Interface("re", re).Msg("not implemented")
		}
	}()

	return
}

func (s *State) repeat(count int, re *syntax.Regexp, others ...*syntax.Regexp) (ch chan string) {
	ch = make(chan string, 1)
	go func() {
		defer close(ch)

		switch count {
		case 0:
			return
		case 1:
			for word := range s.next(re, others...) {
				ch <- word
			}
		default:
			for word := range s.next(re, others...) {
				for next := range s.repeat(count-1, re, others...) {
					ch <- word + next
				}
			}
		}
	}()
	return ch
}
