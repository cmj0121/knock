package knock

import (
	"fmt"
	"net"

	"github.com/google/gopacket"
)

type Target struct {
	gopacket.LayerType

	net.HardwareAddr
	net.IP

	Hostname string
}

func (target *Target) String() (str string) {
	if target.Hostname == "" || target.Hostname == "?" {
		if hostnames, err := net.LookupAddr(target.IP.String()); err == nil && len(hostnames) > 0 {
			// get hostname from IP
			target.Hostname = hostnames[0]
		} else {
			// show the hostname as ?
			target.Hostname = "?"
		}
	}

	str = fmt.Sprintf("%v %v %-16v (%v)", target.LayerType, target.Hostname, target.IP, target.HardwareAddr)
	return
}
