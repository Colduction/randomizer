package randomizer_test

import (
	"bytes"
	cryptorand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/colduction/randomizer-go"
)

type fixedProvider uint64

func (fp fixedProvider) Read(p []byte) (int, error) {
	var i int
	for ; i+8 <= len(p); i += 8 {
		binary.LittleEndian.PutUint64(p[i:i+8], uint64(fp))
	}
	if i < len(p) {
		x := uint64(fp)
		for j := i; j < len(p); j++ {
			p[j] = byte(x)
			x >>= 8
		}
	}
	return len(p), nil
}

func (fp fixedProvider) Sum(b []byte) []byte {
	return binary.LittleEndian.AppendUint64(b, uint64(fp))
}

func (fixedProvider) Sum32() uint32 {
	return 0xabcdef01
}

func (fp fixedProvider) Sum64() uint64 {
	return uint64(fp)
}

type oneShotSource struct {
	calls atomic.Uint32
}

func (oss *oneShotSource) Uint64() uint64 {
	if oss.calls.Add(1) != 1 {
		panic("Uint64 called more than once")
	}
	return 0x0123456789abcdef
}

func TestProviderSetProviderRoutesGenerators(t *testing.T) {
	const value fixedProvider = 0x0123456789abcdef

	previous := randomizer.SetProvider(value)
	defer randomizer.SetProvider(previous)

	if got := randomizer.Uint[uint64](); got != uint64(value) {
		t.Fatalf("Uint[uint64]() = 0x%016x, want 0x%016x", got, uint64(value))
	}
	first := randomizer.Word.String(randomizer.HexLowerAlphabet, 16)
	if second := randomizer.Word.String(randomizer.HexLowerAlphabet, 16); first != second || len(first) != 16 {
		t.Fatalf("Word.String under a fixed provider = %q then %q, want equal 16-byte strings", first, second)
	}
	if got := randomizer.Network.Value(randomizer.VLANID); got != 0x12 {
		t.Fatalf("Network.Value(VLANID) = 0x%x, want 0x12", got)
	}
	var buf [12]byte
	if n, err := randomizer.Read(buf[:]); n != len(buf) || err != nil {
		t.Fatalf("Read length/error = %d/%v, want %d/nil", n, err, len(buf))
	}
	want := []byte{
		0xef, 0xcd, 0xab, 0x89, 0x67, 0x45, 0x23, 0x01,
		0xef, 0xcd, 0xab, 0x89,
	}
	if !bytes.Equal(buf[:], want) {
		t.Fatalf("Read bytes = %x, want %x", buf, want)
	}
}

func TestProviderSetProviderNilRestoresDefault(t *testing.T) {
	previous := randomizer.SetProvider(fixedProvider(1))
	defer randomizer.SetProvider(previous)

	if got := randomizer.SetProvider(nil); got != fixedProvider(1) {
		t.Fatalf("SetProvider(nil) previous provider = %v, want %v", got, fixedProvider(1))
	}
	if got := randomizer.SetProvider(previous); got != randomizer.DefaultProvider {
		t.Fatalf("SetProvider(previous) replaced %T, want DefaultProvider", got)
	}
}

func TestProviderSetProviderConcurrent(t *testing.T) {
	previous := randomizer.SetProvider(fixedProvider(1))
	defer randomizer.SetProvider(previous)

	const goroutines = 32
	var group sync.WaitGroup
	group.Add(goroutines)
	for i := range goroutines {
		go func() {
			defer group.Done()
			provider := fixedProvider(i + 1)
			for range 128 {
				randomizer.SetProvider(provider)
				_ = randomizer.Uint[uint64]()
			}
		}()
	}
	group.Wait()
}

func TestUint64ProviderSupportsMathRand(t *testing.T) {
	const seed int64 = 42

	first := randomizer.NewUint64Provider(rand.New(rand.NewSource(seed)))
	second := randomizer.NewUint64Provider(rand.New(rand.NewSource(seed)))
	if first == nil || second == nil {
		t.Fatal("NewUint64Provider returned nil")
	}

	for range 16 {
		if got, want := first.Sum64(), second.Sum64(); got != want {
			t.Fatalf("matching math/rand sources produced 0x%016x and 0x%016x", got, want)
		}
	}
}

