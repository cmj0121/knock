package worker

import (
	"fmt"
	"net"

	"github.com/cmj0121/knock/progress"
	"github.com/rs/zerolog/log"
)

func init() {
	worker := &SubDomain{
		wildcard_ips: map[string]struct{}{},
	}
	Register(worker)
}

// the debugger worker and just show the word in STDOUT
type SubDomain struct {
	// the target hostname/IP
	hostname string

	// the wildcard IP lists
	wildcard_ips map[string]struct{}
}

// the unique name of worker
func (s SubDomain) Name() string {
	return "subd"
}

// show the help message
func (s SubDomain) Help() string {
	return "list possible sub-domain"
}

// the dummy open method
func (s *SubDomain) Open(args ...string) (err error) {
	// check the wildcard IP address
	switch len(args) {
	case 0:
		err = fmt.Errorf("should pass hostname to the command %#v", s.Name())
	case 1:
		s.hostname = args[0]
		ips := s.check(fmt.Sprintf("IT_SHOULD_NOT_EXIST.%v", s.hostname))

		for _, ip := range ips {
			s.wildcard_ips[ip] = struct{}{}
		}
	default:
		err = fmt.Errorf("should pass one and only one hostname to the command %#v", s.Name())
	}
	return
}

// the dummy close method
func (s SubDomain) Close() (err error) {
	log.Debug().Msg("dummy close")
	return
}

// execute the worker
func (s *SubDomain) Run(producer <-chan string) (err error) {
	for word := range producer {
		log.Debug().Str("word", word).Msg("handle producer")
		progress.AddProgress(word)

		hostname := fmt.Sprintf("%v.%v", word, s.hostname)
		ips := s.check(hostname)
		for _, ip := range ips {
			if _, ok := s.wildcard_ips[ip]; !ok {
				log.Debug().Str("ip", ip).Str("hostname", hostname).Msg("find sub-domain")
				progress.AddText("[A/AAAA] %-22v %v", hostname, ip)
			}
		}
	}
	return
}

// copy the current worker settings and generate a new instance
func (s *SubDomain) Dup() (worker Worker) {
	worker = &SubDomain{
		hostname:     s.hostname,
		wildcard_ips: s.wildcard_ips,
	}
	return
}

// get the possible ip address, by local resolver
func (s *SubDomain) check(domain string) (ips []string) {
	ips, _ = net.LookupHost(domain)
	return
}
