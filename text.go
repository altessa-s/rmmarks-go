// Copyright 2026 ALTESSA SOLUTIONS INC.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

package rmmarks

/*
#include "rmmarks/rmmarks.h"
*/
import "C"

import (
	"strconv"
	"unsafe"
)

// Passages answers the text each band marks: at most one passage per band,
// ordered by (Place.Y0, Place.X0, Band).
//
// Lines are the page's lines, each as its words where they are known. A band
// covers a line when it reaches the line's vertical middle, within
// [PastTheMiddle] of its height. Covering [WholeLine] of the line's width
// selects the whole line; otherwise the words covered by [LeastWidth] of their
// width; failing those, the words touched by [LeastTouch] of the narrower
// span. Selected words are joined in input order by one space, and the
// passage is trimmed at its ends.
//
// Two bands selecting the same pieces of the page answer one passage, with the
// later bands in [MarkedPassage.Again]. Equal text in another place is not a
// repeat.
//
// Alone holds separate readings of single bands, keyed by band index; a key
// with no band is ignored, and two readings of one key are refused with
// [ErrInvalidArgument]. The page's reading wins; a separate reading that
// differs, other than by canonical Unicode equivalence, is kept in
// [MarkedPassage.Second], and one that alone found words becomes the passage.
// The answer of [Brought] can be passed as alone as it is.
//
// A box that is not finite is refused with [ErrInvalidArgument]; text that is
// not well-formed UTF-8, with [ErrInvalidUTF8].
func Passages(bands []PageBox, lines []Line, alone []BandReading) ([]MarkedPassage, error) {
	for i, r := range alone {
		if r.Band < 0 {
			return nil, errInvalid("alone[" + strconv.Itoa(i) + "].Band is negative")
		}
	}
	var m cMemory
	defer m.free()
	cBands := m.boxes(bands)
	cLines := m.lines(lines)
	cAlone := cArray[C.rmm_band_reading](&m, len(alone))
	for i := range cAlone {
		cAlone[i] = C.rmm_band_reading{
			band:       C.size_t(alone[i].Band),
			lines:      m.lines(alone[i].Lines),
			line_count: C.size_t(len(alone[i].Lines)),
		}
	}
	if m.err != nil {
		return nil, m.err
	}
	var out *C.rmm_passages
	var cerr C.rmm_error
	status := C.rmm_text_passages(cBands, C.size_t(len(bands)), cLines, C.size_t(len(lines)),
		first(cAlone), C.size_t(len(alone)), &out, &cerr)
	if err := errorOf(status, &cerr); err != nil {
		return nil, err
	}
	return takePassages(out)
}

// PassagesFromReading is [Passages] for a reading that knows lines and not
// their words: each line is one piece, placed by [PlaceLine].
func PassagesFromReading(bands []PageBox, lines []RecognizedLine) ([]MarkedPassage, error) {
	var m cMemory
	defer m.free()
	cBands := m.boxes(bands)
	cLines := cArray[C.rmm_recognized_line](&m, len(lines))
	for i := range cLines {
		cLines[i] = C.rmm_recognized_line{text: m.str(lines[i].Text), box: lines[i].Box.toC()}
	}
	if m.err != nil {
		return nil, m.err
	}
	var out *C.rmm_passages
	var cerr C.rmm_error
	status := C.rmm_text_passages_from_reading(cBands, C.size_t(len(bands)),
		first(cLines), C.size_t(len(lines)), &out, &cerr)
	if err := errorOf(status, &cerr); err != nil {
		return nil, err
	}
	return takePassages(out)
}

// takePassages copies a C result into Go values and releases it.
func takePassages(r *C.rmm_passages) ([]MarkedPassage, error) {
	defer C.rmm_passages_free(r)
	var c converter
	n := c.length(r.count, "passage count")
	out := make([]MarkedPassage, n)
	if n > 0 {
		for i, p := range unsafe.Slice(r.items, n) {
			out[i] = MarkedPassage{
				Text:   c.str(p.text, "text length"),
				Band:   c.length(p.band, "band index"),
				Place:  boxFromC(p.place),
				Second: c.optional(p.second, "text length"),
			}
			if k := c.length(p.again_count, "repeat count"); k > 0 && p.again != nil {
				again := make([]int, k)
				for j, band := range unsafe.Slice(p.again, k) {
					again[j] = c.length(band, "band index")
				}
				out[i].Again = again
			}
		}
	}
	if c.err != nil {
		return nil, c.err
	}
	return out, nil
}

