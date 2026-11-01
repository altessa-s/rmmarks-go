// Copyright 2026 ALTESSA SOLUTIONS INC.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

package rmmarks

/*
#include <stdlib.h>
#include "rmmarks/rmmarks.h"
*/
import "C"

import (
	"math"
	"unsafe"
)

// cMemory is the C memory one call builds its arguments in. The arguments are
// arrays of records that point at further arrays and at text, and cgo forbids
// handing C Go memory that holds Go pointers, so every array and every string
// lives in C memory and is freed as one when the call returns.
type cMemory struct {
	blocks []unsafe.Pointer
	// err is the first allocation that failed; once set, nothing more is
	// allocated and the call must not reach C.
	err error
}

// alloc answers n zeroed elements of the given size, or nil when n is zero or
// C is out of memory, in which case err is set.
func (m *cMemory) alloc(n int, size uintptr) unsafe.Pointer {
	if m.err != nil || n == 0 {
		return nil
	}
	p := C.calloc(C.size_t(n), C.size_t(size))
	if p == nil {
		m.err = &Error{Status: StatusOutOfMemory, Message: "C memory for the arguments"}
		return nil
	}
	m.blocks = append(m.blocks, p)
	return p
}

func (m *cMemory) free() {
	for _, p := range m.blocks {
		C.free(p)
	}
	m.blocks = nil
}

// cArray answers n zeroed C records, or nil when n is zero or allocation
// failed. Ranging over the answer is safe either way.
func cArray[T any](m *cMemory, n int) []T {
	var zero T
	p := m.alloc(n, unsafe.Sizeof(zero))
	if p == nil {
		return nil
	}
	return unsafe.Slice((*T)(p), n)
}

// first answers the address of the first element, nil for an empty slice: a
// null array is what C takes with a count of zero.
func first[T any](s []T) *T {
	if len(s) == 0 {
		return nil
	}
	return &s[0]
}

// str copies s into C memory as counted bytes, embedded NULs included. The
// bytes cross as they are: ill-formed UTF-8 is C's to refuse.
func (m *cMemory) str(s string) C.rmm_str {
	p := m.alloc(len(s), 1)
	if p == nil {
		return C.rmm_str{}
	}
	copy(unsafe.Slice((*byte)(p), len(s)), s)
	return C.rmm_str{bytes: (*C.char)(p), length: C.size_t(len(s))}
}

func (m *cMemory) boxes(boxes []PageBox) *C.rmm_page_box {
	out := cArray[C.rmm_page_box](m, len(boxes))
	for i := range out {
		out[i] = boxes[i].toC()
	}
	return first(out)
}

func (m *cMemory) lines(lines []Line) *C.rmm_line {
	out := cArray[C.rmm_line](m, len(lines))
	for i := range out {
		pieces := cArray[C.rmm_placed_text](m, len(lines[i]))
		for j := range pieces {
			pieces[j] = C.rmm_placed_text{text: m.str(lines[i][j].Text), box: lines[i][j].Box.toC()}
		}
		out[i] = C.rmm_line{pieces: first(pieces), count: C.size_t(len(lines[i]))}
	}
	return first(out)
}

// goLength converts a C count or length to a Go int, refusing one past
// math.MaxInt.
func goLength(n C.size_t, what string) (int, error) {
	if uint64(n) > math.MaxInt {
		return 0, errTooLarge(what)
	}
	return int(n), nil
}

// converter copies a C result into Go values. The first count it cannot
// represent is kept in err, and everything after it is left empty.
type converter struct {
	err error
}

func (c *converter) length(n C.size_t, what string) int {
	if c.err != nil {
		return 0
	}
	v, err := goLength(n, what)
	if err != nil {
		c.err = err
	}
	return v
}

// str copies counted C text into a Go string, by its length and not up to a
// NUL.
func (c *converter) str(s C.rmm_str, what string) string {
	n := c.length(s.length, what)
	if n == 0 || s.bytes == nil {
		return ""
	}
	return string(unsafe.Slice((*byte)(unsafe.Pointer(s.bytes)), n))
}

// optional copies text C marks absent with a null pointer.
func (c *converter) optional(s C.rmm_str, what string) *string {
	if s.bytes == nil {
		return nil
	}
	v := c.str(s, what)
	return &v
}

func (c *converter) lines(lines *C.rmm_line, count C.size_t) []Line {
	n := c.length(count, "line count")
	if n == 0 || lines == nil {
		return nil
	}
	out := make([]Line, n)
	for i, line := range unsafe.Slice(lines, n) {
		k := c.length(line.count, "piece count")
		if k == 0 || line.pieces == nil {
			continue
		}
		pieces := make(Line, k)
		for j, p := range unsafe.Slice(line.pieces, k) {
			pieces[j] = PlacedText{Text: c.str(p.text, "text length"), Box: boxFromC(p.box)}
		}
		out[i] = pieces
	}
	return out
}
