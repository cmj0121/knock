package task

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cmj0121/stropt"
)

type Web struct {
	stropt.Model

	// the base URL path of the target
	Base *string `desc:"the base URL"`

	*http.Client `-` //nolint
}

// show the unique name of the task
func (web Web) Name() (name string) {
	name = "web"
	return
}

// initial the http.Client
func (web *Web) Prologue(ctx *Context) (err error) {
	// The TLS setting
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	// The default http client
	web.Client = &http.Client{
		Transport: tr,
	}

	return
}

// close everything
func (web *Web) Epilogue(ctx *Context) {
}

// check the web link exist
func (web *Web) Execute(ctx *Context) (err error) {
	for {
		select {
		case token, running := <-ctx.Producer:
			if !running {
				// no-more token, close the task
				return
			}

			path := fmt.Sprintf("%v/%v", *web.Base, token)
			// print the token
			ctx.Collector <- Message{
				Status: TRACE,
				Msg:    path,
			}

			code, _, err := web.Do("GET", path)
			switch {
			case err != nil:
				ctx.Collector <- Message{
					Status: ERROR,
					Msg:    err.Error(),
				}
			default:
				ctx.Collector <- Message{
					Status: RESULT,
					Msg:    fmt.Sprintf("[%v] %v", code, path),
				}
			}
		case <-ctx.Closed:
			// closed by the main thread
			return
		}
	}
}

func (web Web) Do(method, url string) (code int, html []byte, err error) {
	var req *http.Request

	if req, err = http.NewRequest(method, url, nil); err == nil {
		var resp *http.Response
		if resp, err = web.Client.Do(req); err == nil {
			code = resp.StatusCode
			html, _ = ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
		}
	}

	return
}
