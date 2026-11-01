// Copyright 2026 ALTESSA SOLUTIONS INC.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

package rmmarks_test

import (
	"slices"
	"testing"

	"github.com/altessa-s/rmmarks-go"
)

func notes() []rmmarks.Line {
	return []rmmarks.Line{
		{w("потом", .131, .208, .295, .327), w("ещё", .211, .278, .295, .327),
			w("нужно", .281, .410, .295, .327), w("проверить", .413, .673, .295, .327),
			w("отчёт", .676, .829, .295, .327), w("у", .832, .863, .295, .327)},
		{w("открыть", .142, .284, .475, .515), w("первый", .289, .487, .475, .515),
			w("раздел", .491, .625, .475, .515), w("для", .629, .703, .475, .515),
			w("записей", .708, .906, .475, .515)},
		{w("таблицами", .158, .442, .511, .528), w("и", .445, .485, .511, .528),
			w("графиками", .488, .688, .511, .528)},
	}
}

func TestBandsSelectWholeLinesWordsAndTouchedWords(t *testing.T) {
	lines := notes()
	bands := []rmmarks.PageBox{b(.464, .700, .508, .530), b(.119, .880, .295, .319), b(.129, .289, .481, .502)}
	got := mustPassages(t, bands, lines, nil)
	want := []string{"потом ещё нужно проверить отчёт у", "открыть", "и графиками"}
	if !slices.Equal(texts(got), want) {
		t.Fatalf("passages %q, want %q", texts(got), want)
	}
	for i, band := range []int{1, 2, 0} {
		if got[i].Band != band {
			t.Errorf("passage %d is of band %d, want %d", i, got[i].Band, band)
		}
	}
	if got[0].Again != nil || got[0].Second != nil {
		t.Errorf("a repeat or a second reading: %+v", got[0])
	}
}

func TestBandsSelectingTheSamePiecesAnswerOnePassage(t *testing.T) {
	lines := notes()
	first := b(.119, .880, .295, .319)
	bands := []rmmarks.PageBox{first, b(.129, .870, .297, .319), b(.464, .700, .508, .530), first}
	got := mustPassages(t, bands, lines, nil)
	if len(got) != 2 {
		t.Fatalf("%d passages, want 2", len(got))
	}
	if got[0].Band != 0 || !slices.Equal(got[0].Again, []int{1, 3}) {
		t.Errorf("band %d again %v, want band 0 again [1 3]", got[0].Band, got[0].Again)
	}
}

func TestABandOverNothingAnswersNothing(t *testing.T) {
	if got := mustPassages(t, []rmmarks.PageBox{b(.1, .5, .8, .84)}, notes(), nil); len(got) != 0 {
		t.Errorf("passages %q", texts(got))
	}
}

func TestOneWordOfALineAndTheWholeReading(t *testing.T) {
	words := rmmarks.Line{
		w("Здесь", .152, .176, .259, .298), w("отмечено", .181, .287, .259, .298),
		w("одно", .292, .461, .259, .298), w("последнее", .466, .616, .259, .298),
		w("слово.", .621, .756, .259, .298),
	}
	band := []rmmarks.PageBox{b(.629, .755, .270, .286)}
	if got := texts(mustPassages(t, band, []rmmarks.Line{words}, nil)); !slices.Equal(got, []string{"слово."}) {
		t.Errorf("words: %q", got)
	}
	reading := []rmmarks.RecognizedLine{{Text: "Здесь отмечено одно последнее слово.", Box: b(.152, .756, .259, .298)}}
	got, err := rmmarks.PassagesFromReading(band, reading)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(texts(got), []string{"Здесь отмечено одно последнее слово."}) {
		t.Errorf("reading: %q", texts(got))
	}
}

func TestColumnsAndSplitLines(t *testing.T) {
	lines := each(
		w("OPŠTE ODREDBE UGOVORA O SARADNJI", .090, .472, .446, .465),
		w("ОБЩИЕ ПОЛОЖЕНИЯ", .536, .873, .448, .465),
		w("РАЗДЕЛА О ПОРЯДКЕ РАБОТЫ", .541, .871, .467, .483),
	)
	bands := []rmmarks.PageBox{b(.535, .899, .442, .465), b(.524, .896, .458, .482)}
	want := []string{"ОБЩИЕ ПОЛОЖЕНИЯ", "РАЗДЕЛА О ПОРЯДКЕ РАБОТЫ"}
	if got := texts(mustPassages(t, bands, lines, nil)); !slices.Equal(got, want) {
		t.Errorf("columns: %q, want %q", got, want)
	}

	split := each(
		w("Strane su saglasne da je ovaj tekst", .077, .489, .557, .576),
		w("Первая", .497, .601, .561, .576),
		w("строка", .664, .735, .562, .573),
		w("раздела", .797, .914, .562, .571),
	)
	got := mustPassages(t, []rmmarks.PageBox{b(.499, .916, .557, .581)}, split, nil)
	if len(got) != 1 || got[0].Text != "Первая строка раздела" {
		t.Fatalf("split: %q", texts(got))
	}
	if !boxEqual(got[0].Place, b(.497, .914, .561, .576)) {
		t.Errorf("place %v", got[0].Place)
	}
}

