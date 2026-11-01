// Copyright 2026 ALTESSA SOLUTIONS INC.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

package rmmarks_test

import (
	"errors"
	"testing"

	"github.com/altessa-s/rmmarks-go"
)

func TestPassagesAreTrimmedOfUnicodeWhitespace(t *testing.T) {
	unit := b(0, 1, 0, 1)
	for _, c := range []struct{ input, want string }{
		{"\t\r\n слово \t", "слово"},
		{"\u00A0\u1680\u2003слово\u202F\u205F\u3000", "слово"},
		{" \u0301слово \u0301", "слово"},
		{" \u200Dслово \u200D", "слово"},
		{"\u0085\u2028\u2029", ""},
		{"\u200Bword\u200B", "\u200Bword\u200B"},
		{"\u0600 word", "\u0600 word"},
		{"\n\u0301word\n\u0301", "\u0301word\n\u0301"},
	} {
		got := mustPassages(t, []rmmarks.PageBox{unit}, []rmmarks.Line{{{Text: c.input, Box: unit}}}, nil)
		switch {
		case c.want == "" && len(got) != 0:
			t.Errorf("%q: %q, want nothing", c.input, texts(got))
		case c.want != "" && (len(got) != 1 || got[0].Text != c.want):
			t.Errorf("%q: %q, want %q", c.input, texts(got), c.want)
		}
	}
}

func TestCanonicallyEquivalentReadingsAgree(t *testing.T) {
	unit := b(0, 1, 0, 1)
	for _, c := range []struct {
		page, cut string
		agree     bool
	}{
		{"café", "cafe\u0301", true},
		{"й", "и\u0306", true},
		{"각", "\u1100\u1161\u11A8", true},
		{"x\u0301\u0323", "x\u0323\u0301", true},
		{"café", "cafe", false},
		{"\uFF21", "A", false}, // canonical, not compatibility, equivalence
		{"Word", "word", false},
	} {
		alone := []rmmarks.BandReading{{Band: 0, Lines: []rmmarks.Line{{{Text: c.cut, Box: unit}}}}}
		got := mustPassages(t, []rmmarks.PageBox{unit}, []rmmarks.Line{{{Text: c.page, Box: unit}}}, alone)
		if len(got) != 1 {
			t.Fatalf("%q and %q: %d passages", c.page, c.cut, len(got))
		}
		if agree := got[0].Second == nil; agree != c.agree {
			t.Errorf("%q and %q agree: %v, want %v", c.page, c.cut, agree, c.agree)
		}
	}
}

func TestIllFormedUTF8IsRefusedNotRepaired(t *testing.T) {
	unit := b(0, 1, 0, 1)
	for _, bad := range []string{"\xC0\x80", "\xED\xA0\x80", "\xF4\x90\x80\x80", "\xE2\x82", "\xFF", "\x80"} {
		_, err := rmmarks.Passages([]rmmarks.PageBox{unit}, []rmmarks.Line{{{Text: bad, Box: unit}}}, nil)
		if !errors.Is(err, rmmarks.ErrInvalidUTF8) {
			t.Errorf("%q: %v, want ErrInvalidUTF8", bad, err)
		}
	}
	good := "😀 Привет 你好 العربية"
	got := mustPassages(t, []rmmarks.PageBox{unit}, []rmmarks.Line{{{Text: good, Box: unit}}}, nil)
	if len(got) != 1 || got[0].Text != good {
		t.Errorf("%q: %q", good, texts(got))
	}
}

func TestEmbeddedNULsCrossBothWays(t *testing.T) {
	unit := b(0, 1, 0, 1)
	withZero := "a\x00b"
	got := mustPassages(t, []rmmarks.PageBox{unit}, []rmmarks.Line{{{Text: withZero, Box: unit}}}, nil)
	if len(got) != 1 || got[0].Text != withZero {
		t.Errorf("passages %q, want %q", texts(got), withZero)
	}
	readings := mustBring(t, []rmmarks.Line{{{Text: withZero, Box: unit}}},
		[]rmmarks.MarkStrip{{OnPage: unit, InSheet: unit}}, []int{0})
	if len(readings) != 1 || readings[0].Lines[0][0].Text != withZero {
		t.Errorf("readings %+v", readings)
	}
}