func TestUint64ProviderConcurrent(t *testing.T) {
	const (
		goroutines = 32
		values     = 128
	)

	source := new(oneShotSource)
	provider := randomizer.NewUint64Provider(source)
	if provider == nil {
		t.Fatal("NewUint64Provider returned nil")
	}

	results := make(chan uint64, goroutines*values)
	var group sync.WaitGroup
	group.Add(goroutines)
	for range goroutines {
		go func() {
			defer group.Done()
			for range values {
				results <- provider.Sum64()
			}
		}()
	}
	group.Wait()
	close(results)

	seen := make(map[uint64]struct{}, goroutines*values)
	for value := range results {
		if _, exists := seen[value]; exists {
			t.Fatalf("Sum64 returned duplicate value 0x%016x", value)
		}
		seen[value] = struct{}{}
	}
	if got := source.calls.Load(); got != 1 {
		t.Fatalf("source Uint64 calls = %d, want 1", got)
	}
}

func TestReaderProviderValues(t *testing.T) {
	data := []byte{
		0x01, 0x02, 0x03, 0x04,
		0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c,
		0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14,
	}
	provider := randomizer.NewReaderProvider(bytes.NewReader(data))
	if provider == nil {
		t.Fatal("NewReaderProvider returned nil")
	}

	if got := provider.Sum32(); got != 0x04030201 {
		t.Fatalf("Sum32() = 0x%08x, want 0x04030201", got)
	}
	if got := provider.Sum64(); got != 0x0c0b0a0908070605 {
		t.Fatalf("Sum64() = 0x%016x, want 0x0c0b0a0908070605", got)
	}

	got := provider.Sum([]byte{0xaa, 0xbb})
	want := []byte{0xaa, 0xbb, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14}
	if !bytes.Equal(got, want) {
		t.Fatalf("Sum() = %x, want %x", got, want)
	}
}

func TestReaderProviderRead(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	provider := randomizer.NewReaderProvider(bytes.NewReader(data))
	if provider == nil {
		t.Fatal("NewReaderProvider returned nil")
	}

	buf := make([]byte, len(data))
	n, err := provider.Read(buf)
	if n != len(data) || err != nil {
		t.Fatalf("Read length/error = %d/%v, want %d/nil", n, err, len(data))
	}
	if !bytes.Equal(buf, data) {
		t.Fatalf("Read bytes = %x, want %x", buf, data)
	}

	n, err = provider.Read(buf[:1])
	if n != 0 || err == nil {
		t.Fatalf("Read exhausted length/error = %d/%v, want 0/error", n, err)
	}
}

func TestReaderProviderSupportsCryptoRand(t *testing.T) {
	const goroutines = 32

	provider := randomizer.NewReaderProvider(cryptorand.Reader)
	if provider == nil {
		t.Fatal("NewReaderProvider returned nil")
	}

	var group sync.WaitGroup
	group.Add(goroutines)
	for range goroutines {
		go func() {
			defer group.Done()
			if got := provider.Sum(nil); len(got) != 8 {
				t.Errorf("Sum(nil) length = %d, want 8", len(got))
			}
		}()
	}
	group.Wait()
}

func TestProviderConstructorsRejectNil(t *testing.T) {
	if got := randomizer.NewUint64Provider(nil); got != nil {
		t.Fatalf("NewUint64Provider(nil) = %v, want nil", got)
	}
	if got := randomizer.NewReaderProvider(nil); got != nil {
		t.Fatalf("NewReaderProvider(nil) = %v, want nil", got)
	}

}

var benchProviderReadBuf [256]byte

func BenchmarkProviderReadDefault(b *testing.B) {
	buf := benchProviderReadBuf[:]
	b.ReportAllocs()
	for b.Loop() {
		_, _ = randomizer.Read(buf)
	}
}

func BenchmarkProviderReadDefaultParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var buf [256]byte
		for pb.Next() {
			_, _ = randomizer.Read(buf[:])
		}
	})
}

// seqProvider replays scripted words and fails the test when they run out, so
// a rejection loop that draws more than expected is reported instead of
// hanging on a constant provider.
type seqProvider struct {
	t    testing.TB
	vals []uint64
	i    int
}

func newSeqProvider(t testing.TB, vals ...uint64) *seqProvider {
	return &seqProvider{t: t, vals: vals}
}

func (sp *seqProvider) Sum64() uint64 {
	if sp.i >= len(sp.vals) {
		sp.t.Fatalf("seqProvider exhausted after %d words", len(sp.vals))
	}
	v := sp.vals[sp.i]
	sp.i++
	return v
}

func (sp *seqProvider) Sum32() uint32 {
	return uint32(sp.Sum64() >> 32)
}

func (sp *seqProvider) Sum(b []byte) []byte {
	return binary.LittleEndian.AppendUint64(b, sp.Sum64())
}

func (sp *seqProvider) Read(p []byte) (int, error) {
	var i int
	for ; i+8 <= len(p); i += 8 {
		binary.LittleEndian.PutUint64(p[i:i+8], sp.Sum64())
	}
	if i < len(p) {
		x := sp.Sum64()
		for j := i; j < len(p); j++ {
			p[j] = byte(x)
			x >>= 8
		}
	}
	return len(p), nil
}

func TestRuntimeProviderIsDefault(t *testing.T) {
	if randomizer.DefaultProvider == nil {
		t.Fatal("DefaultProvider is nil")
	}
	if _, ok := randomizer.DefaultProvider.(interface{ Get() any }); ok {
		t.Fatal("DefaultProvider must not be the hash pool")
	}
	previous := randomizer.SetProvider(nil)
	defer randomizer.SetProvider(previous)
	if got := randomizer.SetProvider(nil); got != randomizer.DefaultProvider {
		t.Fatalf("active provider = %T, want DefaultProvider", got)
	}
}

func TestRuntimeProviderRead(t *testing.T) {
	for _, n := range []int{0, 1, 7, 8, 9, 256} {
		buf := make([]byte, n)
		got, err := randomizer.DefaultProvider.Read(buf)
		if got != n || err != nil {
			t.Fatalf("Read(%d) = %d, %v", n, got, err)
		}
		if n >= 8 && bytes.Equal(buf, make([]byte, n)) {
			t.Fatalf("Read(%d) left the buffer zero", n)
		}
	}
	if got := randomizer.DefaultProvider.Sum([]byte{1}); len(got) != 9 || got[0] != 1 {
		t.Fatalf("Sum length/prefix = %d/%v, want 9/1", len(got), got[:1])
	}
	first := randomizer.DefaultProvider.Sum64()
	for i := 0; ; i++ {
		if randomizer.DefaultProvider.Sum64() != first {
			break
		}
		if i == 64 {
			t.Fatal("Sum64 appears constant")
		}
	}
	first32 := randomizer.DefaultProvider.Sum32()
	for i := 0; ; i++ {
		if randomizer.DefaultProvider.Sum32() != first32 {
			break
		}
		if i == 64 {
			t.Fatal("Sum32 appears constant")
		}
	}
}

func TestRuntimeProviderConcurrentDistinct(t *testing.T) {
	const goroutines, values = 16, 256
	results := make(chan uint64, goroutines*values)
	var group sync.WaitGroup
	group.Add(goroutines)
	for range goroutines {
		go func() {
			defer group.Done()
			for range values {
				results <- randomizer.Uint[uint64]()
			}
		}()
	}
	group.Wait()
	close(results)
	seen := make(map[uint64]struct{}, goroutines*values)
	for value := range results {
		if _, exists := seen[value]; exists {
			t.Fatalf("Uint returned duplicate value 0x%016x", value)
		}
		seen[value] = struct{}{}
	}
}
