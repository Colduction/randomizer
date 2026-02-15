package randomizer_test

import (
	"net"
	"testing"

	"github.com/colduction/randomizer"
)

var (
	benchIP  net.IP
	benchMAC net.HardwareAddr
)

func TestNetworkIPv4Addr(t *testing.T) {
	ip := randomizer.Network.IPv4Addr()
	if len(ip) != net.IPv4len {
		t.Fatalf("IPv4Addr length = %d, want %d", len(ip), net.IPv4len)
	}
	if ip.To4() == nil {
		t.Fatal("IPv4Addr returned non-IPv4 address")
	}
}

func TestNetworkIPv6Addr(t *testing.T) {
	ip := randomizer.Network.IPv6Addr()
	if len(ip) != net.IPv6len {
		t.Fatalf("IPv6Addr length = %d, want %d", len(ip), net.IPv6len)
	}
	if ip.To16() == nil {
		t.Fatal("IPv6Addr returned non-IPv6 address")
	}
}

func TestNetworkMACAddrBits(t *testing.T) {
	cases := []struct {
		local     bool
		multicast bool
	}{
		{local: false, multicast: false},
		{local: true, multicast: false},
		{local: false, multicast: true},
		{local: true, multicast: true},
	}
	for _, tc := range cases {
		mac := randomizer.Network.MACAddr(tc.local, tc.multicast)
		if len(mac) != 6 {
			t.Fatalf("MACAddr length = %d, want 6", len(mac))
		}
		gotLocal := (mac[0] & 0x02) != 0
		if gotLocal != tc.local {
			t.Fatalf("MACAddr local bit = %t, want %t (mac=%v)", gotLocal, tc.local, mac)
		}
		gotMulticast := (mac[0] & 0x01) != 0
		if gotMulticast != tc.multicast {
			t.Fatalf("MACAddr multicast bit = %t, want %t (mac=%v)", gotMulticast, tc.multicast, mac)
		}
	}
}

func TestNetworkIPv6UnicastPrefixes(t *testing.T) {
	global := randomizer.Network.IPv6UnicastAddr(randomizer.GlobalType)
	if len(global) != net.IPv6len {
		t.Fatalf("Global unicast length = %d, want %d", len(global), net.IPv6len)
	}
	if global[0]&0xE0 != 0x20 {
		t.Fatalf("Global unicast prefix mismatch: first byte=0x%02X", global[0])
	}

	linkLocal := randomizer.Network.IPv6UnicastAddr(randomizer.LinkLocalType)
	if linkLocal[0] != 0xFE || (linkLocal[1]&0xC0) != 0x80 {
		t.Fatalf("Link-local prefix mismatch: first two bytes=0x%02X 0x%02X", linkLocal[0], linkLocal[1])
	}

	siteLocal := randomizer.Network.IPv6UnicastAddr(randomizer.SiteLocalType)
	if siteLocal[0] != 0xFE || (siteLocal[1]&0xC0) != 0xC0 {
		t.Fatalf("Site-local prefix mismatch: first two bytes=0x%02X 0x%02X", siteLocal[0], siteLocal[1])
	}

	uniqueLocal := randomizer.Network.IPv6UnicastAddr(randomizer.UniqueLocalType)
	if uniqueLocal[0] != 0xFD {
		t.Fatalf("Unique-local prefix mismatch: first byte=0x%02X", uniqueLocal[0])
	}

	privateLocal := randomizer.Network.IPv6UnicastAddr(randomizer.PrivateType)
	if privateLocal[0] != 0xFD {
		t.Fatalf("PrivateType prefix mismatch: first byte=0x%02X", privateLocal[0])
	}
}

func TestNetworkIPv6MulticastScope(t *testing.T) {
	scopes := []randomizer.MulticastScope{
		randomizer.InterfaceLocalScope,
		randomizer.LinkLocalScope,
		randomizer.AdminLocalScope,
		randomizer.SiteLocalScope,
		randomizer.OrgLocalScope,
		randomizer.GlobalScope,
	}
	for _, scope := range scopes {
		ip := randomizer.Network.IPv6MulticastAddr(scope)
		if len(ip) != net.IPv6len {
			t.Fatalf("IPv6MulticastAddr length = %d, want %d", len(ip), net.IPv6len)
		}
		if ip[0] != 0xFF {
			t.Fatalf("IPv6MulticastAddr prefix byte = 0x%02X, want 0xFF", ip[0])
		}
		if ip[1]&0x0F != uint8(scope) {
			t.Fatalf("IPv6MulticastAddr scope nibble = 0x%X, want 0x%X", ip[1]&0x0F, uint8(scope))
		}
		if ip[1]&0xF0 != 0x00 {
			t.Fatalf("IPv6MulticastAddr flags nibble = 0x%X, want 0x0", ip[1]>>4)
		}
	}
}

func BenchmarkNetworkIPv4Addr(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchIP = randomizer.Network.IPv4Addr()
	}
}

func BenchmarkNetworkIPv6Addr(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchIP = randomizer.Network.IPv6Addr()
	}
}

func BenchmarkNetworkMACAddr(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchMAC = randomizer.Network.MACAddr(true, true)
	}
}

func BenchmarkNetworkIPv6UnicastAddr(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchIP = randomizer.Network.IPv6UnicastAddr(randomizer.GlobalType)
	}
}

func BenchmarkNetworkIPv6MulticastAddr(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchIP = randomizer.Network.IPv6MulticastAddr(randomizer.GlobalScope)
	}
}
