package knock

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/cmj0121/argparse"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// several network protocol scan
type Scan struct {
	argparse.Help

	sync.Once    `-`
	*pcap.Handle `-`
	stop         chan struct{} `-`

	RTT        int            `default:"4000" help:"the maximal assumption RTT in ms"`
	IFace      *net.Interface `args:"option" short:"i" help:"scan on specified iface"`
	MaxPkgSize int32          `short:"P" name:"pkg-size" default:"65536" help:"maximal packet size"`

	// IP meta
	IPv6 bool `short:"6" help:"only scan IPv6"`

	ARP bool `help:"scan via the ARP protocol"`

	CIDR *string `help:"scan the specified CIDR (IP/mask)"`
}

func (scan *Scan) Open() (err error) {
	if scan.IFace == nil {
		err = fmt.Errorf("should specified IFace")
		return
	}

	if scan.CIDR == nil {
		// load the iface default IP/mask
		var addrs []net.Addr

		if addrs, err = scan.IFace.Addrs(); err != nil {
			err = fmt.Errorf("cannot get addr on %#v: %v", scan.IFace.Name, err)
			return
		}
		for _, addr := range addrs {
			if inet, ok := addr.(*net.IPNet); ok {
				switch {
				case !scan.IPv6 && inet.IP.To4() != nil:
					cidr := inet.String()
					scan.CIDR = &cidr
					break
				}
			}
			if scan.CIDR != nil {
				// already set the CIDR
				break
			}
		}
	}

	if scan.Handle, err = pcap.OpenLive(scan.IFace.Name, scan.MaxPkgSize, true, pcap.BlockForever); err != nil {
		err = fmt.Errorf("open %v: %v", scan.IFace.Name, err)
		return
	}

	scan.stop = make(chan struct{}, 1)
	switch {
	case scan.ARP:
	default:
		scan.ARP = true
	}
	return
}

func (scan *Scan) Close() (err error) {
	close(scan.stop)
	scan.Handle.Close()
	return
}

func (scan *Scan) Run(receiver chan<- Response, broker <-chan string) {
	scan.Once.Do(func() {
		// start receive the package by the *Scan config
		go scan.RecvPkg(receiver)
	})

	_, ipnet, _ := net.ParseCIDR(*scan.CIDR)
	if addrs, err := scan.IFace.Addrs(); err == nil {
		for {
			ip, ok := <-broker
			if !ok {
				// end-of-task
				break
			}

			// show the progress
			receiver <- Response{
				Type:    RESP_PROGRESS,
				Message: ip,
			}

			var srcIP net.IP
			for _, addr := range addrs {
				if inet, ok := addr.(*net.IPNet); ok {
					if ipnet.Contains(inet.IP) {
						srcIP = inet.IP
						break
					}
				}
			}

			// build the ARP request
			pkg_eth := layers.Ethernet{
				SrcMAC:       scan.IFace.HardwareAddr,
				DstMAC:       net.HardwareAddr{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
				EthernetType: layers.EthernetTypeARP,
			}
			pkg_arp := layers.ARP{
				AddrType:          layers.LinkTypeEthernet,
				Protocol:          layers.EthernetTypeIPv4,
				HwAddressSize:     6,
				ProtAddressSize:   4,
				Operation:         layers.ARPRequest,
				SourceHwAddress:   []byte(scan.IFace.HardwareAddr),
				SourceProtAddress: []byte(srcIP.To4()),
				DstHwAddress:      []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			}
			// raw packet
			pkg := gopacket.NewSerializeBuffer()
			pkg_opts := gopacket.SerializeOptions{
				FixLengths:       true,
				ComputeChecksums: true,
			}

			pkg_arp.DstProtAddress = []byte(net.ParseIP(ip).To4())
			if err := gopacket.SerializeLayers(pkg, pkg_opts, &pkg_eth, &pkg_arp); err != nil {
				// show the progress
				receiver <- Response{
					Type:    RESP_PROGRESS,
					Message: fmt.Sprintf("cannot build pkg: %v", err),
				}
				return
			} else if err := scan.Handle.WritePacketData(pkg.Bytes()); err != nil {
				// show the progress
				receiver <- Response{
					Type:    RESP_PROGRESS,
					Message: fmt.Sprintf("cannot send pkg: %v", err),
				}
				return
			}
		}
	}

	// wait for response
	receiver <- Response{
		Type:    RESP_PROGRESS,
		Message: fmt.Sprintf("wait %d ms for RTT", scan.RTT),
	}
	time.Sleep(time.Duration(scan.RTT) * time.Millisecond)
}

func (scan *Scan) Broker(ctx context.Context) (ch <-chan string) {
	broker := make(chan string, 1)

	go func() {
		defer close(broker)

		ip, ipnet, _ := net.ParseCIDR(*scan.CIDR)
		// reset the IP in the mask
		ip = ip.Mask(ipnet.Mask)

		// enumerate all IP in the subnet
		for ipnet.Contains(ip) {
			broker <- ip.String()
			// increase IP
			for idx := len(ip) - 1; idx >= 0; idx-- {
				ip[idx]++
				if ip[idx] > 0 {
					break
				}
			}
		}
	}()

	ch = broker
	return
}

// process the receive package and reply to receiver
func (scan *Scan) RecvPkg(receiver chan<- Response) {
	src := gopacket.NewPacketSource(scan.Handle, layers.LayerTypeEthernet)
	in := src.Packets()

	for {
		select {
		case <-scan.stop:
			return
		case pkg := <-in:
			arp_layer := pkg.Layer(layers.LayerTypeARP)

			switch {
			case scan.ARP && arp_layer != nil:
				// receive ARP packet
				pkg_arp := arp_layer.(*layers.ARP)

				switch pkg_arp.Operation {
				case layers.ARPReply:
					ip := net.IP(pkg_arp.SourceProtAddress).String()
					hostnames, _ := net.LookupAddr(ip)
					hostname := "<UNKNOWN>"
					if len(hostnames) > 0 {
						// load the first possible hostname
						hostname = hostnames[0]
					}
					receiver <- Response{
						Type: RESP_RESULT,
						Message: fmt.Sprintf(
							"%-8s %-22v %-22v (%v)",
							layers.LayerTypeARP,
							ip,
							hostname, // show the first possible hostname
							net.HardwareAddr(pkg_arp.SourceHwAddress),
						),
					}
				}
			}
		}
	}
}
