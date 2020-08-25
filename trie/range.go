package trie

import (
	"fmt"
	"net"
	"sort"

	"github.com/vmihailenco/msgpack"
)

type Range struct {
	Lower Sub
	Upper Sup
	ASN   string
}

func (r Range) Contains(k *Key) bool {
	return !SubLess(Sub(k.Number()), r.Lower) && !SupLess(r.Upper, Sup(k.Number()))
}

func MarshalRanges(r []Range) ([]byte, error) {
	return msgpack.Marshal(r)
}

func UnmarshalRanges(raw []byte) (RangeLookup, error) {
	var r []Range
	if err := msgpack.Unmarshal(raw, &r); err != nil {
		return nil, err
	}
	return RangeLookup(r), nil
}

type RangeLookup []Range

var ErrNoMatch = fmt.Errorf("No matching networks")

func (r RangeLookup) AsnForIPv6(ip net.IP) (string, error) {
	if len(r) == 0 {
		return "", ErrNoMatch
	}
	key := ipToKey(ip)
	j := sort.Search(len(r),
		func(i int) bool {
			// key < r[i].Lower
			return SubLess(Sub(key.Number()), r[i].Lower)
		})
	if j >= len(r) {
		if r[len(r)-1].Contains(key) {
			return r[len(r)-1].ASN, nil
		} else {
			return "", ErrNoMatch
		}
	}
	if j == 0 {
		return "", ErrNoMatch
	}
	if r[j-1].Contains(key) {
		return r[j-1].ASN, nil
	}
	return "", ErrNoMatch
}
