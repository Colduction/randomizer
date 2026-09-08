package randomizer_test

import (
	"net"
	"testing"

	"github.com/colduction/randomizer-go"
)

var (
	benchIP     net.IP
	benchMAC    net.HardwareAddr
	benchPort   uint64
	benchIPNet  *net.IPNet
	benchValue  uint64
	benchBuffer [64]byte
)

func TestNetworkAllocs(t *testing.T) {
	if allocs := testing.AllocsPerRun(1000, func() {
		var buf [16]byte
		ip := randomizer.Network.IP(buf[:0], randomizer.IPv6Any)
		if len(ip) != net.IPv6len {
			t.Fatal("bad IPv6 length")
		}
	}); allocs != 0 {
		t.Fatalf("IP with caller buffer allocs/op = %v, want 0", allocs)
	}
	if allocs := testing.AllocsPerRun(1000, func() {
		var buf [8]byte
		mac := randomizer.Network.Hardware(buf[:0], randomizer.HardwareEUI64, randomizer.HardwareOptions{})
		if len(mac) != 8 {
			t.Fatal("bad EUI64 length")
		}
	}); allocs != 0 {
		t.Fatalf("Hardware with caller buffer allocs/op = %v, want 0", allocs)
	}
	kinds := []randomizer.HardwareKind{
		randomizer.HardwareMACRealOUI, randomizer.HardwareMACAAI, randomizer.HardwareMACSAI,
		randomizer.HardwareEUI64RealOUI, randomizer.HardwareBluetoothStatic,
		randomizer.HardwareBluetoothNRPA, randomizer.HardwareBluetoothRPA,
	}
	for _, kind := range kinds {
		if allocs := testing.AllocsPerRun(1000, func() {
			var buf [8]byte
			if randomizer.Network.Hardware(buf[:0], kind, randomizer.HardwareOptions{}) == nil {
				t.Fatal("nil address")
			}
		}); allocs != 0 {
			t.Fatalf("Hardware(%d) with caller buffer allocs/op = %v, want 0", kind, allocs)
		}
	}
	if allocs := testing.AllocsPerRun(1000, func() {
		var buf [8]byte
		options := randomizer.HardwareOptions{IRK: []byte("0123456789abcdef")}
		if randomizer.Network.Hardware(buf[:0], randomizer.HardwareBluetoothRPA, options) == nil {
			t.Fatal("nil address")
		}
	}); allocs != 0 {
		t.Fatalf("Hardware(RPA with IRK) steady-state allocs/op = %v, want 0", allocs)
	}
	if allocs := testing.AllocsPerRun(1000, func() {
		benchValue = randomizer.Network.Value(randomizer.ASN)
	}); allocs != 0 {
		t.Fatalf("Value allocs/op = %v, want 0", allocs)
	}
	if allocs := testing.AllocsPerRun(1000, func() {
		benchIPNet = randomizer.Network.CIDR(randomizer.IPv6ULAPrefix, 0)
	}); allocs > 1 {
		t.Fatalf("CIDR(IPv6ULAPrefix) allocs/op = %v, want <= 1", allocs)
	}
	if allocs := testing.AllocsPerRun(1000, func() {
		var buf [16]byte
		mac := net.HardwareAddr{0, 1, 2, 3, 4, 5}
		if randomizer.Network.IPv6EUI64(buf[:0], randomizer.IPv6LinkLocal, mac) == nil {
			t.Fatal("nil address")
		}
	}); allocs != 0 {
		t.Fatalf("IPv6EUI64 with caller buffer allocs/op = %v, want 0", allocs)
	}
}
