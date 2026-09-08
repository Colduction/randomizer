package randomizer

import (
	"encoding/binary"
	"net"
	"net/netip"
)

// Addr returns a random address selected by kind as a [netip.Addr].
// It allocates nothing.
func (network) Addr(kind IPKind) netip.Addr {
	var buf [net.IPv6len]byte
	b := buf[:ipLen(kind)]
	fillIP(b, kind, currentProvider())
	if len(b) == net.IPv4len {
		return netip.AddrFrom4([net.IPv4len]byte(b))
	}
	return netip.AddrFrom16(buf)
}

// Prefix returns a random network prefix selected by kind as a
// [netip.Prefix]. It clamps IPv4 prefix lengths to [0, 32] and IPv6 prefix
// lengths to [0, 128] and allocates nothing.
func (network) Prefix(kind CIDRKind, bits uint8) netip.Prefix {
	bits, n := prefixLen(kind, bits)
	var buf [net.IPv6len]byte
	b := buf[:n]
	fillPrefix(b, kind, bits, currentProvider())
	if n == net.IPv4len {
		return netip.PrefixFrom(netip.AddrFrom4([net.IPv4len]byte(b)), int(bits))
	}
	return netip.PrefixFrom(netip.AddrFrom16(buf), int(bits))
}

// AddrIn returns a random address inside prefix. It returns the zero
// [netip.Addr] when prefix is invalid and allocates nothing.
func (network) AddrIn(prefix netip.Prefix) netip.Addr {
	if !prefix.IsValid() {
		return netip.Addr{}
	}
	var (
		p    = currentProvider()
		bits = prefix.Bits()
		addr = prefix.Addr()
	)
	if addr.Is4() {
		base := addr.As4()
		var out [net.IPv4len]byte
		putIPv4(out[:], hostBits32(binary.BigEndian.Uint32(base[:]), ^uint32(0)<<(32-bits), uint32(p.Sum64())))
		return netip.AddrFrom4(out)
	}
	base := addr.As16()
	var hiMask, loMask uint64
	if bits <= 64 {
		hiMask = ^uint64(0) << (64 - bits)
	} else {
		hiMask, loMask = ^uint64(0), ^uint64(0)<<(128-bits)
	}
	var out [net.IPv6len]byte
	binary.BigEndian.PutUint64(out[:], hostBits64(binary.BigEndian.Uint64(base[:]), hiMask, p.Sum64()))
	binary.BigEndian.PutUint64(out[8:], hostBits64(binary.BigEndian.Uint64(base[8:]), loMask, p.Sum64()))
	return netip.AddrFrom16(out)
}
