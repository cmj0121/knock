package worker

import (
	"context"
	"fmt"
	"net"

	"github.com/cmj0121/knock/progress"
	"github.com/rs/zerolog/log"
)

func init() {
	worker := &DNS{
		Resolver: &net.Resolver{
			// preferred Go's built-in DNS resolver
			PreferGo: true,
		},
		wildcard_resp: map[string]string{},
	}
	Register(worker)
}

// the debugger worker and just show the word in STDOUT
type DNS struct {
	*net.Resolver

	// the target hostname/IP
	hostname string

	// the wildcard IP lists
	wildcard_resp map[string]string
}

// the unique name of worker
func (d DNS) Name() string {
	return "dns"
}

// show the help message
func (d DNS) Help() string {
	return "list possible DNS records"
}

// the dummy open method
func (d *DNS) Open(args ...string) (err error) {
	// check the wildcard IP address
	switch len(args) {
	case 0:
		err = fmt.Errorf("should pass hostname to the command %#v", d.Name())
	case 1:
		d.hostname = args[0]
		d.run(d.hostname)

		d.prologue()
	default:
		err = fmt.Errorf("should pass one and only one hostname to the command %#v", d.Name())
	}
	return
}

func (d *DNS) prologue() {
	hostname := fmt.Sprintf("IT_SHOULD_NOT_EXIST.%v", d.hostname)

	if resp, err := d.LookupIP(context.Background(), "ip4", hostname); err == nil {
		for _, ip := range resp {
			d.wildcard_resp["A"] = fmt.Sprintf("%v", ip)
		}
	}

	if resp, err := d.LookupIP(context.Background(), "ip6", hostname); err == nil {
		for _, ip := range resp {
			d.wildcard_resp["AAAA"] = fmt.Sprintf("%v", ip)
		}
	}

	if resp, err := d.LookupCNAME(context.Background(), hostname); err == nil {
		d.wildcard_resp["CNAME"] = fmt.Sprintf("%v", resp)
	}
}

// the dummy close method
func (d DNS) Close() (err error) {
	log.Debug().Msg("dummy close")
	return
}

// execute the worker
func (d *DNS) Run(producer <-chan string) (err error) {
	for word := range producer {
		log.Debug().Str("word", word).Msg("handle producer")
		progress.AddProgress(word)

		hostname := fmt.Sprintf("%v.%v", word, d.hostname)
		d.run(hostname)
	}
	return
}

func (d *DNS) run(hostname string) {
	d.lookupIPv4(hostname)
	d.lookupIPv6(hostname)
	d.lookupCNAME(hostname)
	d.lookupMX(hostname)
	d.lookupNS(hostname)
	d.lookupTXT(hostname)
}

// copy the current worker settings and generate a new instance
func (d *DNS) Dup() (worker Worker) {
	worker = &DNS{
		hostname:      d.hostname,
		wildcard_resp: d.wildcard_resp,
	}
	return
}

func (d *DNS) lookupIPv4(hostname string) {
	resp, err := d.LookupIP(context.Background(), "ip4", hostname)
	if err == nil && len(resp) > 0 {
		for _, ip := range resp {
			d.addProgress("A", hostname, ip)
		}
	}
}

func (d *DNS) lookupIPv6(hostname string) {
	resp, err := d.LookupIP(context.Background(), "ip6", hostname)
	if err == nil && len(resp) > 0 {
		for _, ip := range resp {
			d.addProgress("AAAA", hostname, ip)
		}
	}
}

func (d *DNS) lookupCNAME(hostname string) {
	resp, err := d.LookupCNAME(context.Background(), hostname)
	if err == nil && len(resp) > 0 {
		d.addProgress("CNAME", hostname, resp)
	}
}

func (d *DNS) lookupMX(hostname string) {
	resp, err := d.LookupMX(context.Background(), hostname)
	if err == nil && len(resp) > 0 {
		for _, mx := range resp {
			d.addProgress("MX", hostname, mx.Host)
		}
	}
}

func (d *DNS) lookupNS(hostname string) {
	resp, err := d.LookupNS(context.Background(), hostname)
	if err == nil && len(resp) > 0 {
		for _, ns := range resp {
			d.addProgress("NS", hostname, ns.Host)
		}
	}
}

func (d *DNS) lookupTXT(hostname string) {
	resp, err := d.LookupTXT(context.Background(), hostname)
	if err == nil && len(resp) > 0 {
		for _, txt := range resp {
			d.addProgress("TXT", hostname, txt)
		}
	}
}

func (d *DNS) addProgress(qtype, hostname string, resp interface{}) {
	if ans, ok := d.wildcard_resp[qtype]; ok {
		if text := fmt.Sprintf("%v", resp); text == ans {
			log.Debug().Str("query", qtype).Str("hostname", hostname).Msg("wildcard DNS record")
			return
		}
	}

	progress.AddText("%-6v %-18v %v", qtype, hostname, resp)
}
