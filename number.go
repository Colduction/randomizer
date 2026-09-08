package randomizer

import "math/bits"

// Integers combines [SignedIntegers] and [UnsignedIntegers].
type Integers interface {
	SignedIntegers | UnsignedIntegers
}

// SignedIntegers identifies all supported signed integer types.
type SignedIntegers interface {
	~int8 | ~int16 | ~int | ~int32 | ~int64
}

// UnsignedIntegers identifies all supported unsigned integer types.
type UnsignedIntegers interface {
	~uint8 | ~uint16 | ~uint | ~uint32 | ~uint64 | ~uintptr
}

const (
	signMask uint64 = 1 << 63

	inv24 float32 = 1.0 / (1 << 24)
	inv53 float64 = 1.0 / (1 << 53)
)

// boundedFrom maps the random word x to a uniform value in [0, n) with
// Lemire's nearly divisionless method (ACM TOMACS 2019): one 64x64 multiply
// in the common case. It draws further words from p only when the low product
// half falls below n, which happens with probability under n/2^64.
// It returns 0 when n == 0.
func boundedFrom(x, n uint64, p Provider) uint64 {
	if hi, lo := bits.Mul64(x, n); lo >= n {
		return hi
	}
	return boundedSlow(x, n, p)
}

// boundedSlow finishes [boundedFrom] with the one division and rejection loop
// that the fast path avoids. It stays out of line so boundedFrom inlines.
//
//go:noinline
func boundedSlow(x, n uint64, p Provider) uint64 {
	threshold := -n % n
	hi, lo := bits.Mul64(x, n)
	for lo < threshold {
		hi, lo = bits.Mul64(p.Sum64(), n)
	}
	return hi
}

// uniformUint64n returns a uniform random value in [0, n) drawn from p.
func uniformUint64n(n uint64, p Provider) uint64 {
	return boundedFrom(p.Sum64(), n, p)
}

// Int returns a random integer constrained by [SignedIntegers].
func Int[T SignedIntegers]() T {
	return T(currentProvider().Sum64())
}

// IntInterval returns a random [SignedIntegers] value in [min, max).
// min and max are swapped automatically if min > max.
func IntInterval[T SignedIntegers](min, max T) T {
	if min == max {
		return min
	}
	if min > max {
		min, max = max, min
	}
	var (
		minU = uint64(int64(min)) ^ signMask
		maxU = uint64(int64(max)) ^ signMask
	)
	v := uniformUint64n(maxU-minU, currentProvider())
	return T(int64((minU + v) ^ signMask))
}

// Uint returns a random integer constrained by [UnsignedIntegers].
func Uint[T UnsignedIntegers]() T {
	return T(currentProvider().Sum64())
}

// UintInterval returns a random [UnsignedIntegers] value in [min, max).
// min and max are swapped automatically if min > max.
func UintInterval[T UnsignedIntegers](min, max T) T {
	if min == max {
		return min
	}
	if min > max {
		min, max = max, min
	}
	return min + T(uniformUint64n(uint64(max-min), currentProvider()))
}

// Float32 returns a uniformly distributed random float32 in [0, 1) using the
// active [Provider].
func Float32() float32 {
	return float32(currentProvider().Sum32()>>8) * inv24
}

// Float64 returns a uniformly distributed random float64 in [0, 1) using the
// active [Provider].
func Float64() float64 {
	return float64(currentProvider().Sum64()>>11) * inv53
}
