package main

import (
	"net"

	asnutil "github.com/libp2p/go-libp2p-asn-util"
)

func main() {
	_, ipNet, _ := net.ParseCIDR("1.2.3.4/22")
	asnutil.Store.AsnForIPv6(ipNet.IP)
}
