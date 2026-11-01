// Copyright 2026 ALTESSA SOLUTIONS INC.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

package rmmarks_test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/altessa-s/rmmarks-go"
)

// Every operation runs on many goroutines at once over shared inputs and
// answers what it answers alone. Under -race this also shows the calls only
// read their inputs.
func TestCallsRunConcurrently(t *testing.T) {
	lines := notes()
	bands := []rmmarks.PageBox{b(.464, .700, .508, .530), b(.119, .880, .295, .319), b(.129, .289, .481, .502)}
	strips := twoStrips()
	numbered := []int{0, 2}

	run := func() ([]rmmarks.MarkedPassage, []rmmarks.PageBox, []rmmarks.BandReading, error) {
		readings, err := rmmarks.Brought(lines, strips, numbered)
		if err != nil {
			return nil, nil, nil, err
		}
		passages, err := rmmarks.Passages(bands, lines, readings)
		if err != nil {
			return nil, nil, nil, err
		}
		blocks, err := rmmarks.Blocks(passages)
		return passages, blocks, readings, err
	}
	wantPassages, wantBlocks, wantReadings, err := run()
	if err != nil {
		t.Fatal(err)
	}

	const goroutines, rounds = 16, 50
	var wg sync.WaitGroup
	errs := make(chan string, goroutines)
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range rounds {
				passages, blocks, readings, err := run()
				switch {
				case err != nil:
					errs <- err.Error()
					return
				case !reflect.DeepEqual(passages, wantPassages),
					!reflect.DeepEqual(blocks, wantBlocks),
					!reflect.DeepEqual(readings, wantReadings):
					errs <- "a concurrent call answered differently"
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}
