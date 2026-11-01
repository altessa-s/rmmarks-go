// Copyright 2026 ALTESSA SOLUTIONS INC.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

package rmmarks_test

import (
	"math"
	"testing"

	"github.com/altessa-s/rmmarks-go"
)

// b is a box given as its horizontal span and then its vertical one.
func b(x0, x1, y0, y1 float64) rmmarks.PageBox {
	return rmmarks.PageBox{X0: x0, Y0: y0, X1: x1, Y1: y1}
}

// w is a word and its box, given the way b takes one.
func w(text string, x0, x1, y0, y1 float64) rmmarks.PlacedText {
	return rmmarks.PlacedText{Text: text, Box: b(x0, x1, y0, y1)}
}

func boxEqual(p, q rmmarks.PageBox) bool {
	return math.Abs(p.X0-q.X0) < 1e-12 && math.Abs(p.Y0-q.Y0) < 1e-12 &&
		math.Abs(p.X1-q.X1) < 1e-12 && math.Abs(p.Y1-q.Y1) < 1e-12
}

// each makes every piece a line of its own.
func each(pieces ...rmmarks.PlacedText) []rmmarks.Line {
	lines := make([]rmmarks.Line, len(pieces))
	for i, p := range pieces {
		lines[i] = rmmarks.Line{p}
	}
	return lines
}

func mustPassages(t testing.TB, bands []rmmarks.PageBox, lines []rmmarks.Line, alone []rmmarks.BandReading) []rmmarks.MarkedPassage {
	t.Helper()
	passages, err := rmmarks.Passages(bands, lines, alone)
	if err != nil {
		t.Fatalf("passages: %v", err)
	}
	return passages
}

func texts(passages []rmmarks.MarkedPassage) []string {
	out := make([]string, len(passages))
	for i, p := range passages {
		out[i] = p.Text
	}
	return out
}
