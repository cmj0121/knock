package knock

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/cmj0121/argparse"
)

var (
	WEB_GROKER_CHANNEL_LEN = 64
	WEB_NOT_EXIST_PAGE     = "KNOCK_NOT_EXIST_PAGE"
	// git reference
	RE_GIT_REF = regexp.MustCompile(`ref: ([a-zA-Z0-9/]+)`)
	// git config
	RE_GIT_CONFIG_REF = regexp.MustCompile(`merge = (refs/heads/\w+)`)
	// git object
	RE_GIT_OBJECT_HASH = regexp.MustCompile(`([a-fA-F0-9]{2})([a-fA-F0-9]{38})`)
	RE_GIT_OBJECT_TREE = regexp.MustCompile(`(?:parent|tree) ([a-fA-F0-9]{2})([a-fA-F0-9]{38})`)
	RE_GIT_OBJECT_BLOB = regexp.MustCompile(`\d{5,6} .*?` + "\x00")
)

// fetch the git to local repo
type Web struct {
	argparse.Help

	Timeout int `default:"4" help:"the HTTP request timeout in seconds"`

	// git-related scanner
	Git          bool   `help:"fetch the git repo"`
	GitLocalRepo string `name:"local" default:"git" help:"the local repository location"`

	Scheme string  `short:"s" default:"http" choices:"http https" help:"the URI scheme"`
	URI    *string `help:"the target URI"`

	*os.File `args:"option" short:"F" help:"specified the customized word-list"`

	Skip404 bool `name:"404" default:"true" help:"skip 404 page"`

	*http.Client `-`
	broker       chan string
	commits      sync.Map

	// default success / failure html page
	html_success string
	html_failure string
}

func (web *Web) Open() (err error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	web.Client = &http.Client{
		Transport: tr,
		Timeout:   time.Duration(web.Timeout) * time.Second,
	}

	switch {
	case web.URI == nil || *web.URI == "":
		err = fmt.Errorf("should pass URI")
		return
	case web.Git:
	default:
		// html success
		url := fmt.Sprintf("%s://%s", web.Scheme, *web.URI)
		if resp, err := web.Get(url); err == nil {
			defer resp.Body.Close()

			if data, err := ioutil.ReadAll(resp.Body); err == nil {
				// save the success page
				web.html_success = string(data)
			}
		}

		// html failure
		url = fmt.Sprintf("%s://%s/%s", web.Scheme, *web.URI, WEB_NOT_EXIST_PAGE)
		if resp, err := web.Get(url); err == nil {
			defer resp.Body.Close()

			if data, err := ioutil.ReadAll(resp.Body); err == nil {
				// save the success page
				web.html_success = string(data)
			}
		}
	}

	return
}

func (web *Web) Close() (err error) {
	err = nil
	return
}

func (web *Web) Run(receiver chan<- Response, broker <-chan string) {
	for {
		hash, ok := <-broker

		if !ok {
			// no other hash need process
			break
		}

		switch {
		case web.Git:
			// run the Git repo
			web.runGit(hash, receiver, broker)
		default:
			url := fmt.Sprintf("%s://%s/%s", web.Scheme, *web.URI, hash)
			receiver <- Response{
				Type:    RESP_PROGRESS,
				Message: url,
			}
			resp, err := web.Get(url)
			if err != nil {
				receiver <- Response{
					Type:    RESP_ERR,
					Message: fmt.Sprintf("fetch %v: %v", hash, err),
				}
				return
			}

			data, err := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()

			switch {
			case resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices:
				if string(data) != web.html_success {
					receiver <- Response{
						Type:    RESP_RESULT,
						Message: fmt.Sprintf("[%d] %s", resp.StatusCode, url),
					}
				}
			case resp.StatusCode >= http.StatusMultipleChoices && resp.StatusCode < http.StatusBadRequest:
				receiver <- Response{
					Type:    RESP_RESULT,
					Message: fmt.Sprintf("[%d] %s", resp.StatusCode, url),
				}
			case resp.StatusCode >= http.StatusBadRequest && resp.StatusCode < http.StatusInternalServerError:
				if resp.StatusCode == http.StatusNotFound {
					if string(data) != web.html_failure && web.Skip404 {
						continue
					}
				}
				receiver <- Response{
					Type:    RESP_RESULT,
					Message: fmt.Sprintf("[%d] %s", resp.StatusCode, url),
				}
			default:
				receiver <- Response{
					Type:    RESP_RESULT,
					Message: fmt.Sprintf("[%d] %s", resp.StatusCode, url),
				}
			}
		}
	}
}

