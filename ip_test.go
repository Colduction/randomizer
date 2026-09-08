package randomizer_test

import (
	"net"
	"net/netip"
	"testing"

	"github.com/colduction/randomizer-go"
)

func TestNetworkIPKindValues(t *testing.T) {
	if randomizer.IPv6InterfaceLocalMulticast != 11 || randomizer.IPv6GlobalMulticast != 16 ||
		randomizer.IPv6Private != randomizer.IPv6UniqueLocal || randomizer.IPv4Loopback != 17 {
		t.Fatal("IPKind constants changed value")
	}
}

func TestNetworkIP(t *testing.T) {
	ip4 := randomizer.Network.IP(nil, randomizer.IPv4Any)
	if len(ip4) != net.IPv4len || ip4.To4() == nil {
		t.Fatalf("IP(IPv4Any) = %v, want 4-byte IPv4", ip4)
	}

	ip6 := randomizer.Network.IP(nil, randomizer.IPv6Any)
	if len(ip6) != net.IPv6len || ip6.To16() == nil {
		t.Fatalf("IP(IPv6Any) = %v, want 16-byte IPv6", ip6)
	}

	var buf [16]byte
	got := randomizer.Network.IP(buf[:0], randomizer.IPv6Any)
	if len(got) != net.IPv6len || &got[0] != &buf[0] {
		t.Fatal("IP did not reuse caller buffer")
	}

	for _, kind := range []randomizer.IPKind{10, 200} {
		if ip := randomizer.Network.IP(nil, kind); len(ip) != net.IPv4len {
			t.Fatalf("IP(%d) length = %d, want IPv4 fallback", kind, len(ip))
		}
	}
}

func TestNetworkIPv4Kinds(t *testing.T) {
	var documentation [3]int
	for range 1000 {
		private := randomizer.Network.IP(nil, randomizer.IPv4Private)
		if private.To4() == nil || !private.IsPrivate() {
			t.Fatalf("IP(IPv4Private) = %v, want private IPv4", private)
		}

		linkLocal := randomizer.Network.IP(nil, randomizer.IPv4LinkLocal)
		if linkLocal.To4() == nil || !linkLocal.IsLinkLocalUnicast() || linkLocal[0] != 169 || linkLocal[1] != 254 ||
			linkLocal[2] == 0 || linkLocal[2] == 255 {
			t.Fatalf("IP(IPv4LinkLocal) = %v, want 169.254.1.0-169.254.254.255", linkLocal)
		}

		multicast := randomizer.Network.IP(nil, randomizer.IPv4Multicast)
		if multicast.To4() == nil || !multicast.IsMulticast() || multicast[0] < 224 || multicast[0] > 239 {
			t.Fatalf("IP(IPv4Multicast) = %v, want multicast IPv4", multicast)
		}

		public := randomizer.Network.IP(nil, randomizer.IPv4Public)
		if public.To4() == nil || public.IsPrivate() || public.IsLoopback() ||
			public.IsLinkLocalUnicast() || public.IsMulticast() || public.IsUnspecified() ||
			public[0] == 0 || public[0] >= 240 {
			t.Fatalf("IP(IPv4Public) = %v, want public IPv4", public)
		}

		loopback := randomizer.Network.IP(nil, randomizer.IPv4Loopback)
		if !loopback.IsLoopback() || loopback.Equal(net.IPv4(127, 0, 0, 0)) || loopback.Equal(net.IPv4(127, 255, 255, 255)) {
			t.Fatalf("IP(IPv4Loopback) = %v, want loopback host", loopback)
		}

		cgnat := randomizer.Network.IP(nil, randomizer.IPv4CGNAT)
		if len(cgnat) != net.IPv4len || cgnat[0] != 100 || cgnat[1]&0xc0 != 64 {
			t.Fatalf("IP(IPv4CGNAT) = %v, want 100.64.0.0/10", cgnat)
		}

		doc := randomizer.Network.IP(nil, randomizer.IPv4Documentation)
		switch {
		case doc[0] == 192 && doc[1] == 0 && doc[2] == 2:
			documentation[0]++
		case doc[0] == 198 && doc[1] == 51 && doc[2] == 100:
			documentation[1]++
		case doc[0] == 203 && doc[1] == 0 && doc[2] == 113:
			documentation[2]++
		default:
			t.Fatalf("IP(IPv4Documentation) = %v, want RFC 5737 block", doc)
		}
	}
	for i, n := range documentation {
		if n == 0 {
			t.Fatalf("IPv4Documentation block %d never selected", i)
		}
	}
}

