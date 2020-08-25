package cidrasn

import (
	"fmt"
	"net"

	"github.com/libp2p/go-libp2p-asn-util/trie"
	"go.mongodb.org/mongo-driver/bson"
)

type CIDRASN struct {
	IPv6 *trie.Trie
}

func NewCIDRASN() *CIDRASN {
	return &CIDRASN{
		IPv6: &trie.Trie{},
	}
}

func Unmarshal(raw []byte) (*CIDRASN, error) {
	m := &CIDRASN{}
	if err := bson.Unmarshal(raw, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *CIDRASN) Marshal() ([]byte, error) {
	return bson.Marshal(m)
}

func (m *CIDRASN) Add(ipNet net.IPNet, asn string) {
	m.IPv6.Add(cidrToKey(ipNet, asn))
}

func (m *CIDRASN) AsnForIPv6(ip net.IP) (string, error) {
	netKeys := m.containingNetworksIPv6(ip)
	if len(netKeys) == 0 {
		return "", fmt.Errorf("No matching networks")
	}
	return netKeys[0].ASN, nil
}

func (m *CIDRASN) containingNetworksIPv6(ip net.IP) []cidrKey {
	_, found := m.IPv6.FindSubKeys(ipToKey(ip))
	q := []cidrKey{}
	for _, f := range found {
		k := f.(cidrKey)
		if k.Net.Contains(ip) {
			q = append(q, k)
		}
	}
	return q
}
