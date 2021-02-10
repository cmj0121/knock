package knock

import (
	"context"
	"fmt"
	"net"
	"sync"
)

type Info struct {
	sync.Once `-`
}

func (info *Info) Open() (err error) {
	err = nil
	return
}

func (info *Info) Close() (err error) {
	err = nil
	return
}

func (info *Info) Run(receiver chan<- Response, broker <-chan string) {
	// just do exactly once, no matter how many worker
	info.Once.Do(func() {
		if ifaces, err := net.Interfaces(); err == nil {
			for _, iface := range ifaces {
				receiver <- Response{
					Type:    RESP_PROGRESS,
					Message: iface.Name,
				}
				info.showIface(receiver, iface)
			}
		}
	})
}

func (info *Info) showIface(receiver chan<- Response, iface net.Interface) {
	switch {
	case iface.Flags&net.FlagUp == 0:
	case iface.Flags&net.FlagLoopback == net.FlagLoopback:
	default:
		if addrs, err := iface.Addrs(); err == nil {
			for _, addr := range addrs {
				switch addr.(type) {
				case *net.IPNet, *net.IPAddr:
					receiver <- Response{
						Type:    RESP_RESULT,
						Message: fmt.Sprintf("%-16v %-42v (%v)", iface.Name, addr.String(), iface.Flags),
					}
				}
			}
		}
	}
}

// does not provide the customized word-list
func (info *Info) Broker(ctx context.Context) (ch <-chan string) {
	ch = nil
	return
}
