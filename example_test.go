// Copyright 2026 ALTESSA SOLUTIONS INC.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

package rmmarks_test

import (
	"errors"
	"fmt"
	"log"

	"github.com/altessa-s/rmmarks-go"
)

func ExamplePassages() {
	band := rmmarks.PageBox{X0: .1, Y0: .2, X1: .5, Y1: .24}
	line := rmmarks.Line{
		{Text: "marked", Box: rmmarks.PageBox{X0: .1, Y0: .2, X1: .3, Y1: .24}},
		{Text: "words", Box: rmmarks.PageBox{X0: .31, Y0: .2, X1: .5, Y1: .24}},
	}
	passages, err := rmmarks.Passages([]rmmarks.PageBox{band}, []rmmarks.Line{line}, nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range passages {
		fmt.Printf("band %d: %q\n", p.Band, p.Text)
	}
	// Output:
	// band 0: "marked words"
}

func ExampleBrought() {
	// Two bands cut out and stacked into one picture: the first strip is the
	// top half of the picture, the second the bottom half.
	strips := []rmmarks.MarkStrip{
		{OnPage: rmmarks.PageBox{X0: .1, Y0: .2, X1: .5, Y1: .24}, InSheet: rmmarks.PageBox{X0: 0, Y0: 0, X1: 1, Y1: .5}},
		{OnPage: rmmarks.PageBox{X0: .1, Y0: .6, X1: .5, Y1: .64}, InSheet: rmmarks.PageBox{X0: 0, Y0: .5, X1: 1, Y1: 1}},
	}
	recognized := []rmmarks.Line{
		{{Text: "first", Box: rmmarks.PageBox{X0: 0, Y0: .1, X1: 1, Y1: .4}}},
		{{Text: "second", Box: rmmarks.PageBox{X0: 0, Y0: .6, X1: 1, Y1: .9}}},
	}
	readings, err := rmmarks.Brought(recognized, strips, []int{0, 1})
	if err != nil {
		log.Fatal(err)
	}
	bands := []rmmarks.PageBox{strips[0].OnPage, strips[1].OnPage}
	passages, err := rmmarks.Passages(bands, nil, readings)
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range passages {
		fmt.Printf("band %d: %q\n", p.Band, p.Text)
	}
	// Output:
	// band 0: "first"
	// band 1: "second"
}

func ExampleBlocks() {
	passages := []rmmarks.MarkedPassage{
		{Text: "a heading", Place: rmmarks.PageBox{X0: .536, Y0: .448, X1: .873, Y1: .465}},
		{Text: "its second line", Place: rmmarks.PageBox{X0: .541, Y0: .467, X1: .871, Y1: .483}},
		{Text: "far below", Place: rmmarks.PageBox{X0: .1, Y0: .8, X1: .5, Y1: .83}},
	}
	blocks, err := rmmarks.Blocks(passages)
	if err != nil {
		log.Fatal(err)
	}
	for _, b := range blocks {
		fmt.Printf("%.3f %.3f %.3f %.3f\n", b.X0, b.Y0, b.X1, b.Y1)
	}
	// Output:
	// 0.536 0.448 0.873 0.483
	// 0.100 0.800 0.500 0.830
}

func ExamplePassages_invalidUTF8() {
	box := rmmarks.PageBox{X0: 0, Y0: 0, X1: 1, Y1: 1}
	_, err := rmmarks.Passages([]rmmarks.PageBox{box}, []rmmarks.Line{{{Text: "caf\xe9", Box: box}}}, nil)
	if errors.Is(err, rmmarks.ErrInvalidUTF8) {
		fmt.Println(err)
	}
	// Output:
	// rmmarks: invalid UTF-8
}
