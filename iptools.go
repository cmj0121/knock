package knock

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/google/gopacket"
)

type Target struct {
	gopacket.LayerType

	net.HardwareAddr
	net.IP
}

func (target *Target) String() (str string) {
	str = fmt.Sprintf("%v %-16v (%v)", target.LayerType, target.IP, target.HardwareAddr)
	return
}

func IterateIPRange(ctx context.Context, inet *net.IPNet) (ch <-chan net.IP) {
	tmp := make(chan net.IP, 1)

	go func() {
		switch {
		case inet.IP.To4() != nil:
			ip := binary.BigEndian.Uint32([]byte(inet.IP.To4()))
			mask := binary.BigEndian.Uint32([]byte(inet.Mask))
			// the first IP in the mask
			ip &= mask

			for mask < ^uint32(0) {
				select {
				case <-ctx.Done():
					// closed by the context
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

		close(tmp)
	}()

	ch = tmp
	return
}
