package producer

import (
	"net"

	"github.com/rs/zerolog/log"
)

func NewCIDRProducer(s string) (producer *CIDRProducer, err error) {
	var cidr *net.IPNet

	if _, cidr, err = net.ParseCIDR(s); err != nil {
		// invalid CIDR format
		return
	}

	producer = &CIDRProducer{
		IPNet:  cidr,
		Closed: make(chan struct{}, 1),
	}

	return
}

// the IP address producer by the IP/mask
type CIDRProducer struct {
	// the target IP/mask
	*net.IPNet

	// the signle for close the current connection and the subscriber
	// should close all allocated resources.
	Closed chan struct{}
}

// produce the IP address
func (ctx *CIDRProducer) Produce() (ch <-chan string) {
	tmp := make(chan string, 1)

	ip_inc := func(ip net.IP) {
		for i := len(ip) - 1; i >= 0; i-- {
			ip[i]++
			if ip[i] > 0 {
				break
			}
		}
	}

	go func() {
		defer close(tmp)

		for ip := ctx.IP.Mask(ctx.Mask); ctx.Contains(ip); ip_inc(ip) {
			select {
			case tmp <- ip.String():
			case <-ctx.Closed:
				log.Debug().Msg("explicitly stop the word producer")
				return
			}
		}
	}()

	ch = tmp
	return
}

// explicitly close the current producer
func (ctx *CIDRProducer) Close() {
	close(ctx.Closed)
}