func TestNetworkIPv6Kinds(t *testing.T) {
	cases := []struct {
		kind randomizer.IPKind
		fn   func(net.IP) bool
	}{
		{randomizer.IPv6Global, func(ip net.IP) bool { return ip[0]&0xe0 == 0x20 }},
		{randomizer.IPv6LinkLocal, func(ip net.IP) bool {
			return ip[0] == 0xfe && ip[1] == 0x80 && ip[2]|ip[3]|ip[4]|ip[5]|ip[6]|ip[7] == 0
		}},
		{randomizer.IPv6SiteLocal, func(ip net.IP) bool { return ip[0] == 0xfe && ip[1]&0xc0 == 0xc0 }},
		{randomizer.IPv6UniqueLocal, func(ip net.IP) bool { return ip[0] == 0xfd }},
		{randomizer.IPv6Private, func(ip net.IP) bool { return ip[0] == 0xfd }},
		{randomizer.IPv6Documentation, func(ip net.IP) bool {
			return ip[0] == 0x20 && ip[1] == 0x01 && ip[2] == 0x0d && ip[3] == 0xb8
		}},
	}
	for _, tc := range cases {
		for range 100 {
			ip := randomizer.Network.IP(nil, tc.kind)
			if len(ip) != net.IPv6len || !tc.fn(ip) {
				t.Fatalf("IP(%v) = %v, prefix mismatch", tc.kind, ip)
			}
		}
	}
}

func TestNetworkIPv6Public(t *testing.T) {
	special := []netip.Prefix{
		netip.MustParsePrefix("2001::/23"), netip.MustParsePrefix("2001:db8::/32"),
		netip.MustParsePrefix("2002::/16"), netip.MustParsePrefix("2620:4f:8000::/48"),
		netip.MustParsePrefix("3fff::/20"), netip.MustParsePrefix("5f00::/16"),
	}
	for range 10000 {
		ip := randomizer.Network.IP(nil, randomizer.IPv6Public)
		addr, ok := netip.AddrFromSlice(ip)
		if !ok || !addr.Is6() || !addr.IsGlobalUnicast() || ip[0]&0xe0 != 0x20 {
			t.Fatalf("IP(IPv6Public) = %v, want 2000::/3 global unicast", ip)
		}
		for _, prefix := range special {
			if prefix.Contains(addr) {
				t.Fatalf("IP(IPv6Public) = %v inside special-purpose %v", ip, prefix)
			}
		}
		if ip[8]|ip[9]|ip[10]|ip[11]|ip[12]|ip[13]|ip[14]|ip[15] == 0 {
			t.Fatalf("IP(IPv6Public) = %v has subnet-router anycast IID", ip)
		}
	}
}

func TestNetworkIPv6Multicast(t *testing.T) {
	cases := []struct {
		kind  randomizer.IPKind
		scope byte
	}{
		{randomizer.IPv6InterfaceLocalMulticast, 0x1},
		{randomizer.IPv6LinkLocalMulticast, 0x2},
		{randomizer.IPv6RealmLocalMulticast, 0x3},
		{randomizer.IPv6AdminLocalMulticast, 0x4},
		{randomizer.IPv6SiteLocalMulticast, 0x5},
		{randomizer.IPv6OrgLocalMulticast, 0x8},
		{randomizer.IPv6GlobalMulticast, 0xe},
	}
	for _, tc := range cases {
		ip := randomizer.Network.IP(nil, tc.kind)
		if len(ip) != net.IPv6len || ip[0] != 0xff || ip[1] != 0x10|tc.scope || !ip.IsMulticast() {
			t.Fatalf("IP(%v) = %v, want ff1%x::/16 transient multicast", tc.kind, ip, tc.scope)
		}
	}
}

