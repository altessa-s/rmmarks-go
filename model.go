// Copyright 2026 ALTESSA SOLUTIONS INC.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

package rmmarks

/*
#include "rmmarks/rmmarks.h"
*/
import "C"

// PageBox is a rectangle in fractions of the page, origin top left. It is
// never clamped to [0, 1].
type PageBox struct {
	X0, Y0, X1, Y1 float64
}

func (b PageBox) toC() C.rmm_page_box {
	return C.rmm_page_box{x0: C.double(b.X0), y0: C.double(b.Y0), x1: C.double(b.X1), y1: C.double(b.Y1)}
}

func boxFromC(b C.rmm_page_box) PageBox {
	return PageBox{X0: float64(b.x0), Y0: float64(b.y0), X1: float64(b.x1), Y1: float64(b.y1)}
}

// Inside answers b in the coordinates of frame: frame's top left is 0 and its
// bottom right is 1. A frame with no positive width or height answers b
// unchanged. Nothing is checked; untrusted geometry may answer infinities.
func (b PageBox) Inside(frame PageBox) PageBox {
	return boxFromC(C.rmm_box_inside(b.toC(), frame.toC()))
}

// Within is the inverse of [PageBox.Inside]: b, given in the coordinates of
// frame, in the coordinates frame is given in. Nothing is checked.
func (b PageBox) Within(frame PageBox) PageBox {
	return boxFromC(C.rmm_box_within(b.toC(), frame.toC()))
}

// PlacedText is a piece of recognized text, a word or a whole line, and
// where it stands on the page.
type PlacedText struct {
	// Text is UTF-8 and may hold NUL bytes.
	Text string
	// Box is where the text stands.
	Box PageBox
}

// RecognizedLine is one line of text recognition, read as a whole.
type RecognizedLine struct {
	// Text is UTF-8 and may hold NUL bytes.
	Text string
	// Box is where the line stands.
	Box PageBox
}

// PlaceLine answers the line as one piece of placed text.
func PlaceLine(line RecognizedLine) PlacedText {
	// The C function answers the same text view it is given; only the
	// geometry crosses, and the Go string is kept as it is.
	placed := C.rmm_place_line(C.rmm_recognized_line{box: line.Box.toC()})
	return PlacedText{Text: line.Text, Box: boxFromC(placed.box)}
}

// Line is one line of a page as its pieces, in reading order.
type Line []PlacedText

// MarkBand is one highlighter band drawn on a page.
type MarkBand struct {
	// Box is where the band lies.
	Box PageBox
	// Color is the resolved RGB, 0xRRGGBB, not a palette index.
	Color uint32
	// Item names the item the band was drawn as; nil when unknown.
	Item *string
}

// MarkStrip is one strip of a stacked picture: where its pixels came from on
// the page, and where they stand in the picture.
type MarkStrip struct {
	OnPage, InSheet PageBox
}

// MarkedHighlight is a highlight the page carries with its text.
type MarkedHighlight struct {
	Text string
	Band MarkBand
}

// MarkedPage is what one page holds of marks.
type MarkedPage struct {
	Bands      []MarkBand
	Highlights []MarkedHighlight
	// PageID names the page; nil when unknown.
	PageID *string
}

// IsEmpty reports whether the page holds neither bands nor highlights. A nil
// page is empty.
func (p *MarkedPage) IsEmpty() bool {
	if p == nil {
		return bool(C.rmm_page_is_empty(nil))
	}
	// The C function reads the two counts and nothing else, so only they
	// cross; the record holds no pointer.
	page := C.rmm_marked_page{
		band_count:      C.size_t(len(p.Bands)),
		highlight_count: C.size_t(len(p.Highlights)),
	}
	return bool(C.rmm_page_is_empty(&page))
}

// MarkedPassage is the text one band marks.
type MarkedPassage struct {
	// Text is the selected pieces, joined by one space and trimmed at the
	// ends.
	Text string
	// Band is the index of the band in the call's bands.
	Band int
	// Place is the union of the selected pieces' boxes.
	Place PageBox
	// Second is the band's separate reading where it differs from the page's;
	// nil where the two agree or only one of them found words.
	Second *string
	// Again holds the later bands that selected the same pieces, in order;
	// nil when none did.
	Again []int
}

// BandReading is a separate reading of one band: lines recognized on its own
// crop, in page coordinates.
type BandReading struct {
	// Band is the key of the band; it must not be negative.
	Band int
	// Lines are the reading's lines.
	Lines []Line
}
