package knock

import (
	"fmt"
	"net"
	"os"

	"github.com/cmj0121/argparse"
)

type Info struct {
	argparse.Help

	Hostname string              `args:"-"`
	IFaces   map[string][]string `args:"-"`
}

func (info *Info) Load() {
	var err error
	// ---- get the hostname ----
	if info.Hostname, err = os.Hostname(); err != nil {
		// cannot get hostname, set as error message
		info.Hostname = fmt.Sprintf("[%s]", err)
	}

	// ---- get all interfaces info ----
	info.IFaces = map[string][]string{}
	if ifaces, err := net.Interfaces(); err == nil {
		for _, iface := range ifaces {
			if addrs, err := iface.Addrs(); err == nil {
				for _, addr := range addrs {
					switch addr.(type) {
					case *net.IPNet:
						info.IFaces[iface.Name] = append(info.IFaces[iface.Name], addr.String())
					case *net.IPAddr:
						info.IFaces[iface.Name] = append(info.IFaces[iface.Name], addr.String())
					}
				}
			}
		}
	}

	return
}
