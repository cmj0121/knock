package worker

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/cmj0121/knock/progress"
	"github.com/rs/zerolog/log"
)

func init() {
	worker := &S3{
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
type S3 struct {
	*http.Client

	// the prefix of the bucket name
	prefix string
}

// the unique name of worker
func (S3) Name() string {
	return "s3"
}

// show the help message
func (S3) Help() string {
	return "the s3 bucket emuerator"
}

// the dummy open method
func (s *S3) Open(args ...string) (err error) {
	// check the wildcard IP address
	switch len(args) {
	case 0:
	case 1:
		s.prefix = args[0]
	default:
		err = fmt.Errorf("should pass only one prefix to the command %#v", s.Name())
	}
	return
}

// the dummy close method
func (S3) Close() (err error) {
	log.Debug().Msg("dummy close")
	return
}

// execute the worker
func (s *S3) Run(producer <-chan string) (err error) {
	for word := range producer {
		word = fmt.Sprintf("%v%v", s.prefix, word)

		log.Debug().Str("word", word).Msg("handle producer")
		progress.AddProgress(word)

		word = strings.ToLower(word)
		if !s.validator(word) {
			continue
		}

		url := fmt.Sprintf("https://%v.s3.amazonaws.com", word)
		switch resp, err := s.Client.Get(url); err {
		case nil:
			switch resp.StatusCode {
			case 400:
			case 403:
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
func (s *S3) Dup() (worker Worker) {
	worker = &S3{
		Client: s.Client,
		prefix: s.prefix,
	}
	return
}

// validate the S3 bucket name
// ref: https://docs.aws.amazon.com/AmazonS3/latest/userguide/bucketnamingrules.html
func (S3) validator(bucket string) (valid bool) {
	if matched, _ := regexp.MatchString(`[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]`, bucket); matched {
		log.Info().Str("bucket", bucket).Msg("not valid bucket bucket")
		return
	}

	switch {
	case strings.HasPrefix(bucket, `xn--`):
		log.Info().Str("bucket", bucket).Msg("not valid bucket bucket")
		return
	case strings.HasSuffix(bucket, `-s3alias`):
		log.Info().Str("bucket", bucket).Msg("not valid bucket bucket")
		return
	case strings.HasSuffix(bucket, `--ol-s3`):
		log.Info().Str("bucket", bucket).Msg("not valid bucket bucket")
		return
	}

	valid = true
	return
}
