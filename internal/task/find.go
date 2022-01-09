package task

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/cmj0121/stropt"
	"github.com/go-ping/ping"
)

type Find struct {
	stropt.Model

	Timeout time.Duration `default:"100ms" desc:"timeout when try to connect to host"`

	Port  int  `shortcut:"p" desc:"the port of the web request"`
	HTTP  bool `shortcut:"w" desc:"try to get from web / HTTP"`
	HTTPS bool `shortcut:"s" desc:"try to get from web / HTTPS"`

	UserAgent string `name:"user-agent" default:"knock/web" desc:"customized user agent"`

	// the target CIDR want to search
	CIDR *net.IPNet `default:"127.0.0.1/24" desc:"the target CIDR"`

	*http.Client `-` //nolint
}

// show the unique name of the task
func (find Find) Name() (name string) {
	name = "find"
	return
}

// run the necessary prepared actions before executed
func (find *Find) Prologue(ctx *Context) (mode TaskMode, err error) {
	switch find.CIDR {
	case nil:
	default:
		// set the customized producer
		ctx.Producer = CIDRProducer(ctx, find.CIDR)
	}

	// The TLS setting
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	// The default http client
	timeout := find.Timeout
	if timeout < time.Second {
		// HTTP timeout should >= 1s
		timeout = time.Second
	}

	find.Client = &http.Client{
		Transport: tr,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return
}

// run the necessary clean-up actions after task finished
func (find Find) Epilogue(ctx *Context) {
}

// find the host by the given token
func (find *Find) Execute(ctx *Context) (err error) {
	for {
		select {
		case token, running := <-ctx.Producer:
			if !running {
				// no-more token, close the task
				return
			}
			// print the token
			ctx.Collector <- Message{
				Status: TRACE,
				Msg:    token,
			}

			switch {
			case find.HTTP:
				switch find.Port {
				case 0:
					url := fmt.Sprintf("http://%v", token)
					find.HTTPGet(ctx, url)
				default:
					url := fmt.Sprintf("http://%v:%v", token, find.Port)
					find.HTTPGet(ctx, url)
				}
			case find.HTTPS:
				switch find.Port {
				case 0:
					url := fmt.Sprintf("https://%v", token)
					find.HTTPGet(ctx, url)
				default:
					url := fmt.Sprintf("https://%v:%v", token, find.Port)
					find.HTTPGet(ctx, url)
				}
			default:
				find.Ping(ctx, token)
			}
		case <-ctx.Closed:
			// closed by the main thread
			return
		}
	}
}

// ping the target and show the IP if host alive
func (find Find) Ping(ctx *Context, ip string) {
	switch pinger, err := ping.NewPinger(ip); err {
	case nil:
		pinger.Count = 1
		pinger.Timeout = find.Timeout
		pinger.OnRecv = func(pkt *ping.Packet) {
			ctx.Collector <- Message{
				Status: RESULT,
				Msg:    ip,
			}
		}
		pinger.Run() //nolint
	default:
		ctx.Collector <- Message{
			Status: ERROR,
			Msg:    fmt.Sprintf("cannot build ping %v: %v", ip, err),
		}
	}
}

func (find *Find) HTTPGet(ctx *Context, url string) {
	if req, err := http.NewRequest("GET", url, nil); err == nil {
		// override the useragent
		req.Header.Set("User-Agent", find.UserAgent)

		var resp *http.Response
		if resp, err = find.Client.Do(req); err == nil {
			switch {
			case resp.StatusCode >= 300 && resp.StatusCode < 400:
				location := resp.Header.Get("Location")
				ctx.Collector <- Message{
					Status: RESULT,
					Msg:    fmt.Sprintf("[%v] %-22v -> %v", resp.StatusCode, url, location),
				}
			default:
				ctx.Collector <- Message{
					Status: RESULT,
					Msg:    fmt.Sprintf("[%v] %v", resp.StatusCode, url),
				}
			}
		}
	}
}
