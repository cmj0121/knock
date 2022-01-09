package task

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/cmj0121/stropt"
)

type Web struct {
	stropt.Model

	// the base URL path of the target
	Base *string `desc:"the base URL"`

	// skip the 404 webpage
	Show404 bool `name:"404" desc:"show result of the 404 web page"`
	// scan the sensitive page
	Sensitive bool `shortcut:"s" desc:"only scan the sensitive info in target URL"`
	// only works when set sensitive
	NoComment bool `shortcut:"C" name:"no-comment" desc:"show sensitive comment"`
	NoHost    bool `shortcut:"H" name:"no-host" desc:"show sensitive host"`

	UserAgent string `name:"user-agent" default:"knock/web" desc:"customized user agent"`

	*http.Client   `-` //nolint
	base_url       string
	html_main_page []byte
}

// show the unique name of the task
func (web Web) Name() (name string) {
	name = "web"
	return
}

// initial the http.Client
func (web *Web) Prologue(ctx *Context) (mode TaskMode, err error) {
	// The TLS setting
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	// The default http client
	web.Client = &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	switch {
	case strings.HasPrefix(*web.Base, "http://"):
	case strings.HasPrefix(*web.Base, "https://"):
	default:
		// no http schema, append http by-default
		*web.Base = fmt.Sprintf("http://%v", *web.Base)
	}

	var u *url.URL
	if u, err = url.ParseRequestURI(*web.Base); err == nil {
		// set the HTTP request URL
		web.base_url = u.String()
	}

	_, _, web.html_main_page, _ = web.Do("GET", *web.Base)

	switch {
	case web.Sensitive:
		mode = M_NO_PRODUCER
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

			switch {
			case web.Sensitive:
				web.scan_web_sensitive(ctx)
			default:
				path := fmt.Sprintf("%v/%v", web.base_url, token)
				web.scan_web_path(ctx, path)
			}
		case <-ctx.Closed:
			// closed by the main thread
			return
		}
	}
}

func (web *Web) scan_web_path(ctx *Context, path string) {
	// print the token
	ctx.Collector <- Message{
		Status: TRACE,
		Msg:    path,
	}

	code, header, html, err := web.Do("GET", path)
	switch {
	case err != nil:
		ctx.Collector <- Message{
			Status: ERROR,
			Msg:    err.Error(),
		}
	default:
		location := header.Get("Location")

		switch code {
		case 301, 302:
			if path+"/" == location {
				ctx.Collector <- Message{
					Status: RESULT,
					Msg:    fmt.Sprintf("[%v] %v", code, path),
				}
			} else {
				ctx.Collector <- Message{
					Status: RESULT,
					Msg:    fmt.Sprintf("[%v] %-42v -> %v", code, path, location),
				}
			}
		case 404:
			if web.Show404 {
				ctx.Collector <- Message{
					Status: RESULT,
					Msg:    fmt.Sprintf("[%v] %v", code, path),
				}
			}
		default:
			switch {
			case bytes.Equal(html, web.html_main_page):
			default:
				ctx.Collector <- Message{
					Status: RESULT,
					Msg:    fmt.Sprintf("[%v] %v", code, path),
				}
			}
		}
	}
}

func (web *Web) scan_web_sensitive(ctx *Context) {
	if !web.NoComment {
		re_comment := regexp.MustCompile(`<!--.+?-->`)
		for _, comment := range re_comment.FindAll(web.html_main_page, -1) {
			ctx.Collector <- Message{
				Status: RESULT,
				Msg:    fmt.Sprintf("[comment] %v", string(comment)),
			}
		}
	}

	if !web.NoHost {
		re_link := regexp.MustCompile(`(src|href|action)=['"]((http|https):)?//.*?['"]`)
		for _, comment := range re_link.FindAll(web.html_main_page, -1) {
			ctx.Collector <- Message{
				Status: RESULT,
				Msg:    fmt.Sprintf("[link] %v", string(comment)),
			}
		}
	}
}

func (web Web) Do(method, url string) (code int, header http.Header, html []byte, err error) {
	var req *http.Request

	if req, err = http.NewRequest(method, url, nil); err == nil {
		// override the useragent
		req.Header.Set("User-Agent", web.UserAgent)

		var resp *http.Response
		if resp, err = web.Client.Do(req); err == nil {
			code = resp.StatusCode
			header = resp.Header
			html, _ = ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
		}
	}

	return
}
