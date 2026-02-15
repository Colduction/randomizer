package randomizer

import "net"

type network struct{}

var Network network

func fillRandomBytes(out []byte, rng *wordRNG) {
	i := 0
	for ; i+8 <= len(out); i += 8 {
		x := rng.next64()
		out[i+0] = byte(x >> 0)
		out[i+1] = byte(x >> 8)
		out[i+2] = byte(x >> 16)
		out[i+3] = byte(x >> 24)
		out[i+4] = byte(x >> 32)
		out[i+5] = byte(x >> 40)
		out[i+6] = byte(x >> 48)
		out[i+7] = byte(x >> 56)
	}
	if i < len(out) {
		x := rng.next64()
		for j := i; j < len(out); j++ {
			out[j] = byte(x)
			x >>= 8
		}
	}
}

// ref: https://datatracker.ietf.org/doc/html/rfc3513#section-2.5
type UnicastType uint8

const (
	GlobalType UnicastType = iota + 1
	LinkLocalType
	SiteLocalType
	UniqueLocalType
	PrivateType UnicastType = UniqueLocalType
)

// ref: https://datatracker.ietf.org/doc/html/rfc3513#section-2.7
type MulticastScope uint8

const (
	InterfaceLocalScope MulticastScope = 0x1
	LinkLocalScope      MulticastScope = 0x2
	AdminLocalScope     MulticastScope = 0x4
	SiteLocalScope      MulticastScope = 0x5
	OrgLocalScope       MulticastScope = 0x8
	GlobalScope         MulticastScope = 0xE
)

// IPv4Addr generates a random IPv4 address by creating a 4-byte IP
// using a hash-based approach for randomness, ensuring a unique address.
func (network) IPv4Addr() net.IP {
	b := make(net.IP, net.IPv4len)
	rng := newWordRNG()
	fillRandomBytes(b, &rng)
	return b
}

// IPv6Addr generates a random IPv6 address by creating a 16-byte IP
// through a hash-based approach, ensuring a unique 128-bit address.
func (network) IPv6Addr() net.IP {
	b := make(net.IP, net.IPv6len)
	rng := newWordRNG()
	fillRandomBytes(b, &rng)
	return b
}

// MACAddr generates a random MAC address with configurable local and multicast
// bits. The U/L bit controls whether the address is locally administered, and the
// I/G bit controls whether the address is intended for multicast traffic.
func (network) MACAddr(local, multicast bool) net.HardwareAddr {
	b := make(net.HardwareAddr, 6)
	rng := newWordRNG()
	fillRandomBytes(b, &rng)
	// Set the U/L bit in the first byte
	if local {
		b[0] = b[0] | 0x02
	} else {
		b[0] = b[0] &^ 0x02
	}
	// Set the I/G bit in the first byte
	if multicast {
		b[0] = b[0] | 0x01
	} else {
		b[0] = b[0] &^ 0x01
	}
	return net.HardwareAddr(b)
}

// IPv6UnicastAddr generates a random IPv6 unicast address of a specified
// unicast type by configuring address prefixes.
func (network) IPv6UnicastAddr(unicastType UnicastType) net.IP {
	b := make(net.IP, net.IPv6len)
	rng := newWordRNG()
	fillRandomBytes(b, &rng)
	switch unicastType {
	case GlobalType:
		b[0] = (b[0] & 0x1F) | 0x20
	case LinkLocalType:
		b[0] = 0xFE
		b[1] = (b[1] & 0x3F) | 0x80
	case SiteLocalType:
		b[0] = 0xFE
		b[1] = (b[1] & 0x3F) | 0xC0
	case UniqueLocalType:
		b[0] = 0xFD
	}
	return b
}

// IPv6MulticastAddr generates a random IPv6 multicast address with a
// specified multicast scope, setting the appropriate prefix and scope bits.
func (network) IPv6MulticastAddr(scope MulticastScope) net.IP {
	b := make(net.IP, net.IPv6len)
	rng := newWordRNG()
	fillRandomBytes(b, &rng)
	b[0] = 0xFF
	b[1] = uint8(scope) & 0x0F
	return b
}
