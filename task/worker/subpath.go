package worker

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/cmj0121/knock/progress"
	"github.com/rs/zerolog/log"
)

func init() {
	worker := &SubPath{
		Client: &http.Client{
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

	url       *url.URL
	final_url string
	final_txt []byte
}

// the unique name of worker
func (ctx SubPath) Name() string {
	return "subp"
}

// show the help message
func (ctx SubPath) Help() string {
	return "list possible path"
}

// the dummy open method
func (ctx *SubPath) Open(args ...string) (err error) {
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

	ctx.epologue()
	return
}

func (ctx *SubPath) epologue() {
	path := fmt.Sprintf("%v/abcdefg", ctx.url)
	if resp, err := http.Get(path); err == nil {
		final_url := resp.Request.URL.String()
		if resp.StatusCode == 200 && path != final_url {
			defer resp.Body.Close()

			data, _ := io.ReadAll(resp.Body)

			ctx.final_url = final_url
			ctx.final_txt = data
		}
	}
}

// the dummy close method
func (ctx SubPath) Close() (err error) {
	log.Debug().Msg("dummy close")
	return
}

// execute the worker
func (ctx *SubPath) Run(producer <-chan string) (err error) {
	for word := range producer {
		log.Debug().Str("word", word).Msg("handle producer")
		progress.AddProgress(word)

		url := fmt.Sprintf("%v/%v", ctx.url, word)
		ctx.check(url, http.MethodGet)
		ctx.check(url, http.MethodPost)
		ctx.check(url, http.MethodPut)
		ctx.check(url, http.MethodDelete)
	}
	return
}

// copy the current worker settings and generate a new instance
func (ctx *SubPath) Dup() (worker Worker) {
	worker = &SubPath{
		Client:    ctx.Client,
		url:       ctx.url,
		final_url: ctx.final_url,
	}
	return
}

// check the folder with multiple possible HTTP method
func (ctx *SubPath) check(url, method string) (code int) {
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
			final_url := resp.Request.URL.String()
			data, _ := io.ReadAll(resp.Body)

			if final_url != ctx.final_url && bytes.Equal(data, ctx.final_txt) {
				progress.AddText("%-6v %v %v (%v)", method, resp.StatusCode, url, len(data))
			}

			resp.Body.Close()
		}
	default:
		progress.AddError(err)
	}

	return
}
