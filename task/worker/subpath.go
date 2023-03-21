package worker

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	"github.com/cmj0121/knock/progress"
	"github.com/rs/zerolog/log"
)

func init() {
	worker := &SubPath{
		Client: &http.Client{
			// disable auto-redirect
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			// allow insecure HTTPs
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
	Register(worker)
}

// the debugger worker and just show the word in STDOUT
type SubPath struct {
	*http.Client
	// the target hostname/IP
	hostname string
}

// the unique name of worker
func (s SubPath) Name() string {
	return "subp"
}

// show the help message
func (s SubPath) Help() string {
	return "list possible path"
}

// the dummy open method
func (s *SubPath) Open(args ...string) (err error) {
	// check the wildcard IP address
	switch len(args) {
	case 0:
		err = fmt.Errorf("should pass hostname to the command %#v", s.Name())
	case 1:
		s.hostname = args[0]
	default:
		err = fmt.Errorf("should pass one and only one hostname to the command %#v", s.Name())
	}
	return
}

// the dummy close method
func (s SubPath) Close() (err error) {
	log.Debug().Msg("dummy close")
	return
}

// execute the worker
func (s *SubPath) Run(producer <-chan string) (err error) {
	for word := range producer {
		log.Debug().Str("word", word).Msg("handle producer")
		progress.AddProgress(word)

		url := fmt.Sprintf("%v/%v", s.hostname, word)
		switch resp, err := s.Client.Get(url); err {
		case nil:
			switch resp.StatusCode {
			case 404:
			default:
				size, _ := io.Copy(io.Discard, resp.Body)
				resp.Body.Close()

				progress.AddText("%v %v (%v)", resp.StatusCode, url, size)
			}
		default:
			progress.AddError(err)
		}
	}
	return
}

// copy the current worker settings and generate a new instance
func (s *SubPath) Dup() (worker Worker) {
	worker = &SubPath{
		Client:   s.Client,
		hostname: s.hostname,
	}
	return
}
