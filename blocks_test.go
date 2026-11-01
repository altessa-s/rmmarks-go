// Copyright 2026 ALTESSA SOLUTIONS INC.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

package rmmarks_test

import (
	"testing"

	"github.com/altessa-s/rmmarks-go"
)

func mustBlocks(t *testing.T, places ...rmmarks.PageBox) []rmmarks.PageBox {
	t.Helper()
	passages := make([]rmmarks.MarkedPassage, len(places))
	for i, p := range places {
		passages[i] = rmmarks.MarkedPassage{Place: p}
	}
	blocks, err := rmmarks.Blocks(passages)
	if err != nil {
		t.Fatalf("blocks: %v", err)
	}
	return blocks
}

func TestAdjoiningPassagesMakeOneBlock(t *testing.T) {
	got := mustBlocks(t, b(.536, .873, .448, .465), b(.541, .871, .467, .483))
	if len(got) != 1 || !boxEqual(got[0], b(.536, .873, .448, .483)) {
		t.Errorf("blocks %v", got)
	}
}

func TestPassagesApartStayApart(t *testing.T) {
	if got := mustBlocks(t, b(.142, .284, .475, .515), b(.445, .688, .511, .528)); len(got) != 2 {
		t.Errorf("blocks %v", got)
	}
}

func TestBlocksAreInPageOrder(t *testing.T) {
	got := mustBlocks(t, b(.1, .5, .8, .83), b(.1, .5, .2, .23))
	if len(got) != 2 || got[0].Y0 != .2 || got[1].Y0 != .8 {
		t.Errorf("blocks %v", got)
	}
}

func TestABridgeJoinsEveryBlockItReaches(t *testing.T) {
	bridged := []rmmarks.PageBox{b(0, .3, .10, .14), b(.6, .9, .15, .19), b(.2, .7, .16, .20), b(.1, .5, .70, .74)}
	got := mustBlocks(t, bridged[:3]...)
	if len(got) != 1 || !boxEqual(got[0], b(0, .9, .10, .20)) {
		t.Errorf("bridged: %v", got)
	}
	got = mustBlocks(t, bridged...)
	if len(got) != 2 {
		t.Fatalf("with one apart: %v", got)
	}
	for i := range got {
		for j := i + 1; j < len(got); j++ {
			p, q := got[i], got[j]
			if p.X0 < q.X1 && q.X0 < p.X1 && p.Y0 < q.Y1 && q.Y0 < p.Y1 {
				t.Errorf("blocks %v and %v overlap", p, q)
			}
		}
	}
}

func TestTheGapDoesNotGrowWithTheBlock(t *testing.T) {
	if got := mustBlocks(t, b(0, 1, 0, 1), b(0, 1, 1.75, 2.75)); len(got) != 2 {
		t.Errorf("blocks %v", got)
	}
}

func TestNoPassagesNoBlocks(t *testing.T) {
	if got := mustBlocks(t); len(got) != 0 {
		t.Errorf("blocks %v", got)
	}
}
