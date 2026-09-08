package randomizer_test

import (
	"strings"
	"testing"

	"github.com/colduction/randomizer-go"
)

var (
	benchWordString string
	benchWordBytes  []byte
)

func hasAdjacentDuplicate(b []byte) bool {
	if len(b) < 2 {
		return false
	}
	last := b[0]
	for i := 1; i < len(b); i++ {
		if b[i] == last {
			return true
		}
		last = b[i]
	}
	return false
}

func allInAlphabet(b []byte, allow [256]bool) bool {
	for _, c := range b {
		if !allow[c] {
			return false
		}
	}
	return true
}

func makeAlphabet(chars string) [256]bool {
	var allow [256]bool
	for i := 0; i < len(chars); i++ {
		allow[chars[i]] = true
	}
	return allow
}

func TestWordBuiltInAlphabets(t *testing.T) {
	const n = 4096
	cases := []struct {
		alphabet randomizer.Alphabet
		chars    string
	}{
		{randomizer.DecimalAlphabet, "0123456789"},
		{randomizer.HexLowerAlphabet, "0123456789abcdef"},
		{randomizer.HexUpperAlphabet, "0123456789ABCDEF"},
		{randomizer.OctalAlphabet, "01234567"},
		{randomizer.LowerAlphabet, "abcdefghijklmnopqrstuvwxyz"},
		{randomizer.UpperAlphabet, "ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{randomizer.AlphaAlphabet, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{randomizer.AlphaNumericAlphabet, "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{randomizer.Base32Alphabet, "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"},
		{randomizer.Base64URLAlphabet, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"},
	}

	for _, tc := range cases {
		allow := makeAlphabet(tc.chars)

		s := randomizer.Word.String(tc.alphabet, n)
		if len(s) != n {
			t.Fatalf("String(%v) length = %d, want %d", tc.alphabet, len(s), n)
		}
		sb := []byte(s)
		if !allInAlphabet(sb, allow) {
			t.Fatalf("String(%v) produced invalid character", tc.alphabet)
		}
		if hasAdjacentDuplicate(sb) {
			t.Fatalf("String(%v) produced adjacent duplicate character", tc.alphabet)
		}

		b := randomizer.Word.Bytes(tc.alphabet, n)
		if len(b) != n {
			t.Fatalf("Bytes(%v) length = %d, want %d", tc.alphabet, len(b), n)
		}
		if !allInAlphabet(b, allow) {
			t.Fatalf("Bytes(%v) produced invalid character", tc.alphabet)
		}
		if hasAdjacentDuplicate(b) {
			t.Fatalf("Bytes(%v) produced adjacent duplicate character", tc.alphabet)
		}

		var buf [n + 3]byte
		dst := append(buf[:0], "id:"...)
		out := randomizer.Word.Append(dst, tc.alphabet, n)
		if len(out) != n+3 || string(out[:3]) != "id:" || &out[3] != &buf[3] {
			t.Fatalf("Append(%v) did not preserve prefix or reuse caller buffer", tc.alphabet)
		}
		if !allInAlphabet(out[3:], allow) {
			t.Fatalf("Append(%v) produced invalid character", tc.alphabet)
		}
		if hasAdjacentDuplicate(out[3:]) {
			t.Fatalf("Append(%v) produced adjacent duplicate character", tc.alphabet)
		}
	}
}

func TestWordZeroLength(t *testing.T) {
	if got := randomizer.Word.String(randomizer.DecimalAlphabet, 0); got != "" {
		t.Fatalf("String(DecimalAlphabet,0) = %q, want empty string", got)
	}
	if got := randomizer.Word.String(randomizer.AlphaNumericAlphabet, 0); got != "" {
		t.Fatalf("String(AlphaNumericAlphabet,0) = %q, want empty string", got)
	}
	if got := randomizer.Word.String(randomizer.HexLowerAlphabet, 0); got != "" {
		t.Fatalf("String(HexLowerAlphabet,0) = %q, want empty string", got)
	}
	if got := randomizer.Word.String(randomizer.OctalAlphabet, 0); got != "" {
		t.Fatalf("String(OctalAlphabet,0) = %q, want empty string", got)
	}
	if got := randomizer.Word.Bytes(randomizer.DecimalAlphabet, 0); got != nil {
		t.Fatalf("Bytes(DecimalAlphabet,0) = %v, want nil", got)
	}
	if got := randomizer.Word.Bytes(randomizer.AlphaNumericAlphabet, 0); got != nil {
		t.Fatalf("Bytes(AlphaNumericAlphabet,0) = %v, want nil", got)
	}
	if got := randomizer.Word.Append([]byte("x"), randomizer.AlphaNumericAlphabet, 0); string(got) != "x" {
		t.Fatalf("Append zero length = %q, want %q", got, "x")
	}
	if got := randomizer.Word.Bytes(randomizer.HexLowerAlphabet, 0); got != nil {
		t.Fatalf("Bytes(HexLowerAlphabet,0) = %v, want nil", got)
	}
	if got := randomizer.Word.Bytes(randomizer.OctalAlphabet, 0); got != nil {
		t.Fatalf("Bytes(OctalAlphabet,0) = %v, want nil", got)
	}
}

func TestWordDecimalOutput(t *testing.T) {
	const n = 4096
	allow := makeAlphabet("0123456789")

	s := randomizer.Word.String(randomizer.DecimalAlphabet, n)
	if len(s) != n {
		t.Fatalf("Decimal length = %d, want %d", len(s), n)
	}
	sb := []byte(s)
	if !allInAlphabet(sb, allow) {
		t.Fatal("Decimal produced invalid character")
	}
	if hasAdjacentDuplicate(sb) {
		t.Fatal("Decimal produced adjacent duplicate character")
	}

	b := randomizer.Word.Bytes(randomizer.DecimalAlphabet, n)
	if len(b) != n {
		t.Fatalf("DecimalBytes length = %d, want %d", len(b), n)
	}
	if !allInAlphabet(b, allow) {
		t.Fatal("DecimalBytes produced invalid character")
	}
	if hasAdjacentDuplicate(b) {
		t.Fatal("DecimalBytes produced adjacent duplicate character")
	}
}

func TestWordHexOutput(t *testing.T) {
	const n = 4096
	allowLower := makeAlphabet("0123456789abcdef")
	allowUpper := makeAlphabet("0123456789ABCDEF")

	sLower := randomizer.Word.String(randomizer.HexLowerAlphabet, n)
	if len(sLower) != n {
		t.Fatalf("Hex length = %d, want %d", len(sLower), n)
	}
	sbLower := []byte(sLower)
	if !allInAlphabet(sbLower, allowLower) {
		t.Fatal("Hex lower produced invalid character")
	}
	if hasAdjacentDuplicate(sbLower) {
		t.Fatal("Hex lower produced adjacent duplicate character")
	}

	sUpper := randomizer.Word.String(randomizer.HexUpperAlphabet, n)
	if len(sUpper) != n {
		t.Fatalf("Hex uppercase length = %d, want %d", len(sUpper), n)
	}
	sbUpper := []byte(sUpper)
	if !allInAlphabet(sbUpper, allowUpper) {
		t.Fatal("Hex uppercase produced invalid character")
	}
	if hasAdjacentDuplicate(sbUpper) {
		t.Fatal("Hex uppercase produced adjacent duplicate character")
	}

	bLower := randomizer.Word.Bytes(randomizer.HexLowerAlphabet, n)
	if len(bLower) != n {
		t.Fatalf("HexBytes lower length = %d, want %d", len(bLower), n)
	}
	if !allInAlphabet(bLower, allowLower) {
		t.Fatal("HexBytes lower produced invalid character")
	}
	if hasAdjacentDuplicate(bLower) {
		t.Fatal("HexBytes lower produced adjacent duplicate character")
	}

	bUpper := randomizer.Word.Bytes(randomizer.HexUpperAlphabet, n)
	if len(bUpper) != n {
		t.Fatalf("HexBytes upper length = %d, want %d", len(bUpper), n)
	}
	if !allInAlphabet(bUpper, allowUpper) {
		t.Fatal("HexBytes upper produced invalid character")
	}
	if hasAdjacentDuplicate(bUpper) {
		t.Fatal("HexBytes upper produced adjacent duplicate character")
	}
}

func TestWordOctalOutput(t *testing.T) {
	const n = 4096
	allow := makeAlphabet("01234567")

	s := randomizer.Word.String(randomizer.OctalAlphabet, n)
	if len(s) != n {
		t.Fatalf("Octal length = %d, want %d", len(s), n)
	}
	sb := []byte(s)
	if !allInAlphabet(sb, allow) {
		t.Fatal("Octal produced invalid character")
	}
	if hasAdjacentDuplicate(sb) {
		t.Fatal("Octal produced adjacent duplicate character")
	}

	b := randomizer.Word.Bytes(randomizer.OctalAlphabet, n)
	if len(b) != n {
		t.Fatalf("OctalBytes length = %d, want %d", len(b), n)
	}
	if !allInAlphabet(b, allow) {
		t.Fatal("OctalBytes produced invalid character")
	}
	if hasAdjacentDuplicate(b) {
		t.Fatal("OctalBytes produced adjacent duplicate character")
	}
}

func TestWordCustomOutput(t *testing.T) {
	const n = 4096
	const dict = "AABCXYZ9"
	allow := makeAlphabet(dict)

	s := randomizer.Word.Custom(dict, n)
	if len(s) != n {
		t.Fatalf("Custom length = %d, want %d", len(s), n)
	}
	sb := []byte(s)
	if !allInAlphabet(sb, allow) {
		t.Fatal("Custom produced invalid character")
	}
	if hasAdjacentDuplicate(sb) {
		t.Fatal("Custom produced adjacent duplicate character")
	}

	fromBytes := randomizer.Word.CustomFromBytes([]byte(dict), n)
	if len(fromBytes) != n {
		t.Fatalf("CustomFromBytes length = %d, want %d", len(fromBytes), n)
	}
	fromBytesBytes := []byte(fromBytes)
	if !allInAlphabet(fromBytesBytes, allow) {
		t.Fatal("CustomFromBytes produced invalid character")
	}
	if hasAdjacentDuplicate(fromBytesBytes) {
		t.Fatal("CustomFromBytes produced adjacent duplicate character")
	}

	b := randomizer.Word.CustomBytes([]byte(dict), n)
	if len(b) != n {
		t.Fatalf("CustomBytes length = %d, want %d", len(b), n)
	}
	if !allInAlphabet(b, allow) {
		t.Fatal("CustomBytes produced invalid character")
	}
	if hasAdjacentDuplicate(b) {
		t.Fatal("CustomBytes produced adjacent duplicate character")
	}

	fromString := randomizer.Word.CustomBytesFromString(dict, n)
	if len(fromString) != n {
		t.Fatalf("CustomBytesFromString length = %d, want %d", len(fromString), n)
	}
	if !allInAlphabet(fromString, allow) {
		t.Fatal("CustomBytesFromString produced invalid character")
	}
	if hasAdjacentDuplicate(fromString) {
		t.Fatal("CustomBytesFromString produced adjacent duplicate character")
	}
}

func TestWordCustomInvalidOutput(t *testing.T) {
	if got := randomizer.Word.Custom("AB", 0); got != "" {
		t.Fatalf("Custom zero length = %q, want empty string", got)
	}
	if got := randomizer.Word.Custom("", 1); got != "" {
		t.Fatalf("Custom empty dictionary = %q, want empty string", got)
	}
	if got := randomizer.Word.Custom("AA", 2); got != "" {
		t.Fatalf("Custom duplicate-only dictionary = %q, want empty string", got)
	}
	if got := randomizer.Word.Custom("AA", 1); len(got) != 1 || got[0] != 'A' {
		t.Fatalf("Custom duplicate-only dictionary length 1 = %q, want %q", got, "A")
	}
	if got := randomizer.Word.CustomBytes([]byte("AB"), 0); got != nil {
		t.Fatalf("CustomBytes zero length = %v, want nil", got)
	}
	if got := randomizer.Word.CustomBytes(nil, 1); got != nil {
		t.Fatalf("CustomBytes nil dictionary = %v, want nil", got)
	}
	if got := randomizer.Word.CustomBytes([]byte("AA"), 2); got != nil {
		t.Fatalf("CustomBytes duplicate-only dictionary = %v, want nil", got)
	}
	if got := randomizer.Word.CustomBytes([]byte("AA"), 1); len(got) != 1 || got[0] != 'A' {
		t.Fatalf("CustomBytes duplicate-only dictionary length 1 = %v, want %v", got, []byte("A"))
	}
}

func TestWordAppendCustom(t *testing.T) {
	const n = 256
	const dict = "ABC123"
	allow := makeAlphabet(dict)

	var buf [n + 4]byte
	dst := append(buf[:0], "key:"...)
	out := randomizer.Word.AppendCustom(dst, dict, n)
	if len(out) != n+4 || string(out[:4]) != "key:" || &out[4] != &buf[4] {
		t.Fatal("AppendCustom did not preserve prefix or reuse caller buffer")
	}
	if !allInAlphabet(out[4:], allow) {
		t.Fatal("AppendCustom produced invalid character")
	}
	if hasAdjacentDuplicate(out[4:]) {
		t.Fatal("AppendCustom produced adjacent duplicate character")
	}

	fromBytes := randomizer.Word.AppendCustomFromBytes(dst[:4], []byte(dict), n)
	if len(fromBytes) != n+4 || string(fromBytes[:4]) != "key:" || &fromBytes[4] != &buf[4] {
		t.Fatal("AppendCustomFromBytes did not preserve prefix or reuse caller buffer")
	}
	if !allInAlphabet(fromBytes[4:], allow) {
		t.Fatal("AppendCustomFromBytes produced invalid character")
	}
	if hasAdjacentDuplicate(fromBytes[4:]) {
		t.Fatal("AppendCustomFromBytes produced adjacent duplicate character")
	}

	invalid := randomizer.Word.AppendCustom([]byte("x"), "AA", 2)
	if string(invalid) != "x" {
		t.Fatalf("AppendCustom invalid dictionary = %q, want %q", invalid, "x")
	}
}

func BenchmarkWordDecimal(b *testing.B) {
	const n = 256
	b.ReportAllocs()
	for b.Loop() {
		benchWordString = randomizer.Word.String(randomizer.DecimalAlphabet, n)
	}
}

func BenchmarkWordDecimalBytes(b *testing.B) {
	const n = 256
	b.ReportAllocs()
	for b.Loop() {
		benchWordBytes = randomizer.Word.Bytes(randomizer.DecimalAlphabet, n)
	}
}

func BenchmarkWordHex(b *testing.B) {
	const n = 256
	b.ReportAllocs()
	for b.Loop() {
		benchWordString = randomizer.Word.String(randomizer.HexLowerAlphabet, n)
	}
}

func BenchmarkWordHexBytes(b *testing.B) {
	const n = 256
	b.ReportAllocs()
	for b.Loop() {
		benchWordBytes = randomizer.Word.Bytes(randomizer.HexLowerAlphabet, n)
	}
}

func BenchmarkWordOctal(b *testing.B) {
	const n = 256
	b.ReportAllocs()
	for b.Loop() {
		benchWordString = randomizer.Word.String(randomizer.OctalAlphabet, n)
	}
}

func BenchmarkWordOctalBytes(b *testing.B) {
	const n = 256
	b.ReportAllocs()
	for b.Loop() {
		benchWordBytes = randomizer.Word.Bytes(randomizer.OctalAlphabet, n)
	}
}

func BenchmarkWordAlphaNumeric(b *testing.B) {
	const n = 256
	b.ReportAllocs()
	for b.Loop() {
		benchWordString = randomizer.Word.String(randomizer.AlphaNumericAlphabet, n)
	}
}

func BenchmarkWordAppendAlphaNumeric(b *testing.B) {
	const n = 256
	buf := make([]byte, 0, n)
	b.ReportAllocs()
	for b.Loop() {
		benchWordBytes = randomizer.Word.Append(buf[:0], randomizer.AlphaNumericAlphabet, n)
	}
}

func BenchmarkWordBase64URL(b *testing.B) {
	const n = 256
	b.ReportAllocs()
	for b.Loop() {
		benchWordString = randomizer.Word.String(randomizer.Base64URLAlphabet, n)
	}
}

func BenchmarkWordLower(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchWordString = randomizer.Word.String(randomizer.LowerAlphabet, 256)
	}
}

func BenchmarkWordAlphaNumeric16(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchWordString = randomizer.Word.String(randomizer.AlphaNumericAlphabet, 16)
	}
}

