package asnutil

import (
	"math/rand"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAsnIpv6(t *testing.T) {
	tcs := map[string]struct {
		ip          net.IP
		expectedASN string
	}{
		"google": {
			ip:          net.ParseIP("2001:4860:4860::8888"),
			expectedASN: "15169",
		},
		"facebook": {
			ip:          net.ParseIP("2a03:2880:f003:c07:face:b00c::2"),
			expectedASN: "32934",
		},
		"comcast": {
			ip:          net.ParseIP("2601::"),
			expectedASN: "7922",
		},
		"does not exist": {
			ip:          net.ParseIP("::"),
			expectedASN: "",
		},
	}

	for name, tc := range tcs {
		require.NotEmpty(t, tc.ip, name)
		t.Run(name, func(t *testing.T) {
			n, err := Store.AsnForIPv6(tc.ip)
			require.NoError(t, err)
			require.Equal(t, tc.expectedASN, n, name)
		})
	}
}

var leak uint32

func BenchmarkAsnForIPv6(b *testing.B) {
	r := rand.New(rand.NewSource(0))
	var doNotOptimize uint32
	for i := b.N; i != 0; i-- {
		doNotOptimize = AsnForIPv6Network(r.Uint64())
	}
	leak = doNotOptimize
}
