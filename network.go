package randomizer

//go:generate go run ./internal/ouigen -o oui_table.go

type network struct{}

// Network provides random network addresses, identifiers, and port values
// using the active [Provider] selected by [SetProvider].
var Network network

// fixedBytes returns dst resliced to n bytes, or a fresh n-byte slice when dst
// lacks the capacity.
func fixedBytes(dst []byte, n int) []byte {
	if cap(dst) < n {
		return make([]byte, n)
	}
	return dst[:n]
}
