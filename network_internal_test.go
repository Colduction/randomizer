package randomizer

import (
	"bytes"
	"math/bits"
	"net/netip"
	"testing"
)

func TestRealOUIsTable(t *testing.T) {
	if len(realOUIs) < 128 {
		t.Fatalf("realOUIs has %d entries, want at least 128", len(realOUIs))
	}
	for i, oui := range realOUIs {
		if oui[0]&0x03 != 0 {
			t.Fatalf("realOUIs[%d] = %x is not universal unicast", i, oui)
		}
		if i > 0 && bytes.Compare(realOUIs[i-1][:], oui[:]) >= 0 {
			t.Fatalf("realOUIs[%d] = %x is not sorted after %x", i, oui, realOUIs[i-1])
		}
	}
	if realOUI(0) != realOUIs[0] || realOUI(uint64(len(realOUIs)-1)<<32) != realOUIs[len(realOUIs)-1] {
		t.Fatal("realOUI does not span the table")
	}
}

func TestIsPublicIPv6(t *testing.T) {
	cases := map[string]bool{
		"2001::1": false, "2001:2::1": false, "2001:1ff::1": false, "2001:200::1": true,
		"2001:db8::1": false, "2001:db9::1": true, "2002::1": false, "3fff::1": false,
		"3fff:fff::1": false, "3fff:1000::1": true, "5f00::1": false, "2620:4f:8000::1": false,
		"2620:4f:8001::1": true, "2a00::1": true, "2606:4700::1": true,
	}
	for s, want := range cases {
		hi := uint64(0)
		a := netip.MustParseAddr(s).As16()
		for _, b := range a[:8] {
			hi = hi<<8 | uint64(b)
		}
		if got := isPublicIPv6(hi); got != want {
			t.Fatalf("isPublicIPv6(%s) = %t, want %t", s, got, want)
		}
	}
}

func TestIsPublicASN(t *testing.T) {
	cases := map[uint32]bool{
		0: false, 1: true, 23455: true, 23456: false, 23457: true, 64495: true, 64496: false,
		65535: false, 65552: false, 100000: false, 131071: false, 131072: true,
		4199999999: true, 4200000000: false, 4294967295: false,
	}
	for a, want := range cases {
		if got := isPublicASN(a); got != want {
			t.Fatalf("isPublicASN(%d) = %t, want %t", a, got, want)
		}
	}
}

func TestBluetoothHashVector(t *testing.T) {
	// Bluetooth Core Specification Vol 3 Part H Appendix D.7.
	irk := []byte{0xec, 0x02, 0x34, 0xa3, 0x57, 0xc8, 0xad, 0x05, 0x34, 0x10, 0x10, 0xa6, 0x0a, 0x39, 0x7d, 0x9b}
	if got := bluetoothHash(irk, 0x708194); got != 0x0dfbaa {
		t.Fatalf("bluetoothHash = %06x, want 0dfbaa", got)
	}
}

func TestBatchSpecs(t *testing.T) {
	for n := uint64(1); n <= 256; n++ {
		spec := batchSpecs[n]
		prod := uint64(1)
		for range spec.k {
			prod *= n
		}
		if spec.k == 0 || prod != spec.prod {
			t.Fatalf("batchSpecs[%d] = %+v, want n^k == prod", n, spec)
		}
		if n > 1 && -spec.prod%spec.prod > 1<<60 {
			t.Fatalf("batchSpecs[%d] rejects more than 1/16 of words", n)
		}
		if hi, _ := bits.Mul64(spec.prod, n); n > 1 && hi == 0 && -(spec.prod*n)%(spec.prod*n) <= 1<<60 {
			t.Fatalf("batchSpecs[%d] could use k+1", n)
		}
	}
	if batchSpecs[9].k < 18 || batchSpecs[61].k < 10 || batchSpecs[25].k < 13 {
		t.Fatalf("batches smaller than expected: %+v %+v %+v", batchSpecs[9], batchSpecs[61], batchSpecs[25])
	}
}
