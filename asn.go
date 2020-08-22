package asnutil

import (
	"encoding/binary"
	"errors"
	"math"
	"net"
	"sort"
)

type asn struct {
	prefix uint64
	asn    string
}

// AsnForIPv6 returns the AS number for the given IPv6 address.
// If no mapping exists for the given IP, this function will
// return "" and a nil error.
func AsnForIPv6(ip net.IP) (string, error) {
	ip = ip.To16()
	if ip == nil {
		return "", errors.New("ONLY IPv6 addresses supported for now")
	}

	targetPrefix := binary.BigEndian.Uint64(ip)

	idx := sort.Search(len(ipv6CidrToAsnMap), func(i int) bool {
		return ipv6CidrToAsnMap[i].prefix&^uint64(0xFF) > targetPrefix
	})
	if idx == 0 {
		return "", nil
	}
	a := ipv6CidrToAsnMap[idx-1]
	prefixLen := a.prefix & 0xFF
	prefix := a.prefix & ^uint64(0xFF)
	if prefix == targetPrefix&(math.MaxUint64<<(64-prefixLen)) {
		return a.asn, nil
	}
	return "", nil
}

type asnStore struct{}

func (a asnStore) AsnForIPv6(ip net.IP) (string, error) {
	return AsnForIPv6(ip)
}

var Store = asnStore{}
