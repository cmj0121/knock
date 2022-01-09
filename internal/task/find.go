package task

import (
	"fmt"
	"net"
	"time"

	"github.com/cmj0121/stropt"
	"github.com/go-ping/ping"
)

type Find struct {
	stropt.Model

	Timeout time.Duration `default:"100ms" desc:"timeout when try to connect to host"`

	// the target CIDR want to search
	CIDR *net.IPNet `default:"127.0.0.1/24" desc:"the target CIDR"`
}

// show the unique name of the task
func (find Find) Name() (name string) {
	name = "find"
	return
}

// run the necessary prepared actions before executed
func (find *Find) Prologue(ctx *Context) (mode TaskMode, err error) {
	switch find.CIDR {
	case nil:
	default:
		// set the customized producer
		ctx.Producer = CIDRProducer(ctx, find.CIDR)
	}
	return
}

// run the necessary clean-up actions after task finished
func (find Find) Epilogue(ctx *Context) {
}

// find the host by the given token
func (find *Find) Execute(ctx *Context) (err error) {
	for {
		select {
		case token, running := <-ctx.Producer:
			if !running {
				// no-more token, close the task
				return
			}
			// print the token
			ctx.Collector <- Message{
				Status: TRACE,
				Msg:    token,
			}

			switch pinger, err := ping.NewPinger(token); err {
			case nil:
				pinger.Count = 1
				pinger.Timeout = find.Timeout
				pinger.OnRecv = func(pkt *ping.Packet) {
					ctx.Collector <- Message{
						Status: RESULT,
						Msg:    token,
					}
				}
				pinger.Run() //nolint
			default:
				ctx.Collector <- Message{
					Status: ERROR,
					Msg:    fmt.Sprintf("cannot build ping %v: %v", token, err),
				}
			}
		case <-ctx.Closed:
			// closed by the main thread
			return
		}
	}
}
