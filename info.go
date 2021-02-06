package knock

import (
	"fmt"
	"net"
	"sync"
)

type Info struct {
	sync.Once `-`
}

func (info *Info) Run(broker <-chan string, receiver chan<- Response) {
	// just do exactly once, no matter how many worker
	info.Once.Do(func() {
		if ifaces, err := net.Interfaces(); err == nil {
			for _, iface := range ifaces {
				receiver <- Response{
					Type:    RESP_PROGRESS,
					Message: iface.Name,
				}

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
		}
	})
}
