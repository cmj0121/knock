package knock

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"sort"
	"strings"

	"github.com/cmj0121/argparse"
)

// fetch via the DNS record
type DNS struct {
	argparse.Help

	Hostname *string `help:"the target hostname"`

	IPs []string `-`
}

func (dns *DNS) Open() (err error) {
	switch {
	case dns.Hostname == nil || *dns.Hostname == "":
		err = fmt.Errorf("should specified HOSTNAME")
		return
	}

	// get the wild-card DNS addresses
	dns.IPs, _ = net.LookupHost(fmt.Sprintf("ITSHOULDNOTEXIST.%s", *dns.Hostname))
	sort.Strings(dns.IPs)
	return
}

func (dns *DNS) Close() (err error) {
	return
}

func (dns *DNS) Run(receiver chan<- Response, broker <-chan string) {
	for {
		switch name, ok := <-broker; ok {
		case false:
			// no other message
			return
		case true:
			hostname := fmt.Sprintf("%s.%s", name, *dns.Hostname)
			receiver <- Response{
				Type:    RESP_PROGRESS,
				Message: hostname,
			}

			// find A/AAAA
			if addrs, err := net.LookupHost(hostname); err == nil {
				sort.Strings(addrs)

				if reflect.DeepEqual(addrs, dns.IPs) {
					// wildcard subdomain
					continue
				}
				receiver <- Response{
					Type:    RESP_RESULT,
					Message: fmt.Sprintf("[A/AAAA] %-22s -> %v", hostname, addrs),
				}
			}

			// find CNAME
			if cname, err := net.LookupCNAME(hostname); err == nil {
				switch {
				case cname == hostname:
				case cname == hostname+".":
				default:
					receiver <- Response{
						Type:    RESP_RESULT,
						Message: fmt.Sprintf("[CNAME]  %-22s -> %s", hostname, cname),
					}
				}
			}

			// find TXT
			if txts, err := net.LookupTXT(hostname); err == nil {
				if txt := strings.Join(txts, " "); txt != "v=spf1 -all" {
					receiver <- Response{
						Type:    RESP_RESULT,
						Message: fmt.Sprintf("[TXT]    %-22s -> %v", hostname, txt),
					}
				}
			}

			// find MX
			if mxs, err := net.LookupMX(hostname); err == nil {
				for idx := range mxs {
					receiver <- Response{
						Type:    RESP_RESULT,
						Message: fmt.Sprintf("[MX]     %-22s -> %v", hostname, mxs[idx].Host),
					}
				}
			}
		}
	}
}

func (dns *DNS) Broker(ctx context.Context) (broker <-chan string) {
	broker = nil
	return
}
