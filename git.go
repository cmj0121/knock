package knock

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

var (
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
type Git struct {
	Scheme string  `default:"http" choices:"http https" help:"the URI scheme"`
	URI    *string `help:"the target URI"`
	Repo   string  `default:"git" help:"the local repository location"`

	Timeout int `default:"4" help:"the HTTP request timeout in seconds"`

	*http.Client `-`
	broker       chan string
	commits      sync.Map
}

func (git *Git) Open() (err error) {
	if git.URI == nil || *git.URI == "" {
		err = fmt.Errorf("should pass URI")
		return
	}

	err = nil

	git.Client = &http.Client{
		Timeout: time.Duration(git.Timeout) * time.Second,
	}

	return
}

func (git *Git) Close() (err error) {
	err = nil
	return
}

func (git *Git) Run(receiver chan<- Response, broker <-chan string) {
	for {
		var new_hash []string
		hash, ok := <-broker

		if !ok {
			// no other hash need process
			break
		} else if _, ok := git.commits.Load(hash); ok {
			// already processed
			continue
		} else if hash == "objects/00/00000000000000000000000000000000000000" {
			// initial commit, skip
			continue
		}

		git.commits.Store(hash, true)
		receiver <- Response{
			Type: RESP_DEBUG,
			//Type:    RESP_PROGRESS,
			Message: hash,
		}

		// fetch the git file
		resp, err := git.Client.Get(fmt.Sprintf("%s://%s/%s", git.Scheme, *git.URI, hash))
		if err != nil {
			receiver <- Response{
				Type:    RESP_ERR,
				Message: fmt.Sprintf("fetch %v: %v", hash, err),
			}
			continue
		} else if resp.StatusCode != 200 {
			receiver <- Response{
				Type:    RESP_ERR,
				Message: fmt.Sprintf("fetch %v: %v", hash, resp.Status),
			}
			continue
		}

		data, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err == nil {
			// save to the local repo
			name := fmt.Sprintf("%s/%s", git.Repo, hash)
			dirname := filepath.Dir(name)
			if err := os.MkdirAll(dirname, 0755); err != nil {
				receiver <- Response{
					Type:    RESP_ERR,
					Message: fmt.Sprintf("cannot mkdir -p %v: %v", dirname, err),
				}

				continue
			} else if ioutil.WriteFile(name, data, 0644); err != nil {
				receiver <- Response{
					Type:    RESP_ERR,
					Message: fmt.Sprintf("cannot write to %v: %v", name, err),
				}
				continue
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
				new_hash = append(new_hash, git.processGitObject(receiver, data)...)
			default:
				receiver <- Response{
					Type:    RESP_ERR,
					Message: fmt.Sprintf("not implement process %#v", hash),
				}
			}
		}

		go func() {
			for idx := range new_hash {
				git.broker <- new_hash[idx]
			}
		}()
	}
}

func (git *Git) processGitObject(receiver chan<- Response, blob []byte) (new_hash []string) {
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

func (git *Git) Broker(ctx context.Context) (broker <-chan string) {
	git.broker = make(chan string, 64)
	go func() {
		defer close(git.broker)

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
			git.broker <- commit
		}
		// wait few
		time.Sleep(time.Second)
		for len(git.broker) > 0 {
			// buffer not empty
			time.Sleep(time.Second)
		}
	}()

	broker = git.broker
	return
}
