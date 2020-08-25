package main

import (
	"fmt"
	"net"
	"runtime"

	asnutil "github.com/libp2p/go-libp2p-asn-util"
)

func main() {
	_, ipNet, _ := net.ParseCIDR("1.2.3.4/22")
	asnutil.Store.AsnForIPv6(ipNet.IP)
	runtime.GC()
	PrintMemUsage()
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
