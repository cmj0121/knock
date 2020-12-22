package knock

import (
	"fmt"
	"net"
	"encoding/json"

	"github.com/google/gopacket"
)

type Target struct {
	gopacket.LayerType

	net.HardwareAddr
	net.IP

	Hostname string
}

func (target *Target) String() (str string) {
	target.loadHostname()
	str = fmt.Sprintf("%v %v %-16v (%v)", target.LayerType, target.Hostname, target.IP, target.HardwareAddr)
	return
}

func (target *Target) MarshalYAML() (out interface{}, err error) {
	target.loadHostname()
	out = map[string]string{
		"layer":    target.LayerType.String(),
		"MAC":      target.HardwareAddr.String(),
		"ip":       target.IP.String(),
		"hostname": target.Hostname,
	}
	return
}

func (target *Target) MarshalJSON() (out []byte, err error) {
	target.loadHostname()

	out, err = json.Marshal(map[string]string{
		"layer":    target.LayerType.String(),
		"MAC":      target.HardwareAddr.String(),
		"ip":       target.IP.String(),
		"hostname": target.Hostname,
	})
	return
}

func (target *Target) loadHostname() {
	if target.Hostname == "" || target.Hostname == "?" {
		if hostnames, err := net.LookupAddr(target.IP.String()); err == nil && len(hostnames) > 0 {
			// get hostname from IP
			target.Hostname = hostnames[0]
		} else {
			// show the hostname as ?
			target.Hostname = "?"
		}
	}
}
