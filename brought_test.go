// Copyright 2026 ALTESSA SOLUTIONS INC.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

package rmmarks_test

import (
	"math"
	"testing"

	"github.com/altessa-s/rmmarks-go"
)

func mustBring(t *testing.T, lines []rmmarks.Line, strips []rmmarks.MarkStrip, numbered []int) []rmmarks.BandReading {
	t.Helper()
	readings, err := rmmarks.Brought(lines, strips, numbered)
	if err != nil {
		t.Fatalf("brought: %v", err)
	}
	return readings
}

func twoStrips() []rmmarks.MarkStrip {
	return []rmmarks.MarkStrip{
		{OnPage: b(.1, .9, .1, .2), InSheet: b(0, 1, 0, .4)},
		{OnPage: b(.1, .9, .5, .6), InSheet: b(0, 1, .6, 1)},
	}
}

func TestTheWholeLinesMiddleChoosesTheStrip(t *testing.T) {
	lines := each(w("вторая", 0, 1, .35, .95), w("ничья", 0, 1, .45, .55))
	got := mustBring(t, lines, twoStrips(), []int{3, 7})
	if len(got) != 1 || got[0].Band != 7 || len(got[0].Lines) != 1 {
		t.Fatalf("readings %+v", got)
	}
	p := got[0].Lines[0][0]
	if p.Text != "вторая" || math.Abs(p.Box.Y0-.4375) > 1e-12 || math.Abs(p.Box.X1-.9) > 1e-12 {
		t.Errorf("piece %+v", p)
	}
}

func TestALineStaysWhole(t *testing.T) {
	strips := twoStrips()
	strips[0].InSheet.Y1 = .5
	strips[1].InSheet.Y0 = .5
	line := []rmmarks.Line{{w("то", .1, .4, .40, .48), w("ру", .5, .8, .44, .72)}}
	got := mustBring(t, line, strips, []int{3, 7})
	if len(got) != 1 || got[0].Band != 7 || len(got[0].Lines[0]) != 2 || got[0].Lines[0][1].Text != "ру" {
		t.Errorf("readings %+v", got)
	}
	if got := mustBring(t, line, strips, []int{3}); len(got) != 0 {
		t.Errorf("an unnumbered strip is read: %+v", got)
	}
	if got := mustBring(t, line, nil, []int{3, 7}); len(got) != 0 {
		t.Errorf("no strips answer %+v", got)
	}
	boundary := []rmmarks.Line{{w("boundary", .1, .4, .49, .51)}}
	if got := mustBring(t, boundary, strips, []int{3, 7}); len(got) != 1 || got[0].Band != 3 {
		t.Errorf("the first strip does not win: %+v", got)
	}
}

func TestStripsMayShareABand(t *testing.T) {
	strips := twoStrips()
	strips[0].InSheet.Y1 = .5
	strips[1].InSheet.Y0 = .5
	got := mustBring(t, each(w("a", 0, 1, .1, .2), w("b", 0, 1, .7, .8)), strips, []int{3, 3})
	if len(got) != 1 || len(got[0].Lines) != 2 || got[0].Lines[1][0].Text != "b" {
		t.Errorf("readings %+v", got)
	}
}

func TestBroughtReadingsAreReadingsOfBands(t *testing.T) {
	strips := twoStrips()
	readings := mustBring(t, each(w("слово", 0, 1, .1, .3)), strips, []int{0, 1})
	bands := []rmmarks.PageBox{strips[0].OnPage, strips[1].OnPage}
	got := mustPassages(t, bands, nil, readings)
	if len(got) != 1 || got[0].Band != 0 || got[0].Text != "слово" {
		t.Errorf("passages %+v", got)
	}
}

func TestNothingToBring(t *testing.T) {
	if got := mustBring(t, nil, nil, nil); len(got) != 0 {
		t.Errorf("readings %+v", got)
	}
}
