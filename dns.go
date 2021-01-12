package knock

import (
	"net"

	"github.com/cmj0121/argparse"
	"github.com/cmj0121/logger"
)

type DNSRecord struct {
	A     []string
	AAAA  []string
	TXT   []string
	CNAME string
	MX    []string
	NS    []string
}

type DNS struct {
	argparse.Help

	*logger.Logger `-`

	Hostname *string `help:"target hostname"`
}

func (dns *DNS) Run(log *logger.Logger) (result interface{}) {
	dns.Logger = log

	record := DNSRecord{}
	if ips, err := net.LookupIP(*dns.Hostname); err == nil {
		// get CNAME record
		for _, ip := range ips {
			switch {
			case ip.To4() != nil:
				record.A = append(record.A, ip.To4().String())
			case ip.To16() != nil:
				record.AAAA = append(record.AAAA, ip.To16().String())
			}
		}
	}

	if cname, err := net.LookupCNAME(*dns.Hostname); err == nil && len(cname) > 0 {
		// get CNAME record
		record.CNAME = cname
	}

	if txt, err := net.LookupTXT(*dns.Hostname); err == nil && len(txt) > 0 {
		// get CNAME record
		record.TXT = txt
	}

	if mxs, err := net.LookupMX(*dns.Hostname); err == nil && len(mxs) > 0 {
		// get MX record
		for _, mx := range mxs {
			record.MX = append(record.MX, mx.Host)
		}
	}

	if nss, err := net.LookupNS(*dns.Hostname); err == nil && len(nss) > 0 {
		// get NS record
		for _, ns := range nss {
			record.MX = append(record.NS, ns.Host)
		}
	}

	result = record
	return
}