func BenchmarkWordCustom(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchWordString = randomizer.Word.Custom("AABCXYZ9", 256)
	}
}

func BenchmarkWordAppendAlphaNumericParallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var buf [256]byte
		for pb.Next() {
			_ = randomizer.Word.Append(buf[:0], randomizer.AlphaNumericAlphabet, 256)
		}
	})
}

func TestWordSmallDictionaries(t *testing.T) {
	if got := randomizer.Word.Custom("A", 1); got != "A" {
		t.Fatalf("Custom(A, 1) = %q, want A", got)
	}
	ab := randomizer.Word.Custom("AB", 64)
	if len(ab) != 64 || hasAdjacentDuplicate([]byte(ab)) || !strings.Contains(ab, "A") || !strings.Contains(ab, "B") {
		t.Fatalf("Custom(AB, 64) = %q, want alternating A and B", ab)
	}
	for range 1000 {
		if got := randomizer.Word.Custom("AAB", 2); got == "AA" || len(got) != 2 {
			t.Fatalf("Custom(AAB, 2) = %q", got)
		}
	}
	abc := randomizer.Word.Custom("ABC", 4096)
	if hasAdjacentDuplicate([]byte(abc)) || !allInAlphabet([]byte(abc), makeAlphabet("ABC")) {
		t.Fatal("Custom(ABC) produced invalid output")
	}
}

