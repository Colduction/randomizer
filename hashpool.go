package randomizer

import (
	"hash/maphash"
	"sync/atomic"
)

// hashPool is a custom pool that limits the number of maphash.Hash objects.
type hashPool struct {
	pool  chan *maphash.Hash
	state atomic.Uint64
}

const splitMixGamma uint64 = 0x9e3779b97f4a7c15

func splitMix64(x uint64) uint64 {
	z := x
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// NewHashPool creates a new hashPool with the specified size.
// If size is 0, it returns nil. The pool will preallocate size maphash.Hash
// objects for reuse.
func NewHashPool(size int) *hashPool {
	if size <= 0 {
		return nil
	}
	p := &hashPool{
		pool: make(chan *maphash.Hash, size),
	}
	seed := maphash.Bytes(maphash.MakeSeed(), nil)
	if seed == 0 {
		seed = splitMixGamma
	}
	p.state.Store(seed)
	for range size {
		h := new(maphash.Hash)
		h.SetSeed(maphash.MakeSeed())
		p.pool <- h
	}
	return p
}

// Get retrieves a maphash.Hash object from the pool.
// If the pool is empty, it creates and returns a new maphash.Hash instance.
func (p *hashPool) Get() *maphash.Hash {
	if p == nil {
		h := new(maphash.Hash)
		h.SetSeed(maphash.MakeSeed())
		return h
	}
	select {
	case h := <-p.pool:
		return h
	default:
		h := new(maphash.Hash)
		h.SetSeed(maphash.MakeSeed())
		return h
	}
}

// Put returns a maphash.Hash object to the pool.
// If the pool is full, the hash object is discarded.
func (p *hashPool) Put(h *maphash.Hash) {
	if p == nil || h == nil {
		return
	}
	h.Reset()
	select {
	case p.pool <- h:
	default:
	}
}

func (p *hashPool) next64() uint64 {
	if p == nil {
		return splitMix64(maphash.Bytes(maphash.MakeSeed(), nil) + splitMixGamma)
	}
	return splitMix64(p.state.Add(splitMixGamma))
}

// Sum appends 8 random bytes to b and returns the extended slice.
func (p *hashPool) Sum(b []byte) []byte {
	x := p.next64()
	return append(b,
		byte(x>>0),
		byte(x>>8),
		byte(x>>16),
		byte(x>>24),
		byte(x>>32),
		byte(x>>40),
		byte(x>>48),
		byte(x>>56))
}

// Sum32 generates a random 32-bit number using the hashPool.
func (p *hashPool) Sum32() uint32 {
	return uint32(p.next64() >> 32)
}

// Sum64 generates a random 64-bit number using the hashPool.
func (p *hashPool) Sum64() uint64 {
	return p.next64()
}

// DefaultHashPool is a globally accessible hashPool with a preallocated size.
var DefaultHashPool = NewHashPool(64)
