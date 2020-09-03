package asnutil

import (
	"net"
	"testing"
)

// TestStartup tests that the global store is preloaded successfully.
func TestStartup(t *testing.T) {
	// preloading is triggered on first use.
	asn, err := Store.AsnForIPv6(net.IP{0x20, 0x01, 0x00, 0xc0, 0x00, 0x03})
	if err != nil {
		t.Errorf("expected success (%v)", err)
	}
	expected := "22884"
	if asn != expected {
		t.Errorf("expected %v, got %v", expected, asn)
	}
}