func TestNetworkIPv6EUI64(t *testing.T) {
	mac := net.HardwareAddr{0x00, 0x1a, 0x2b, 0x3c, 0x4d, 0x5e}
	iid := []byte{0x02, 0x1a, 0x2b, 0xff, 0xfe, 0x3c, 0x4d, 0x5e}
	cases := []struct {
		kind   randomizer.IPKind
		prefix netip.Prefix
	}{
		{randomizer.IPv6LinkLocal, netip.MustParsePrefix("fe80::/64")},
		{randomizer.IPv6Global, netip.MustParsePrefix("2000::/3")},
		{randomizer.IPv6Public, netip.MustParsePrefix("2000::/3")},
		{randomizer.IPv6UniqueLocal, netip.MustParsePrefix("fd00::/8")},
		{randomizer.IPv6SiteLocal, netip.MustParsePrefix("fec0::/10")},
		{randomizer.IPv6Documentation, netip.MustParsePrefix("2001:db8::/32")},
	}
	for _, tc := range cases {
		var buf [16]byte
		ip := randomizer.Network.IPv6EUI64(buf[:0], tc.kind, mac)
		addr, _ := netip.AddrFromSlice(ip)
		if len(ip) != net.IPv6len || &ip[0] != &buf[0] || !tc.prefix.Contains(addr) || string(ip[8:]) != string(iid) {
			t.Fatalf("IPv6EUI64(%v) = %v, want %v with IID %x", tc.kind, ip, tc.prefix, iid)
		}
	}
	eui := net.HardwareAddr{0x00, 0x1a, 0x2b, 0x11, 0x22, 0x3c, 0x4d, 0x5e}
	if ip := randomizer.Network.IPv6EUI64(nil, randomizer.IPv6LinkLocal, eui); string(ip[8:]) != "\x02\x1a\x2b\x11\x22\x3c\x4d\x5e" {
		t.Fatalf("IPv6EUI64 from EUI-64 = %v, want U/L flipped only", ip)
	}
	if got := randomizer.Network.IPv6EUI64(nil, randomizer.IPv4Any, mac); got != nil {
		t.Fatalf("IPv6EUI64(IPv4Any) = %v, want nil", got)
	}
	if got := randomizer.Network.IPv6EUI64(nil, randomizer.IPv6LinkLocal, mac[:5]); got != nil {
		t.Fatalf("IPv6EUI64 with 5-byte MAC = %v, want nil", got)
	}
}

func BenchmarkNetworkIPIPv4(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchIP = randomizer.Network.IP(nil, randomizer.IPv4Any)
	}
}

func BenchmarkNetworkIPIPv4Buffered(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchIP = randomizer.Network.IP(benchBuffer[:0], randomizer.IPv4Any)
	}
}

func BenchmarkNetworkIPIPv4BufferedParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var buf [16]byte
		for pb.Next() {
			_ = randomizer.Network.IP(buf[:0], randomizer.IPv4Any)
		}
	})
}

func BenchmarkNetworkIPIPv4Public(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchIP = randomizer.Network.IP(benchBuffer[:0], randomizer.IPv4Public)
	}
}

func BenchmarkNetworkIPIPv6(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchIP = randomizer.Network.IP(nil, randomizer.IPv6Any)
	}
}

func BenchmarkNetworkIPIPv6Buffered(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchIP = randomizer.Network.IP(benchBuffer[:0], randomizer.IPv6Global)
	}
}

func BenchmarkNetworkIPIPv6Public(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchIP = randomizer.Network.IP(benchBuffer[:0], randomizer.IPv6Public)
	}
}

func BenchmarkNetworkIPv6EUI64(b *testing.B) {
	mac := net.HardwareAddr{0x00, 0x1a, 0x2b, 0x3c, 0x4d, 0x5e}
	b.ReportAllocs()
	for b.Loop() {
		benchIP = randomizer.Network.IPv6EUI64(benchBuffer[:0], randomizer.IPv6LinkLocal, mac)
	}
}
