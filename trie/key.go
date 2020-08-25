package trie

import (
	"fmt"
	"net"
)

type Key struct {
	IP  net.IP
	Net net.IPNet
	ASN string
}

func cidrToKey(ipNet net.IPNet, asn string) *Key {
	return &Key{Net: ipNet, ASN: asn}
}

func ipToKey(ip net.IP) *Key {
	return &Key{IP: ip}
}

func (k *Key) Len() int {
	if len(k.IP) > 0 {
		return len(k.IP) * 8
	} else {
		s, _ := k.Net.Mask.Size()
		return s
	}
}

func (k *Key) String() string {
	if len(k.IP) == 0 {
		return fmt.Sprintf("%v/%v-->%s", k.Net.IP, k.Len(), k.ASN)
	} else {
		return fmt.Sprintf("%v", k.IP)
	}
}

func (k *Key) asIP() net.IP {
	if len(k.IP) > 0 {
		return k.IP
	} else {
		return k.Net.IP
	}
}

func (k *Key) Equal(r *Key) bool {
	if k.Len() != r.Len() {
		return false
	} else {
		return commonPrefixLen(k.asIP(), r.asIP()) >= k.Len()
	}
}

func (k *Key) BitAt(i int) byte {
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
