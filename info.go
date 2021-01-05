package knock

import (
	"fmt"
	"net"
	"os"

	"github.com/cmj0121/argparse"
	"github.com/cmj0121/logger"
	"gopkg.in/yaml.v3"
)

type Info struct {
	argparse.Help

	*logger.Logger `-`

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

func (info *Info) Run(log *logger.Logger) {
	info.Logger = log

	info.Load()
	if data, err := yaml.Marshal(info); err != nil {
		info.Logger.Warn("cannot marshal info")
		return
	} else {
		// show on the STDOUT
		os.Stdout.Write(data)
	}
}