func TestWordWeightedDictionaryNoAdjacent(t *testing.T) {
	const n = 100000
	out := randomizer.Word.CustomBytes([]byte("AABC"), n)
	if len(out) != n || hasAdjacentDuplicate(out) || !allInAlphabet(out, makeAlphabet("ABC")) {
		t.Fatal("CustomBytes(AABC) produced invalid output")
	}
	var counts [256]int
	for _, c := range out {
		counts[c]++
	}
	// After A the next byte is B or C with equal odds; after B or C the
	// next byte is A with probability 2/3. The stationary distribution of
	// that chain is A 40%, B 30%, C 30%.
	within := func(count, want int) bool { return count > want*97/100 && count < want*103/100 }
	if !within(counts['A'], n*4/10) || !within(counts['B'], n*3/10) || !within(counts['C'], n*3/10) {
		t.Fatalf("CustomBytes(AABC) counts A=%d B=%d C=%d, want 40%%/30%%/30%%", counts['A'], counts['B'], counts['C'])
	}
}

func TestWordLargeDictionary(t *testing.T) {
	dict := make([]byte, 300)
	for i := range dict {
		dict[i] = byte(i)
	}
	out := randomizer.Word.CustomBytes(dict, 4096)
	if len(out) != 4096 || hasAdjacentDuplicate(out) {
		t.Fatal("CustomBytes(300-byte dictionary) produced invalid output")
	}
}

