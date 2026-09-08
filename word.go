package randomizer

import (
	"math/bits"
	"slices"
	"unsafe"
)

const (
	deci     string = "0123456789"
	octi     string = "01234567"
	lhexdict string = "0123456789abcdef"
	uhexdict string = "0123456789ABCDEF"
	lower    string = "abcdefghijklmnopqrstuvwxyz"
	upper    string = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alpha    string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alnum    string = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	base32   string = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	base64u  string = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
)

type word struct{}

// Word provides random strings and byte sequences using the active [Provider]
// selected by [SetProvider].
var Word word

// Alphabet identifies a built-in byte alphabet used by [Word].
type Alphabet uint8

const (
	// DecimalAlphabet selects decimal digits 0-9.
	DecimalAlphabet Alphabet = iota
	// HexLowerAlphabet selects lowercase hexadecimal digits 0-9 and a-f.
	HexLowerAlphabet
	// HexUpperAlphabet selects uppercase hexadecimal digits 0-9 and A-F.
	HexUpperAlphabet
	// OctalAlphabet selects octal digits 0-7.
	OctalAlphabet
	// LowerAlphabet selects lowercase ASCII letters a-z.
	LowerAlphabet
	// UpperAlphabet selects uppercase ASCII letters A-Z.
	UpperAlphabet
	// AlphaAlphabet selects ASCII letters a-z and A-Z.
	AlphaAlphabet
	// AlphaNumericAlphabet selects ASCII digits and letters.
	AlphaNumericAlphabet
	// Base32Alphabet selects RFC 4648 base32 bytes without padding.
	Base32Alphabet
	// Base64URLAlphabet selects RFC 4648 URL-safe base64 bytes without padding.
	Base64URLAlphabet
)

// batchSpec describes how many uniform base-n positions one 64-bit word
// yields with the batched method of Brackett-Rozinsky and Lemire (Software:
// Practice and Experience, 2024): k successive multiplications by n peel k
// digits from the word, and the word is rejected when the final low product
// half falls below 2^64 mod n^k, which keeps every digit exactly uniform.
type batchSpec struct {
	prod uint64 // n^k
	k    uint8
}

// newBatchSpec returns the batch for base n: the largest k with n^k <= 2^64,
// reduced while more than one word in sixteen would be rejected.
func newBatchSpec(n uint64) batchSpec {
	if n < 2 {
		return batchSpec{prod: 1, k: 64}
	}
	prod, k := uint64(1), uint8(0)
	for {
		hi, lo := bits.Mul64(prod, n)
		if hi != 0 {
			break
		}
		prod, k = lo, k+1
	}
	for k > 1 && -prod%prod > 1<<60 {
		prod, k = prod/n, k-1
	}
	return batchSpec{prod: prod, k: k}
}

// batchSpecs holds the batch for every base up to 256, indexed by base.
var batchSpecs = func() (specs [257]batchSpec) {
	for n := range specs {
		specs[n] = newBatchSpec(uint64(n))
	}
	return specs
}()

// alphabets holds the built-in dictionaries indexed by [Alphabet].
var alphabets = [...]string{
	DecimalAlphabet:      deci,
	HexLowerAlphabet:     lhexdict,
	HexUpperAlphabet:     uhexdict,
	OctalAlphabet:        octi,
	LowerAlphabet:        lower,
	UpperAlphabet:        upper,
	AlphaAlphabet:        alpha,
	AlphaNumericAlphabet: alnum,
	Base32Alphabet:       base32,
	Base64URLAlphabet:    base64u,
}

func alphabetData(alphabet Alphabet) string {
	if int(alphabet) < len(alphabets) {
		return alphabets[alphabet]
	}
	return deci
}

// String returns a random string of the given length from alphabet.
// Unknown alphabet values use [DecimalAlphabet].
func (w word) String(alphabet Alphabet, length int) string {
	out := w.Bytes(alphabet, length)
	if len(out) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(out), len(out))
}

// Bytes returns a random byte slice of the given length from alphabet.
// Unknown alphabet values use [DecimalAlphabet].
func (word) Bytes(alphabet Alphabet, length int) []byte {
	if length <= 0 {
		return nil
	}
	out := make([]byte, length)
	fillBatchedNoRepeat(out, alphabetData(alphabet), currentProvider())
	return out
}

// Append appends a random byte sequence of the given length from alphabet to dst.
// It allocates only when dst lacks capacity. Unknown alphabet values use [DecimalAlphabet].
func (word) Append(dst []byte, alphabet Alphabet, length int) []byte {
	if length <= 0 {
		return dst
	}
	offset := len(dst)
	dst = slices.Grow(dst, length)[:offset+length]
	fillBatchedNoRepeat(dst[offset:], alphabetData(alphabet), currentProvider())
	return dst
}

// Custom returns a random string of the given length, or an empty string when
// dictionary cannot avoid adjacent duplicates.
func (word) Custom(dictionary string, length int) string {
	out := appendCustom(nil, dictionary, length)
	if len(out) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(out), len(out))
}

