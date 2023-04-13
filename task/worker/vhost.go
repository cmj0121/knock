package worker

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"

	"github.com/cmj0121/knock/progress"
	"github.com/rs/zerolog/log"
)

func init() {
	worker := &VirtualHost{
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
type VirtualHost struct {
	*http.Client

	url *url.URL
}

// the unique name of worker
func (ctx VirtualHost) Name() string {
	return "vhost"
}

// show the help message
func (ctx VirtualHost) Help() string {
	return "list possible virtual host"
}

// the dummy open method
func (ctx *VirtualHost) Open(args ...string) (err error) {
	// check the wildcard IP address
	switch len(args) {
	case 0:
		err = fmt.Errorf("should pass hostname to the command %#v", ctx.Name())
	case 1:
		if ctx.url, err = url.Parse(args[0]); err == nil {
			switch ctx.url.Scheme {
			case "http":
			case "https":
			default:
				err = fmt.Errorf("should provides the protocol: %v", args[0])
				return
			}
		}
	default:
		err = fmt.Errorf("should pass one and only one hostname to the command %#v", ctx.Name())
	}
	return
}

// the dummy close method
func (ctx VirtualHost) Close() (err error) {
	log.Debug().Msg("dummy close")
	return
}

// execute the worker
func (ctx *VirtualHost) Run(producer <-chan string) (err error) {
	for word := range producer {
		if !ctx.validator(word) {
			continue
		}

		url := fmt.Sprintf("%v://%v.%v", ctx.url.Scheme, word, ctx.url.Host)

		log.Debug().Str("word", url).Msg("handle producer")
		progress.AddProgress(url)

		ctx.check(url, http.MethodGet)
	}
	return
}

// copy the current worker settings and generate a new instance
func (ctx *VirtualHost) Dup() (worker Worker) {
	worker = &VirtualHost{
		Client: ctx.Client,
		url:    ctx.url,
	}
	return
}

// check the folder with multiple possible HTTP method
func (ctx *VirtualHost) check(url, method string) (code int) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		progress.AddError(err)
		return
	}

	switch resp, err := ctx.Client.Do(req); err {
	case nil:
		switch code = resp.StatusCode; code {
		case 404:
		case 405:
		default:
			size, _ := io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			progress.AddText("%v %v (%v)", resp.StatusCode, url, size)
		}
	}

	return
}

// validate the hostname
func (ctx VirtualHost) validator(word string) (ok bool) {
	ok, _ = regexp.MatchString(`^[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+){0,254}$`, word)
	return
}
