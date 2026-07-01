package randomizer

import "unsafe"

const (
	deci     string = "0123456789"
	octi     string = "01234567"
	lhexdict string = "0123456789abcdef"
	uhexdict string = "0123456789ABCDEF"
)

type word struct{}

// Word provides random decimal, hexadecimal, octal, and custom strings using the
// active [Provider] selected by [SetProvider].
var Word word

// Decimal returns a random decimal string of the given length.
func (word) Decimal(length int) string {
	if length <= 0 {
		return ""
	}
	out := make([]byte, length)
	fillAlphabetNoRepeat(out, deci, 0, currentProvider())
	return unsafe.String(unsafe.SliceData(out), len(out))
}

// DecimalBytes returns a random decimal byte slice of the given length.
func (word) DecimalBytes(length int) []byte {
	if length <= 0 {
		return nil
	}
	out := make([]byte, length)
	fillAlphabetNoRepeat(out, deci, 0, currentProvider())
	return out
}

// Hex returns a random hexadecimal string of the given length.
// If uppercase is true, A-F are used; otherwise a-f.
func (word) Hex(length int, uppercase bool) string {
	if length <= 0 {
		return ""
	}
	dict := lhexdict
	if uppercase {
		dict = uhexdict
	}
	out := make([]byte, length)
	fillAlphabetNoRepeat(out, dict, 4, currentProvider())
	return unsafe.String(unsafe.SliceData(out), len(out))
}

// HexBytes returns a random hexadecimal byte slice of the given length.
// If uppercase is true, A-F are used; otherwise a-f.
func (word) HexBytes(length int, uppercase bool) []byte {
	if length <= 0 {
		return nil
	}
	dict := lhexdict
	if uppercase {
		dict = uhexdict
	}
	out := make([]byte, length)
	fillAlphabetNoRepeat(out, dict, 4, currentProvider())
	return out
}

// Octal returns a random octal string of the given length.
func (word) Octal(length int) string {
	if length <= 0 {
		return ""
	}
	out := make([]byte, length)
	fillAlphabetNoRepeat(out, octi, 3, currentProvider())
	return unsafe.String(unsafe.SliceData(out), len(out))
}

// OctalBytes returns a random octal byte slice of the given length.
func (word) OctalBytes(length int) []byte {
	if length <= 0 {
		return nil
	}
	out := make([]byte, length)
	fillAlphabetNoRepeat(out, octi, 3, currentProvider())
	return out
}

// Custom returns a random string with the same length as dictionary, or an empty
// string when dictionary cannot avoid adjacent duplicates.
func (word) Custom(dictionary string) string {
	out := customBytes(dictionary)
	if len(out) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(out), len(out))
}

// CustomFromBytes returns a random string with the same length as dictionary, or
// an empty string when dictionary cannot avoid adjacent duplicates.
func (word) CustomFromBytes(dictionary []byte) string {
	out := customBytes(dictionary)
	if len(out) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(out), len(out))
}

// CustomBytes returns a random byte slice with the same length as dictionary, or
// nil when dictionary cannot avoid adjacent duplicates.
func (word) CustomBytes(dictionary []byte) []byte {
	return customBytes(dictionary)
}

// CustomBytesFromString returns a random byte slice with the same length as
// dictionary, or nil when dictionary cannot avoid adjacent duplicates.
func (word) CustomBytesFromString(dictionary string) []byte {
	return customBytes(dictionary)
}

func customBytes[B ~string | ~[]byte](dictionary B) []byte {
	if len(dictionary) == 0 {
		return nil
	}
	if len(dictionary) > 1 {
		first := dictionary[0]
		for i := 1; ; i++ {
			if i == len(dictionary) {
				return nil
			}
			if dictionary[i] != first {
				break
			}
		}
	}
	out := make([]byte, len(dictionary))
	fillAlphabetNoRepeat(out, dictionary, 0, currentProvider())
	return out
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
