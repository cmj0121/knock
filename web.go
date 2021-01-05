package knock

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/cmj0121/logger"
)

type Web struct {
	*logger.Logger `-`

	Scheme string `short:"S" default:"https" help:"the scheme identifies the protocol"`
	Host   string `short:"H" help:"the target hostname"`
	Path   string `short:"P" default:"/" help:"the base filepath"`

	Config  string `short:"c" help:"the path-list file"`
	Timeout int    `default:"4" help:"timeout"`
}

func (web *Web) Run(log *logger.Logger) {
	web.Logger = log

	client := &http.Client{
		Timeout: time.Duration(web.Timeout) * time.Second,
	}

	if web.Config != "" {
		if file, err := os.Open(web.Config); err != nil {
			web.Crit("cannot open %#v: %v", web.Config, err)
			return
		} else {
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				URL := fmt.Sprintf("%s://%s%s%s", web.Scheme, web.Host, web.Path, scanner.Text())
				web.Scan(client, URL)
			}
		}
	}

}

func (web *Web) Scan(client *http.Client, URL string) {
	web.Verbose("try %#v", URL)
	resp, err := client.Get(URL)
	if err != nil {
		web.Crit("[GET] %#v: %v", URL, err)
		return
	}

	switch {
	case resp.StatusCode >= 100 && resp.StatusCode < 200:
		// HTTP status 1XX
		web.Info("[GET] %#-44v: %s", URL, resp.Status)
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		// HTTP status 2XX
		web.Info("[GET] %#-44v: %s", URL, resp.Status)
	case resp.StatusCode >= 300 && resp.StatusCode < 400:
		// HTTP status 3XX
		web.Info("[GET] %#-44v: %s", URL, resp.Status)
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		// HTTP status 4XX
		web.Debug("[GET] %#-44v: %s", URL, resp.Status)
	case resp.StatusCode >= 500 && resp.StatusCode < 600:
		// HTTP status 4XX
		web.Info("[GET] %#-44v: %s", URL, resp.Status)
	default:
		// HTTP status which less than 100
		web.Info("[GET] %#-44v: %s", URL, resp.Status)
	}
	return
}
