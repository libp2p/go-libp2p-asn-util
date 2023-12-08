package internal

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math/bits"
	"net"

	u "github.com/ipfs/go-ipfs-util"
)

func ReadMappingFile(f io.Reader) (map[string]string, error) {
	m := make(map[string]string)
	r := csv.NewReader(f)
	r.Comma = '\t'
	for {
		record, err := r.Read()
		// Stop at EOF.
		if err == io.EOF {
			return m, nil
		}

		if len(record) < 3 {
			return nil, fmt.Errorf("invalid record %s", record)
		}

		startIP := record[0]
		endIP := record[1]
		asn := record[2]
		if asn == "0" {
			continue
		}

		s := net.ParseIP(startIP)
		e := net.ParseIP(endIP)
		if s.To16() == nil || e.To16() == nil {
			return nil, errors.New("IP should be v6")
		}

		prefixLen := zeroPrefixLen(u.XOR(s.To16(), e.To16()))
		cn := fmt.Sprintf("%s/%d", startIP, prefixLen)
		m[cn] = asn
	}
}

func zeroPrefixLen(id []byte) int {
	for i, b := range id {
		if b != 0 {
			return i*8 + bits.LeadingZeros8(uint8(b))
		}
	}
	return len(id) * 8
}
