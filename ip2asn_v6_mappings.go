package asnutil

import (
	"compress/gzip"
	"embed"
	"fmt"

	"github.com/libp2p/go-libp2p-asn-util/internal"
)

//go:embed ip2asn-v6.tsv.gz
var ip2asnv6_Raw embed.FS

func loadIPv6ASNMappings() ([]struct{ cidr, asn string }, error) {
	f, err := ip2asnv6_Raw.Open("ip2asn-v6.tsv.gz")
	if err != nil {
		return nil, fmt.Errorf("failed to open ipv6 ASN data file: %w", err)
	}
	defer f.Close()

	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	m, err := internal.ReadMappingFile(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read ipv6 ASN data file: %w", err)
	}

	s := make([]struct{ cidr, asn string }, 0, len(m))
	for k, v := range m {
		s = append(s, struct {
			cidr string
			asn  string
		}{k, v})
	}

	return s, nil
}
