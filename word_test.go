package randomizer_test

import (
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

func TestWordZeroLength(t *testing.T) {
	if got := randomizer.Word.Decimal(0); got != "" {
		t.Fatalf("Decimal(0) = %q, want empty string", got)
	}
	if got := randomizer.Word.Hex(0, false); got != "" {
		t.Fatalf("Hex(0,false) = %q, want empty string", got)
	}
	if got := randomizer.Word.Octal(0); got != "" {
		t.Fatalf("Octal(0) = %q, want empty string", got)
	}
	if got := randomizer.Word.DecimalBytes(0); got != nil {
		t.Fatalf("DecimalBytes(0) = %v, want nil", got)
	}
	if got := randomizer.Word.HexBytes(0, false); got != nil {
		t.Fatalf("HexBytes(0,false) = %v, want nil", got)
	}
	if got := randomizer.Word.OctalBytes(0); got != nil {
		t.Fatalf("OctalBytes(0) = %v, want nil", got)
	}
}

func TestWordDecimalOutput(t *testing.T) {
	const n = 4096
	allow := makeAlphabet("0123456789")

	s := randomizer.Word.Decimal(n)
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

	b := randomizer.Word.DecimalBytes(n)
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

	sLower := randomizer.Word.Hex(n, false)
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

	sUpper := randomizer.Word.Hex(n, true)
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

	bLower := randomizer.Word.HexBytes(n, false)
	if len(bLower) != n {
		t.Fatalf("HexBytes lower length = %d, want %d", len(bLower), n)
	}
	if !allInAlphabet(bLower, allowLower) {
		t.Fatal("HexBytes lower produced invalid character")
	}
	if hasAdjacentDuplicate(bLower) {
		t.Fatal("HexBytes lower produced adjacent duplicate character")
	}

	bUpper := randomizer.Word.HexBytes(n, true)
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

	s := randomizer.Word.Octal(n)
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

	b := randomizer.Word.OctalBytes(n)
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

func BenchmarkWordDecimal(b *testing.B) {
	const n = 256
	b.ReportAllocs()
	for b.Loop() {
		benchWordString = randomizer.Word.Decimal(n)
	}
}

func BenchmarkWordDecimalBytes(b *testing.B) {
	const n = 256
	b.ReportAllocs()
	for b.Loop() {
		benchWordBytes = randomizer.Word.DecimalBytes(n)
	}
}

func BenchmarkWordHex(b *testing.B) {
	const n = 256
	b.ReportAllocs()
	for b.Loop() {
		benchWordString = randomizer.Word.Hex(n, false)
	}
}

func BenchmarkWordHexBytes(b *testing.B) {
	const n = 256
	b.ReportAllocs()
	for b.Loop() {
		benchWordBytes = randomizer.Word.HexBytes(n, false)
	}
}

func BenchmarkWordOctal(b *testing.B) {
	const n = 256
	b.ReportAllocs()
	for b.Loop() {
		benchWordString = randomizer.Word.Octal(n)
	}
}

func BenchmarkWordOctalBytes(b *testing.B) {
	const n = 256
	b.ReportAllocs()
	for b.Loop() {
		benchWordBytes = randomizer.Word.OctalBytes(n)
	}
}
