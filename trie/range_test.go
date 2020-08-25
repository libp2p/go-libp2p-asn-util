package trie

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const N = 1000000

func TestWithAWSRangesAgainstRandom(t *testing.T) {
	testContains(t, N, randIPv6Gen)
}

func TestWithAWSRangesAgainstCurated(t *testing.T) {
	testContains(t, N, curatedAWSIPv6Gen)
}

func testContains(t *testing.T, iterations int, ipGen ipGenerator) {
	if testing.Short() {
		t.Skip("Skipping memory test in `-short` mode")
	}
	ds := NewCIDRASN()
	loadAWSRangesIntoDS(t, ds)

	for i := 0; i < iterations; i++ {
		ip := ipGen()
		expected, expectedErr := AWSAsnForIPv6(ip)
		actual, actualErr := ds.AsnForIPv6(ip)
		assert.Equal(t, expected, actual)
		assert.Equal(t, expectedErr, actualErr)
	}
}

// helper routines

type ipGenerator func() net.IP

func randIPv4Gen() net.IP {
	return net.IPv4(byte(rand.Uint32()), byte(rand.Uint32()), byte(rand.Uint32()), byte(rand.Uint32()))
}
func randIPv6Gen() net.IP {
	ip := make(net.IP, 16)
	for i := 0; i < 16; i++ {
		ip[i] = byte(rand.Uint32())
	}
	return ip
}

func curatedAWSIPv6Gen() net.IP {
	randIdx := rand.Intn(len(ipV6AWSRangesIPNets))

	// Randomly generate an IP somewhat near the range.
	network := ipV6AWSRangesIPNets[randIdx].Net
	ones, bits := network.Mask.Size()
	zeros := bits - ones
	nnPartIdx := zeros / 8
	nn := dupIP(network.IP)
	nn[nnPartIdx] = byte(rand.Uint32())
	return nn
}

func dupIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}

// baseline map

func AWSAsnForIPv6(ip net.IP) (string, error) {
	found := []*Key{}
	for _, r := range ipV6AWSRangesIPNets {
		if r.Net.Contains(ip) {
			found = append(found, r)
		}
	}
	if len(found) == 0 {
		return "", ErrNoMatch
	}
	best := found[0]
	for _, f := range found {
		fOnes, _ := f.Net.Mask.Size()
		bestOnes, _ := best.Net.Mask.Size()
		if fOnes > bestOnes {
			best = f
		}
	}
	return best.ASN, nil
}

// test harness

func loadAWSRangesIntoDS(tb testing.TB, ds *CIDRASN) {
	for _, prefix := range awsRanges.Prefixes {
		_, network, err := net.ParseCIDR(prefix.IPPrefix)
		assert.NoError(tb, err)
		ds.Add(*network, network.String())
	}
	for _, prefix := range awsRanges.IPv6Prefixes {
		_, network, err := net.ParseCIDR(prefix.IPPrefix)
		assert.NoError(tb, err)
		ds.Add(*network, network.String())
	}
}

type AWSRanges struct {
	Prefixes     []Prefix     `json:"prefixes"`
	IPv6Prefixes []IPv6Prefix `json:"ipv6_prefixes"`
}

type Prefix struct {
	IPPrefix string `json:"ip_prefix"`
	Region   string `json:"region"`
	Service  string `json:"service"`
}

type IPv6Prefix struct {
	IPPrefix string `json:"ipv6_prefix"`
	Region   string `json:"region"`
	Service  string `json:"service"`
}

var awsRanges *AWSRanges
var ipV4AWSRangesIPNets []*Key
var ipV6AWSRangesIPNets []*Key

func loadAWSRanges() *AWSRanges {
	file, err := ioutil.ReadFile("./testdata/aws_ip_ranges.json")
	if err != nil {
		panic(err)
	}
	var ranges AWSRanges
	err = json.Unmarshal(file, &ranges)
	if err != nil {
		panic(err)
	}
	return &ranges
}

func init() {
	awsRanges = loadAWSRanges()
	for _, prefix := range awsRanges.IPv6Prefixes {
		_, network, _ := net.ParseCIDR(prefix.IPPrefix)
		ipV6AWSRangesIPNets = append(ipV6AWSRangesIPNets, &Key{Net: *network, ASN: network.String()})
	}
	for _, prefix := range awsRanges.Prefixes {
		_, network, _ := net.ParseCIDR(prefix.IPPrefix)
		ipV4AWSRangesIPNets = append(ipV4AWSRangesIPNets, &Key{Net: *network, ASN: network.String()})
	}
	rand.Seed(time.Now().Unix())
}
