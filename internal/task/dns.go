package task

import (
	"fmt"
	"net"
	"reflect"
	"sort"
	"strings"

	"github.com/cmj0121/stropt"
)

// The knock for DNS records.
type DNS struct {
	stropt.Model

	Hostname *string `default:"example.com" help:"The target hostname"`

	// pre-load the possible wildcard ip
	wildcard_ips []string
}

// show the unique name of the task
func (dns DNS) Name() (name string) {
	name = "dns"
	return
}

// initial the resource and load the possible wildcard IPs
func (dns *DNS) Prologue(ctx *Context) (err error) {
	dns.wildcard_ips, _ = net.LookupHost(fmt.Sprintf("IT_SHOULD_NOT_EXIST.%v", *dns.Hostname))
	sort.Strings(dns.wildcard_ips)

	if len(dns.wildcard_ips) > 0 {
		ctx.Collector <- Message{
			Status: RESULT,
			Msg:    fmt.Sprintf("[A/AAAA] *%-22s  %v", *dns.Hostname, dns.wildcard_ips),
		}
	}

	return
}

// close everything
func (dns *DNS) Epilogue(ctx *Context) {
}

// check the dns recoed exist
func (dns *DNS) Execute(ctx *Context) (err error) {
	for {
		select {
		case token, running := <-ctx.Producer:
			if !running {
				// no-more token, close the task
				return
			}

			hostname := strings.ToLower(fmt.Sprintf("%v.%v", token, *dns.Hostname))
			ctx.Collector <- Message{
				Status: TRACE,
				Msg:    hostname,
			}

			dns.execute(ctx, hostname)
		case <-ctx.Closed:
			// closed by the main thread
			return
		}
	}
}

func (dns *DNS) execute(ctx *Context, hostname string) {
	if txts, err := net.LookupTXT(hostname); err == nil {
		if txt := strings.Join(txts, " "); txt != "v=spf1 -all" {
			ctx.Collector <- Message{
				Status: RESULT,
				Msg:    fmt.Sprintf("[TXT]    %-22s -> %v", hostname, txt),
			}
		}
	}

	if mxs, err := net.LookupMX(hostname); err == nil {
		for idx := range mxs {
			ctx.Collector <- Message{
				Status: RESULT,
				Msg:    fmt.Sprintf("[MX]     %-22s -> %v", hostname, mxs[idx].Host),
			}
		}
	}

	if cname, err := net.LookupCNAME(hostname); err == nil {
		switch {
		case cname == hostname:
		case cname == hostname+".":
		default:
			ctx.Collector <- Message{
				Status: RESULT,
				Msg:    fmt.Sprintf("[CNAME]  %-22s -> %s", hostname, cname),
			}

			// CNAME will no need to find the A/AAAA record
		}
	}

	if addrs, err := net.LookupHost(hostname); err == nil {
		sort.Strings(addrs)

		if !reflect.DeepEqual(addrs, dns.wildcard_ips) {
			ctx.Collector <- Message{
				Status: RESULT,
				Msg:    fmt.Sprintf("[A/AAAA] %-22s -> %v", hostname, addrs),
			}
		}
	}
}
