package knock

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"net"
	"os"
	"time"

	"github.com/cmj0121/argparse"
	"github.com/cmj0121/logger"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"gopkg.in/yaml.v3"
)

const (
	// maximal packet size in the PCAP handler
	MAX_PACKAGE_SIZE = 65536
)

type Scan struct {
	argparse.Help

	*logger.Logger `-`

	Format     string `short:"f" default:"yaml" choices:"yaml json" help:"output format"`
	Timeout    int    `default:"60" help:"set timeout on all run"`
	IPv6       bool   `help:"scan IPv6 only"`
	MaxPkgSize int32  `default:"65536" help:"maximal packet size"`

	IFace *net.Interface `args:"option"`

	IP *[]string `help:"scan IP list"`

	Targets []*Target `-`
}

func (scan *Scan) Run(log *logger.Logger) {
	// set the logger
	scan.Logger = log

	// load the iface by default
	if scan.IFace == nil {
		// load the first available interface
		ifaces, err := net.Interfaces()
		if err != nil {
			scan.Crit("cannot load iface")
			return
		}

		for _, iface := range ifaces {
			if iface.Flags&net.FlagUp == net.FlagUp && iface.Flags&net.FlagLoopback == 0 {
				scan.IFace = &iface
				break
			}
		}

		if scan.IFace == nil {
			scan.Crit("cannot found iface")
			return
		}
	}

	if scan.MaxPkgSize <= 0 {
		// override maximal packet size
		scan.MaxPkgSize = MAX_PACKAGE_SIZE
	}

	// open PCAP handle for packet R/W (blocking)
	scan.Info("scan the iface: %#v", scan.IFace.Name)
	handler, err := pcap.OpenLive(scan.IFace.Name, scan.MaxPkgSize, true, pcap.BlockForever)
	if err != nil {
		scan.Crit("open PCAP handler: %v", err)
		return
	}
	defer handler.Close()

	// open the context
	timeout := time.Second * time.Duration(scan.Timeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// receive the packet on the goroutine
	go scan.Recv(ctx, handler)
	for ip := range scan.ListIP(ctx) {
		// send the ARP request
		scan.sendARP(handler, scan.IFace, ip)
	}

	// always wait few seconds
	time.Sleep(TASK_WAIT_SECONDS)

	// show the output and show in the STDOUT
	switch scan.Format {
	case "yaml":
		data, err := yaml.Marshal(scan.Targets)
		if err != nil {
			scan.Warn("cannot marshal as %#v: %v", scan.Format, err)
		}
		os.Stdout.Write(data)
	case "json":
		data, err := json.Marshal(scan.Targets)
		if err != nil {
			scan.Warn("cannot marshal as %#v: %v", scan.Format, err)
		}
		os.Stdout.Write(data)
	default:
		scan.Crit("not implement format: %#v", scan.Format)
		return
	}
}

// receive the packet and save the result
func (scan *Scan) Recv(ctx context.Context, handler *pcap.Handle) {
	src := gopacket.NewPacketSource(handler, layers.LayerTypeEthernet)
	in := src.Packets()

	scan.Info("start receive on %#v", scan.IFace.Name)
	for {
		// blocking wait until 1) context Done or 2) receive packet
		select {
		case <-ctx.Done():
			// stop read ARP packets
			scan.Info("force stop receive packet")
			return
		case pkg := <-in:
			// receive packet
			scan.Debug("receive packet: %v", pkg)

			if target := scan.recvARP(pkg); target != nil {
				// append to the result
				scan.Targets = append(scan.Targets, target)
			}
		}
	}
}

// list the possible IP list, set iface net if scan.IP is empty
func (scan *Scan) ListIP(ctx context.Context) (ch <-chan net.IP) {
	tmp := make(chan net.IP, 1)

	scan.Info("scan ip list: #%v", scan.IP)
	go func() {
		if scan.IP == nil {
			// list all subnet in the iface
			addrs, err := scan.IFace.Addrs()
			if err != nil {
				scan.Crit("get %#v address: %v", scan.IFace.Name, err)
				return
			}

			for _, addr := range addrs {
				if inet, ok := addr.(*net.IPNet); ok {
					switch {
					case !scan.IPv6 && inet.IP.To4() != nil:
						scan.Debug("enumerate all IP in subnet %v", inet)

						ip := binary.BigEndian.Uint32([]byte(inet.IP.To4()))
						mask := binary.BigEndian.Uint32([]byte(inet.Mask))
						// the first IP in the mask
						ip &= mask

						for mask < ^uint32(0) {
							select {
							case <-ctx.Done():
								scan.Info("stop all IP list")
								close(tmp)
								return
							default:
								var buff [4]byte

								binary.BigEndian.PutUint32(buff[:], ip)
								tmp <- net.IP(buff[:])

								ip++
								mask++
							}
						}
					}
				}
			}
		}

		close(tmp)
		scan.Info("stop list IP")
	}()

	ch = tmp
	return
}
