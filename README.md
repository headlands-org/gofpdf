# GoFPDF document generator

[![MIT
licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/phpdave11/gofpdf/master/LICENSE)
[![Report](https://goreportcard.com/badge/github.com/phpdave11/gofpdf)](https://goreportcard.com/report/github.com/phpdave11/gofpdf)
[![GoDoc](https://img.shields.io/badge/godoc-GoFPDF-blue.svg)](https://godoc.org/github.com/phpdave11/gofpdf)

![](https://github.com/phpdave11/gofpdf/raw/master/image/logo_gofpdf.jpg?raw=true)

Package gofpdf implements a PDF document generator with high level
support for text, drawing and images.

## Features

  - UTF-8 support (full 4-byte sequence support)
  - Emoji rendering with BMP Unicode support (U+2000-U+2FFF)
  - Grapheme cluster handling for complex Unicode text
  - Choice of measurement unit, page format and margins
  - Page header and footer management
  - Automatic page breaks, line breaks, and text justification
  - Inclusion of JPEG, PNG, GIF, TIFF and basic path-only SVG images
  - Colors, gradients and alpha channel transparency
  - Outline bookmarks
  - Internal and external links
  - TrueType, Type1 and encoding support
  - Page compression
  - Lines, Bézier curves, arcs, and ellipses
  - Rotation, scaling, skewing, translation, and mirroring
  - Clipping
  - Document protection
  - Layers
  - Templates
  - Barcodes
  - Charting facility
  - Import PDFs as templates

gofpdf has no dependencies other than the Go standard library. All tests
pass on Linux, Mac and Windows platforms.

gofpdf supports UTF-8 TrueType fonts and “right-to-left” languages. Note
that Chinese, Japanese, and Korean characters may not be included in
many general purpose fonts. For these languages, a specialized font (for
example,
[NotoSansSC](https://github.com/jsntn/webfonts/blob/master/NotoSansSC-Regular.ttf)
for simplified Chinese) can be used.

Also, support is provided to automatically translate UTF-8 runes to code
page encodings for languages that have fewer than 256 glyphs.

## Installation

To install the package on your system, run

``` shell
go get github.com/phpdave11/gofpdf
```

Later, to receive updates, run

``` shell
go get -u -v github.com/phpdave11/gofpdf/...
```

## Quick Start

The following Go code generates a simple PDF file.

``` go
pdf := gofpdf.New("P", "mm", "A4", "")
pdf.AddPage()
pdf.SetFont("Arial", "B", 16)
pdf.Cell(40, 10, "Hello, world")
err := pdf.OutputFileAndClose("hello.pdf")
```

See the functions in the
[fpdf\_test.go](https://github.com/phpdave11/gofpdf/blob/master/fpdf_test.go)
file (shown as examples in this documentation) for more advanced PDF
examples.

## Errors

If an error occurs in an Fpdf method, an internal error field is set.
After this occurs, Fpdf method calls typically return without performing
any operations and the error state is retained. This error management
scheme facilitates PDF generation since individual method calls do not
need to be examined for failure; it is generally sufficient to wait
until after `Output()` is called. For the same reason, if an error
occurs in the calling application during PDF generation, it may be
desirable for the application to transfer the error to the Fpdf instance
by calling the `SetError()` method or the `SetErrorf()` method. At any
time during the life cycle of the Fpdf instance, the error state can be
determined with a call to `Ok()` or `Err()`. The error itself can be
retrieved with a call to `Error()`.

## Conversion Notes

This package is a relatively straightforward translation from the
original [FPDF](http://www.fpdf.org/) library written in PHP (despite
the caveat in the introduction to [Effective
Go](https://golang.org/doc/effective_go.html)). The API names have been
retained even though the Go idiom would suggest otherwise (for example,
`pdf.GetX()` is used rather than simply `pdf.X()`). The similarity of
the two libraries makes the original FPDF website a good source of
information. It includes a forum and FAQ.

However, some internal changes have been made. Page content is built up
using buffers (of type bytes.Buffer) rather than repeated string
concatenation. Errors are handled as explained above rather than
panicking. Output is generated through an interface of type io.Writer or
io.WriteCloser. A number of the original PHP methods behave differently
based on the type of the arguments that are passed to them; in these
cases additional methods have been exported to provide similar
functionality. Font definition files are produced in JSON rather than
PHP.

## Example PDFs

A side effect of running `go test ./...` is the production of a number
of example PDFs. These can be found in the gofpdf/pdf directory after
the tests complete.

Please note that these examples run in the context of a test. In order
run an example as a standalone application, you’ll need to examine
[fpdf\_test.go](https://github.com/phpdave11/gofpdf/blob/master/fpdf_test.go)
for some helper routines, for example `exampleFilename()` and
`summary()`.

Example PDFs can be compared with reference copies in order to verify
that they have been generated as expected. This comparison will be
performed if a PDF with the same name as the example PDF is placed in
the gofpdf/pdf/reference directory and if the third argument to
`ComparePDFFiles()` in internal/example/example.go is true. (By default
it is false.) The routine that summarizes an example will look for this
file and, if found, will call `ComparePDFFiles()` to check the example
PDF for equality with its reference PDF. If differences exist between
the two files they will be printed to standard output and the test will
fail. If the reference file is missing, the comparison is considered to
succeed. In order to successfully compare two PDFs, the placement of
internal resources must be consistent and the internal creation
timestamps must be the same. To do this, the methods `SetCatalogSort()`
and `SetCreationDate()` need to be called for both files. This is done
automatically for all examples.

## Nonstandard Fonts

Nothing special is required to use the standard PDF fonts (courier,
helvetica, times, zapfdingbats) in your documents other than calling
`SetFont()`.

You should use `AddUTF8Font()` or `AddUTF8FontFromBytes()` to add a
TrueType UTF-8 encoded font. Use `RTL()` and `LTR()` methods switch
between “right-to-left” and “left-to-right” mode.

In order to use a different non-UTF-8 TrueType or Type1 font, you will
need to generate a font definition file and, if the font will be
embedded into PDFs, a compressed version of the font file. This is done
by calling the MakeFont function or using the included makefont command
line utility. To create the utility, cd into the makefont subdirectory
and run “go build”. This will produce a standalone executable named
makefont. Select the appropriate encoding file from the font
subdirectory and run the command as in the following example.

``` shell
./makefont --embed --enc=../font/cp1252.map --dst=../font ../font/calligra.ttf
```

In your PDF generation code, call `AddFont()` to load the font and, as
with the standard fonts, SetFont() to begin using it. Most examples,
including the package example, demonstrate this method. Good sources of
free, open-source fonts include [Google
Fonts](https://fonts.google.com/) and [DejaVu
Fonts](http://dejavu-fonts.org/).

## Emoji Support

gofpdf supports emoji rendering using monochrome TrueType fonts such as
Noto Emoji. The library includes full UTF-8 support with 4-byte sequence
handling and grapheme cluster processing, enabling proper rendering of
many Unicode symbols and emoji characters.

### What Works

The library successfully renders emoji from the Basic Multilingual Plane
(BMP), specifically in the Unicode range U+2000 to U+2FFF. This includes:

  - Weather symbols: ☀ ☁ ☂ ☃ ⛄ ⚡
  - Common symbols: ❤ ⭐ ✔ ✘ ☕ ✉
  - Arrows and geometric shapes: ← → ↑ ↓ ▲ ▼ ◆ ●
  - Zodiac and miscellaneous: ♈ ♉ ♊ ♋ ☮ ☯ ⚠
  - Full UTF-8 support for 4-byte character sequences
  - Grapheme cluster handling for complex character combinations
  - Right-to-left (RTL) and left-to-right (LTR) text rendering

### Limitations

Due to technical constraints in the PDF font subsetting implementation,
certain emoji types are not supported:

  - **CMAP Format 4 Limitation**: gofpdf uses CMAP format 4 for font
    character mapping, which only supports the Basic Multilingual Plane
    (BMP, U+0000 to U+FFFF). Modern emoji in the Supplementary Multilingual
    Plane (U+1F300 and above) cannot be rendered with the current
    implementation.

  - **No Color Emoji**: Only monochrome (single-color) emoji rendering is
    supported. Color emoji fonts (such as Apple Color Emoji or Noto Color
    Emoji) will not display colors in the generated PDF.

  - **Supplementary Plane Emoji**: Popular modern emoji like 😀 🎉 🚀
    (U+1F600+) may not render correctly or may appear as missing glyphs,
    since they fall outside the BMP range.

### Workarounds

For emoji outside the BMP range, consider these alternatives:

  - Use BMP emoji equivalents when available (e.g., ☺ instead of 😀)
  - Embed emoji as small PNG/JPEG images using `ImageOptions()`
  - Use pictographic symbols from the supported Unicode ranges
  - Consider SVG-to-image conversion for complex emoji needs

### Quick Start with Emoji

To use emoji in your PDFs, you'll need a TrueType font that includes emoji
glyphs. Noto Emoji is a free, open-source option that works well with
gofpdf.

#### 1. Download Noto Emoji Font

Download the Noto Emoji font from Google Fonts:

``` shell
# Download Noto Emoji Regular font
wget https://github.com/googlefonts/noto-emoji/raw/main/fonts/NotoEmoji-Regular.ttf
```

Alternatively, download it manually from the [Noto Emoji GitHub
repository](https://github.com/googlefonts/noto-emoji).

#### 2. Basic Usage Example

Here's a minimal example showing how to render emoji in a PDF:

``` go
package main

import (
    "log"
    "github.com/phpdave11/gofpdf"
)

func main() {
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()

    // Add Noto Emoji font
    pdf.AddUTF8Font("notoemoji", "", "NotoEmoji-Regular.ttf")
    pdf.SetFont("notoemoji", "", 16)

    // Render BMP emoji (these work!)
    pdf.Cell(0, 10, "Weather: ☀ ☁ ☂ ☃ ⛄")
    pdf.Ln(12)
    pdf.Cell(0, 10, "Symbols: ❤ ⭐ ✔ ☕ ✉")
    pdf.Ln(12)
    pdf.Cell(0, 10, "Arrows: → ↑ ← ↓")
    pdf.Ln(12)
    pdf.Cell(0, 10, "Shapes: ● ▲ ◆ ■")

    // Output the PDF
    err := pdf.OutputFileAndClose("emoji.pdf")
    if err != nil {
        log.Fatal(err)
    }
}
```

#### 3. Expected Output

Running the above code will generate a PDF file named `emoji.pdf`
containing the rendered emoji symbols in monochrome (black). The symbols
will appear as outlined or filled shapes, depending on the font design.

### Troubleshooting

**Problem**: Emoji appear as blank squares or missing glyphs

**Solutions**:
  - Ensure you're using a font that includes emoji glyphs (e.g., Noto
    Emoji)
  - Verify the font file path is correct in `AddUTF8Font()`
  - Check that the emoji you're using are in the BMP range (U+2000-U+2FFF)
  - Modern emoji (U+1F300+) are not supported due to CMAP format 4
    limitations

**Problem**: Font file not found error

**Solutions**:
  - Use an absolute path to the font file
  - Ensure the font file exists at the specified location
  - Check file permissions

**Problem**: Some emoji render, others don't

**Solutions**:
  - This is expected behavior. Only BMP emoji (U+2000-U+2FFF) are supported
  - Check the Unicode code point of non-rendering emoji. If they're above
    U+FFFF, they're in the supplementary plane and won't render
  - Use a Unicode character lookup tool to find BMP alternatives

### Unicode Range Reference

For reference, here are the Unicode ranges and their support status:

  - **U+0000 to U+FFFF** (BMP): Supported via CMAP format 4
  - **U+2000 to U+2FFF** (General Punctuation, Symbols): Full emoji support
  - **U+1F300 to U+1F9FF** (Supplementary Plane): Not supported
  - **U+10000+** (Supplementary Planes): Not supported

To check if a specific emoji is supported, look up its Unicode code point.
If it falls within U+0000 to U+FFFF (and your font includes it), it should
render correctly.

## Related Packages

The [draw2d](https://github.com/llgcode/draw2d) package is a two
dimensional vector graphics library that can generate output in
different forms. It uses gofpdf for its document production mode.

## Contributing Changes

gofpdf is a global community effort and you are invited to make it even
better. If you have implemented a new feature or corrected a problem,
please consider contributing your change to the project. A contribution
that does not directly pertain to the core functionality of gofpdf
should be placed in its own directory directly beneath the `contrib`
directory.

Here are guidelines for making submissions. Your change should

  - be compatible with the MIT License
  - be properly documented
  - be formatted with `go fmt`
  - include an example in
    [fpdf\_test.go](https://github.com/phpdave11/gofpdf/blob/master/fpdf_test.go)
    if appropriate
  - conform to the standards of [golint](https://github.com/golang/lint)
    and [go vet](https://golang.org/cmd/vet/), that is, `golint .` and
    `go vet .` should not generate any warnings
  - not diminish [test coverage](https://blog.golang.org/cover)

[Pull requests](https://help.github.com/articles/using-pull-requests/)
are the preferred means of accepting your changes.

## License

gofpdf is released under the MIT License. It is copyrighted by Dave Barnes
and the contributors acknowledged below.

## Acknowledgments

Thank you to Kurt Jung who originally wrote gofpdf in 2013 - 2019.
This package’s code and documentation are closely derived from the
[FPDF](http://www.fpdf.org/) library created by Olivier Plathey, and a
number of font and image resources are copied directly from it. Bruno
Michel has provided valuable assistance with the code. Drawing support
is adapted from the FPDF geometric figures script by David Hernández
Sanz. Transparency support is adapted from the FPDF transparency script
by Martin Hall-May. Support for gradients and clipping is adapted from
FPDF scripts by Andreas Würmser. Support for outline bookmarks is
adapted from Olivier Plathey by Manuel Cornes. Layer support is adapted
from Olivier Plathey. Support for transformations is adapted from the
FPDF transformation script by Moritz Wagner and Andreas Würmser. PDF
protection is adapted from the work of Klemen Vodopivec for the FPDF
product. Lawrence Kesteloot provided code to allow an image’s extent to
be determined prior to placement. Support for vertical alignment within
a cell was provided by Stefan Schroeder. Ivan Daniluk generalized the
font and image loading code to use the Reader interface while
maintaining backward compatibility. Anthony Starks provided code for the
Polygon function. Robert Lillack provided the Beziergon function and
corrected some naming issues with the internal curve function. Claudio
Felber provided implementations for dashed line drawing and generalized
font loading. Stani Michiels provided support for multi-segment path
drawing with smooth line joins, line join styles, enhanced fill modes,
and has helped greatly with package presentation and tests. Templating
is adapted by Marcus Downing from the FPDF\_Tpl library created by Jan
Slabon and Setasign. Jelmer Snoeck contributed packages that generate a
variety of barcodes and help with registering images on the web. Jelmer
Snoek and Guillermo Pascual augmented the basic HTML functionality with
aligned text. Kent Quirk implemented backwards-compatible support for
reading DPI from images that support it, and for setting DPI manually
and then having it properly taken into account when calculating image
size. Paulo Coutinho provided support for static embedded fonts. Dan
Meyers added support for embedded JavaScript. David Fish added a generic
alias-replacement function to enable, among other things, table of
contents functionality. Andy Bakun identified and corrected a problem in
which the internal catalogs were not sorted stably. Paul Montag added
encoding and decoding functionality for templates, including images that
are embedded in templates; this allows templates to be stored
independently of gofpdf. Paul also added support for page boxes used in
printing PDF documents. Wojciech Matusiak added supported for word
spacing. Artem Korotkiy added support of UTF-8 fonts. Dave Barnes added
support for imported objects and templates. Brigham Thompson added
support for rounded rectangles. Joe Westcott added underline
functionality and optimized image storage. Benoit KUGLER contributed
support for rectangles with corners of unequal radius, modification
times, and for file attachments and annotations.

## Roadmap

  - Remove all legacy code page font support; use UTF-8 exclusively
  - Improve test coverage as reported by the coverage tool.
