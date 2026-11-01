// Copyright 2026 ALTESSA SOLUTIONS INC.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

package rmmarks_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/altessa-s/rmmarks-go"
)

func TestBoxesGoIntoAFrameAndBack(t *testing.T) {
	frame, cut, page := b(.25, .75, .25, .75), b(0, .5, 0, 1), b(.25, .5, .25, .75)
	if got := cut.Within(frame); !boxEqual(got, page) {
		t.Errorf("within: %v, want %v", got, page)
	}
	if got := page.Inside(frame); !boxEqual(got, cut) {
		t.Errorf("inside: %v, want %v", got, cut)
	}
	frame, page = b(.1, .8, .2, .6), b(.3, .5, .3, .4)
	if got := page.Inside(frame).Within(frame); !boxEqual(got, page) {
		t.Errorf("round trip: %v, want %v", got, page)
	}
}

func TestAFlatFrameLeavesTheBoxAlone(t *testing.T) {
	page, flat := b(.3, .5, .3, .4), b(.5, .5, .5, .5)
	if got := page.Inside(flat); !boxEqual(got, page) {
		t.Errorf("inside a flat frame: %v", got)
	}
	if got := page.Within(flat); !boxEqual(got, flat) {
		t.Errorf("within a flat frame: %v", got)
	}
	if got := page.Inside(b(.8, .2, 0, 1)); !boxEqual(got, page) {
		t.Errorf("inside an inverted frame: %v", got)
	}
	if got := b(-1, 2, -1, 2).Within(b(.1, .8, .2, .6)); got.X0 >= 0 {
		t.Errorf("a box past its frame is clamped: %v", got)
	}
}

func TestAPageWithoutMarksIsEmpty(t *testing.T) {
	var none *rmmarks.MarkedPage
	if !none.IsEmpty() {
		t.Error("a nil page is not empty")
	}
	if !(&rmmarks.MarkedPage{}).IsEmpty() {
		t.Error("a page without marks is not empty")
	}
	band := rmmarks.MarkBand{Box: b(.25, .5, .25, .75), Color: 0xFEFD60}
	if band.Item != nil {
		t.Error("an item by default")
	}
	if (&rmmarks.MarkedPage{Bands: []rmmarks.MarkBand{band}}).IsEmpty() {
		t.Error("a page with a band is empty")
	}
	highlight := rmmarks.MarkedHighlight{Text: "слово", Band: band}
	if (&rmmarks.MarkedPage{Highlights: []rmmarks.MarkedHighlight{highlight}}).IsEmpty() {
		t.Error("a page with a highlight is empty")
	}
}

func TestARecognizedLineIsPlacedAsOnePiece(t *testing.T) {
	line := rmmarks.RecognizedLine{Text: "Здесь отмечено одно", Box: b(.25, .5, .25, .75)}
	placed := rmmarks.PlaceLine(line)
	if placed.Text != line.Text || !boxEqual(placed.Box, line.Box) {
		t.Errorf("placed: %+v, want %+v", placed, line)
	}
}

func TestConstantsAreTheLibrarys(t *testing.T) {
	version, thresholds := rmmarks.CConstants()
	if version != rmmarks.Version {
		t.Errorf("version: C %q, Go %q", version, rmmarks.Version)
	}
	want := [5]float64{rmmarks.PastTheMiddle, rmmarks.WholeLine, rmmarks.LeastWidth, rmmarks.LeastTouch, rmmarks.WithinALine}
	if thresholds != want {
		t.Errorf("thresholds: C %v, Go %v", thresholds, want)
	}
	vendored, err := os.ReadFile(filepath.Join("internal", "clib", "VERSION"))
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(vendored)); got != "v"+rmmarks.Version {
		t.Errorf("vendored tree is %q, Version %q", got, rmmarks.Version)
	}
}

func TestStatusesHaveNames(t *testing.T) {
	for _, s := range []rmmarks.Status{rmmarks.StatusOK, rmmarks.StatusInvalidArgument, rmmarks.StatusOutOfMemory, rmmarks.StatusInvalidUTF8} {
		if name := s.String(); name == "" || name == "unknown status" {
			t.Errorf("status %d is called %q", int(s), name)
		}
	}
	if got := rmmarks.Status(999).String(); got != "unknown status" {
		t.Errorf("status 999 is called %q", got)
	}
}

func TestErrorIsMatchesByStatus(t *testing.T) {
	err := error(&rmmarks.Error{Status: rmmarks.StatusInvalidUTF8, Message: "anything"})
	if !errors.Is(err, rmmarks.ErrInvalidUTF8) {
		t.Error("does not match its sentinel")
	}
	if errors.Is(err, rmmarks.ErrInvalidArgument) {
		t.Error("matches another sentinel")
	}
	if !errors.Is(fmt.Errorf("wrapped: %w", err), rmmarks.ErrInvalidUTF8) {
		t.Error("does not match through a wrap")
	}
	var none *rmmarks.Error
	if none.Is(rmmarks.ErrInvalidUTF8) || rmmarks.ErrInvalidUTF8.Is(none) {
		t.Error("a nil *Error matches")
	}
	if got := none.Error(); got != "rmmarks: <nil>" {
		t.Errorf("nil spells %q", got)
	}
	if got := err.Error(); got != "rmmarks: invalid UTF-8: anything" {
		t.Errorf("spells %q", got)
	}
	if got := rmmarks.ErrOutOfMemory.Error(); got != "rmmarks: out of memory" {
		t.Errorf("sentinel spells %q", got)
	}
}
