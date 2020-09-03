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

func KeysEqual(x, y *Key) bool {
	return NumbersEqual(x.Number(), y.Number())
}

func (k *Key) Len() int {
	if len(k.IP) > 0 {
		return len(k.IP) * 8
	} else {
		s, _ := k.Net.Mask.Size()
		return s
	}
}

func (k *Key) BitAt(i int) byte {
	return k.Number().BitAt(i)
}

func (k *Key) Number() Number {
	if len(k.IP) > 0 {
		return Number{Bytes: k.IP, Len: k.Len()}
	} else {
		return Number{Bytes: k.Net.IP, Len: k.Len()}
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