// Blocks gathers neighbouring passages into the boxes to crop for a second
// reading. A passage joins every block it reaches: one closer vertically than
// [WithinALine] of the passage's own height, and overlapping it horizontally.
// Joining repeats until nothing more joins; the gap does not grow with the
// block. The blocks are ordered by (Y0, X0).
//
// Only each passage's Place is read. A place that is not finite is refused
// with [ErrInvalidArgument].
func Blocks(passages []MarkedPassage) ([]PageBox, error) {
	var m cMemory
	defer m.free()
	cPassages := cArray[C.rmm_marked_passage](&m, len(passages))
	for i := range cPassages {
		cPassages[i].place = passages[i].Place.toC()
	}
	if m.err != nil {
		return nil, m.err
	}
	var out *C.rmm_blocks
	var cerr C.rmm_error
	status := C.rmm_text_blocks(first(cPassages), C.size_t(len(passages)), &out, &cerr)
	if err := errorOf(status, &cerr); err != nil {
		return nil, err
	}
	defer C.rmm_blocks_free(out)
	var c converter
	n := c.length(out.count, "block count")
	if c.err != nil {
		return nil, c.err
	}
	blocks := make([]PageBox, n)
	if n > 0 {
		for i, b := range unsafe.Slice(out.items, n) {
			blocks[i] = boxFromC(b)
		}
	}
	return blocks, nil
}

// Brought takes lines recognized on a stacked picture of strips back to the
// page, as separate readings of the bands the strips were cut for.
//
// Strips and numbered are paired in order, and the shorter decides how many
// pairs there are. A line belongs to the first strip whose InSheet holds the
// vertical middle of the whole line, and its pieces are carried from InSheet
// to OnPage; a line no strip holds is dropped. Readings are in order of first
// encounter, their lines in input order, and several strips may give lines to
// one band.
//
// A negative band number, a box that is not finite or a transform that
// overflows is refused with [ErrInvalidArgument]; text that is not well-formed
// UTF-8, with [ErrInvalidUTF8].
func Brought(lines []Line, strips []MarkStrip, numbered []int) ([]BandReading, error) {
	for i, n := range numbered {
		if n < 0 {
			return nil, errInvalid("numbered[" + strconv.Itoa(i) + "] is negative")
		}
	}
	var m cMemory
	defer m.free()
	cLines := m.lines(lines)
	cStrips := cArray[C.rmm_mark_strip](&m, len(strips))
	for i := range cStrips {
		cStrips[i] = C.rmm_mark_strip{on_page: strips[i].OnPage.toC(), in_sheet: strips[i].InSheet.toC()}
	}
	cNumbered := cArray[C.size_t](&m, len(numbered))
	for i := range cNumbered {
		cNumbered[i] = C.size_t(numbered[i])
	}
	if m.err != nil {
		return nil, m.err
	}
	var out *C.rmm_readings
	var cerr C.rmm_error
	status := C.rmm_text_brought(cLines, C.size_t(len(lines)), first(cStrips), C.size_t(len(strips)),
		first(cNumbered), C.size_t(len(numbered)), &out, &cerr)
	if err := errorOf(status, &cerr); err != nil {
		return nil, err
	}
	defer C.rmm_readings_free(out)
	var c converter
	n := c.length(out.count, "reading count")
	readings := make([]BandReading, n)
	if n > 0 {
		for i, r := range unsafe.Slice(out.items, n) {
			readings[i] = BandReading{Band: c.length(r.band, "band number"), Lines: c.lines(r.lines, r.line_count)}
		}
	}
	if c.err != nil {
		return nil, c.err
	}
	return readings, nil
}
