package task

import (
	"net"
	"time"
)

// the task mode, return from prologue
type TaskMode int

const (
	M_INIT TaskMode = 1 << iota
	// run with no producer
	M_NO_PRODUCER
)

// the interface of the task executed by the knock.
type Task interface {
	// the name of the task, should be unique in the task pools
	Name() string

	// run the necessary prepared actions before executed
	Prologue(ctx *Context) (TaskMode, error)

	// run the necessary clean-up actions after task finished
	Epilogue(ctx *Context)

	// execute the task with passed context, return error when fail
	Execute(ctx *Context) error
}

func CIDRProducer(ctx *Context, cidr *net.IPNet) (producer <-chan string) {
	tmp := make(chan string)

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

		for ip := cidr.IP.Mask(cidr.Mask); cidr.Contains(ip); ip_inc(ip) {
			select {
			case <-ctx.Closed:
				return
			case tmp <- ip.String():
			}

			time.Sleep(ctx.Wait)
		}
	}()

	producer = tmp
	return
}
