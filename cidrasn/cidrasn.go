package cidrasn

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"

	"github.com/libp2p/go-libp2p-asn-util/trie"
)

type CIDRASN struct {
	ipv6 *trie.Trie
}

func NewCIDRASN() *CIDRASN {
	return &CIDRASN{
		ipv6: &trie.Trie{},
	}
}

func (m *CIDRASN) Marshal() ([]byte, error) {
	var w bytes.Buffer
	if err := gob.NewEncoder(&w).Encode(m); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (m *CIDRASN) Add(ipNet net.IPNet, asn string) {
	m.ipv6.Add(cidrToKey(ipNet, asn))
}

func (m *CIDRASN) AsnForIPv6(ip net.IP) (string, error) {
	netKeys := m.containingNetworksIPv6(ip)
	if len(netKeys) == 0 {
		return "", fmt.Errorf("No matching networks")
	}
	return netKeys[0].ASN, nil
}

func (m *CIDRASN) containingNetworksIPv6(ip net.IP) []cidrKey {
	_, found := m.ipv6.FindSubKeys(ipToKey(ip))
	q := []cidrKey{}
	for _, f := range found {
		k := f.(cidrKey)
		if k.Net.Contains(ip) {
			q = append(q, k)
		}
	}
	return q
}
