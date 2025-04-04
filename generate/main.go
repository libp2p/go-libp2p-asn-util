package main

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

const outputFile = "sorted-network-list.bin"

const defaultFile = "https://iptoasn.com/data/ip2asn-v6.tsv.gz"

func main() {
	// file with the ASN mappings for IPv6 CIDRs.
	// See ipv6_asn.tsv
	ipv6File := os.Getenv("ASN_IPV6_FILE")

	if len(ipv6File) == 0 {
		ipv6File = defaultFile
	}
	if strings.Contains(ipv6File, "://") {
		local, err := getMappingFile(ipv6File)
		if err != nil {
			panic(err)
		}
		ipv6File = local
	}

	networks := readMappingFile(ipv6File)

	// Keep fixing and optimizing until we get a stable configuration.
	for {
		before := slices.Clone(networks)
		// Ensure the networks are sorted, first smallest start then biggest end.
		slices.SortStableFunc(networks, func(a, b entry) int {
			switch {
			case a.start < b.start:
				return -1
			case a.start == b.start:
				switch {
				case a.end < b.end:
					return 1
				case a.end == b.end:
					return 0
				default:
					return -1
				}
			default:
				return 1
			}
		})

		// Merge adjacent ranges
		{
			var new []entry
			var old entry
			for _, n := range networks {
				if n.start <= old.end || n.start <= old.end+1 {
					// mergeable
					if n.asn == old.asn {
						// merge
						if n.end > old.end {
							old.end = n.end
						}
						continue
					}
					// We have an overlap of different networks :'(
					// Real example:
					//	2403:8080::	2403:8080:ffff:ffff:ffff:ffff:ffff:ffff	17964	CN	DXTNET Beijing Dian-Xin-Tong Network Technologies Co., Ltd.
					//	2403:8080:101::	2403:8080:101:ffff:ffff:ffff:ffff:ffff	4847	CN	CNIX-AP China Networks Inter-Exchange
					// Split the networks.
					if n.start < old.start {
						n, old = old, n
					}
					switch {
					case n.start > old.start && n.end < old.end:
						// |   old   |
						//   | new |
						new = append(new, entry{old.start, n.start - 1, old.asn}, n)
						old.start = n.end + 1
					case n.start > old.start && n.end == old.end:
						// |   old   |
						//     | new |
						fallthrough
					case n.start > old.start && n.end > old.end:
						// |   old   |
						//       | new |
						new = append(new, entry{old.start, n.start - 1, old.asn})
						old = n
					case n.start == old.start && n.end > old.end:
						// |   old   |
						// |    new    |
						n, old = old, n
						fallthrough
					case n.start == old.start && n.end < old.end:
						// |   old   |
						// | new |
						new = append(new, n)
						old.start = n.end + 1
					case n.start == old.start && n.end == old.end:
						// |   old   |
						// |   new   |
						// ¯\_(ツ)_/¯ here we are kinda fucked
						// in theory we could try merging boths with the next elements and keep the smallest one
						// but that is tricky since the next elements could also need the same treatment.
						// Instead drop old, it most likely is a bigger range which was truncated down, and then we should most often prioritize the smallest one.
						old = n
					default:
						panic(fmt.Sprintf("unreachabe: old: %v, new: %v", old, n))
					}
					continue
				}

				// save
				if old.asn != 0 {
					new = append(new, old)
				}
				old = n
			}
			if old.asn != 0 {
				// trailing save
				new = append(new, old)
			}
			networks = new
		}

		if slices.Equal(before, networks) {
			// stabilized
			break
		}
	}

	// Sanity check, we should have increasing networks without overlap
	var old uint64
	for i, n := range networks {
		if n.start > n.end {
			panic(fmt.Sprintf("at %d; %d has backward range %v %v", i, n.asn, networkToIp(n.start), networkToIp(n.end)))
		}
		if n.start <= old {
			panic(fmt.Sprintf("at %d; %d isn't correctly ordered %v vs %v", i, n.asn, networkToIp(n.start), networkToIp(old)))
		}
		old = n.end
	}

	f, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer func() {
		if err := w.Flush(); err != nil {
			panic(err)
		}
	}()
	var b [8]byte
	for _, n := range networks {
		// Write ips "backward" as little endian since all archs we care about natively use little endian.
		// We could in theory produce an le and be version of the file and use build tags but I don't care enough.
		// It will still work, but might add a nano second or two to flip the endianness at runtime on be arches.

		// Only store 48 most significant bits since on public BGP networks smaller than 48 bits are not allowed.
		binary.LittleEndian.PutUint64(b[:], n.start)
		_, err := w.Write(b[2:])
		if err != nil {
			panic(err)
		}
		binary.LittleEndian.PutUint64(b[:], n.end)
		_, err = w.Write(b[2:])
		if err != nil {
			panic(err)
		}
		binary.LittleEndian.PutUint32(b[:], n.asn)
		_, err = w.Write(b[:4])
		if err != nil {
			panic(err)
		}
	}
}

func readMappingFile(path string) (l []entry) {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comma = '\t'
	for {
		record, err := r.Read()
		// Stop at EOF.
		if err == io.EOF {
			return
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
			panic(errors.New("IP should be v6"))
		}

		as, err := strconv.ParseUint(asn, 10, 32)
		if err != nil {
			panic(err)
		}

		l = append(l, entry{
			start: binary.BigEndian.Uint64(s),
			end:   binary.BigEndian.Uint64(e),
			asn:   uint32(as),
		})
	}
}

// Get a url, return file it's downloaded to. optionally gzip decode.
func getMappingFile(url string) (path string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	baseFile, err := os.CreateTemp("", "ip-map-download-*")
	if err != nil {
		return
	}
	defer baseFile.Close()
	_, err = io.Copy(baseFile, resp.Body)
	if err != nil {
		return
	}
	initBuf := make([]byte, 512)
	_, err = baseFile.ReadAt(initBuf, 0)
	if err != nil {
		return
	}
	if strings.Contains(http.DetectContentType(initBuf), "application/x-gzip") {
		// gunzip it.
		_, err = baseFile.Seek(0, io.SeekStart)
		if err != nil {
			return
		}
		var gzr *gzip.Reader
		gzr, err = gzip.NewReader(baseFile)
		if err != nil {
			return
		}
		var rawFile *os.File
		rawFile, err = os.CreateTemp("", "ip-map-download-*")
		if err != nil {
			return
		}
		defer os.Remove(baseFile.Name())
		defer rawFile.Close()
		_, err = io.Copy(rawFile, gzr)
		if err != nil {
			return
		}
		path = rawFile.Name()
		return
	}
	path = baseFile.Name()
	return
}

type entry struct {
	// networks
	start, end uint64

	asn uint32
}

func (e entry) String() string {
	return fmt.Sprintf("{%v, %v, %d}", networkToIp(e.start), networkToIp(e.end), e.asn)
}

func networkToIp(net uint64) net.IP {
	var ip [16]byte
	binary.BigEndian.PutUint64(ip[:], net)
	return ip[:]
}