func (web *Web) runGit(hash string, receiver chan<- Response, broker <-chan string) {
	var new_hash []string

	if _, ok := web.commits.Load(hash); ok {
		// already processed
		return
	} else if hash == "objects/00/00000000000000000000000000000000000000" {
		// initial commit, skip
		return
	}

	web.commits.Store(hash, true)
	receiver <- Response{
		Type:    RESP_PROGRESS,
		Message: hash,
	}

	// fetch the git file
	resp, err := web.Get(fmt.Sprintf("%s://%s/%s", web.Scheme, *web.URI, hash))
	if err != nil {
		receiver <- Response{
			Type:    RESP_ERR,
			Message: fmt.Sprintf("fetch %v: %v", hash, err),
		}
		return
	}

	switch resp.StatusCode {
	case http.StatusOK:
	default:
		receiver <- Response{
			Type:    RESP_ERR,
			Message: fmt.Sprintf("fetch %v: %v", hash, resp.Status),
		}
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err == nil {
		// save to the local repo
		name := fmt.Sprintf("%s/%s", web.GitLocalRepo, hash)
		dirname := filepath.Dir(name)
		if err := os.MkdirAll(dirname, 0755); err != nil {
			receiver <- Response{
				Type:    RESP_ERR,
				Message: fmt.Sprintf("cannot mkdir -p %v: %v", dirname, err),
			}

			return
		} else if ioutil.WriteFile(name, data, 0644); err != nil {
			receiver <- Response{
				Type:    RESP_ERR,
				Message: fmt.Sprintf("cannot write to %v: %v", name, err),
			}
			return
		}

		switch {
		case hash == "HEAD":
			matched := RE_GIT_REF.FindAllSubmatch(data, -1)
			for idx := range matched {
				// save the ref
				ref := string(matched[idx][1])
				new_hash = append(new_hash, ref)
			}
		case hash == "config":
			matched := RE_GIT_CONFIG_REF.FindAllSubmatch(data, -1)
			for idx := range matched {
				ref := string(matched[idx][1])
				new_hash = append(new_hash, ref)
			}
		case hash == "ORIG_HEAD" || hash == "FETCH_HEAD" || hash[:5] == "logs/" || hash[:5] == "refs/":
			matched := RE_GIT_OBJECT_HASH.FindAllSubmatch(data, -1)
			for idx := range matched {
				// save the object hash
				object_hash := fmt.Sprintf("objects/%s/%s", string(matched[idx][1]), string(matched[idx][2]))
				new_hash = append(new_hash, object_hash)
			}
		case hash[:8] == "objects/":
			new_hash = append(new_hash, web.processGitObject(receiver, data)...)
		default:
			receiver <- Response{
				Type:    RESP_ERR,
				Message: fmt.Sprintf("not implement process %#v", hash),
			}
		}
	}

	go func() {
		for idx := range new_hash {
			web.broker <- new_hash[idx]
		}
	}()
}

func (web *Web) processGitObject(receiver chan<- Response, blob []byte) (new_hash []string) {
	r := bytes.NewReader(blob)
	reader, err := zlib.NewReader(r)
	if err != nil {
		receiver <- Response{
			Type:    RESP_ERR,
			Message: fmt.Sprintf("cannot open zlib reader: %v", err),
		}
		return
	}
	defer reader.Close()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		receiver <- Response{
			Type:    RESP_ERR,
			Message: fmt.Sprintf("cannot read git object file: %v", err),
		}
		return
	}

	index := bytes.Index(data, []byte{0x00})
	heads := string(data[:index])
	switch {
	case len(heads) > 5 && heads[:5] == "blob ":
		// nop
	case len(heads) > 5 && heads[:5] == "tree ":
		matched_index := RE_GIT_OBJECT_BLOB.FindAllIndex(data[index:], -1)
		if len(matched_index) == 0 {
			receiver <- Response{
				Type:    RESP_ERR,
				Message: fmt.Sprintf("invalid tree: %#v", data),
			}
		}

		for _, loc := range matched_index {
			blob := data[index+loc[1] : index+loc[1]+20]
			hash := fmt.Sprintf(
				"%08x%08x%08x%08x%08x",
				binary.BigEndian.Uint32(blob[0x00:0x04]),
				binary.BigEndian.Uint32(blob[0x04:0x08]),
				binary.BigEndian.Uint32(blob[0x08:0x0C]),
				binary.BigEndian.Uint32(blob[0x0C:0x10]),
				binary.BigEndian.Uint32(blob[0x10:0x14]),
			)
			hash = fmt.Sprintf("objects/%v/%v", hash[:2], hash[2:])
			new_hash = append(new_hash, hash)
		}
	case len(heads) > 7 && heads[:7] == "commit ":
		matched := RE_GIT_OBJECT_TREE.FindAllSubmatch(data[index:], -1)
		if len(matched) == 0 {
			receiver <- Response{
				Type:    RESP_ERR,
				Message: fmt.Sprintf("invalid commit: %#v", data),
			}
		}

		for idx := range matched {
			hash := fmt.Sprintf("objects/%s/%s", string(matched[idx][1]), string(matched[idx][2]))
			new_hash = append(new_hash, hash)
		}
	default:
		receiver <- Response{
			Type:    RESP_ERR,
			Message: fmt.Sprintf("not implement blob: %#v", heads),
		}
	}

	return
}

func (web *Web) Broker(ctx context.Context) (broker <-chan string) {
	switch {
	case web.Git:
		web.broker = make(chan string, WEB_GROKER_CHANNEL_LEN)

		go func() {
			defer close(web.broker)

			// default GIT files
			default_commit_hashs := []string{
				"HEAD",
				"config",
				"ORIG_HEAD",
				"FETCH_HEAD",
				"logs/HEAD",
				"refs/remotes/origin/master",
			}

			for _, commit := range default_commit_hashs {
				web.broker <- commit
			}
			// wait few
			time.Sleep(time.Second)
			for len(web.broker) > 0 {
				// buffer not empty
				time.Sleep(time.Second)
			}
		}()

		broker = web.broker
	case web.File != nil:
		tmp := make(chan string, 1)
		go func() {
			defer close(tmp)

			web.File.Seek(0, os.SEEK_SET)
			scanner := bufio.NewScanner(web.File)
			for scanner.Scan() {
				select {
				case <-ctx.Done():
					return
				default:
					text := scanner.Text()
					switch {
					case len(text) > 0 && text[0] == '#':
					default:
						tmp <- text
					}
				}
			}
		}()
		broker = tmp
	}
	return
}

func (web *Web) Get(url string) (resp *http.Response, err error) {
	var req *http.Request

	if req, err = http.NewRequest("GET", url, nil); err == nil {
		// customized header
		req.Header.Set("User-Agent", Version())
		resp, err = web.Client.Do(req)
	}

	return
}
