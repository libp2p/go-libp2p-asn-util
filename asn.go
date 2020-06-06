package asnutil

import (
	"errors"
	"fmt"
	"net"

	"github.com/libp2p/go-cidranger"
)

var Store *asnStore

func init() {
	s, err := NewAsnStore()
	if err != nil {
		panic(err)
	}
	Store = s
}

type networkWithAsn struct {
	nn  net.IPNet
	asn string
}

func (e *networkWithAsn) Network() net.IPNet {
	return e.nn
}

type asnStore struct {
	cr cidranger.Ranger
}

// AsnForIPv6 returns the AS number for the given IPv6 address.
// If no mapping exists for the given IP, this function will
// return an empty ASN and a nil error.
func (a *asnStore) AsnForIPv6(ip net.IP) (string, error) {
	if ip.To16() == nil {
		return "", errors.New("ONLY IPv6 addresses supported for now")
	}

	ns, err := a.cr.ContainingNetworks(ip)
	if err != nil {
		return "", fmt.Errorf("failed to find matching networks for the given ip: %w", err)
	}

	if len(ns) == 0 {
		return "", nil
	}

	// longest prefix match
	n := ns[len(ns)-1].(*networkWithAsn)
	return n.asn, nil
}

// NewAsnStore returns a `asnStore` that can be queried for the Autonomous System Numbers
// for a given IP address or a multiaddress which contains an IP address.
func NewAsnStore() (*asnStore, error) {
	cr := cidranger.NewPCTrieRanger()

	for k, v := range ipv6CidrToAsnMap {
		_, nn, err := net.ParseCIDR(k)
		if err != nil {
			return nil, fmt.Errorf("failed to parse CIDR %s: %w", k, err)
		}

		if err := cr.Insert(&networkWithAsn{*nn, v}); err != nil {
			return nil, fmt.Errorf("failed to insert CIDR %s in Trie store: %w", k, err)
		}
	}

	return &asnStore{cr}, nil
}
