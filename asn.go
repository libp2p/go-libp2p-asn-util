package asnutil

import (
	"net"

	"github.com/libp2p/go-libp2p-asn-util/trie"
)

var Store *store

func init() {
	m, err := trie.Unmarshal([]byte(cidrASNRaw))
	if err != nil {
		panic(err)
	}
	Store = &store{m}
}

type store struct {
	m *trie.CIDRASN
}

func (s *store) AsnForIPv6(ip net.IP) (string, error) {
	return s.m.AsnForIPv6(ip)
}
