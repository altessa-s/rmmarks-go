// Copyright 2026 ALTESSA SOLUTIONS INC.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

package rmmarks_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"unicode/utf8"

	"github.com/altessa-s/rmmarks-go"
)

// FuzzPassages builds up to sixteen bands and one-word lines from arbitrary
// bytes, seeded with testdata/fuzz-seeds. The first byte is the count less
// one; each band is four coordinate bytes, a length byte and that many bytes
// of text, and the word stands in its band's box.
//
// The boxes are always finite, so the only refusal is ErrInvalidUTF8, and
// only when some word is ill-formed. What succeeds must be self-consistent:
// passages of real bands in well-formed UTF-8, blocks for them, and strips
// that carry the words back.
func FuzzPassages(f *testing.F) {
	seeds, err := filepath.Glob(filepath.Join("testdata", "fuzz-seeds", "*"))
	if err != nil {
		f.Fatal(err)
	}
	for _, path := range seeds {
		data, err := os.ReadFile(path)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 4 {
			return
		}
		n := int(data[0])%16 + 1
		at := 1
		next := func() byte {
			if at < len(data) {
				at++
				return data[at-1]
			}
			return 0
		}
		bands := make([]rmmarks.PageBox, n)
		lines := make([]rmmarks.Line, n)
		reading := make([]rmmarks.RecognizedLine, n)
		wellFormed := true
		for i := range n {
			var c [4]float64
			for j := range c {
				c[j] = float64(next())/128 - .5
			}
			bands[i] = rmmarks.PageBox{X0: c[0], Y0: c[1], X1: c[2], Y1: c[3]}
			length := 0
			if at < len(data) {
				length = int(next()) % 32
			}
			length = min(length, len(data)-at)
			text := string(data[at : at+length])
			at += length
			wellFormed = wellFormed && utf8.ValidString(text)
			lines[i] = rmmarks.Line{{Text: text, Box: bands[i]}}
			reading[i] = rmmarks.RecognizedLine{Text: text, Box: bands[i]}
		}
		refusal := func(what string, err error) bool {
			t.Helper()
			switch {
			case wellFormed && err != nil:
				t.Fatalf("%s refuses well-formed text: %v", what, err)
			case !wellFormed && !errors.Is(err, rmmarks.ErrInvalidUTF8):
				t.Fatalf("%s answers %v for ill-formed text", what, err)
			}
			return err != nil
		}

		alone := []rmmarks.BandReading{{Band: 0, Lines: lines}}
		passages, err := rmmarks.Passages(bands, lines, alone)
		if !refusal("Passages", err) {
			for _, p := range passages {
				if p.Band < 0 || p.Band >= n || p.Text == "" || !utf8.ValidString(p.Text) {
					t.Fatalf("passage %+v of %d bands", p, n)
				}
				if p.Second != nil && !utf8.ValidString(*p.Second) {
					t.Fatalf("second reading %q", *p.Second)
				}
				for _, again := range p.Again {
					if again <= p.Band || again >= n {
						t.Fatalf("repeat %d of band %d", again, p.Band)
					}
				}
			}
			blocks, err := rmmarks.Blocks(passages)
			if err != nil || len(blocks) > len(passages) {
				t.Fatalf("%d blocks of %d passages: %v", len(blocks), len(passages), err)
			}
		}

		_, err = rmmarks.PassagesFromReading(bands, reading)
		refusal("PassagesFromReading", err)

		strips := []rmmarks.MarkStrip{{OnPage: bands[0], InSheet: bands[n-1]}}
		readings, err := rmmarks.Brought(lines, strips, []int{0})
		if !refusal("Brought", err) && len(readings) > 1 {
			t.Fatalf("%d readings of one band", len(readings))
		}
	})
}
