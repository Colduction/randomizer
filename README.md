# randomizer-go

[![Go Reference](https://pkg.go.dev/badge/github.com/colduction/randomizer-go.svg)](https://pkg.go.dev/github.com/colduction/randomizer-go)
![GitHub License](https://img.shields.io/github/license/Colduction/randomizer-go)

**randomizer-go** is a fast, allocation-aware, goroutine-safe random data generation library for Go.
It covers numbers, byte/string alphabets, network values, network addresses, UUIDs, and configurable random providers.

## Features

- **Numbers**: signed and unsigned integers, bounded intervals, and floats in `[0, 1)`
- **Byte fill**: `Read(p []byte)` fills caller-owned buffers with zero allocations
- **Strings and bytes**: built-in alphabets, custom dictionaries, append APIs, and no adjacent duplicate bytes
- **Network**: compact enum-based API for IPv4, IPv6, MAC, EUI-64, ports, VLAN IDs, ASNs, VNI, flow labels, MPLS labels, CIDRs, and UUIDs
- **Providers**: default lock-free provider, `math/rand`, `math/rand/v2`, `crypto/rand`, `math/rand/v2.ChaCha8`, or custom providers
- **Hash pool**: optional `maphash.Hash` reuse through `NewHashPool`

## Requirements

- Go **1.26** or later

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

	ip := randomizer.Network.IP(nil, randomizer.IPv4Public)
	fmt.Println(ip)

	var macBuf [8]byte
	mac := randomizer.Network.Hardware(macBuf[:0], randomizer.HardwareMAC, randomizer.HardwareOptions{
		Local: true,
	})
	fmt.Println(net.HardwareAddr(mac))

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

`IntInterval` and `UintInterval` return values in `[min, max)`. Equal bounds return the bound. Swapped bounds are corrected.

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

Network APIs use small kind enums instead of many single-purpose methods. Slice results allocate only when `dst` capacity is too small.

### IP

```go
ip4 := randomizer.Network.IP(nil, randomizer.IPv4Any)
public4 := randomizer.Network.IP(nil, randomizer.IPv4Public)

var ipBuf [16]byte
ip6 := randomizer.Network.IP(ipBuf[:0], randomizer.IPv6Global)
mc6 := randomizer.Network.IP(ipBuf[:0], randomizer.IPv6LinkLocalMulticast)
```

`IPKind` values:

| Family         | Constants                                                                                                                                                    |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| IPv4           | `IPv4Any`, `IPv4Private`, `IPv4LinkLocal`, `IPv4Multicast`, `IPv4Public`                                                                                     |
| IPv6 unicast   | `IPv6Any`, `IPv6Global`, `IPv6LinkLocal`, `IPv6SiteLocal`, `IPv6UniqueLocal`, `IPv6Private`                                                                  |
| IPv6 multicast | `IPv6InterfaceLocalMulticast`, `IPv6LinkLocalMulticast`, `IPv6AdminLocalMulticast`, `IPv6SiteLocalMulticast`, `IPv6OrgLocalMulticast`, `IPv6GlobalMulticast` |

### Hardware

```go
mac := randomizer.Network.Hardware(nil, randomizer.HardwareMAC, randomizer.HardwareOptions{
	Local:     true,
	Multicast: false,
})

ouiMAC := randomizer.Network.Hardware(nil, randomizer.HardwareMACOUI, randomizer.HardwareOptions{
	OUI: [3]byte{0x3c, 0x22, 0xfb},
})

realOUI := randomizer.Network.Hardware(nil, randomizer.HardwareMACRealOUI, randomizer.HardwareOptions{})
eui64 := randomizer.Network.Hardware(nil, randomizer.HardwareEUI64, randomizer.HardwareOptions{})
fromMAC := randomizer.Network.Hardware(nil, randomizer.HardwareEUI64FromMAC, randomizer.HardwareOptions{MAC: mac})
```

`HardwareKind` values: `HardwareMAC`, `HardwareMACOUI`, `HardwareMACRealOUI`, `HardwareEUI64`, `HardwareEUI64FromMAC`.

### Values

```go
port := randomizer.Network.Value(randomizer.RegisteredPort)
vlan := randomizer.Network.Value(randomizer.VLANID)
asn := randomizer.Network.Value(randomizer.ASN)
vni := randomizer.Network.Value(randomizer.VNI)
flow := randomizer.Network.Value(randomizer.FlowLabel)
mpls := randomizer.Network.Value(randomizer.MPLSLabel)
iid := randomizer.Network.Value(randomizer.IPv6InterfaceID)
```

`ValueKind` values:

| Constant          | Range                                                   |
| ----------------- | ------------------------------------------------------- |
| `AnyPort`         | `[0, 65535]`                                            |
| `PrivilegedPort`  | `[1, 1023]`                                             |
| `RegisteredPort`  | `[1024, 49151]`                                         |
| `EphemeralPort`   | `[49152, 65535]`                                        |
| `VLANID`          | `[0, 4095]`                                             |
| `ASN`             | `[0, 4294967295]`                                       |
| `VNI`             | `[0, 16777215]`                                         |
| `FlowLabel`       | `[0, 1048575]`                                          |
| `MPLSLabel`       | `[0, 1048575]`                                          |
| `IPv6InterfaceID` | 64-bit interface identifier with U/L and I/G bits clear |

### CIDR and Hosts

```go
net4 := randomizer.Network.CIDR(randomizer.IPv4CIDR, 24)
net6 := randomizer.Network.CIDR(randomizer.IPv6CIDR, 48)
ula := randomizer.Network.CIDR(randomizer.IPv6ULAPrefix, 0)

host4 := randomizer.Network.IPInCIDR(nil, net4)

var hostBuf [16]byte
host6 := randomizer.Network.IPInCIDR(hostBuf[:0], net6)
```

`CIDR` clamps IPv4 prefix lengths to `[0, 32]` and IPv6 prefix lengths to `[0, 128]`. `IPv6ULAPrefix` always returns a random RFC 4193 `fd00::/8` `/48` prefix.

### UUID

```go
uuid := randomizer.Network.UUID()

buf := make([]byte, 0, 36)
buf = randomizer.Network.AppendUUID(buf)
```

`UUID` returns a `[16]byte` RFC 4122 version-4 UUID. `AppendUUID` appends lowercase standard text form and can be zero-allocation with enough capacity.

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

`DefaultProvider` is lock-free for number and byte generation. It also supports `io.Reader` style byte fill:

```go
var buf [256]byte
_, _ = randomizer.Read(buf[:])

n64 := randomizer.DefaultProvider.Sum64()
n32 := randomizer.DefaultProvider.Sum32()
dst := randomizer.DefaultProvider.Sum(existing)
```

### NewUint64Provider

`NewUint64Provider(source interface{ Uint64() uint64 }) Provider` seeds a lock-free provider by calling `source.Uint64()` once.

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

Use this constructor when high concurrent access matters.

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

`NewHashPool(size int) *hashPool` creates an independent provider and a `sync.Pool` of `maphash.Hash` values. It returns `nil` for `size <= 0`; nil receiver methods are safe.

```go
pool := randomizer.NewHashPool(16)

h := pool.Get()
h.WriteString("seed-data")
fmt.Printf("%016x\n", h.Sum64())
pool.Put(h)
```

## Performance

Representative benchmarks on AMD Ryzen 9 7950X, Go 1.26, `GOMAXPROCS=32`:

| Benchmark                                | Time/op  | Allocs/op |
| ---------------------------------------- | -------- | --------- |
| `Int[int64]`                             | ~4.9 ns  | 0         |
| `IntInterval`                            | ~7.8 ns  | 0         |
| `Float64`                                | ~6.3 ns  | 0         |
| `Read(256 bytes)`                        | ~60 ns   | 0         |
| `Word.String(DecimalAlphabet, 256)`      | ~800 ns  | 1         |
| `Word.String(Base64URLAlphabet, 256)`    | ~380 ns  | 1         |
| `Word.Append(AlphaNumericAlphabet, 256)` | ~490 ns  | 0         |
| `Network.IP(nil, IPv4Any)`               | ~10 ns   | 1         |
| `Network.IP(buf, IPv4Any)`               | ~6.5 ns  | 0         |
| `Network.Hardware(nil, HardwareMAC)`     | ~14.5 ns | 1         |
| `Network.Hardware(buf, HardwareMAC)`     | ~7.2 ns  | 0         |
| `Network.Value(AnyPort)`                 | ~6.1 ns  | 0         |
| `Network.UUID()`                         | ~9.3 ns  | 0         |
| `Network.AppendUUID(buf)`                | ~24 ns   | 0         |
| `Network.CIDR(IPv6CIDR, 48)`             | ~33 ns   | 1         |

Slice-returning functions allocate when they create returned storage. Append and buffer-taking APIs can be zero-allocation when caller capacity is enough.

## Thread Safety

Package generators and provider replacement are safe for concurrent use. `DefaultProvider` and providers from `NewUint64Provider` use atomic state. Providers installed with `NewReaderProvider` are only as concurrent-safe as their wrapped reader.

## License

This project is licensed under the terms of the [MIT License](LICENSE).
