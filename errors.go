// Copyright 2026 ALTESSA SOLUTIONS INC.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

package rmmarks

/*
#include "rmmarks/rmmarks.h"
*/
import "C"

import "strings"

// Status says what went wrong. The numbers are the C library's.
type Status int

// The statuses a call can answer with.
const (
	// StatusOK: no error. No *Error carries it.
	StatusOK Status = iota
	// StatusInvalidArgument: a box that is not finite or whose extents are
	// not, duplicate band keys among the separate readings, a negative band
	// key, or a coordinate transform that overflows.
	StatusInvalidArgument
	// StatusOutOfMemory: an allocation failed, or a count or a length is
	// larger than a Go int can hold.
	StatusOutOfMemory
	// StatusInvalidUTF8: an input text is not well-formed UTF-8.
	StatusInvalidUTF8
)

// String answers the status as a bare English phrase, and a phrase of its own
// for a number the C library does not name.
func (s Status) String() string {
	return C.GoString(C.rmm_status_string(C.rmm_status(s)))
}

// Error is what every call of this package fails with.
type Error struct {
	// Status says which failure this is.
	Status Status
	// Message is what went wrong in words: the C library's phrase, or what
	// the Go side refused.
	Message string
}

// Sentinels for [errors.Is]: an *Error matches the sentinel of its Status,
// whatever its message.
var (
	ErrInvalidArgument = &Error{Status: StatusInvalidArgument}
	ErrOutOfMemory     = &Error{Status: StatusOutOfMemory}
	ErrInvalidUTF8     = &Error{Status: StatusInvalidUTF8}
)

// Error spells the status, then the message when it says more.
func (e *Error) Error() string {
	if e == nil {
		return "rmmarks: <nil>"
	}
	var b strings.Builder
	b.WriteString("rmmarks: ")
	status := e.Status.String()
	b.WriteString(status)
	if e.Message != "" && e.Message != status {
		b.WriteString(": ")
		b.WriteString(e.Message)
	}
	return b.String()
}

// Is reports whether target is an *Error of the same Status. A nil *Error,
// on either side, matches nothing.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	return ok && t != nil && e != nil && t.Status == e.Status
}

// errorOf turns a C status and the error record beside it into a Go error,
// nil for success.
func errorOf(status C.rmm_status, cerr *C.rmm_error) error {
	if status == C.RMM_OK {
		return nil
	}
	return &Error{Status: Status(status), Message: C.GoString(&cerr.message[0])}
}

// errInvalid answers an argument the Go side refuses before C sees it.
func errInvalid(message string) error {
	return &Error{Status: StatusInvalidArgument, Message: message}
}

// errTooLarge answers a count or a length C states and a Go int cannot hold.
// No Go value of that size can be allocated, so it is reported the way a
// failed allocation is.
func errTooLarge(what string) error {
	return &Error{Status: StatusOutOfMemory, Message: what + " too large for a Go int"}
}
