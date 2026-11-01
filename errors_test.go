// Copyright 2026 ALTESSA SOLUTIONS INC.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

package rmmarks_test

import (
	"errors"
	"math"
	"testing"

	"github.com/altessa-s/rmmarks-go"
)

func TestInvalidInputIsRefused(t *testing.T) {
	unit := b(0, 1, 0, 1)
	word := rmmarks.Line{{Text: "word", Box: unit}}
	nan := rmmarks.Line{{Text: "word", Box: rmmarks.PageBox{X0: 0, Y0: math.NaN(), X1: 1, Y1: 1}}}
	infinite := rmmarks.PageBox{X0: 0, Y0: 0, X1: math.Inf(1), Y1: 1}
	cases := []struct {
		name string
		call func() error
		want error
	}{
		{"ill-formed text", func() error {
			_, err := rmmarks.Passages([]rmmarks.PageBox{unit}, []rmmarks.Line{{{Text: "\xff", Box: unit}}}, nil)
			return err
		}, rmmarks.ErrInvalidUTF8},
		{"ill-formed separate reading", func() error {
			alone := []rmmarks.BandReading{{Band: 0, Lines: []rmmarks.Line{{{Text: "a\xc0", Box: unit}}}}}
			_, err := rmmarks.Passages([]rmmarks.PageBox{unit}, nil, alone)
			return err
		}, rmmarks.ErrInvalidUTF8},
		{"ill-formed recognized line", func() error {
			_, err := rmmarks.PassagesFromReading([]rmmarks.PageBox{unit}, []rmmarks.RecognizedLine{{Text: "\xed\xa0\x80", Box: unit}})
			return err
		}, rmmarks.ErrInvalidUTF8},
		{"ill-formed brought line", func() error {
			_, err := rmmarks.Brought([]rmmarks.Line{{{Text: "\x80", Box: unit}}}, nil, nil)
			return err
		}, rmmarks.ErrInvalidUTF8},
		{"NaN in a word", func() error {
			_, err := rmmarks.Passages([]rmmarks.PageBox{unit}, []rmmarks.Line{nan}, nil)
			return err
		}, rmmarks.ErrInvalidArgument},
		{"band extent past a double", func() error {
			_, err := rmmarks.Passages([]rmmarks.PageBox{b(-math.MaxFloat64, math.MaxFloat64, 0, 1)}, []rmmarks.Line{word}, nil)
			return err
		}, rmmarks.ErrInvalidArgument},
		{"duplicate reading keys", func() error {
			alone := []rmmarks.BandReading{{Band: 0, Lines: []rmmarks.Line{word}}, {Band: 0, Lines: []rmmarks.Line{word}}}
			_, err := rmmarks.Passages([]rmmarks.PageBox{unit}, []rmmarks.Line{word}, alone)
			return err
		}, rmmarks.ErrInvalidArgument},
		{"negative reading key", func() error {
			_, err := rmmarks.Passages([]rmmarks.PageBox{unit}, nil, []rmmarks.BandReading{{Band: -1}})
			return err
		}, rmmarks.ErrInvalidArgument},
		{"infinite recognized box", func() error {
			_, err := rmmarks.PassagesFromReading([]rmmarks.PageBox{unit}, []rmmarks.RecognizedLine{{Text: "x", Box: infinite}})
			return err
		}, rmmarks.ErrInvalidArgument},
		{"infinite passage place", func() error {
			_, err := rmmarks.Blocks([]rmmarks.MarkedPassage{{Place: infinite}})
			return err
		}, rmmarks.ErrInvalidArgument},
		{"infinite strip", func() error {
			_, err := rmmarks.Brought([]rmmarks.Line{word}, []rmmarks.MarkStrip{{OnPage: infinite, InSheet: unit}}, []int{0})
			return err
		}, rmmarks.ErrInvalidArgument},
		{"negative band number", func() error {
			_, err := rmmarks.Brought([]rmmarks.Line{word}, []rmmarks.MarkStrip{{OnPage: unit, InSheet: unit}}, []int{-2})
			return err
		}, rmmarks.ErrInvalidArgument},
		{"transform past a double", func() error {
			wide := []rmmarks.Line{{w("wide", 0, 1e300, .1, .2)}}
			_, err := rmmarks.Brought(wide, []rmmarks.MarkStrip{{OnPage: unit, InSheet: b(0, 1e-10, 0, 1)}}, []int{0})
			return err
		}, rmmarks.ErrInvalidArgument},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.call()
			if !errors.Is(err, c.want) {
				t.Fatalf("%v, want %v", err, c.want)
			}
			var rmErr *rmmarks.Error
			if !errors.As(err, &rmErr) || rmErr.Message == "" {
				t.Errorf("%#v carries no message", err)
			}
		})
	}
}

func TestAStripPastItsPairIsNotChecked(t *testing.T) {
	// Only paired strips are read, so an unpaired one may hold anything.
	strips := []rmmarks.MarkStrip{
		{OnPage: b(0, 1, 0, 1), InSheet: b(0, 1, 0, 1)},
		{OnPage: rmmarks.PageBox{X0: math.NaN()}, InSheet: b(0, 1, 0, 1)},
	}
	if _, err := rmmarks.Brought(nil, strips, []int{0}); err != nil {
		t.Errorf("an unpaired strip is checked: %v", err)
	}
}
