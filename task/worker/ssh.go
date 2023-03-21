package worker

import (
	"fmt"

	"github.com/cmj0121/knock/progress"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

func init() {
	worker := &SSH{}
	Register(worker)
}

// the debugger worker and just show the word in STDOUT
type SSH struct {
	target   string
	username string
}

// the unique name of worker
func (SSH) Name() string {
	return "ssh"
}

// show the help message
func (SSH) Help() string {
	return "ssh brute-force cracker"
}

// the dummy open method
func (s *SSH) Open(args ...string) (err error) {
	switch len(args) {
	case 2:
		s.target = args[0]
		s.username = args[1]
	default:
		err = fmt.Errorf("should pass `target` `username` to the command %#v", s.Name())
	}
	return
}

// the dummy close method
func (s *SSH) Close() (err error) {
	log.Error().Msg("dummy close")
	return
}

// execute the worker
func (s *SSH) Run(producer <-chan string) (err error) {
	for word := range producer {
		log.Debug().Str("word", word).Msg("handle producer")
		progress.AddProgress(word)

		if s.check(word) {
			progress.AddText("%v:%v on %v", s.username, word, s.target)
		}
	}

	return
}

// copy the current worker settings and generate a new instance
func (s *SSH) Dup() (worker Worker) {
	worker = &SSH{
		target:   s.target,
		username: s.username,
	}
	return
}

func (s *SSH) check(password string) bool {
	config := &ssh.ClientConfig{
		User: s.username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	switch _, err := ssh.Dial("tcp", s.target, config); err {
	case nil:
		return true
	default:
		log.Info().Err(err).Msg("cannot access to SSH server")
		return false
	}
}
