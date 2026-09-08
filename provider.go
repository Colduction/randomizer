package randomizer

import (
	"encoding/binary"
	"io"
	"sync/atomic"
	_ "unsafe" // for go:linkname
)

// runtimeRand returns a random 64-bit value from the Go runtime's per-thread
// ChaCha8 generator, the source behind the math/rand/v2 top-level functions.
// It never locks: every OS thread owns its own state, so throughput scales
// linearly with the number of threads.
//
//go:linkname runtimeRand runtime.rand
func runtimeRand() uint64

// DefaultProvider is the package-level [Provider] restored by [SetProvider]
// when no custom provider is supplied. It draws from the Go runtime's
// per-thread ChaCha8 generator (the design described in "Secure Randomness in
// Go 1.22"): lock-free under any level of concurrency, seeded by the operating
// system, and not reproducible. Use [NewUint64Provider] or [NewReaderProvider]
// for seeded or cryptographic sources.
var DefaultProvider Provider = runtimeProvider{}

// Provider supplies random bytes and integer values to package generators.
// Use [NewUint64Provider], [NewReaderProvider], or a concurrency-safe custom
// implementation with [SetProvider].
type Provider interface {
	// Read fills p with random bytes and returns len(p), nil on success.
	Read(p []byte) (n int, err error)
	// Sum appends eight random bytes to b using the same source as
	// [Provider.Sum32] and [Provider.Sum64].
	Sum(b []byte) []byte
	// Sum32 returns a random 32-bit value from the provider.
	Sum32() uint32
	// Sum64 returns a random 64-bit value from the provider.
	Sum64() uint64
}

// runtimeProvider implements [Provider] over [runtimeRand]. It must stay an
// empty comparable struct: callers compare it against [DefaultProvider]
// through the interface.
type runtimeProvider struct{}

// Read fills p with random bytes and returns len(p), nil.
func (runtimeProvider) Read(p []byte) (n int, err error) {
	fillRuntimeBytes(p)
	return len(p), nil
}

// Sum appends eight random bytes to b and returns the extended slice.
func (runtimeProvider) Sum(b []byte) []byte {
	return binary.LittleEndian.AppendUint64(b, runtimeRand())
}

// Sum32 returns a random 32-bit value.
func (runtimeProvider) Sum32() uint32 {
	return uint32(runtimeRand() >> 32)
}

// Sum64 returns a random 64-bit value.
func (runtimeProvider) Sum64() uint64 {
	return runtimeRand()
}

// isRuntime reports whether p is the runtime provider. Loops hoist the check
// and call [runtimeRand] directly, skipping the interface dispatch per word.
func isRuntime(p Provider) bool {
	_, ok := p.(runtimeProvider)
	return ok
}

// putTail writes the low len(out) bytes of x into out, little-endian.
// len(out) must be at most 8.
func putTail(out []byte, x uint64) {
	for i := range out {
		out[i] = byte(x)
		x >>= 8
	}
}

func fillRuntimeBytes(out []byte) {
	i := 0
	for ; i+8 <= len(out); i += 8 {
		binary.LittleEndian.PutUint64(out[i:i+8], runtimeRand())
	}
	if i < len(out) {
		putTail(out[i:], runtimeRand())
	}
}

// fillRandomBytes fills out from p, eight bytes per word. It never hands out
// to an interface method, so callers may pass stack buffers.
func fillRandomBytes(out []byte, p Provider) {
	if isRuntime(p) {
		fillRuntimeBytes(out)
		return
	}
	i := 0
	for ; i+8 <= len(out); i += 8 {
		binary.LittleEndian.PutUint64(out[i:i+8], p.Sum64())
	}
	if i < len(out) {
		putTail(out[i:], p.Sum64())
	}
}

func fillAtomicRandomBytes(out []byte, state *atomic.Uint64) {
	i := 0
	for ; i+8 <= len(out); i += 8 {
		binary.LittleEndian.PutUint64(out[i:i+8], splitMix64(state.Add(splitMixGamma)))
	}
	if i < len(out) {
		putTail(out[i:], splitMix64(state.Add(splitMixGamma)))
	}
}

type uint64Provider struct {
	state atomic.Uint64
}

// NewUint64Provider returns a seeded [Provider] that advances one shared
// SplitMix64 counter from source.Uint64(), called once. Values are
// reproducible for equal seeds; the shared counter serializes concurrent
// callers on one cache line, unlike [DefaultProvider].
// Pass the result to [SetProvider]. It returns nil when source is nil.
func NewUint64Provider(source interface{ Uint64() uint64 }) Provider {
	if source == nil {
		return nil
	}
	provider := new(uint64Provider)
	provider.state.Store(source.Uint64())
	return provider
}

// Read fills p with random bytes and returns len(p), nil.
func (u64p *uint64Provider) Read(p []byte) (n int, err error) {
	fillAtomicRandomBytes(p, &u64p.state)
	return len(p), nil
}

// Sum appends eight random bytes to b and returns the extended slice.
func (u64p *uint64Provider) Sum(b []byte) []byte {
	return binary.LittleEndian.AppendUint64(b, u64p.Sum64())
}

// Sum32 returns a random 32-bit value.
func (u64p *uint64Provider) Sum32() uint32 {
	return uint32(u64p.Sum64() >> 32)
}

// Sum64 returns a random 64-bit value.
func (u64p *uint64Provider) Sum64() uint64 {
	return splitMix64(u64p.state.Add(splitMixGamma))
}

type readerProvider struct {
	reader io.Reader
}

// NewReaderProvider returns a [Provider] backed by an [io.Reader].
// The reader must be safe for concurrent use, as [crypto/rand.Reader] is.
// Pass the result to [SetProvider]. It returns nil when reader is nil.
func NewReaderProvider(reader io.Reader) Provider {
	if reader == nil {
		return nil
	}
	return &readerProvider{reader: reader}
}

// Read fills p from the provider reader.
func (rp *readerProvider) Read(p []byte) (n int, err error) {
	return io.ReadFull(rp.reader, p)
}

// Sum appends eight random bytes to b and returns the extended slice.
func (rp *readerProvider) Sum(b []byte) []byte {
	offset := len(b)
	b = append(b, 0, 0, 0, 0, 0, 0, 0, 0)
	if _, err := rp.Read(b[offset:]); err != nil {
		panic(err)
	}
	return b
}

// Sum32 returns a random 32-bit value.
func (rp *readerProvider) Sum32() uint32 {
	var b [4]byte
	if _, err := rp.Read(b[:]); err != nil {
		panic(err)
	}
	return binary.LittleEndian.Uint32(b[:])
}

// Sum64 returns a random 64-bit value.
func (rp *readerProvider) Sum64() uint64 {
	var b [8]byte
	if _, err := rp.Read(b[:]); err != nil {
		panic(err)
	}
	return binary.LittleEndian.Uint64(b[:])
}

// activeProvider publishes the provider used by all package generators.
var activeProvider atomic.Pointer[Provider]

func init() {
	provider := DefaultProvider
	activeProvider.Store(&provider)
}

// SetProvider sets the [Provider] used by all package generators and returns
// the prior provider. A nil provider restores [DefaultProvider].
func SetProvider(provider Provider) Provider {
	if provider == nil {
		provider = DefaultProvider
	}
	return *activeProvider.Swap(&provider)
}

// Read fills p with random bytes from the active [Provider].
func Read(p []byte) (n int, err error) {
	return currentProvider().Read(p)
}

func currentProvider() Provider {
	return *activeProvider.Load()
}
