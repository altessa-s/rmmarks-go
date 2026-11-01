// Copyright 2026 ALTESSA SOLUTIONS INC.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

// Package rmmarks is geometry for marked text: where a highlight lies on a
// page, which words it covers, and how text recognized on a stack of cropped
// strips returns to the page. It reads no file and runs no recognition; it
// takes boxes and text and answers boxes and text.
//
// The package binds rmmarks-lib, a C23 library vendored into this module and
// compiled by cgo; building it needs a C23 compiler (clang 19 or newer, GCC 13
// or newer) and nothing else. Every value the API returns is a Go value: the
// C memory behind a call is released before the call returns, and nothing the
// caller holds points into it.
//
// # Coordinates and text
//
// A [PageBox] is in fractions of the page, origin top left, and is never
// clamped to [0, 1]. Text is UTF-8 and crosses to C with its length, so a NUL
// byte is text like any other. Ill-formed UTF-8 is not repaired: the call
// fails with [ErrInvalidUTF8].
//
// # Operations
//
// [Passages] finds the text each band marks; [PassagesFromReading] does the
// same for whole recognized lines. [Blocks] gathers neighbouring passages
// into boxes to crop, and [Brought] carries text recognized on the cropped
// strips back to the page, as readings [Passages] takes.
//
// # Errors
//
// Every failure is an [*Error]. Its [Status] says which, and the sentinel
// values ([ErrInvalidArgument], [ErrOutOfMemory], [ErrInvalidUTF8]) match it
// through [errors.Is]. A count or a length the C library states and a Go int
// cannot hold is [ErrOutOfMemory]: no Go value of that size can be allocated.
//
// # Concurrency
//
// The C library holds no mutable global state and emits no logs. Each call
// builds its arguments in C memory of its own and frees them and the result
// before it returns, so any number of calls may run at once, from any
// goroutine. Inputs are read, never changed.
package rmmarks
