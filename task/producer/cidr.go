package producer

import (
	"fmt"
	"net"
	"time"

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
	ProducerBase

	// the target IP/mask
	*net.IPNet

	// the prefix
	prefix string

	// the signle for close the current connection and the subscriber
	// should close all allocated resources.
	Closed chan struct{}
}

// produce the IP address
func (ctx *CIDRProducer) Produce(wait time.Duration) (ch <-chan string) {
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
			ip := fmt.Sprintf("%v%v", ctx.prefix, ip)
			select {
			case tmp <- ip:
			case <-ctx.Closed:
				log.Debug().Msg("explicitly stop the word producer")
				return
			}

			time.Sleep(wait)
		}
	}()

	ch = tmp
	return
}

// explicitly close the current producer
func (ctx *CIDRProducer) Close() {
	close(ctx.Closed)
}
