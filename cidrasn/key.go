package cidrasn

import (
	"fmt"
	"net"

	"github.com/libp2p/go-libp2p-asn-util/trie"
)

type cidrKey struct {
	IP  net.IP
	Net net.IPNet
	ASN string
}

func cidrToKey(ipNet net.IPNet, asn string) cidrKey {
	return cidrKey{Net: ipNet, ASN: asn}
}

func ipToKey(ip net.IP) cidrKey {
	return cidrKey{IP: ip}
}

func (k cidrKey) Len() int {
	if len(k.IP) > 0 {
		return len(k.IP) * 8
	} else {
		s, _ := k.Net.Mask.Size()
		return s
	}
}

func (k cidrKey) String() string {
	if len(k.IP) == 0 {
		return fmt.Sprintf("%v/%v-->%s", k.Net.IP, k.Len(), k.ASN)
	} else {
		return fmt.Sprintf("%v", k.IP)
	}
}

func (k cidrKey) asIP() net.IP {
	if len(k.IP) > 0 {
		return k.IP
	} else {
		return k.Net.IP
	}
}

func (k cidrKey) Equal(r trie.Key) bool {
	if k2, ok := r.(cidrKey); ok {
		if k.Len() != k2.Len() {
			return false
		} else {
			return commonPrefixLen(k.asIP(), k2.asIP()) >= k.Len()
		}
	} else {
		return false
	}
}

func (k cidrKey) BitAt(i int) byte {
	b := []byte(k.asIP())
	// the most significant byte in an IP address is the first one
	d := b[i/8] & (byte(1) << (7 - (i % 8)))
	if d == 0 {
		return 0
	} else {
		return 1
	}
}

func commonPrefixLen(a, b []byte) (cpl int) {
	if len(a) > len(b) {
		a = a[:len(b)]
	}
	if len(b) > len(a) {
		b = b[:len(a)]
	}
	for len(a) > 0 {
		if a[0] == b[0] {
			cpl += 8
			a = a[1:]
			b = b[1:]
			continue
		}
		bits := 8
		ab, bb := a[0], b[0]
		for {
			ab >>= 1
			bb >>= 1
			bits--
			if ab == bb {
				cpl += bits
				return
			}
		}
	}
	return
}