// CustomFromBytes returns a random string of the given length, or
// an empty string when dictionary cannot avoid adjacent duplicates.
func (word) CustomFromBytes(dictionary []byte, length int) string {
	out := appendCustom(nil, dictionary, length)
	if len(out) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(out), len(out))
}

// CustomBytes returns a random byte slice of the given length, or
// nil when dictionary cannot avoid adjacent duplicates.
func (word) CustomBytes(dictionary []byte, length int) []byte {
	return appendCustom(nil, dictionary, length)
}

// CustomBytesFromString returns a random byte slice of the given length, or nil
// when dictionary cannot avoid adjacent duplicates.
func (word) CustomBytesFromString(dictionary string, length int) []byte {
	return appendCustom(nil, dictionary, length)
}

// AppendCustom appends a random byte sequence from dictionary to dst.
// It returns dst unchanged when dictionary cannot avoid adjacent duplicates.
func (word) AppendCustom(dst []byte, dictionary string, length int) []byte {
	return appendCustom(dst, dictionary, length)
}

// AppendCustomFromBytes appends a random byte sequence from dictionary to dst.
// It returns dst unchanged when dictionary cannot avoid adjacent duplicates.
func (word) AppendCustomFromBytes(dst []byte, dictionary []byte, length int) []byte {
	return appendCustom(dst, dictionary, length)
}

func appendCustom[B ~string | ~[]byte](dst []byte, dictionary B, length int) []byte {
	if length <= 0 || len(dictionary) == 0 {
		return dst
	}
	if length > 1 {
		first := dictionary[0]
		for i := 1; ; i++ {
			if i == len(dictionary) {
				return dst
			}
			if dictionary[i] != first {
				break
			}
		}
	}
	offset := len(dst)
	dst = slices.Grow(dst, length)[:offset+length]
	out := dst[offset:]
	p := currentProvider()
	if len(dictionary) <= 256 {
		fillBatchedNoRepeat(out, dictionary, p)
	} else {
		fillWideNoRepeat(out, dictionary, p)
	}
	return dst
}

// skipPast returns 1 when idx >= lastIdx and 0 otherwise, without a branch:
// the outcome is a coin flip, so a conditional jump would mispredict half
// the time. Both arguments are below 2^63.
func skipPast(idx, lastIdx uint64) uint64 {
	return uint64(int64(lastIdx-1-idx)>>63) & 1
}

// fillBatchedNoRepeat fills out with symbols from a dict of at most 256
// bytes so that no two adjacent bytes are equal. The first symbol is drawn
// uniformly from all n positions; every later symbol draws a position from
// n-1 candidates and shifts it past the previous position, a bijection onto
// the other positions, so unique dictionaries never retry. A dictionary with
// repeated bytes keeps its weighting: positions holding the previous byte are
// skipped, leaving every other position equally likely. Positions come in
// batches of up to batchSpecs[n-1].k per random word; a rejected word
// restores the state from before its batch and is redrawn.
func fillBatchedNoRepeat[S ~string | ~[]byte](out []byte, dict S, p Provider) {
	if len(out) == 0 {
		return
	}
	n := uint64(len(dict))
	lastIdx := boundedFrom(p.Sum64(), n, p)
	last := dict[lastIdx]
	out[0] = last
	if len(out) == 1 || n < 2 {
		return
	}
	var (
		n1    = n - 1
		batch = int(batchSpecs[n1].k)
		fast  = isRuntime(p)
	)
	for i := 1; i < len(out); {
		var lo uint64
		if fast {
			lo = runtimeRand()
		} else {
			lo = p.Sum64()
		}
		var (
			start                    = i
			saveLast, saveIdx        = last, lastIdx
			prod              uint64 = 1
		)
		for d := 0; d < batch && i < len(out); d++ {
			var hi uint64
			hi, lo = bits.Mul64(lo, n1)
			prod *= n1
			idx := hi + skipPast(hi, lastIdx)
			c := dict[idx]
			if c == last {
				continue
			}
			out[i] = c
			last, lastIdx = c, idx
			i++
		}
		if lo < prod && lo < -prod%prod {
			last, lastIdx = saveLast, saveIdx
			i = start
		}
	}
}

// fillWideNoRepeat fills out from any dictionary so that no two adjacent
// bytes are equal, sampling each position with [boundedFrom] and retrying a
// position whose byte equals the previous byte; that keeps the weighting of
// dictionaries with repeated bytes, which the batched sampler cannot.
func fillWideNoRepeat[S ~string | ~[]byte](out []byte, dict S, p Provider) {
	var (
		last    byte
		lastIdx uint64
		n       = uint64(len(dict))
	)
	for i := 0; i < len(out); {
		var idx uint64
		if i == 0 {
			idx = boundedFrom(p.Sum64(), n, p)
		} else {
			idx = boundedFrom(p.Sum64(), n-1, p)
			idx += skipPast(idx, lastIdx)
			if dict[idx] == last {
				continue
			}
		}
		out[i] = dict[idx]
		last, lastIdx = dict[idx], idx
		i++
	}
}
