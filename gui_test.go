//go:build !windows

package main

import (
	"reflect"
	"testing"

	"gioui.org/font/gofont"
	"gioui.org/text"
	"golang.org/x/image/math/fixed"
)

// Gio v0.9.0 needs github.com/go-text/typesetting v0.3.0. With v0.3.4 every
// rune shapes to the .notdef glyph and the GUI draws only empty boxes.
func TestShaperFindsGlyphs(t *testing.T) {
	sh := text.NewShaper(text.WithCollection(gofont.Collection()))
	sh.LayoutString(text.Parameters{PxPerEm: fixed.I(16), MaxWidth: 1000}, "Pa5$")
	seen := map[text.GlyphID]bool{}
	for {
		g, ok := sh.NextGlyph()
		if !ok {
			break
		}
		seen[g.ID] = true
	}
	if len(seen) != 4 {
		t.Fatalf("got %d distinct glyphs for 4 distinct runes; fonts aren't shaping (check go-text/typesetting version)", len(seen))
	}
}

func TestFitWidths(t *testing.T) {
	tests := []struct {
		name  string
		pref  []int
		mins  []int
		avail int
		want  []int
	}{
		{"fits at preferred", []int{60, 80}, []int{20, 20}, 200, []int{60, 80}},
		{"exact fit", []int{60, 80}, []int{20, 20}, 140, []int{60, 80}},
		{"widest capped first", []int{60, 100}, []int{20, 20}, 140, []int{60, 80}},
		{"all capped evenly", []int{60, 100}, []int{20, 20}, 100, []int{50, 50}},
		{"minimum respected", []int{60, 100}, []int{40, 20}, 70, []int{40, 30}},
		{"minimums overflow", []int{60, 100}, []int{40, 50}, 50, []int{40, 50}},
	}
	for _, tt := range tests {
		got := fitWidths(tt.pref, tt.mins, tt.avail)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: fitWidths(%v, %v, %d) = %v, want %v", tt.name, tt.pref, tt.mins, tt.avail, got, tt.want)
		}
	}
}

func TestNetWMIconData(t *testing.T) {
	data := netWMIconData()
	if len(data) == 0 {
		t.Fatal("no embedded icons")
	}
	sizes := 0
	for i := 0; i < len(data); {
		w, h := int(data[i]), int(data[i+1])
		if w == 0 || w != h {
			t.Fatalf("icon %d: bad size %dx%d", sizes, w, h)
		}
		i += 2 + w*h
		if i > len(data) {
			t.Fatalf("icon %d (%dx%d) runs past end of data", sizes, w, h)
		}
		sizes++
	}
	if sizes < 2 {
		t.Errorf("got %d icon sizes, want several", sizes)
	}
}