func TestASecondReading(t *testing.T) {
	band := b(.1, .9, .29, .33)
	page := rmmarks.Line{w("потом ещё нужно проверить отчёт и", .13, .86, .295, .327)}
	cut := rmmarks.Line{w("потом ещё нужно проворить отчёт и", .13, .86, .295, .327)}
	alone := []rmmarks.BandReading{{Band: 0, Lines: []rmmarks.Line{cut}}}

	got := mustPassages(t, []rmmarks.PageBox{band}, []rmmarks.Line{page}, alone)
	if len(got) != 1 || got[0].Text != page[0].Text || got[0].Second == nil || *got[0].Second != cut[0].Text {
		t.Errorf("differing readings: %+v", got)
	}

	got = mustPassages(t, []rmmarks.PageBox{band}, nil, alone)
	if len(got) != 1 || got[0].Text != cut[0].Text || got[0].Second != nil {
		t.Errorf("the cut alone: %+v", got)
	}

	same := []rmmarks.BandReading{{Band: 0, Lines: []rmmarks.Line{page}}}
	got = mustPassages(t, []rmmarks.PageBox{band}, []rmmarks.Line{page}, same)
	if len(got) != 1 || got[0].Second != nil {
		t.Errorf("agreeing readings: %+v", got)
	}

	empty := []rmmarks.BandReading{{Band: 0}}
	if got = mustPassages(t, []rmmarks.PageBox{band}, nil, empty); len(got) != 0 {
		t.Errorf("an empty reading: %+v", got)
	}

	cuts := []rmmarks.BandReading{{Band: 0, Lines: []rmmarks.Line{cut}}, {Band: 1, Lines: []rmmarks.Line{cut}}}
	if got = mustPassages(t, []rmmarks.PageBox{band, band}, nil, cuts); len(got) != 2 {
		t.Errorf("cut-only passages are deduplicated: %+v", got)
	}

	far := []rmmarks.BandReading{{Band: 7, Lines: []rmmarks.Line{cut}}}
	if got = mustPassages(t, []rmmarks.PageBox{band}, nil, far); len(got) != 0 {
		t.Errorf("a reading of no band is read: %+v", got)
	}
}

func TestTheSameSpellingElsewhereIsNotARepeat(t *testing.T) {
	lines := each(w("word", .1, .4, .1, .2), w("word", .1, .4, .5, .6))
	got := mustPassages(t, []rmmarks.PageBox{b(0, 1, .1, .2), b(0, 1, .5, .6)}, lines, nil)
	if len(got) != 2 {
		t.Errorf("%d passages, want 2", len(got))
	}
}

func TestThresholds(t *testing.T) {
	line := []rmmarks.Line{{w("word", 0, 1, 0, 1), w("tail", 1.1, 2, 0, 1)}}
	for _, c := range []struct {
		band rmmarks.PageBox
		want []string
	}{
		{b(0, .5, .49, .51), []string{"word"}},
		{b(.2, .3, .49, .51), []string{"word"}},
		{b(5, 6, 0, 1), nil},
	} {
		if got := texts(mustPassages(t, []rmmarks.PageBox{c.band}, line, nil)); !slices.Equal(got, c.want) {
			t.Errorf("band %v: %q, want %q", c.band, got, c.want)
		}
	}

	spaced := rmmarks.Line{w("  ", .1, .2, .3, .34), w("слово", .21, .4, .3, .34), w(" ", .41, .5, .3, .34)}
	band := []rmmarks.PageBox{b(.09, .51, .3, .34)}
	if got := texts(mustPassages(t, band, []rmmarks.Line{spaced}, nil)); !slices.Equal(got, []string{"слово"}) {
		t.Errorf("trimmed: %q", got)
	}
	spaced[1].Text = " "
	if got := mustPassages(t, band, []rmmarks.Line{spaced}, nil); len(got) != 0 {
		t.Errorf("blank: %q", texts(got))
	}
}

func TestTheAnswerSharesNothingWithTheInput(t *testing.T) {
	text := []byte("word")
	lines := []rmmarks.Line{{{Text: string(text), Box: b(0, 1, 0, 1)}}}
	got := mustPassages(t, []rmmarks.PageBox{b(0, 1, 0, 1)}, lines, nil)
	text[0] = 'X'
	lines[0][0].Text = "other"
	if len(got) != 1 || got[0].Text != "word" {
		t.Errorf("passages %q", texts(got))
	}
}

func TestEmptyInputsAnswerNothing(t *testing.T) {
	got, err := rmmarks.Passages(nil, nil, nil)
	if err != nil || len(got) != 0 {
		t.Errorf("passages %v, %v", got, err)
	}
	got, err = rmmarks.PassagesFromReading(nil, nil)
	if err != nil || len(got) != 0 {
		t.Errorf("from reading %v, %v", got, err)
	}
}
