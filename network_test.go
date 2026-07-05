package randomizer_test

import (
	"bytes"
	"net"
	"testing"

	"github.com/colduction/randomizer-go"
)

var (
	benchIP     net.IP
	benchMAC    net.HardwareAddr
	benchPort   uint64
	benchUUID   [16]byte
	benchUUIDB  []byte
	benchIPNet  *net.IPNet
	benchValue  uint64
	benchBuffer [64]byte
)

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
}

func TestNetworkIPv4Kinds(t *testing.T) {
	for range 100 {
		private := randomizer.Network.IP(nil, randomizer.IPv4Private)
		if private.To4() == nil || !private.IsPrivate() {
			t.Fatalf("IP(IPv4Private) = %v, want private IPv4", private)
		}

		linkLocal := randomizer.Network.IP(nil, randomizer.IPv4LinkLocal)
		if linkLocal.To4() == nil || !linkLocal.IsLinkLocalUnicast() || linkLocal[0] != 169 || linkLocal[1] != 254 {
			t.Fatalf("IP(IPv4LinkLocal) = %v, want 169.254.0.0/16", linkLocal)
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
	}
}

func TestNetworkIPv6Kinds(t *testing.T) {
	cases := []struct {
		kind randomizer.IPKind
		fn   func(net.IP) bool
	}{
		{randomizer.IPv6Global, func(ip net.IP) bool { return ip[0]&0xe0 == 0x20 }},
		{randomizer.IPv6LinkLocal, func(ip net.IP) bool { return ip[0] == 0xfe && ip[1]&0xc0 == 0x80 }},
		{randomizer.IPv6SiteLocal, func(ip net.IP) bool { return ip[0] == 0xfe && ip[1]&0xc0 == 0xc0 }},
		{randomizer.IPv6UniqueLocal, func(ip net.IP) bool { return ip[0] == 0xfd }},
		{randomizer.IPv6Private, func(ip net.IP) bool { return ip[0] == 0xfd }},
	}
	for _, tc := range cases {
		ip := randomizer.Network.IP(nil, tc.kind)
		if len(ip) != net.IPv6len || !tc.fn(ip) {
			t.Fatalf("IP(%v) = %v, prefix mismatch", tc.kind, ip)
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
		{randomizer.IPv6AdminLocalMulticast, 0x4},
		{randomizer.IPv6SiteLocalMulticast, 0x5},
		{randomizer.IPv6OrgLocalMulticast, 0x8},
		{randomizer.IPv6GlobalMulticast, 0xe},
	}
	for _, tc := range cases {
		ip := randomizer.Network.IP(nil, tc.kind)
		if len(ip) != net.IPv6len || ip[0] != 0xff || ip[1] != tc.scope {
			t.Fatalf("IP(%v) = %v, want ff%02x::/16 scope byte", tc.kind, ip, tc.scope)
		}
	}
}

func TestNetworkHardware(t *testing.T) {
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
		mac := randomizer.Network.Hardware(nil, randomizer.HardwareMAC, randomizer.HardwareOptions{
			Local:     tc.local,
			Multicast: tc.multicast,
		})
		if len(mac) != 6 {
			t.Fatalf("Hardware(HardwareMAC) length = %d, want 6", len(mac))
		}
		if got := mac[0]&0x02 != 0; got != tc.local {
			t.Fatalf("HardwareMAC local bit = %t, want %t", got, tc.local)
		}
		if got := mac[0]&0x01 != 0; got != tc.multicast {
			t.Fatalf("HardwareMAC multicast bit = %t, want %t", got, tc.multicast)
		}
	}

	var buf [8]byte
	got := randomizer.Network.Hardware(buf[:0], randomizer.HardwareEUI64, randomizer.HardwareOptions{})
	if len(got) != 8 || &got[0] != &buf[0] {
		t.Fatal("Hardware did not reuse caller buffer")
	}
}

func TestNetworkHardwareOUI(t *testing.T) {
	oui := [3]byte{0x02, 0x1a, 0x3f}
	mac := randomizer.Network.Hardware(nil, randomizer.HardwareMACOUI, randomizer.HardwareOptions{OUI: oui})
	if len(mac) != 6 || mac[0] != oui[0] || mac[1] != oui[1] || mac[2] != oui[2] {
		t.Fatalf("Hardware(HardwareMACOUI) = %x, want prefix %x", mac, oui)
	}

	real := randomizer.Network.Hardware(nil, randomizer.HardwareMACRealOUI, randomizer.HardwareOptions{})
	if len(real) != 6 || real[0]&0x03 != 0 {
		t.Fatalf("Hardware(HardwareMACRealOUI) = %x, want universal unicast", real)
	}
}

func TestNetworkHardwareEUI64FromMAC(t *testing.T) {
	mac := net.HardwareAddr{0x00, 0x1a, 0x2b, 0x3c, 0x4d, 0x5e}
	eui := randomizer.Network.Hardware(nil, randomizer.HardwareEUI64FromMAC, randomizer.HardwareOptions{MAC: mac})
	if len(eui) != 8 {
		t.Fatalf("Hardware(HardwareEUI64FromMAC) length = %d, want 8", len(eui))
	}
	if eui[0] != mac[0]^0x02 || eui[1] != mac[1] || eui[2] != mac[2] {
		t.Fatalf("HardwareEUI64FromMAC OUI mismatch: got %x", eui[:3])
	}
	if eui[3] != 0xff || eui[4] != 0xfe || eui[5] != mac[3] || eui[6] != mac[4] || eui[7] != mac[5] {
		t.Fatalf("HardwareEUI64FromMAC suffix mismatch: got %x", eui[3:])
	}
	if got := randomizer.Network.Hardware(nil, randomizer.HardwareEUI64FromMAC, randomizer.HardwareOptions{
		MAC: net.HardwareAddr{0x00},
	}); got != nil {
		t.Fatalf("HardwareEUI64FromMAC invalid MAC = %x, want nil", got)
	}
}

func TestNetworkValueRanges(t *testing.T) {
	cases := []struct {
		kind     randomizer.ValueKind
		min, max uint64
	}{
		{randomizer.AnyPort, 0, 65535},
		{randomizer.PrivilegedPort, 1, 1023},
		{randomizer.RegisteredPort, 1024, 49151},
		{randomizer.EphemeralPort, 49152, 65535},
		{randomizer.VLANID, 0, 4095},
		{randomizer.VNI, 0, 0xffffff},
		{randomizer.FlowLabel, 0, 0xfffff},
		{randomizer.MPLSLabel, 0, 0xfffff},
	}
	for _, tc := range cases {
		for range 1000 {
			v := randomizer.Network.Value(tc.kind)
			if v < tc.min || v > tc.max {
				t.Fatalf("Value(%v) = %d, want [%d, %d]", tc.kind, v, tc.min, tc.max)
			}
		}
	}
}

func TestNetworkValueFixedProvider(t *testing.T) {
	const value fixedProvider = 0x0123456789abcdef
	previous := randomizer.SetProvider(value)
	defer randomizer.SetProvider(previous)

	if got, want := randomizer.Network.Value(randomizer.VLANID), uint64(0x12); got != want {
		t.Fatalf("Value(VLANID) = 0x%x, want 0x%x", got, want)
	}
	if got, want := randomizer.Network.Value(randomizer.ASN), uint64(value)>>32; got != want {
		t.Fatalf("Value(ASN) = 0x%x, want 0x%x", got, want)
	}
	if got, want := randomizer.Network.Value(randomizer.VNI), uint64(value)>>40; got != want {
		t.Fatalf("Value(VNI) = 0x%x, want 0x%x", got, want)
	}
	if got, want := randomizer.Network.Value(randomizer.FlowLabel), uint64(value)>>44; got != want {
		t.Fatalf("Value(FlowLabel) = 0x%x, want 0x%x", got, want)
	}
	if got, want := randomizer.Network.Value(randomizer.MPLSLabel), uint64(value)>>44; got != want {
		t.Fatalf("Value(MPLSLabel) = 0x%x, want 0x%x", got, want)
	}
	if got, want := randomizer.Network.Value(randomizer.IPv6InterfaceID), uint64(value)&^(uint64(0x03)<<56); got != want {
		t.Fatalf("Value(IPv6InterfaceID) = 0x%x, want 0x%x", got, want)
	}
}

func TestNetworkUUID(t *testing.T) {
	uuid := randomizer.Network.UUID()
	if uuid[6]>>4 != 0x4 {
		t.Fatalf("UUID version nibble = 0x%x, want 0x4", uuid[6]>>4)
	}
	if uuid[8]>>6 != 0x2 {
		t.Fatalf("UUID variant bits = 0x%x, want 0x2", uuid[8]>>6)
	}
}

func TestNetworkAppendUUID(t *testing.T) {
	dst := []byte("id:")
	out := randomizer.Network.AppendUUID(dst)
	if len(out) != 39 || string(out[:3]) != "id:" {
		t.Fatalf("AppendUUID length/prefix = %d/%q, want 39/id:", len(out), out[:3])
	}
	uuid := out[3:]
	for _, pos := range []int{8, 13, 18, 23} {
		if uuid[pos] != '-' {
			t.Fatalf("AppendUUID uuid[%d] = %q, want '-'", pos, uuid[pos])
		}
	}
	if uuid[14] != '4' {
		t.Fatalf("AppendUUID version char = %q, want '4'", uuid[14])
	}
	if !bytes.ContainsAny(uuid[19:20], "89ab") {
		t.Fatalf("AppendUUID variant char = %q, want one of 8, 9, a, b", uuid[19])
	}
}

func TestNetworkCIDR(t *testing.T) {
	cases := []struct {
		kind       randomizer.CIDRKind
		prefix     uint8
		ipLen      int
		maskBits   int
		maskLength int
	}{
		{randomizer.IPv4CIDR, 24, net.IPv4len, 24, 32},
		{randomizer.IPv4CIDR, 40, net.IPv4len, 32, 32},
		{randomizer.IPv6CIDR, 48, net.IPv6len, 48, 128},
		{randomizer.IPv6CIDR, 200, net.IPv6len, 128, 128},
	}
	for _, tc := range cases {
		ipNet := randomizer.Network.CIDR(tc.kind, tc.prefix)
		if ipNet == nil || len(ipNet.IP) != tc.ipLen {
			t.Fatalf("CIDR(%v, %d) = %v, want IP len %d", tc.kind, tc.prefix, ipNet, tc.ipLen)
		}
		ones, bits := ipNet.Mask.Size()
		if ones != tc.maskBits || bits != tc.maskLength {
			t.Fatalf("CIDR(%v, %d) mask = %d/%d, want %d/%d", tc.kind, tc.prefix, ones, bits, tc.maskBits, tc.maskLength)
		}
		for i, b := range ipNet.IP {
			if b&^ipNet.Mask[i] != 0 {
				t.Fatalf("CIDR(%v, %d) host bit set in byte %d", tc.kind, tc.prefix, i)
			}
		}
	}
}

func TestNetworkIPv6ULAPrefix(t *testing.T) {
	ipNet := randomizer.Network.CIDR(randomizer.IPv6ULAPrefix, 0)
	if ipNet == nil || len(ipNet.IP) != net.IPv6len || ipNet.IP[0] != 0xfd || ipNet.IP.To4() != nil {
		t.Fatalf("CIDR(IPv6ULAPrefix) = %v, want fd00::/8 IPv6", ipNet)
	}
	for i := 6; i < net.IPv6len; i++ {
		if ipNet.IP[i] != 0 {
			t.Fatalf("CIDR(IPv6ULAPrefix) host bit set in byte %d", i)
		}
	}
	ones, bits := ipNet.Mask.Size()
	if ones != 48 || bits != 128 {
		t.Fatalf("CIDR(IPv6ULAPrefix) mask = %d/%d, want 48/128", ones, bits)
	}
}

func TestNetworkIPInCIDR(t *testing.T) {
	_, ip4Net, _ := net.ParseCIDR("192.168.1.0/24")
	_, ip6Net, _ := net.ParseCIDR("2001:db8::/32")
	for range 100 {
		var buf [16]byte
		ip4 := randomizer.Network.IPInCIDR(buf[:0], ip4Net)
		if ip4 == nil || !ip4Net.Contains(ip4) || &ip4[0] != &buf[0] {
			t.Fatalf("IPInCIDR IPv4 = %v, want address inside %v using caller buffer", ip4, ip4Net)
		}
		ip6 := randomizer.Network.IPInCIDR(buf[:0], ip6Net)
		if ip6 == nil || !ip6Net.Contains(ip6) || &ip6[0] != &buf[0] {
			t.Fatalf("IPInCIDR IPv6 = %v, want address inside %v using caller buffer", ip6, ip6Net)
		}
	}
	if got := randomizer.Network.IPInCIDR(nil, nil); got != nil {
		t.Fatalf("IPInCIDR(nil) = %v, want nil", got)
	}
}

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
	if allocs := testing.AllocsPerRun(1000, func() {
		var buf [36]byte
		out := randomizer.Network.AppendUUID(buf[:0])
		if len(out) != 36 {
			t.Fatal("bad UUID text length")
		}
	}); allocs != 0 {
		t.Fatalf("AppendUUID with caller buffer allocs/op = %v, want 0", allocs)
	}
	if allocs := testing.AllocsPerRun(1000, func() {
		benchValue = randomizer.Network.Value(randomizer.ASN)
	}); allocs != 0 {
		t.Fatalf("Value allocs/op = %v, want 0", allocs)
	}
	if allocs := testing.AllocsPerRun(1000, func() {
		benchUUID = randomizer.Network.UUID()
	}); allocs != 0 {
		t.Fatalf("UUID allocs/op = %v, want 0", allocs)
	}
	if allocs := testing.AllocsPerRun(1000, func() {
		benchIPNet = randomizer.Network.CIDR(randomizer.IPv6ULAPrefix, 0)
	}); allocs > 1 {
		t.Fatalf("CIDR(IPv6ULAPrefix) allocs/op = %v, want <= 1", allocs)
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

func BenchmarkNetworkIPIPv6(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchIP = randomizer.Network.IP(nil, randomizer.IPv6Any)
	}
}

func BenchmarkNetworkHardwareMAC(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchMAC = randomizer.Network.Hardware(nil, randomizer.HardwareMAC, randomizer.HardwareOptions{
			Local:     true,
			Multicast: false,
		})
	}
}

func BenchmarkNetworkHardwareMACBuffered(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchMAC = randomizer.Network.Hardware(benchBuffer[:0], randomizer.HardwareMAC, randomizer.HardwareOptions{
			Local:     true,
			Multicast: false,
		})
	}
}

func BenchmarkNetworkValuePort(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchPort = randomizer.Network.Value(randomizer.AnyPort)
	}
}

func BenchmarkNetworkValueASN(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchValue = randomizer.Network.Value(randomizer.ASN)
	}
}

func BenchmarkNetworkUUID(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchUUID = randomizer.Network.UUID()
	}
}

func BenchmarkNetworkAppendUUID(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchUUIDB = randomizer.Network.AppendUUID(nil)
	}
}

func BenchmarkNetworkAppendUUIDBuffered(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchUUIDB = randomizer.Network.AppendUUID(benchBuffer[:0])
	}
}

func BenchmarkNetworkCIDR(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchIPNet = randomizer.Network.CIDR(randomizer.IPv6CIDR, 48)
	}
}
