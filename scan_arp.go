package knock

import (
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func (scan *Scan) sendARP(handler *pcap.Handle, iface *net.Interface, ip net.IP) {
	addrs, err := iface.Addrs()
	if err != nil {
		scan.Warn("cannot get address on %#v: %v", iface.Name, err)
		return
	}

	for _, addr := range addrs {
		if inet, ok := addr.(*net.IPNet); ok {
			switch {
			case inet.IP.To4() != nil && ip.To4() != nil:
				pkg_eth := layers.Ethernet{
					SrcMAC:       iface.HardwareAddr,
					DstMAC:       net.HardwareAddr{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
					EthernetType: layers.EthernetTypeARP,
				}
				pkg_arp := layers.ARP{
					AddrType:          layers.LinkTypeEthernet,
					Protocol:          layers.EthernetTypeIPv4,
					HwAddressSize:     6,
					ProtAddressSize:   4,
					Operation:         layers.ARPRequest,
					SourceHwAddress:   []byte(iface.HardwareAddr),
					SourceProtAddress: []byte(inet.IP.To4()),
					DstHwAddress:      []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				}
				// raw packet
				pkg := gopacket.NewSerializeBuffer()
				pkg_opts := gopacket.SerializeOptions{
					FixLengths:       true,
					ComputeChecksums: true,
				}

				pkg_arp.DstProtAddress = []byte(ip)
				gopacket.SerializeLayers(pkg, pkg_opts, &pkg_eth, &pkg_arp)
				if err = handler.WritePacketData(pkg.Bytes()); err != nil {
					scan.Logger.Warn("send ARP scan: %v", err)
					return
				}

				scan.Logger.Info("send ARP scan to %v", ip)
			}
		}
	}
}

func (scan *Scan) recvARP(pkg gopacket.Packet) (target *Target) {
	if arp_layer := pkg.Layer(layers.LayerTypeARP); arp_layer != nil {
		// receive ARP packet
		pkg_arp := arp_layer.(*layers.ARP)
		switch pkg_arp.Operation {
		case layers.ARPReply:
			// ARP reply
			target = &Target{
				LayerType:    layers.LayerTypeARP,
				IP:           net.IP(pkg_arp.SourceProtAddress),
				HardwareAddr: net.HardwareAddr(pkg_arp.SourceHwAddress),
			}
		}
	}

	return
}
