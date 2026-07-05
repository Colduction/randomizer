package randomizer

import (
	"encoding/binary"
	"io"
	"sync/atomic"
)

// DefaultProvider is the package-level [Provider] restored by [SetProvider]
// when no custom provider is supplied.
var DefaultProvider Provider = NewHashPool(64)

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

func fillAtomicRandomBytes(out []byte, state *atomic.Uint64) {
	var (
		i int
		n = len(out)
	)
	for ; i+8 <= n; i += 8 {
		binary.LittleEndian.PutUint64(out[i:i+8], splitMix64(state.Add(splitMixGamma)))
	}
	if i < n {
		x := splitMix64(state.Add(splitMixGamma))
		for j := i; j < n; j++ {
			out[j] = byte(x)
			x >>= 8
		}
	}
}

type uint64Provider struct {
	state atomic.Uint64
}

// NewUint64Provider returns a lock-free [Provider] seeded once from a source
// such as [math/rand.Rand] or [math/rand/v2.Source].
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