func TestWordUniformity(t *testing.T) {
	const n = 200000
	cases := []struct {
		alphabet randomizer.Alphabet
		chars    string
	}{
		{randomizer.DecimalAlphabet, "0123456789"},
		{randomizer.LowerAlphabet, "abcdefghijklmnopqrstuvwxyz"},
		{randomizer.AlphaAlphabet, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{randomizer.AlphaNumericAlphabet, "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{randomizer.HexLowerAlphabet, "0123456789abcdef"},
	}
	for _, tc := range cases {
		var counts [256]int
		for _, c := range randomizer.Word.Bytes(tc.alphabet, n) {
			counts[c]++
		}
		mean := n / len(tc.chars)
		for i := 0; i < len(tc.chars); i++ {
			if c := counts[tc.chars[i]]; c < mean*9/10 || c > mean*11/10 {
				t.Fatalf("alphabet %v symbol %q count %d, want within 10%% of %d", tc.alphabet, tc.chars[i], c, mean)
			}
		}
	}
}

func TestWordFixedProviderLanes(t *testing.T) {
	// All-one lanes select the last position, then alternate with the one
	// before it through the skip mapping.
	previous := randomizer.SetProvider(fixedProvider(^uint64(0)))
	defer randomizer.SetProvider(previous)
	if got := randomizer.Word.String(randomizer.DecimalAlphabet, 6); got != "989898" {
		t.Fatalf("Decimal with all-one lanes = %q, want 989898", got)
	}
	if got := randomizer.Word.Custom("xyz", 5); got != "zyzyz" {
		t.Fatalf("Custom(xyz) with all-one lanes = %q, want zyzyz", got)
	}
}
