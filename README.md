# randomizer-go

[![Go Reference](https://pkg.go.dev/badge/github.com/colduction/randomizer-go.svg)](https://pkg.go.dev/github.com/colduction/randomizer-go)
![GitHub License](https://img.shields.io/github/license/Colduction/randomizer-go)

**randomizer-go** is a fast, allocation-aware, goroutine-safe random data generation library for Go.
It covers numbers, byte/string alphabets, network values and addresses, link-layer and Bluetooth identifiers, mobile equipment identifiers, and configurable random providers.

## Features

- **Numbers**: signed and unsigned integers, bounded intervals, and floats in `[0, 1)`
- **Byte fill**: `Read(p []byte)` fills caller-owned buffers with zero allocations
- **Strings and bytes**: built-in alphabets, custom dictionaries, append APIs, and no adjacent duplicate bytes
- **Network**: enum-based API for IPv4, IPv6, `netip` values, MAC, EUI-64, IEEE 802c local addresses, Bluetooth device addresses, ports, VLAN IDs, ASNs, VNI, flow labels, MPLS labels, CIDRs
- **Telecom**: IMEI, IMEISV, ICCID, and MEID with correct check digits
- **Providers**: lock-free per-thread ChaCha8 default, `math/rand`, `math/rand/v2`, `crypto/rand`, or custom providers
- **Hash pool**: optional `maphash.Hash` reuse and a SplitMix64 provider through `NewHashPool`

## Requirements

- Go **1.27** or later

## Installation

```bash
go get github.com/colduction/randomizer-go@latest
```

## Quick Start

```go
package main

import (
	"fmt"
	"net"

	"github.com/colduction/randomizer-go"
)

func main() {
	fmt.Println(randomizer.Int[int64]())
	fmt.Println(randomizer.IntInterval(int64(1), int64(100)))
	fmt.Println(randomizer.Float64())

	fmt.Println(randomizer.Word.String(randomizer.HexLowerAlphabet, 16))
	fmt.Println(randomizer.Word.String(randomizer.AlphaNumericAlphabet, 24))
	fmt.Println(randomizer.Word.Custom("ABCDEFGH", 12))

	fmt.Println(randomizer.Network.IP(nil, randomizer.IPv4Public))
	fmt.Println(randomizer.Network.Addr(randomizer.IPv6Public))

	var macBuf [8]byte
	mac := randomizer.Network.Hardware(macBuf[:0], randomizer.HardwareMACRealOUI, randomizer.HardwareOptions{})
	fmt.Println(net.HardwareAddr(mac))

	fmt.Println(randomizer.Telecom.String(randomizer.IMEI, randomizer.TelecomOptions{}))

	var raw [32]byte
	_, _ = randomizer.Read(raw[:])
}
```

## Numbers

With `DefaultProvider`, number functions are zero-allocation and safe for concurrent use.

```go
n := randomizer.Int[int64]()
u := randomizer.Uint[uint64]()
bounded := randomizer.IntInterval(int64(-50), int64(50))
ubounded := randomizer.UintInterval(uint64(1), uint64(1000))
f32 := randomizer.Float32()
f64 := randomizer.Float64()
```

`IntInterval` and `UintInterval` return values in `[min, max)` using Lemire's nearly divisionless bounded sampling: one multiplication in the common case, a division only with probability below `span / 2^64`. Equal bounds return the bound. Swapped bounds are corrected.

## Words

`Word.String` and `Word.Bytes` allocate the returned output. `Word.Append` writes into caller-owned capacity and can be zero-allocation. All word generators avoid adjacent duplicate bytes.

```go
s := randomizer.Word.String(randomizer.DecimalAlphabet, 12)
b := randomizer.Word.Bytes(randomizer.Base64URLAlphabet, 32)

buf := make([]byte, 0, 32)
buf = randomizer.Word.Append(buf, randomizer.AlphaNumericAlphabet, 32)
```

Built-in alphabets:

| Constant               | Bytes                                    |
| ---------------------- | ---------------------------------------- |
| `DecimalAlphabet`      | `0-9`                                    |
| `HexLowerAlphabet`     | `0-9a-f`                                 |
| `HexUpperAlphabet`     | `0-9A-F`                                 |
| `OctalAlphabet`        | `0-7`                                    |
| `LowerAlphabet`        | `a-z`                                    |
| `UpperAlphabet`        | `A-Z`                                    |
| `AlphaAlphabet`        | `a-zA-Z`                                 |
| `AlphaNumericAlphabet` | `0-9a-zA-Z`                              |
| `Base32Alphabet`       | RFC 4648 base32 without padding          |
| `Base64URLAlphabet`    | RFC 4648 URL-safe base64 without padding |

Symbols are drawn with the batched ranged integer method of Brackett-Rozinsky and Lemire: one 64-bit random word yields up to 19 decimal digits, 15 hexadecimal digits, or 10 alphanumeric symbols through chained multiplications, with a single rejection test per word and no divisions. After the first symbol, each position is drawn from the other `n-1` positions and shifted past the previous one, so unique dictionaries never retry a symbol.

Custom dictionaries sample bytes directly. Repeated dictionary bytes increase weight.

```go
s := randomizer.Word.Custom("AABCXYZ9", 16)
s2 := randomizer.Word.CustomFromBytes([]byte("AABCXYZ9"), 16)
b := randomizer.Word.CustomBytes([]byte("AABCXYZ9"), 16)
b2 := randomizer.Word.CustomBytesFromString("AABCXYZ9", 16)

dst := []byte("id:")
dst = randomizer.Word.AppendCustom(dst, "AABCXYZ9", 16)
dst = randomizer.Word.AppendCustomFromBytes(dst[:3], []byte("AABCXYZ9"), 16)
```

Custom string functions return `""` when `length <= 0`, dictionary is empty, or length is greater than 1 and the dictionary cannot avoid adjacent duplicates. Custom byte functions return `nil` for those invalid cases. Append variants return `dst` unchanged.

## Network

Network APIs use small kind enums instead of many single-purpose methods. Slice results allocate only when `dst` capacity is too small; the `netip` methods never allocate.

### IP

```go
ip4 := randomizer.Network.IP(nil, randomizer.IPv4Any)
public4 := randomizer.Network.IP(nil, randomizer.IPv4Public)

var ipBuf [16]byte
ip6 := randomizer.Network.IP(ipBuf[:0], randomizer.IPv6Public)
mc6 := randomizer.Network.IP(ipBuf[:0], randomizer.IPv6LinkLocalMulticast)

addr := randomizer.Network.Addr(randomizer.IPv6UniqueLocal) // netip.Addr, zero allocations
```

`IPKind` values:

| Family         | Constants                                                                                                                                                                               |
| -------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| IPv4           | `IPv4Any`, `IPv4Private`, `IPv4LinkLocal`, `IPv4Multicast`, `IPv4Public`, `IPv4Loopback`, `IPv4CGNAT`, `IPv4Documentation`                                                              |
| IPv6 unicast   | `IPv6Any`, `IPv6Global`, `IPv6Public`, `IPv6LinkLocal`, `IPv6SiteLocal`, `IPv6UniqueLocal`, `IPv6Private`, `IPv6Documentation`                                                          |
| IPv6 multicast | `IPv6InterfaceLocalMulticast`, `IPv6LinkLocalMulticast`, `IPv6RealmLocalMulticast`, `IPv6AdminLocalMulticast`, `IPv6SiteLocalMulticast`, `IPv6OrgLocalMulticast`, `IPv6GlobalMulticast` |

Validity rules applied by the generators:

- `IPv4Public` excludes every IANA special-purpose, private, and bogon range (RFC 6890).
- `IPv4LinkLocal` stays within `169.254.1.0`–`169.254.254.255` (RFC 3927 reserves the first and last /24).
- `IPv4Documentation` picks one of the three RFC 5737 blocks with equal probability; `IPv4CGNAT` is `100.64.0.0/10` (RFC 6598).
- `IPv6LinkLocal` is a strict `fe80::/64` prefix with a random interface identifier (RFC 4291 §2.5.6).
- `IPv6Public` is `2000::/3` minus the IANA IPv6 Special-Purpose Address Registry blocks (`2001::/23`, `2001:db8::/32`, `2002::/16`, `2620:4f:8000::/48`, `3fff::/20`, `5f00::/16`) and never uses the subnet-router or RFC 2526 reserved anycast interface identifiers.
- IPv6 multicast kinds set the transient flag (`T = 1`), the form required for non-IANA groups, so the second byte is `0x1s` where `s` is the scope.

`IPv6EUI64` builds a SLAAC-style address: a random prefix of the selected kind (`IPv6LinkLocal`, `IPv6Global`, `IPv6Public`, `IPv6UniqueLocal`, `IPv6SiteLocal`, or `IPv6Documentation`) with the RFC 4291 modified EUI-64 of a 6-byte MAC address or an 8-byte EUI-64.

```go
mac := net.HardwareAddr{0x00, 0x1a, 0x2b, 0x3c, 0x4d, 0x5e}
ll := randomizer.Network.IPv6EUI64(nil, randomizer.IPv6LinkLocal, mac) // fe80::21a:2bff:fe3c:4d5e
```

### Hardware

```go
mac := randomizer.Network.Hardware(nil, randomizer.HardwareMAC, randomizer.HardwareOptions{Local: true})
ouiMAC := randomizer.Network.Hardware(nil, randomizer.HardwareMACOUI, randomizer.HardwareOptions{OUI: [3]byte{0x3c, 0x22, 0xfb}})
realOUI := randomizer.Network.Hardware(nil, randomizer.HardwareMACRealOUI, randomizer.HardwareOptions{})

// IEEE MA-M (28-bit) assignment prefix.
mam := randomizer.Network.Hardware(nil, randomizer.HardwareMACPrefix, randomizer.HardwareOptions{
	Prefix: []byte{0x70, 0xb3, 0xd5, 0xa0}, PrefixBits: 28,
})

// IEEE 802c Structured Local Address Plan quadrants.
aai := randomizer.Network.Hardware(nil, randomizer.HardwareMACAAI, randomizer.HardwareOptions{})
eli := randomizer.Network.Hardware(nil, randomizer.HardwareMACELI, randomizer.HardwareOptions{CID: [3]byte{0x0a, 0x1b, 0x2c}})
sai := randomizer.Network.Hardware(nil, randomizer.HardwareMACSAI, randomizer.HardwareOptions{})

eui64 := randomizer.Network.Hardware(nil, randomizer.HardwareEUI64RealOUI, randomizer.HardwareOptions{})
fromMAC := randomizer.Network.Hardware(nil, randomizer.HardwareEUI64FromMAC, randomizer.HardwareOptions{MAC: mac})
iid := randomizer.Network.Hardware(nil, randomizer.HardwareModifiedEUI64FromMAC, randomizer.HardwareOptions{MAC: mac})

// Bluetooth device addresses.
irk := []byte("0123456789abcdef") // 16-byte Identity Resolving Key
public := randomizer.Network.Hardware(nil, randomizer.HardwareBluetoothPublic, randomizer.HardwareOptions{})
static := randomizer.Network.Hardware(nil, randomizer.HardwareBluetoothStatic, randomizer.HardwareOptions{})
nrpa := randomizer.Network.Hardware(nil, randomizer.HardwareBluetoothNRPA, randomizer.HardwareOptions{})
rpa := randomizer.Network.Hardware(nil, randomizer.HardwareBluetoothRPA, randomizer.HardwareOptions{IRK: irk})
```

`HardwareKind` values:

| Constant                       | Result                                                                                                                                |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------- |
| `HardwareMAC`                  | 6 random bytes; `Local` and `Multicast` set the U/L and I/G bits                                                                      |
| `HardwareMACOUI`               | `OUI` followed by 3 random bytes                                                                                                      |
| `HardwareMACRealOUI`           | An IEEE-registered OUI of a well-known vendor and a random extension outside the Bluetooth inquiry LAP range                          |
| `HardwareMACPrefix`            | The first `PrefixBits` (1–48) bits of `Prefix`, then random bits; covers MA-L, MA-M, MA-S, and CID assignments                        |
| `HardwareMACAAI`               | IEEE 802c Administratively Assigned Identifier (local unicast, second hex digit `2`)                                                  |
| `HardwareMACELI`               | IEEE 802c Extended Local Identifier from the 24-bit Company ID in `CID` (second hex digit `A`); `nil` unless `CID` has that form      |
| `HardwareMACSAI`               | IEEE 802c Standard Assigned Identifier (local unicast, second hex digit `E`)                                                          |
| `HardwareEUI64`                | 8 random bytes with the U/L bit set and the I/G bit clear                                                                             |
| `HardwareEUI64RealOUI`         | An IEEE-registered OUI with a 40-bit extension outside the `ff:fe` and `ff:ff` encapsulation values                                   |
| `HardwareEUI64FromMAC`         | The IEEE EUI-64 encapsulation of `MAC`: OUI, `ff fe`, extension, U/L bit kept                                                         |
| `HardwareModifiedEUI64FromMAC` | The RFC 4291 modified EUI-64 of `MAC`, the IPv6 interface identifier form with the U/L bit inverted                                   |
| `HardwareBluetoothPublic`      | Same as `HardwareMACRealOUI`                                                                                                          |
| `HardwareBluetoothStatic`      | Static random address: top two bits `11`, 46 random bits that are neither all zero nor all one                                        |
| `HardwareBluetoothNRPA`        | Non-resolvable private address: top two bits `00`, 46 random bits that are neither all zero nor all one                               |
| `HardwareBluetoothRPA`         | Resolvable private address: 22-bit `prand` under top bits `01`, then `hash = ah(IRK, prand)` (AES-128) when `IRK` is set, else random |

Bluetooth addresses are returned most-significant byte first, the order used in `XX:XX:XX:XX:XX:XX` notation. Resolvable private addresses with an `IRK` are computed with `crypto/aes`; the cipher for the most recent key is cached, so repeated calls with one key allocate nothing.

### Values

```go
port := randomizer.Network.Value(randomizer.RegisteredPort)
vlan := randomizer.Network.Value(randomizer.VLANIDUnreserved)
asn := randomizer.Network.Value(randomizer.ASNPublic)
private := randomizer.Network.Value(randomizer.ASNPrivate16)
vni := randomizer.Network.Value(randomizer.VNI)
flow := randomizer.Network.Value(randomizer.FlowLabel)
mpls := randomizer.Network.Value(randomizer.MPLSLabelUnreserved)
iid := randomizer.Network.Value(randomizer.IPv6InterfaceID)
```

`ValueKind` values:

| Constant              | Range                                                                                                                               |
| --------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `AnyPort`             | `[0, 65535]`                                                                                                                        |
| `PrivilegedPort`      | `[1, 1023]`                                                                                                                         |
| `RegisteredPort`      | `[1024, 49151]`                                                                                                                     |
| `EphemeralPort`       | `[49152, 65535]`                                                                                                                    |
| `VLANID`              | `[0, 4095]`                                                                                                                         |
| `VLANIDUnreserved`    | `[1, 4094]` (802.1Q reserves 0 and 4095)                                                                                            |
| `ASN`                 | `[0, 4294967295]`                                                                                                                   |
| `ASNPublic`           | Assignable 32-bit ASN: excludes 0, 23456, 64496–131071, and 4200000000–4294967295 (RFC 5398, 6793, 6996, 7300, 7607, IANA reserved) |
| `ASNPublic16`         | `[1, 64495]` other than 23456                                                                                                       |
| `ASNPrivate16`        | `[64512, 65534]` (RFC 6996)                                                                                                         |
| `ASNPrivate32`        | `[4200000000, 4294967294]` (RFC 6996)                                                                                               |
| `VNI`                 | `[0, 16777215]`                                                                                                                     |
| `FlowLabel`           | `[0, 1048575]`                                                                                                                      |
| `MPLSLabel`           | `[0, 1048575]`                                                                                                                      |
| `MPLSLabelUnreserved` | `[16, 1048575]` (RFC 7274 special-purpose labels 0–15 excluded)                                                                     |
| `IPv6InterfaceID`     | 64-bit interface identifier with U/L and I/G bits clear                                                                             |

### CIDR and Hosts

```go
net4 := randomizer.Network.CIDR(randomizer.IPv4CIDR, 24)
net6 := randomizer.Network.CIDR(randomizer.IPv6CIDR, 48)
ula := randomizer.Network.CIDR(randomizer.IPv6ULAPrefix, 0)

host4 := randomizer.Network.IPInCIDR(nil, net4)

var hostBuf [16]byte
host6 := randomizer.Network.IPInCIDR(hostBuf[:0], net6)

// netip equivalents, zero allocations.
prefix := randomizer.Network.Prefix(randomizer.IPv6CIDR, 56)
host := randomizer.Network.AddrIn(prefix)
```

`CIDR` and `Prefix` clamp IPv4 prefix lengths to `[0, 32]` and IPv6 prefix lengths to `[0, 128]`. `IPv6ULAPrefix` always returns a random RFC 4193 `fd00::/8` `/48` prefix. `IPInCIDR` and `AddrIn` draw host bits uniformly even when the network address carries host bits; `AddrIn` returns the zero `netip.Addr` for an invalid prefix.

## Telecom

`Telecom.String` allocates; `Telecom.Append` writes into caller-owned capacity. Invalid options make `String` return `""` and `Append` return `dst` unchanged.

```go
imei := randomizer.Telecom.String(randomizer.IMEI, randomizer.TelecomOptions{})
withTAC := randomizer.Telecom.String(randomizer.IMEI, randomizer.TelecomOptions{TAC: "35123456"})
test := randomizer.Telecom.String(randomizer.IMEI, randomizer.TelecomOptions{Test: true})
imeisv := randomizer.Telecom.String(randomizer.IMEISV, randomizer.TelecomOptions{SVN: "07"})
iccid := randomizer.Telecom.String(randomizer.ICCID, randomizer.TelecomOptions{IIN: "8901", Length: 20})
meid := randomizer.Telecom.String(randomizer.MEID, randomizer.TelecomOptions{})
```

| Kind     | Result                                                                                                                               |
| -------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `IMEI`   | 15 digits: 8-digit TAC (a GSMA TS.06 Reporting Body Identifier, or `00` with `Test`), 6-digit serial, Luhn check digit               |
| `IMEISV` | 16 digits: the 14 IMEI digits and a 2-digit software version in `[00, 98]` (3GPP TS 23.003 reserves 99)                              |
| `ICCID`  | ITU-T E.118: `89`, an assigned E.164 country code (or `IIN`), account digits, Luhn check digit; 19 digits by default, `Length` 18–20 |
| `MEID`   | 14 uppercase hexadecimal digits with a regional code in `[A0, FF]`                                                                   |

## Providers

All package generators read from the active `Provider`. `SetProvider` atomically swaps it and returns the prior provider. Passing `nil` restores `DefaultProvider`.

```go
type Provider interface {
	Read([]byte) (int, error)
	Sum([]byte) []byte
	Sum32() uint32
	Sum64() uint64
}
```

### DefaultProvider

`DefaultProvider` draws from the Go runtime's per-thread ChaCha8 generator, the source behind `math/rand/v2` (see _Secure Randomness in Go 1.22_ and the C2SP `chacha8rand` specification). Every OS thread owns its own state, so it never locks and throughput scales linearly with goroutines. It is seeded by the operating system and cannot be reseeded; use `NewUint64Provider` for reproducible sequences.

```go
var buf [256]byte
_, _ = randomizer.Read(buf[:])

n64 := randomizer.DefaultProvider.Sum64()
n32 := randomizer.DefaultProvider.Sum32()
dst := randomizer.DefaultProvider.Sum(existing)
```

### NewUint64Provider

`NewUint64Provider(source interface{ Uint64() uint64 }) Provider` seeds a SplitMix64 counter by calling `source.Uint64()` once. Equal seeds give equal sequences. The counter is one shared atomic word, so heavy concurrent use serializes on a single cache line; prefer `DefaultProvider` when reproducibility is not needed.

```go
import (
	mrand "math/rand"
	randv2 "math/rand/v2"
)

p1 := randomizer.NewUint64Provider(mrand.New(mrand.NewSource(42)))
p2 := randomizer.NewUint64Provider(randv2.New(randv2.NewPCG(1, 2)))

chacha := randv2.NewChaCha8([32]byte{})
p3 := randomizer.NewUint64Provider(chacha)
```

### NewReaderProvider

`NewReaderProvider(reader io.Reader) Provider` reads every value from the supplied reader.

```go
import cryptorand "crypto/rand"

p := randomizer.NewReaderProvider(cryptorand.Reader)
previous := randomizer.SetProvider(p)
defer randomizer.SetProvider(previous)
```

This works with `crypto/rand.Reader` and readers such as `math/rand/v2.ChaCha8`. The reader must be safe for concurrent use when installed as the package provider. `crypto/rand.Reader` is safe. A shared `ChaCha8` value should be protected or used through `NewUint64Provider`.

`Read` returns errors. `Sum`, `Sum32`, and `Sum64` panic on reader errors because their signatures cannot return errors.

## HashPool

`NewHashPool(size int) *hashPool` creates an independent SplitMix64 provider and a `sync.Pool` of `maphash.Hash` values. It returns `nil` for `size <= 0`; nil receiver methods are safe.

```go
pool := randomizer.NewHashPool(16)

h := pool.Get()
h.WriteString("seed-data")
fmt.Printf("%016x\n", h.Sum64())
pool.Put(h)
```

## OUI Table

`HardwareMACRealOUI`, `HardwareEUI64RealOUI`, and `HardwareBluetoothPublic` pick from `oui_table.go`, a generated list of universal unicast OUIs of well-known vendors taken from the IEEE MA-L registry. Regenerate it with:

```bash
go generate ./...
```

The generator (`internal/ouigen`) downloads `https://standards-oui.ieee.org/oui/oui.csv`, keeps the lowest three assignments of every allowlisted vendor, and fails when a vendor no longer matches the registry so renames are noticed.

## Thread Safety

Package generators and provider replacement are safe for concurrent use. `DefaultProvider` uses per-thread runtime state and never contends. Providers from `NewUint64Provider` and `NewHashPool` share one atomic counter. Providers installed with `NewReaderProvider` are only as concurrent-safe as their wrapped reader.

## Breaking Changes

- `DefaultProvider` is the runtime ChaCha8 provider, no longer a `*hashPool`; type assertions on it break. `NewHashPool` still returns the SplitMix64 provider.
- `HardwareEUI64FromMAC` now returns the IEEE EUI-64 (U/L bit kept). The previous output, the IPv6 interface identifier form, is `HardwareModifiedEUI64FromMAC`.
- `IPv6LinkLocal` produces `fe80::/64` addresses; the 54 bits after the prefix are zero as RFC 4291 requires.
- IPv6 multicast kinds set the transient flag: the second byte is `0x11`, `0x12`, ... instead of `0x01`, `0x02`, ...
- `IPv4LinkLocal` no longer produces `169.254.0.x` or `169.254.255.x`.
- Word generators changed their sampling algorithm; sequences produced under a seeded provider differ from earlier versions.
- `UUID` and `AppendUUID` were removed.

## References

- Daniel Lemire, _Fast Random Integer Generation in an Interval_, ACM Transactions on Modeling and Computer Simulation 29(1), 2019.
- Nevin Brackett-Rozinsky and Daniel Lemire, _Batched Ranged Random Integer Generation_, Software: Practice and Experience 55(1), 2024.
- Russ Cox and Filippo Valsorda, _Secure Randomness in Go 1.22_, and the C2SP `chacha8rand` specification.
- Guy L. Steele Jr., Doug Lea, and Christine H. Flood, _Fast Splittable Pseudorandom Number Generators_, OOPSLA 2014 (SplitMix64).
- IEEE Std 802c-2017 and RFC 8948 (SLAP quadrants); RFC 4291, RFC 9542 (EUI-64); Bluetooth Core Specification 5.4 Vol 6 Part B §1.3 and Vol 3 Part H §2.2.2 (device addresses, `ah`).
- RFC 6890, RFC 5737, RFC 6598, RFC 3927, RFC 2526, and the IANA IPv6 Special-Purpose Address Registry; RFC 7274 (MPLS); RFC 5398, 6793, 6996, 7300, 7607 (ASN).
- GSMA TS.06 (IMEI), 3GPP TS 23.003 (IMEISV), ITU-T E.118 (ICCID), 3GPP2 S.R0048 (MEID).

## License

This project is licensed under the terms of the [MIT License](LICENSE).
