package main

import (
	"strings"
	"testing"
)

const fixture = `Registry,Assignment,Organization Name,Organization Address
MA-L,3C22FB,"Apple, Inc.",1 Infinite Loop Cupertino CA US 95014
MA-L,001B21,Intel Corporate,"Lot 8, Jalan Hi-Tech 2/3  Kulim Kedah MY 09000 "
MA-L,000C29,"VMware, Inc.",3401 Hillview Avenue PALO ALTO CA US 94304
MA-L,005056,"VMware, Inc.",3401 Hillview Avenue PALO ALTO CA US 94304
MA-L,001C14,"VMware, Inc.",3401 Hillview Avenue PALO ALTO CA US 94304
MA-L,0A1B2C,Intel Corporate,local bit set must be skipped
MA-M,70B3D5,Intelbras,not MA-L and not an exact match
`

func TestSelectOUIs(t *testing.T) {
	rows, err := readRegistry(strings.NewReader(fixture))
	if err != nil {
		t.Fatal(err)
	}
	list := []vendor{v("Apple", `Apple, Inc\.`), v("Intel", `Intel Corporate`), v("VMware", `VMware, Inc\.`)}
	entries, err := selectOUIs(rows, list, 2)
	if err != nil {
		t.Fatal(err)
	}
	want := [][3]byte{{0x00, 0x0c, 0x29}, {0x00, 0x1b, 0x21}, {0x00, 0x1c, 0x14}, {0x3c, 0x22, 0xfb}}
	if len(entries) != len(want) {
		t.Fatalf("selected %d entries, want %d", len(entries), len(want))
	}
	for i, e := range entries {
		if e.oui != want[i] {
			t.Fatalf("entry %d = %x, want %x", i, e.oui, want[i])
		}
	}
	if _, err := selectOUIs(rows, append(list, v("Missing", `Nobody`)), 2); err == nil {
		t.Fatal("unmatched vendor did not fail")
	}
}

func TestRender(t *testing.T) {
	src, err := render([]entry{{oui: [3]byte{0x3c, 0x22, 0xfb}, vendor: "Apple", org: "Apple, Inc."}}, "u")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), "{0x3c, 0x22, 0xfb}, // Apple (Apple, Inc.)") {
		t.Fatalf("unexpected output:\n%s", src)
	}
}
