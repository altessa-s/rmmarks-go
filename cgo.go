// Copyright 2026 ALTESSA SOLUTIONS INC.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

package rmmarks

// -std=c2x and not -std=c23: GCC 13 knows the standard only by its draft
// name, while GCC 14 and clang 19 or newer take both. The library spells its
// thresholds with constexpr, which clang 18 does not read, so no flag rescues
// an older compiler.

/*
#cgo CFLAGS: -std=c2x -I${SRCDIR}/internal/clib/include -I${SRCDIR}/internal/clib/src
#cgo linux LDFLAGS: -lm

#include "rmmarks/rmmarks.h"

static const char *rmmgo_version(void) { return RMMARKS_VERSION_STRING; }
static double rmmgo_past_the_middle(void) { return RMM_PAST_THE_MIDDLE; }
static double rmmgo_whole_line(void) { return RMM_WHOLE_LINE; }
static double rmmgo_least_width(void) { return RMM_LEAST_WIDTH; }
static double rmmgo_least_touch(void) { return RMM_LEAST_TOUCH; }
static double rmmgo_within_a_line(void) { return RMM_WITHIN_A_LINE; }
*/
import "C"

// Version is the version of the vendored rmmarks-lib.
const Version = "1.0.1"

// The thresholds the matching rules are written in, as fractions.
const (
	// PastTheMiddle is the slack, as a share of the line's height, by which a
	// band may miss the line's vertical middle and still cover it.
	PastTheMiddle = 0.05
	// WholeLine is the share of a line's width a band covers to select the
	// whole line.
	WholeLine = 0.85
	// LeastWidth is the share of a word's width a band covers to select the
	// word.
	LeastWidth = 0.5
	// LeastTouch is the share of the narrower of word and band two must share
	// for a touched word to count, when no word is half covered.
	LeastTouch = 0.2
	// WithinALine is the gap, as a share of an incoming passage's height,
	// across which [Blocks] joins it to a block.
	WithinALine = 0.75
)

// cConstants answers the C library's own spelling of the constants above, so
// that a test can hold the two together.
func cConstants() (version string, thresholds [5]float64) {
	return C.GoString(C.rmmgo_version()), [5]float64{
		float64(C.rmmgo_past_the_middle()),
		float64(C.rmmgo_whole_line()),
		float64(C.rmmgo_least_width()),
		float64(C.rmmgo_least_touch()),
		float64(C.rmmgo_within_a_line()),
	}
}
