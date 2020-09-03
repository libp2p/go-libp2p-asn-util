package asnutil

import (
	"net"
	"sync"

	"github.com/libp2p/go-libp2p-asn-util/trie"
)

var Store *store = &store{}

type store struct {
	sync.Mutex
	m trie.RangeLookup
}

func (s *store) preload() {
	m, err := trie.UnmarshalRanges([]byte(cidrASNRaw))
	if err != nil {
		panic(err)
	}
	s.m = m
}

func (s *store) AsnForIPv6(ip net.IP) (string, error) {
	s.Lock()
	if s.m == nil {
		s.preload()
	}
	s.Unlock()
	return s.m.AsnForIPv6(ip)
}
