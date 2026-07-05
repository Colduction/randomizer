package randomizer

import (
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

func alphabetData(alphabet Alphabet) (string, uint8) {
	switch alphabet {
	case HexLowerAlphabet:
		return lhexdict, 4
	case HexUpperAlphabet:
		return uhexdict, 4
	case OctalAlphabet:
		return octi, 3
	case LowerAlphabet:
		return lower, 0
	case UpperAlphabet:
		return upper, 0
	case AlphaAlphabet:
		return alpha, 0
	case AlphaNumericAlphabet:
		return alnum, 0
	case Base32Alphabet:
		return base32, 5
	case Base64URLAlphabet:
		return base64u, 6
	default:
		return deci, 0
	}
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
	dict, bits := alphabetData(alphabet)
	fillAlphabetNoRepeat(out, dict, bits, currentProvider())
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
	dict, bits := alphabetData(alphabet)
	fillAlphabetNoRepeat(dst[offset:], dict, bits, currentProvider())
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
	fillAlphabetNoRepeat(dst[offset:], dictionary, 0, currentProvider())
	return dst
}

// fillAlphabetNoRepeat fills out with characters from dict ensuring no two adjacent characters are identical.
func fillAlphabetNoRepeat[S ~string | ~[]byte](out []byte, dict S, bits uint8, provider Provider) {
	var (
		raw   uint64
		avail uint8
		last  byte
	)
	if bits > 0 {
		mask := uint64((1 << bits) - 1)
		for i := 0; i < len(out); {
			if avail < bits {
				raw = provider.Sum64()
				avail = 64
			}
			c := dict[int(raw&mask)]
			raw >>= bits
			avail -= bits
			if i > 0 && c == last {
				continue
			}
			out[i] = c
			last = c
			i++
		}
		return
	}
	dn := len(dict)
	if dn == 0 {
		return
	}
	if dn > 256 {
		limit := uint64(dn)
		for i := 0; i < len(out); {
			c := dict[int(uniformUint64n(limit, provider))]
			if i > 0 && c == last {
				continue
			}
			out[i] = c
			last = c
			i++
		}
		return
	}
	cutoff := (256 / dn) * dn
	for i := 0; i < len(out); {
		if avail < 8 {
			raw = provider.Sum64()
			avail = 64
		}
		v := int(uint8(raw))
		raw >>= 8
		avail -= 8
		if v >= cutoff {
			continue
		}
		c := dict[v%dn]
		if i > 0 && c == last {
			continue
		}
		out[i] = c
		last = c
		i++
	}
}
