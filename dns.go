package knock

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/cmj0121/argparse"
)

// fetch via the DNS record
type DNS struct {
	argparse.Help

	A     bool `default:"true" help:"scan the A record"`
	MX    bool `default:"true" help:"scan the MX record"`
	TXT   bool `default:"true" help:"scan the TXT record"`
	CNAME bool `default:"true" help:"scan the CNAME record"`

	Hostname *string `help:"the target hostname"`

	*os.File `args:"option" short:"F" help:"specified the customized word-list"`

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

			// find TXT
			if dns.TXT {
				if txts, err := net.LookupTXT(hostname); err == nil {
					if txt := strings.Join(txts, " "); txt != "v=spf1 -all" {
						receiver <- Response{
							Type:    RESP_RESULT,
							Message: fmt.Sprintf("[TXT]    %-22s -> %v", hostname, txt),
						}
					}
				}
			}

			// find CNAME
			if dns.CNAME {
				if cname, err := net.LookupCNAME(hostname); err == nil {
					switch {
					case cname == hostname:
					case cname == hostname+".":
					default:
						receiver <- Response{
							Type:    RESP_RESULT,
							Message: fmt.Sprintf("[CNAME]  %-22s -> %s", hostname, cname),
						}

						// CNAME will no need to find the A/AAAA record
						continue
					}
				}
			}

			// find A/AAAA
			if dns.A {
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
			}

			// find MX
			if dns.MX {
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
}

func (dns *DNS) Broker(ctx context.Context) (broker <-chan string) {
	switch dns.File {
	case nil:
		broker = nil
	default:
		tmp := make(chan string, 1)
		go func() {
			defer close(tmp)

			dns.File.Seek(0, os.SEEK_SET)
			scanner := bufio.NewScanner(dns.File)
			for scanner.Scan() {
				select {
				case <-ctx.Done():
					return
				default:
					text := scanner.Text()
					tmp <- text
				}
			}
		}()
		broker = tmp
	}
	return
}
