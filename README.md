# rmmarks-go

Geometry for marked text, from Go: where a highlight lies on a page, which words it covers, and how text recognized on a stack of cropped strips
returns to the page. It reads no file and runs no recognition: it takes boxes and text and answers boxes and text.

Idiomatic cgo bindings over **[rmmarks-lib](https://github.com/altessa-s/rmmarks-lib)**, the portable C23 library, whose sources are vendored into
this module: `go get` builds it with no system library to install. The same rules are available for Swift as
**[rmmarks-swift](https://github.com/altessa-s/rmmarks-swift)**.

Part of **[Slate Drive](https://slatedrive.app)** ([repository](https://github.com/altessa-s/slatedrive)), the macOS app that shows a reMarkable
library in Finder.

## Requirements

- Go 1.24 or newer, with cgo enabled (`CGO_ENABLED=1`, the default wherever a C compiler is found).
- A **C23 compiler**: clang 19 or newer, or GCC 13 or newer. Not every compiler that takes a C23 flag is one: the library spells its thresholds with
  `constexpr`, which clang 18 does not read. Apple's clang in a current Xcode qualifies. MSVC does not, so on Windows that means clang.

The C is compiled with `-std=c2x`, the one spelling GCC 13, GCC 14 and clang 19 all take, and linked with `-lm` on Linux. Point `CC` at the
compiler when the default one is too old:

```sh
CC=clang-19 go build ./...
```

## Installing

```sh
go get github.com/altessa-s/rmmarks-go
```

```go
import "github.com/altessa-s/rmmarks-go"   // package rmmarks
```

## Finding the marked text

```go
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
```

Coordinates are fractions of the page, origin top left, and are never clipped to `[0, 1]`. Pass each line as its words where they are known;
`PassagesFromReading` takes whole recognized lines and treats each as one piece.

The matching rules, each threshold a named constant:

- The band must cover the line's vertical middle, within `PastTheMiddle` (5%) of the line's height.
- Covering `WholeLine` (85%) of the line's width selects the whole line.
- Otherwise, words covered by `LeastWidth` (half) of their width are selected.
- If none qualify, words touched by `LeastTouch` (20%) of the narrower span are selected.
- Selected pieces are joined in input order with one space, and the passage is trimmed of Unicode whitespace at its ends.
- Bands selecting the same pieces of the page become one passage; the later bands are in `Again`. Equal text in another place is not a repeat.
- Passages are ordered by `(Place.Y0, Place.X0, Band)`.

## A second reading from cropped strips

```go
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
```

`Brought` pairs strips with band numbers in order; the shorter list decides how many pairs there are. A line belongs to the first strip whose
`InSheet` holds the vertical middle of the *whole* line, and its pieces are carried to `OnPage`. Several strips may give lines to one band.

What `Brought` answers is what `Passages` takes as separate readings, one per band key. The page's own reading wins; a separate reading that
differs is kept in `Second`, and one that alone found words becomes the passage. Canonically equivalent Unicode counts as agreement.

## Blocks to crop

```go
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
```

`Blocks` joins a passage to every block closer vertically than `WithinALine` (75%) of the passage's own height and overlapping it horizontally,
and repeats until nothing more joins. The gap does not grow with the block. Only each passage's `Place` is read.

## Errors

```go
box := rmmarks.PageBox{X0: 0, Y0: 0, X1: 1, Y1: 1}
_, err := rmmarks.Passages([]rmmarks.PageBox{box}, []rmmarks.Line{{{Text: "caf\xe9", Box: box}}}, nil)
if errors.Is(err, rmmarks.ErrInvalidUTF8) {
	fmt.Println(err)
}
```

Every failure is an `*Error` whose `Status` mirrors the C library's, matched with `errors.Is`:

| Error | When |
|---|---|
| `ErrInvalidArgument` | A box that is not finite or whose extents are not, duplicate band keys among separate readings, a negative band key or number, or a coordinate transform that overflows. |
| `ErrOutOfMemory` | An allocation failed, or a count or a length is larger than a Go `int` holds. |
| `ErrInvalidUTF8` | An input text is not well-formed UTF-8. It is refused by the C library, never repaired on the way. |

An empty answer is a success, distinct from failure.

## Design

**Values, not handles.** Every answer is plain Go data. The C result is copied out field by field and released before the call returns: nothing
the caller holds points into C memory, and there is nothing to close.

**Counted text.** Strings cross to C as bytes and a length, in C memory, so a NUL byte is text like any other and no Go pointer is retained.
Answers are copied back by their length, not up to a NUL.

**C memory for C pointers.** The arguments are arrays of records pointing at further arrays and text, and cgo forbids handing C Go memory that
holds Go pointers. Each call builds them in C memory and frees them as one; the tests run clean under `GOEXPERIMENT=cgocheck2`.

**No state between calls.** The C library holds no mutable global state and emits no logs, so there is no logger here. Any number of calls may
run at once, from any goroutine; inputs are only read.

**The whole contract.** All twelve functions rmmarks-lib exports are called: the three operations and the reading variant, the three result
releases, the two box transforms, the empty-page test, line placement and the status names. The thresholds and the version are held against the
C library's own spelling by a test.

## Layout

| Path | What is in it |
|---|---|
| `*.go` | The API: `model.go` (boxes, text, marks), `text.go` (the operations), `errors.go`, `cgo.go`, `memory.go` (C arguments and results) |
| `clib_*.c` | One line each, compiling a vendored C source into the package: cgo builds only the C files of the package directory |
| `clib_digest.h` | Generated: the vendored tag and a digest of `internal/clib`, so that a change there rebuilds the package |
| `internal/clib/` | The vendored rmmarks-lib: `src/`, `include/rmmarks/`, its `LICENSE`, the Unicode data licence and the tag it was taken at in `VERSION` |
| `scripts/update-clib.sh` | Re-vendors rmmarks-lib from a checkout, swapping the tree in whole and rewriting the digest: `scripts/update-clib.sh ../rmmarks-lib` |
| `testdata/fuzz-seeds/` | Seeds for `FuzzPassages` |

```sh
go test -race ./...
go test -run='^$' -fuzz=FuzzPassages -fuzztime=30s .
```

## See also

- **[rmmarks-lib](https://github.com/altessa-s/rmmarks-lib)** — the C23 library these bindings compile.
- **[rmmarks-swift](https://github.com/altessa-s/rmmarks-swift)** — the same rules as a Swift package.
- **[rmlines-go](https://github.com/altessa-s/rmlines-go)** — the `.rm` v6 page format for Go; its highlights are what this package places.
- **[rmlines-lib](https://github.com/altessa-s/rmlines-lib)** and **[rmlines-swift](https://github.com/altessa-s/rmlines-swift)** — the same
  page format for C and Swift.
- **[rmdoc-swift](https://github.com/altessa-s/rmdoc-swift)** — the `.rmdoc` container a document and its pages travel in.
- **[Slate Drive](https://slatedrive.app)** ([repository](https://github.com/altessa-s/slatedrive)) — the app these packages are part of.

## Acknowledgements

reMarkable does not publish its file formats. This package is built on openly available information and on the research of third-party developers
who documented them.

## License

MIT — see [LICENSE](LICENSE). The vendored C under `internal/clib/` is rmmarks-lib, under its own MIT licence beside it; its generated Unicode
tables carry the [Unicode data licence](internal/clib/UNICODE-LICENSE.txt).

reMarkable is a trademark of reMarkable AS. This project is not affiliated with, endorsed by or sponsored by reMarkable AS.
