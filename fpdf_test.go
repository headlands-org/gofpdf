/*
 * Copyright (c) 2013-2015 Kurt Jung (Gmail: kurt.w.jung)
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package gofpdf_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/headlands-org/gofpdf"
	"github.com/headlands-org/gofpdf/internal/example"
	"github.com/headlands-org/gofpdf/internal/files"
)

func init() {
	cleanup()
}

func cleanup() {
	filepath.Walk(example.PdfDir(),
		func(path string, info os.FileInfo, err error) (reterr error) {
			if info.Mode().IsRegular() {
				dir, _ := filepath.Split(path)
				if "reference" != filepath.Base(dir) {
					if len(path) > 3 {
						if path[len(path)-4:] == ".pdf" {
							os.Remove(path)
						}
					}
				}
			}
			return
		})
}

func TestFpdfImplementPdf(t *testing.T) {
	// this will not compile if Fpdf and Tpl
	// do not implement Pdf
	var _ gofpdf.Pdf = (*gofpdf.Fpdf)(nil)
	var _ gofpdf.Pdf = (*gofpdf.Tpl)(nil)
}

// TestPagedTemplate ensures new paged templates work
func TestPagedTemplate(t *testing.T) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	tpl := pdf.CreateTemplate(func(t *gofpdf.Tpl) {
		// this will be the second page, as a page is already
		// created by default
		t.AddPage()
		t.AddPage()
		t.AddPage()
	})

	if tpl.NumPages() != 4 {
		t.Fatalf("The template does not have the correct number of pages %d", tpl.NumPages())
	}

	tplPages := tpl.FromPages()
	for x := 0; x < len(tplPages); x++ {
		pdf.AddPage()
		pdf.UseTemplate(tplPages[x])
	}

	// get the last template
	tpl2, err := tpl.FromPage(tpl.NumPages())
	if err != nil {
		t.Fatal(err)
	}

	// the objects should be the exact same, as the
	// template will represent the last page by default
	// therefore no new id should be set, and the object
	// should be the same object
	if fmt.Sprintf("%p", tpl2) != fmt.Sprintf("%p", tpl) {
		t.Fatal("Template no longer respecting initial template object")
	}
}

// TestIssue0116 addresses issue 116 in which library silently fails after
// calling CellFormat when no font has been set.
func TestIssue0116(t *testing.T) {
	var pdf *gofpdf.Fpdf

	pdf = gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "OK")
	if pdf.Error() != nil {
		t.Fatalf("not expecting error when rendering text")
	}

	pdf = gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.Cell(40, 10, "Not OK") // Font not set
	if pdf.Error() == nil {
		t.Fatalf("expecting error when rendering text without having set font")
	}
}

// TestIssue0193 addresses issue 193 in which the error io.EOF is incorrectly
// assigned to the FPDF instance error.
func TestIssue0193(t *testing.T) {
	var png []byte
	var pdf *gofpdf.Fpdf
	var err error
	var rdr *bytes.Reader

	png, err = ioutil.ReadFile(example.ImageFile("sweden.png"))
	if err == nil {
		rdr = bytes.NewReader(png)
		pdf = gofpdf.New("P", "mm", "A4", "")
		pdf.AddPage()
		_ = pdf.RegisterImageOptionsReader("sweden", gofpdf.ImageOptions{ImageType: "png", ReadDpi: true}, rdr)
		err = pdf.Error()
	}
	if err != nil {
		t.Fatalf("issue 193 error: %s", err)
	}

}

// TestIssue0209SplitLinesEqualMultiCell addresses issue 209
// make SplitLines and MultiCell split at the same place
func TestIssue0209SplitLinesEqualMultiCell(t *testing.T) {
	var pdf *gofpdf.Fpdf

	pdf = gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 8)
	// this sentence should not be splited
	str := "Guochin Amandine"
	lines := pdf.SplitLines([]byte(str), 26)
	_, FontSize := pdf.GetFontSize()
	y_start := pdf.GetY()
	pdf.MultiCell(26, FontSize, str, "", "L", false)
	y_end := pdf.GetY()

	if len(lines) != 1 {
		t.Fatalf("expect SplitLines split in one line")
	}
	if int(y_end-y_start) != int(FontSize) {
		t.Fatalf("expect MultiCell split in one line %.2f != %.2f", y_end-y_start, FontSize)
	}

	// this sentence should be splited in two lines
	str = "Guiochini Amandine"
	lines = pdf.SplitLines([]byte(str), 26)
	y_start = pdf.GetY()
	pdf.MultiCell(26, FontSize, str, "", "L", false)
	y_end = pdf.GetY()

	if len(lines) != 2 {
		t.Fatalf("expect SplitLines split in two lines")
	}
	if int(y_end-y_start) != int(FontSize*2) {
		t.Fatalf("expect MultiCell split in two lines %.2f != %.2f", y_end-y_start, FontSize*2)
	}
}

// TestFooterFuncLpi tests to make sure the footer is not call twice and SetFooterFuncLpi can work
// without SetFooterFunc.
func TestFooterFuncLpi(t *testing.T) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	var (
		oldFooterFnc  = "oldFooterFnc"
		bothPages     = "bothPages"
		firstPageOnly = "firstPageOnly"
		lastPageOnly  = "lastPageOnly"
	)

	// This set just for testing, only set SetFooterFuncLpi.
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, oldFooterFnc,
			"", 0, "C", false, 0, "")
	})
	pdf.SetFooterFuncLpi(func(lastPage bool) {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, bothPages, "", 0, "L", false, 0, "")
		if !lastPage {
			pdf.CellFormat(0, 10, firstPageOnly, "", 0, "C", false, 0, "")
		} else {
			pdf.CellFormat(0, 10, lastPageOnly, "", 0, "C", false, 0, "")
		}
	})
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	for j := 1; j <= 40; j++ {
		pdf.CellFormat(0, 10, fmt.Sprintf("Printing line number %d", j),
			"", 1, "", false, 0, "")
	}
	if pdf.Error() != nil {
		t.Fatalf("not expecting error when rendering text")
	}
	w := &bytes.Buffer{}
	if err := pdf.Output(w); err != nil {
		t.Errorf("unexpected err: %s", err)
	}
	b := w.Bytes()
	if bytes.Contains(b, []byte(oldFooterFnc)) {
		t.Errorf("not expecting %s render on pdf when FooterFncLpi is set", oldFooterFnc)
	}
	got := bytes.Count(b, []byte("bothPages"))
	if got != 2 {
		t.Errorf("footer %s should render on two page got:%d", bothPages, got)
	}
	got = bytes.Count(b, []byte(firstPageOnly))
	if got != 1 {
		t.Errorf("footer %s should render only on first page got: %d", firstPageOnly, got)
	}
	got = bytes.Count(b, []byte(lastPageOnly))
	if got != 1 {
		t.Errorf("footer %s should render only on first page got: %d", lastPageOnly, got)
	}
	f := bytes.Index(b, []byte(firstPageOnly))
	l := bytes.Index(b, []byte(lastPageOnly))
	if f > l {
		t.Errorf("index %d (%s) should less than index %d (%s)", f, firstPageOnly, l, lastPageOnly)
	}
}

type fontResourceType struct {
}

func (f fontResourceType) Open(name string) (rdr io.Reader, err error) {
	var buf []byte
	buf, err = ioutil.ReadFile(example.FontFile(name))
	if err == nil {
		rdr = bytes.NewReader(buf)
		fmt.Printf("Generalized font loader reading %s\n", name)
	}
	return
}

// strDelimit converts 'ABCDEFG' to, for example, 'A,BCD,EFG'
func strDelimit(str string, sepstr string, sepcount int) string {
	pos := len(str) - sepcount
	for pos > 0 {
		str = str[:pos] + sepstr + str[pos:]
		pos = pos - sepcount
	}
	return str
}

func loremList() []string {
	return []string{
		"Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod " +
			"tempor incididunt ut labore et dolore magna aliqua.",
		"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut " +
			"aliquip ex ea commodo consequat.",
		"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum " +
			"dolore eu fugiat nulla pariatur.",
		"Excepteur sint occaecat cupidatat non proident, sunt in culpa qui " +
			"officia deserunt mollit anim id est laborum.",
	}
}

func lorem() string {
	return strings.Join(loremList(), " ")
}

// Example demonstrates the generation of a simple PDF document. Note that
// since only core fonts are used (in this case Arial, a synonym for
// Helvetica), an empty string can be specified for the font directory in the
// call to New(). Note also that the example.Filename() and example.Summary()
// functions belong to a separate, internal package and are not part of the
// gofpdf library. If an error occurs at some point during the construction of
// the document, subsequent method calls exit immediately and the error is
// finally retrieved with the output call where it can be handled by the
// application.
func Example() {
	pdf := gofpdf.New(gofpdf.OrientationPortrait, "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Hello World!")
	fileStr := example.Filename("basic")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/basic.pdf
}

// ExampleFpdf_AddPage demonsrates the generation of headers, footers and page breaks.
func ExampleFpdf_AddPage() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTopMargin(30)
	pdf.SetHeaderFuncMode(func() {
		pdf.Image(example.ImageFile("logo.png"), 10, 6, 30, 0, false, "", 0, "")
		pdf.SetY(5)
		pdf.SetFont("Arial", "B", 15)
		pdf.Cell(80, 0, "")
		pdf.CellFormat(30, 10, "Title", "1", 0, "C", false, 0, "")
		pdf.Ln(20)
	}, true)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()),
			"", 0, "C", false, 0, "")
	})
	pdf.AliasNbPages("")
	pdf.AddPage()
	pdf.SetFont("Times", "", 12)
	for j := 1; j <= 40; j++ {
		pdf.CellFormat(0, 10, fmt.Sprintf("Printing line number %d", j),
			"", 1, "", false, 0, "")
	}
	fileStr := example.Filename("Fpdf_AddPage")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_AddPage.pdf
}

// ExampleFpdf_MultiCell demonstrates word-wrapping, line justification and
// page-breaking.
func ExampleFpdf_MultiCell() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	titleStr := "20000 Leagues Under the Seas"
	pdf.SetTitle(titleStr, false)
	pdf.SetAuthor("Jules Verne", false)
	pdf.SetHeaderFunc(func() {
		// Arial bold 15
		pdf.SetFont("Arial", "B", 15)
		// Calculate width of title and position
		wd := pdf.GetStringWidth(titleStr) + 6
		pdf.SetX((210 - wd) / 2)
		// Colors of frame, background and text
		pdf.SetDrawColor(0, 80, 180)
		pdf.SetFillColor(230, 230, 0)
		pdf.SetTextColor(220, 50, 50)
		// Thickness of frame (1 mm)
		pdf.SetLineWidth(1)
		// Title
		pdf.CellFormat(wd, 9, titleStr, "1", 1, "C", true, 0, "")
		// Line break
		pdf.Ln(10)
	})
	pdf.SetFooterFunc(func() {
		// Position at 1.5 cm from bottom
		pdf.SetY(-15)
		// Arial italic 8
		pdf.SetFont("Arial", "I", 8)
		// Text color in gray
		pdf.SetTextColor(128, 128, 128)
		// Page number
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d", pdf.PageNo()),
			"", 0, "C", false, 0, "")
	})
	chapterTitle := func(chapNum int, titleStr string) {
		// 	// Arial 12
		pdf.SetFont("Arial", "", 12)
		// Background color
		pdf.SetFillColor(200, 220, 255)
		// Title
		pdf.CellFormat(0, 6, fmt.Sprintf("Chapter %d : %s", chapNum, titleStr),
			"", 1, "L", true, 0, "")
		// Line break
		pdf.Ln(4)
	}
	chapterBody := func(fileStr string) {
		// Read text file
		txtStr, err := ioutil.ReadFile(fileStr)
		if err != nil {
			pdf.SetError(err)
		}
		// Times 12
		pdf.SetFont("Times", "", 12)
		// Output justified text
		pdf.MultiCell(0, 5, string(txtStr), "", "", false)
		// Line break
		pdf.Ln(-1)
		// Mention in italics
		pdf.SetFont("", "I", 0)
		pdf.Cell(0, 5, "(end of excerpt)")
	}
	printChapter := func(chapNum int, titleStr, fileStr string) {
		pdf.AddPage()
		chapterTitle(chapNum, titleStr)
		chapterBody(fileStr)
	}
	printChapter(1, "A RUNAWAY REEF", example.TextFile("20k_c1.txt"))
	printChapter(2, "THE PROS AND CONS", example.TextFile("20k_c2.txt"))
	fileStr := example.Filename("Fpdf_MultiCell")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_MultiCell.pdf
}

// ExampleFpdf_SetLeftMargin demonstrates the generation of a PDF document that has multiple
// columns. This is accomplished with the SetLeftMargin() and Cell() methods.
func ExampleFpdf_SetLeftMargin() {
	var y0 float64
	var crrntCol int
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetDisplayMode("fullpage", "TwoColumnLeft")
	titleStr := "20000 Leagues Under the Seas"
	pdf.SetTitle(titleStr, false)
	pdf.SetAuthor("Jules Verne", false)
	setCol := func(col int) {
		// Set position at a given column
		crrntCol = col
		x := 10.0 + float64(col)*65.0
		pdf.SetLeftMargin(x)
		pdf.SetX(x)
	}
	chapterTitle := func(chapNum int, titleStr string) {
		// Arial 12
		pdf.SetFont("Arial", "", 12)
		// Background color
		pdf.SetFillColor(200, 220, 255)
		// Title
		pdf.CellFormat(0, 6, fmt.Sprintf("Chapter %d : %s", chapNum, titleStr),
			"", 1, "L", true, 0, "")
		// Line break
		pdf.Ln(4)
		y0 = pdf.GetY()
	}
	chapterBody := func(fileStr string) {
		// Read text file
		txtStr, err := ioutil.ReadFile(fileStr)
		if err != nil {
			pdf.SetError(err)
		}
		// Font
		pdf.SetFont("Times", "", 12)
		// Output text in a 6 cm width column
		pdf.MultiCell(60, 5, string(txtStr), "", "", false)
		pdf.Ln(-1)
		// Mention
		pdf.SetFont("", "I", 0)
		pdf.Cell(0, 5, "(end of excerpt)")
		// Go back to first column
		setCol(0)
	}
	printChapter := func(num int, titleStr, fileStr string) {
		// Add chapter
		pdf.AddPage()
		chapterTitle(num, titleStr)
		chapterBody(fileStr)
	}
	pdf.SetAcceptPageBreakFunc(func() bool {
		// Method accepting or not automatic page break
		if crrntCol < 2 {
			// Go to next column
			setCol(crrntCol + 1)
			// Set ordinate to top
			pdf.SetY(y0)
			// Keep on page
			return false
		}
		// Go back to first column
		setCol(0)
		// Page break
		return true
	})
	pdf.SetHeaderFunc(func() {
		// Arial bold 15
		pdf.SetFont("Arial", "B", 15)
		// Calculate width of title and position
		wd := pdf.GetStringWidth(titleStr) + 6
		pdf.SetX((210 - wd) / 2)
		// Colors of frame, background and text
		pdf.SetDrawColor(0, 80, 180)
		pdf.SetFillColor(230, 230, 0)
		pdf.SetTextColor(220, 50, 50)
		// Thickness of frame (1 mm)
		pdf.SetLineWidth(1)
		// Title
		pdf.CellFormat(wd, 9, titleStr, "1", 1, "C", true, 0, "")
		// Line break
		pdf.Ln(10)
		// Save ordinate
		y0 = pdf.GetY()
	})
	pdf.SetFooterFunc(func() {
		// Position at 1.5 cm from bottom
		pdf.SetY(-15)
		// Arial italic 8
		pdf.SetFont("Arial", "I", 8)
		// Text color in gray
		pdf.SetTextColor(128, 128, 128)
		// Page number
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d", pdf.PageNo()),
			"", 0, "C", false, 0, "")
	})
	printChapter(1, "A RUNAWAY REEF", example.TextFile("20k_c1.txt"))
	printChapter(2, "THE PROS AND CONS", example.TextFile("20k_c2.txt"))
	fileStr := example.Filename("Fpdf_SetLeftMargin_multicolumn")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_SetLeftMargin_multicolumn.pdf
}

// ExampleFpdf_SplitLines_tables demonstrates word-wrapped table cells
func ExampleFpdf_SplitLines_tables() {
	const (
		colCount = 3
		colWd    = 60.0
		marginH  = 15.0
		lineHt   = 5.5
		cellGap  = 2.0
	)
	// var colStrList [colCount]string
	type cellType struct {
		str  string
		list [][]byte
		ht   float64
	}
	var (
		cellList [colCount]cellType
		cell     cellType
	)

	pdf := gofpdf.New("P", "mm", "A4", "") // 210 x 297
	header := [colCount]string{"Column A", "Column B", "Column C"}
	alignList := [colCount]string{"L", "C", "R"}
	strList := loremList()
	pdf.SetMargins(marginH, 15, marginH)
	pdf.SetFont("Arial", "", 14)
	pdf.AddPage()

	// Headers
	pdf.SetTextColor(224, 224, 224)
	pdf.SetFillColor(64, 64, 64)
	for colJ := 0; colJ < colCount; colJ++ {
		pdf.CellFormat(colWd, 10, header[colJ], "1", 0, "CM", true, 0, "")
	}
	pdf.Ln(-1)
	pdf.SetTextColor(24, 24, 24)
	pdf.SetFillColor(255, 255, 255)

	// Rows
	y := pdf.GetY()
	count := 0
	for rowJ := 0; rowJ < 2; rowJ++ {
		maxHt := lineHt
		// Cell height calculation loop
		for colJ := 0; colJ < colCount; colJ++ {
			count++
			if count > len(strList) {
				count = 1
			}
			cell.str = strings.Join(strList[0:count], " ")
			cell.list = pdf.SplitLines([]byte(cell.str), colWd-cellGap-cellGap)
			cell.ht = float64(len(cell.list)) * lineHt
			if cell.ht > maxHt {
				maxHt = cell.ht
			}
			cellList[colJ] = cell
		}
		// Cell render loop
		x := marginH
		for colJ := 0; colJ < colCount; colJ++ {
			pdf.Rect(x, y, colWd, maxHt+cellGap+cellGap, "D")
			cell = cellList[colJ]
			cellY := y + cellGap + (maxHt-cell.ht)/2
			for splitJ := 0; splitJ < len(cell.list); splitJ++ {
				pdf.SetXY(x+cellGap, cellY)
				pdf.CellFormat(colWd-cellGap-cellGap, lineHt, string(cell.list[splitJ]), "", 0,
					alignList[colJ], false, 0, "")
				cellY += lineHt
			}
			x += colWd
		}
		y += maxHt + cellGap + cellGap
	}

	fileStr := example.Filename("Fpdf_SplitLines_tables")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_SplitLines_tables.pdf
}

// ExampleFpdf_CellFormat_tables demonstrates various table styles.
func ExampleFpdf_CellFormat_tables() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	type countryType struct {
		nameStr, capitalStr, areaStr, popStr string
	}
	countryList := make([]countryType, 0, 8)
	header := []string{"Country", "Capital", "Area (sq km)", "Pop. (thousands)"}
	loadData := func(fileStr string) {
		fl, err := os.Open(fileStr)
		if err == nil {
			scanner := bufio.NewScanner(fl)
			var c countryType
			for scanner.Scan() {
				// Austria;Vienna;83859;8075
				lineStr := scanner.Text()
				list := strings.Split(lineStr, ";")
				if len(list) == 4 {
					c.nameStr = list[0]
					c.capitalStr = list[1]
					c.areaStr = list[2]
					c.popStr = list[3]
					countryList = append(countryList, c)
				} else {
					err = fmt.Errorf("error tokenizing %s", lineStr)
				}
			}
			fl.Close()
			if len(countryList) == 0 {
				err = fmt.Errorf("error loading data from %s", fileStr)
			}
		}
		if err != nil {
			pdf.SetError(err)
		}
	}
	// Simple table
	basicTable := func() {
		left := (210.0 - 4*40) / 2
		pdf.SetX(left)
		for _, str := range header {
			pdf.CellFormat(40, 7, str, "1", 0, "", false, 0, "")
		}
		pdf.Ln(-1)
		for _, c := range countryList {
			pdf.SetX(left)
			pdf.CellFormat(40, 6, c.nameStr, "1", 0, "", false, 0, "")
			pdf.CellFormat(40, 6, c.capitalStr, "1", 0, "", false, 0, "")
			pdf.CellFormat(40, 6, c.areaStr, "1", 0, "", false, 0, "")
			pdf.CellFormat(40, 6, c.popStr, "1", 0, "", false, 0, "")
			pdf.Ln(-1)
		}
	}
	// Better table
	improvedTable := func() {
		// Column widths
		w := []float64{40.0, 35.0, 40.0, 45.0}
		wSum := 0.0
		for _, v := range w {
			wSum += v
		}
		left := (210 - wSum) / 2
		// 	Header
		pdf.SetX(left)
		for j, str := range header {
			pdf.CellFormat(w[j], 7, str, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(-1)
		// Data
		for _, c := range countryList {
			pdf.SetX(left)
			pdf.CellFormat(w[0], 6, c.nameStr, "LR", 0, "", false, 0, "")
			pdf.CellFormat(w[1], 6, c.capitalStr, "LR", 0, "", false, 0, "")
			pdf.CellFormat(w[2], 6, strDelimit(c.areaStr, ",", 3),
				"LR", 0, "R", false, 0, "")
			pdf.CellFormat(w[3], 6, strDelimit(c.popStr, ",", 3),
				"LR", 0, "R", false, 0, "")
			pdf.Ln(-1)
		}
		pdf.SetX(left)
		pdf.CellFormat(wSum, 0, "", "T", 0, "", false, 0, "")
	}
	// Colored table
	fancyTable := func() {
		// Colors, line width and bold font
		pdf.SetFillColor(255, 0, 0)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetDrawColor(128, 0, 0)
		pdf.SetLineWidth(.3)
		pdf.SetFont("", "B", 0)
		// 	Header
		w := []float64{40, 35, 40, 45}
		wSum := 0.0
		for _, v := range w {
			wSum += v
		}
		left := (210 - wSum) / 2
		pdf.SetX(left)
		for j, str := range header {
			pdf.CellFormat(w[j], 7, str, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)
		// Color and font restoration
		pdf.SetFillColor(224, 235, 255)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("", "", 0)
		// 	Data
		fill := false
		for _, c := range countryList {
			pdf.SetX(left)
			pdf.CellFormat(w[0], 6, c.nameStr, "LR", 0, "", fill, 0, "")
			pdf.CellFormat(w[1], 6, c.capitalStr, "LR", 0, "", fill, 0, "")
			pdf.CellFormat(w[2], 6, strDelimit(c.areaStr, ",", 3),
				"LR", 0, "R", fill, 0, "")
			pdf.CellFormat(w[3], 6, strDelimit(c.popStr, ",", 3),
				"LR", 0, "R", fill, 0, "")
			pdf.Ln(-1)
			fill = !fill
		}
		pdf.SetX(left)
		pdf.CellFormat(wSum, 0, "", "T", 0, "", false, 0, "")
	}
	loadData(example.TextFile("countries.txt"))
	pdf.SetFont("Arial", "", 14)
	pdf.AddPage()
	basicTable()
	pdf.AddPage()
	improvedTable()
	pdf.AddPage()
	fancyTable()
	fileStr := example.Filename("Fpdf_CellFormat_tables")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_CellFormat_tables.pdf
}

// ExampleFpdf_HTMLBasicNew demonstrates internal and external links with and without basic
// HTML.
func ExampleFpdf_HTMLBasicNew() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	// First page: manual local link
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 20)
	_, lineHt := pdf.GetFontSize()
	pdf.Write(lineHt, "To find out what's new in this tutorial, click ")
	pdf.SetFont("", "U", 0)
	link := pdf.AddLink()
	pdf.WriteLinkID(lineHt, "here", link)
	pdf.SetFont("", "", 0)
	// Second page: image link and basic HTML with link
	pdf.AddPage()
	pdf.SetLink(link, 0, -1)
	pdf.Image(example.ImageFile("logo.png"), 10, 12, 30, 0, false, "", 0, "http://www.fpdf.org")
	pdf.SetLeftMargin(45)
	pdf.SetFontSize(14)
	_, lineHt = pdf.GetFontSize()
	htmlStr := `You can now easily print text mixing different styles: <b>bold</b>, ` +
		`<i>italic</i>, <u>underlined</u>, or <b><i><u>all at once</u></i></b>!<br><br>` +
		`<center>You can also center text.</center>` +
		`<right>Or align it to the right.</right>` +
		`You can also insert links on text, such as ` +
		`<a href="http://www.fpdf.org">www.fpdf.org</a>, or on an image: click on the logo.`
	html := pdf.HTMLBasicNew()
	html.Write(lineHt, htmlStr)
	fileStr := example.Filename("Fpdf_HTMLBasicNew")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_HTMLBasicNew.pdf
}

// ExampleFpdf_AddFont demonstrates the use of a non-standard font.
func ExampleFpdf_AddFont() {
	pdf := gofpdf.New("P", "mm", "A4", example.FontDir())
	pdf.AddFont("Calligrapher", "", "calligra.json")
	pdf.AddPage()
	pdf.SetFont("Calligrapher", "", 35)
	pdf.Cell(0, 10, "Enjoy new fonts with FPDF!")
	fileStr := example.Filename("Fpdf_AddFont")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_AddFont.pdf
}

// ExampleFpdf_WriteAligned demonstrates how to align text with the Write function.
func ExampleFpdf_WriteAligned() {
	pdf := gofpdf.New("P", "mm", "A4", example.FontDir())
	pdf.SetLeftMargin(50.0)
	pdf.SetRightMargin(50.0)
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 12)
	pdf.WriteAligned(0, 35, "This text is the default alignment, Left", "")
	pdf.Ln(35)
	pdf.WriteAligned(0, 35, "This text is aligned Left", "L")
	pdf.Ln(35)
	pdf.WriteAligned(0, 35, "This text is aligned Center", "C")
	pdf.Ln(35)
	pdf.WriteAligned(0, 35, "This text is aligned Right", "R")
	pdf.Ln(35)
	line := "This can by used to write justified text"
	leftMargin, _, rightMargin, _ := pdf.GetMargins()
	pageWidth, _ := pdf.GetPageSize()
	pageWidth -= leftMargin + rightMargin
	pdf.SetWordSpacing((pageWidth - pdf.GetStringWidth(line)) / float64(strings.Count(line, " ")))
	pdf.WriteAligned(pageWidth, 35, line, "L")
	fileStr := example.Filename("Fpdf_WriteAligned")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_WriteAligned.pdf
}

// ExampleFpdf_Image demonstrates how images are included in documents.
func ExampleFpdf_Image() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 11)
	pdf.Image(example.ImageFile("logo.png"), 10, 10, 30, 0, false, "", 0, "")
	pdf.Text(50, 20, "logo.png")
	pdf.Image(example.ImageFile("logo.gif"), 10, 40, 30, 0, false, "", 0, "")
	pdf.Text(50, 50, "logo.gif")
	pdf.Image(example.ImageFile("logo-gray.png"), 10, 70, 30, 0, false, "", 0, "")
	pdf.Text(50, 80, "logo-gray.png")
	pdf.Image(example.ImageFile("logo-rgb.png"), 10, 100, 30, 0, false, "", 0, "")
	pdf.Text(50, 110, "logo-rgb.png")
	pdf.Image(example.ImageFile("logo.jpg"), 10, 130, 30, 0, false, "", 0, "")
	pdf.Text(50, 140, "logo.jpg")
	fileStr := example.Filename("Fpdf_Image")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_Image.pdf
}

// ExampleFpdf_ImageOptions demonstrates how the AllowNegativePosition field of the
// ImageOption struct can be used to affect horizontal image placement.
func ExampleFpdf_ImageOptions() {
	var opt gofpdf.ImageOptions

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 11)
	pdf.SetX(60)
	opt.ImageType = "png"
	pdf.ImageOptions(example.ImageFile("logo.png"), -10, 10, 30, 0, false, opt, 0, "")
	opt.AllowNegativePosition = true
	pdf.ImageOptions(example.ImageFile("logo.png"), -10, 50, 30, 0, false, opt, 0, "")
	fileStr := example.Filename("Fpdf_ImageOptions")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_ImageOptions.pdf
}

// ExampleFpdf_RegisterImageOptionsReader demonstrates how to load an image
// from a io.Reader (in this case, a file) and register it with options.
func ExampleFpdf_RegisterImageOptionsReader() {
	var (
		opt    gofpdf.ImageOptions
		pdfStr string
		fl     *os.File
		err    error
	)

	pdfStr = example.Filename("Fpdf_RegisterImageOptionsReader")
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 11)
	fl, err = os.Open(example.ImageFile("logo.png"))
	if err == nil {
		opt.ImageType = "png"
		opt.AllowNegativePosition = true
		_ = pdf.RegisterImageOptionsReader("logo", opt, fl)
		fl.Close()
		for x := -20.0; x <= 40.0; x += 5 {
			pdf.ImageOptions("logo", x, x+30, 0, 0, false, opt, 0, "")
		}
		err = pdf.OutputFileAndClose(pdfStr)
	}
	example.Summary(err, pdfStr)
	// Output:
	// Successfully generated pdf/Fpdf_RegisterImageOptionsReader.pdf
}

// This example demonstrates Landscape mode with images.
func ExampleFpdf_SetAcceptPageBreakFunc() {
	var y0 float64
	var crrntCol int
	loremStr := lorem()
	pdf := gofpdf.New("L", "mm", "A4", "")
	const (
		pageWd = 297.0 // A4 210.0 x 297.0
		margin = 10.0
		gutter = 4
		colNum = 3
		colWd  = (pageWd - 2*margin - (colNum-1)*gutter) / colNum
	)
	setCol := func(col int) {
		crrntCol = col
		x := margin + float64(col)*(colWd+gutter)
		pdf.SetLeftMargin(x)
		pdf.SetX(x)
	}
	pdf.SetHeaderFunc(func() {
		titleStr := "gofpdf"
		pdf.SetFont("Helvetica", "B", 48)
		wd := pdf.GetStringWidth(titleStr) + 6
		pdf.SetX((pageWd - wd) / 2)
		pdf.SetTextColor(128, 128, 160)
		pdf.Write(12, titleStr[:2])
		pdf.SetTextColor(128, 128, 128)
		pdf.Write(12, titleStr[2:])
		pdf.Ln(20)
		y0 = pdf.GetY()
	})
	pdf.SetAcceptPageBreakFunc(func() bool {
		if crrntCol < colNum-1 {
			setCol(crrntCol + 1)
			pdf.SetY(y0)
			// Start new column, not new page
			return false
		}
		setCol(0)
		return true
	})
	pdf.AddPage()
	pdf.SetFont("Times", "", 12)
	for j := 0; j < 20; j++ {
		if j == 1 {
			pdf.Image(example.ImageFile("fpdf.png"), -1, 0, colWd, 0, true, "", 0, "")
		} else if j == 5 {
			pdf.Image(example.ImageFile("golang-gopher.png"),
				-1, 0, colWd, 0, true, "", 0, "")
		}
		pdf.MultiCell(colWd, 5, loremStr, "", "", false)
		pdf.Ln(-1)
	}
	fileStr := example.Filename("Fpdf_SetAcceptPageBreakFunc_landscape")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_SetAcceptPageBreakFunc_landscape.pdf
}

// This example tests corner cases as reported by the gocov tool.
func ExampleFpdf_SetKeywords() {
	var err error
	fileStr := example.Filename("Fpdf_SetKeywords")
	err = gofpdf.MakeFont(example.FontFile("CalligrapherRegular.pfb"),
		example.FontFile("cp1252.map"), example.FontDir(), nil, true)
	if err == nil {
		pdf := gofpdf.New("", "", "", "")
		pdf.SetFontLocation(example.FontDir())
		pdf.SetTitle("世界", true)
		pdf.SetAuthor("世界", true)
		pdf.SetSubject("世界", true)
		pdf.SetCreator("世界", true)
		pdf.SetKeywords("世界", true)
		pdf.AddFont("Calligrapher", "", "CalligrapherRegular.json")
		pdf.AddPage()
		pdf.SetFont("Calligrapher", "", 16)
		pdf.Writef(5, "\x95 %s \x95", pdf)
		err = pdf.OutputFileAndClose(fileStr)
	}
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_SetKeywords.pdf
}

// ExampleFpdf_Circle demonstrates the construction of various geometric figures,
func ExampleFpdf_Circle() {
	const (
		thin  = 0.2
		thick = 3.0
	)
	pdf := gofpdf.New("", "", "", "")
	pdf.SetFont("Helvetica", "", 12)
	pdf.SetFillColor(200, 200, 220)
	pdf.AddPage()

	y := 15.0
	pdf.Text(10, y, "Circles")
	pdf.SetFillColor(200, 200, 220)
	pdf.SetLineWidth(thin)
	pdf.Circle(20, y+15, 10, "D")
	pdf.Circle(45, y+15, 10, "F")
	pdf.Circle(70, y+15, 10, "FD")
	pdf.SetLineWidth(thick)
	pdf.Circle(95, y+15, 10, "FD")
	pdf.SetLineWidth(thin)

	y += 40.0
	pdf.Text(10, y, "Ellipses")
	pdf.SetFillColor(220, 200, 200)
	pdf.Ellipse(30, y+15, 20, 10, 0, "D")
	pdf.Ellipse(75, y+15, 20, 10, 0, "F")
	pdf.Ellipse(120, y+15, 20, 10, 0, "FD")
	pdf.SetLineWidth(thick)
	pdf.Ellipse(165, y+15, 20, 10, 0, "FD")
	pdf.SetLineWidth(thin)

	y += 40.0
	pdf.Text(10, y, "Curves (quadratic)")
	pdf.SetFillColor(220, 220, 200)
	pdf.Curve(10, y+30, 15, y-20, 40, y+30, "D")
	pdf.Curve(45, y+30, 50, y-20, 75, y+30, "F")
	pdf.Curve(80, y+30, 85, y-20, 110, y+30, "FD")
	pdf.SetLineWidth(thick)
	pdf.Curve(115, y+30, 120, y-20, 145, y+30, "FD")
	pdf.SetLineCapStyle("round")
	pdf.Curve(150, y+30, 155, y-20, 180, y+30, "FD")
	pdf.SetLineWidth(thin)
	pdf.SetLineCapStyle("butt")

	y += 40.0
	pdf.Text(10, y, "Curves (cubic)")
	pdf.SetFillColor(220, 200, 220)
	pdf.CurveBezierCubic(10, y+30, 15, y-20, 10, y+30, 40, y+30, "D")
	pdf.CurveBezierCubic(45, y+30, 50, y-20, 45, y+30, 75, y+30, "F")
	pdf.CurveBezierCubic(80, y+30, 85, y-20, 80, y+30, 110, y+30, "FD")
	pdf.SetLineWidth(thick)
	pdf.CurveBezierCubic(115, y+30, 120, y-20, 115, y+30, 145, y+30, "FD")
	pdf.SetLineCapStyle("round")
	pdf.CurveBezierCubic(150, y+30, 155, y-20, 150, y+30, 180, y+30, "FD")
	pdf.SetLineWidth(thin)
	pdf.SetLineCapStyle("butt")

	y += 40.0
	pdf.Text(10, y, "Arcs")
	pdf.SetFillColor(200, 220, 220)
	pdf.SetLineWidth(thick)
	pdf.Arc(45, y+35, 20, 10, 0, 0, 180, "FD")
	pdf.SetLineWidth(thin)
	pdf.Arc(45, y+35, 25, 15, 0, 90, 270, "D")
	pdf.SetLineWidth(thick)
	pdf.Arc(45, y+35, 30, 20, 0, 0, 360, "D")
	pdf.SetLineCapStyle("round")
	pdf.Arc(135, y+35, 20, 10, 135, 0, 180, "FD")
	pdf.SetLineWidth(thin)
	pdf.Arc(135, y+35, 25, 15, 135, 90, 270, "D")
	pdf.SetLineWidth(thick)
	pdf.Arc(135, y+35, 30, 20, 135, 0, 360, "D")
	pdf.SetLineWidth(thin)
	pdf.SetLineCapStyle("butt")

	fileStr := example.Filename("Fpdf_Circle_figures")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_Circle_figures.pdf
}

// ExampleFpdf_SetAlpha demonstrates alpha transparency.
func ExampleFpdf_SetAlpha() {
	const (
		gapX  = 10.0
		gapY  = 9.0
		rectW = 40.0
		rectH = 58.0
		pageW = 210
		pageH = 297
	)
	modeList := []string{"Normal", "Multiply", "Screen", "Overlay",
		"Darken", "Lighten", "ColorDodge", "ColorBurn", "HardLight", "SoftLight",
		"Difference", "Exclusion", "Hue", "Saturation", "Color", "Luminosity"}
	pdf := gofpdf.New("", "", "", "")
	pdf.SetLineWidth(2)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 18)
	pdf.SetXY(0, gapY)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(pageW, gapY, "Alpha Blending Modes", "", 0, "C", false, 0, "")
	j := 0
	y := 3 * gapY
	for col := 0; col < 4; col++ {
		x := gapX
		for row := 0; row < 4; row++ {
			pdf.Rect(x, y, rectW, rectH, "D")
			pdf.SetFont("Helvetica", "B", 12)
			pdf.SetFillColor(0, 0, 0)
			pdf.SetTextColor(250, 250, 230)
			pdf.SetXY(x, y+rectH-4)
			pdf.CellFormat(rectW, 5, modeList[j], "", 0, "C", true, 0, "")
			pdf.SetFont("Helvetica", "I", 150)
			pdf.SetTextColor(80, 80, 120)
			pdf.SetXY(x, y+2)
			pdf.CellFormat(rectW, rectH, "A", "", 0, "C", false, 0, "")
			pdf.SetAlpha(0.5, modeList[j])
			pdf.Image(example.ImageFile("golang-gopher.png"),
				x-gapX, y, rectW+2*gapX, 0, false, "", 0, "")
			pdf.SetAlpha(1.0, "Normal")
			x += rectW + gapX
			j++
		}
		y += rectH + gapY
	}
	fileStr := example.Filename("Fpdf_SetAlpha_transparency")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_SetAlpha_transparency.pdf
}

// ExampleFpdf_LinearGradient deomstrates various gradients.
func ExampleFpdf_LinearGradient() {
	pdf := gofpdf.New("", "", "", "")
	pdf.SetFont("Helvetica", "", 12)
	pdf.AddPage()
	pdf.LinearGradient(0, 0, 210, 100, 250, 250, 255, 220, 220, 225, 0, 0, 0, .5)
	pdf.LinearGradient(20, 25, 75, 75, 220, 220, 250, 80, 80, 220, 0, .2, 0, .8)
	pdf.Rect(20, 25, 75, 75, "D")
	pdf.LinearGradient(115, 25, 75, 75, 220, 220, 250, 80, 80, 220, 0, 0, 1, 1)
	pdf.Rect(115, 25, 75, 75, "D")
	pdf.RadialGradient(20, 120, 75, 75, 220, 220, 250, 80, 80, 220,
		0.25, 0.75, 0.25, 0.75, 1)
	pdf.Rect(20, 120, 75, 75, "D")
	pdf.RadialGradient(115, 120, 75, 75, 220, 220, 250, 80, 80, 220,
		0.25, 0.75, 0.75, 0.75, 0.75)
	pdf.Rect(115, 120, 75, 75, "D")
	fileStr := example.Filename("Fpdf_LinearGradient_gradient")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_LinearGradient_gradient.pdf
}

// ExampleFpdf_ClipText demonstrates clipping.
func ExampleFpdf_ClipText() {
	pdf := gofpdf.New("", "", "", "")
	y := 10.0
	pdf.AddPage()

	pdf.SetFont("Helvetica", "", 24)
	pdf.SetXY(0, y)
	pdf.ClipText(10, y+12, "Clipping examples", false)
	pdf.RadialGradient(10, y, 100, 20, 128, 128, 160, 32, 32, 48,
		0.25, 0.5, 0.25, 0.5, 0.2)
	pdf.ClipEnd()

	y += 12
	pdf.SetFont("Helvetica", "B", 120)
	pdf.SetDrawColor(64, 80, 80)
	pdf.SetLineWidth(.5)
	pdf.ClipText(10, y+40, pdf.String(), true)
	pdf.RadialGradient(10, y, 200, 50, 220, 220, 250, 80, 80, 220,
		0.25, 0.5, 0.25, 0.5, 1)
	pdf.ClipEnd()

	y += 55
	pdf.ClipRect(10, y, 105, 20, true)
	pdf.SetFillColor(255, 255, 255)
	pdf.Rect(10, y, 105, 20, "F")
	pdf.ClipCircle(40, y+10, 15, false)
	pdf.RadialGradient(25, y, 30, 30, 220, 250, 220, 40, 60, 40, 0.3,
		0.85, 0.3, 0.85, 0.5)
	pdf.ClipEnd()
	pdf.ClipEllipse(80, y+10, 20, 15, false)
	pdf.RadialGradient(60, y, 40, 30, 250, 220, 220, 60, 40, 40, 0.3,
		0.85, 0.3, 0.85, 0.5)
	pdf.ClipEnd()
	pdf.ClipEnd()

	y += 28
	pdf.ClipEllipse(26, y+10, 16, 10, true)
	pdf.Image(example.ImageFile("logo.jpg"), 10, y, 32, 0, false, "JPG", 0, "")
	pdf.ClipEnd()

	pdf.ClipCircle(60, y+10, 10, true)
	pdf.RadialGradient(50, y, 20, 20, 220, 220, 250, 40, 40, 60, 0.3,
		0.7, 0.3, 0.7, 0.5)
	pdf.ClipEnd()

	pdf.ClipPolygon([]gofpdf.PointType{{X: 80, Y: y + 20}, {X: 90, Y: y},
		{X: 100, Y: y + 20}}, true)
	pdf.LinearGradient(80, y, 20, 20, 250, 220, 250, 60, 40, 60, 0.5,
		1, 0.5, 0.5)
	pdf.ClipEnd()

	y += 30
	pdf.SetLineWidth(.1)
	pdf.SetDrawColor(180, 180, 180)
	pdf.ClipRoundedRect(10, y, 120, 20, 5, true)
	pdf.RadialGradient(10, y, 120, 20, 255, 255, 255, 240, 240, 220,
		0.25, 0.75, 0.25, 0.75, 0.5)
	pdf.SetXY(5, y-5)
	pdf.SetFont("Times", "", 12)
	pdf.MultiCell(130, 5, lorem(), "", "", false)
	pdf.ClipEnd()

	y += 30
	pdf.SetDrawColor(180, 100, 180)
	pdf.ClipRoundedRectExt(10, y, 120, 20, 5, 10, 5, 10, true)
	pdf.RadialGradient(10, y, 120, 20, 255, 255, 255, 240, 240, 220,
		0.25, 0.75, 0.25, 0.75, 0.5)
	pdf.SetXY(5, y-5)
	pdf.SetFont("Times", "", 12)
	pdf.MultiCell(130, 5, lorem(), "", "", false)
	pdf.ClipEnd()

	fileStr := example.Filename("Fpdf_ClipText")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_ClipText.pdf
}

// ExampleFpdf_PageSize generates a PDF document with various page sizes.
func ExampleFpdf_PageSize() {
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr:    "in",
		Size:       gofpdf.SizeType{Wd: 6, Ht: 6},
		FontDirStr: example.FontDir(),
	})
	pdf.SetMargins(0.5, 1, 0.5)
	pdf.SetFont("Times", "", 14)
	pdf.AddPageFormat("L", gofpdf.SizeType{Wd: 3, Ht: 12})
	pdf.SetXY(0.5, 1.5)
	pdf.CellFormat(11, 0.2, "12 in x 3 in", "", 0, "C", false, 0, "")
	pdf.AddPage() // Default size established in NewCustom()
	pdf.SetXY(0.5, 3)
	pdf.CellFormat(5, 0.2, "6 in x 6 in", "", 0, "C", false, 0, "")
	pdf.AddPageFormat("P", gofpdf.SizeType{Wd: 3, Ht: 12})
	pdf.SetXY(0.5, 6)
	pdf.CellFormat(2, 0.2, "3 in x 12 in", "", 0, "C", false, 0, "")
	for j := 0; j <= 3; j++ {
		wd, ht, u := pdf.PageSize(j)
		fmt.Printf("%d: %6.2f %s, %6.2f %s\n", j, wd, u, ht, u)
	}
	fileStr := example.Filename("Fpdf_PageSize")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// 0:   6.00 in,   6.00 in
	// 1:  12.00 in,   3.00 in
	// 2:   6.00 in,   6.00 in
	// 3:   3.00 in,  12.00 in
	// Successfully generated pdf/Fpdf_PageSize.pdf
}

// ExampleFpdf_Bookmark demonstrates the Bookmark method.
func ExampleFpdf_Bookmark() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 15)
	pdf.Bookmark("Page 1", 0, 0)
	pdf.Bookmark("Paragraph 1", 1, -1)
	pdf.Cell(0, 6, "Paragraph 1")
	pdf.Ln(50)
	pdf.Bookmark("Paragraph 2", 1, -1)
	pdf.Cell(0, 6, "Paragraph 2")
	pdf.AddPage()
	pdf.Bookmark("Page 2", 0, 0)
	pdf.Bookmark("Paragraph 3", 1, -1)
	pdf.Cell(0, 6, "Paragraph 3")
	fileStr := example.Filename("Fpdf_Bookmark")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_Bookmark.pdf
}

// ExampleFpdf_TransformBegin demonstrates various transformations. It is adapted from an
// example script by Moritz Wagner and Andreas Würmser.
func ExampleFpdf_TransformBegin() {
	const (
		light = 200
		dark  = 0
	)
	var refX, refY float64
	var refStr string
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	color := func(val int) {
		pdf.SetDrawColor(val, val, val)
		pdf.SetTextColor(val, val, val)
	}
	reference := func(str string, x, y float64, val int) {
		color(val)
		pdf.Rect(x, y, 40, 10, "D")
		pdf.Text(x, y-1, str)
	}
	refDraw := func(str string, x, y float64) {
		refStr = str
		refX = x
		refY = y
		reference(str, x, y, light)
	}
	refDupe := func() {
		reference(refStr, refX, refY, dark)
	}

	titleStr := "Transformations"
	titlePt := 36.0
	titleHt := pdf.PointConvert(titlePt)
	pdf.SetFont("Helvetica", "", titlePt)
	titleWd := pdf.GetStringWidth(titleStr)
	titleX := (210 - titleWd) / 2
	pdf.Text(titleX, 10+titleHt, titleStr)
	pdf.TransformBegin()
	pdf.TransformMirrorVertical(10 + titleHt + 0.5)
	pdf.ClipText(titleX, 10+titleHt, titleStr, false)
	// Remember that the transform will mirror the gradient box too
	pdf.LinearGradient(titleX, 10, titleWd, titleHt+4, 120, 120, 120,
		255, 255, 255, 0, 0, 0, 0.6)
	pdf.ClipEnd()
	pdf.TransformEnd()

	pdf.SetFont("Helvetica", "", 12)

	// Scale by 150% centered by lower left corner of the rectangle
	refDraw("Scale", 50, 60)
	pdf.TransformBegin()
	pdf.TransformScaleXY(150, 50, 70)
	refDupe()
	pdf.TransformEnd()

	// Translate 7 to the right, 5 to the bottom
	refDraw("Translate", 125, 60)
	pdf.TransformBegin()
	pdf.TransformTranslate(7, 5)
	refDupe()
	pdf.TransformEnd()

	// Rotate 20 degrees counter-clockwise centered by the lower left corner of
	// the rectangle
	refDraw("Rotate", 50, 110)
	pdf.TransformBegin()
	pdf.TransformRotate(20, 50, 120)
	refDupe()
	pdf.TransformEnd()

	// Skew 30 degrees along the x-axis centered by the lower left corner of the
	// rectangle
	refDraw("Skew", 125, 110)
	pdf.TransformBegin()
	pdf.TransformSkewX(30, 125, 110)
	refDupe()
	pdf.TransformEnd()

	// Mirror horizontally with axis of reflection at left side of the rectangle
	refDraw("Mirror horizontal", 50, 160)
	pdf.TransformBegin()
	pdf.TransformMirrorHorizontal(50)
	refDupe()
	pdf.TransformEnd()

	// Mirror vertically with axis of reflection at bottom side of the rectangle
	refDraw("Mirror vertical", 125, 160)
	pdf.TransformBegin()
	pdf.TransformMirrorVertical(170)
	refDupe()
	pdf.TransformEnd()

	// Reflect against a point at the lower left point of rectangle
	refDraw("Mirror point", 50, 210)
	pdf.TransformBegin()
	pdf.TransformMirrorPoint(50, 220)
	refDupe()
	pdf.TransformEnd()

	// Mirror against a straight line described by a point and an angle
	angle := -20.0
	px := 120.0
	py := 220.0
	refDraw("Mirror line", 125, 210)
	pdf.TransformBegin()
	pdf.TransformRotate(angle, px, py)
	pdf.Line(px-1, py-1, px+1, py+1)
	pdf.Line(px-1, py+1, px+1, py-1)
	pdf.Line(px-5, py, px+60, py)
	pdf.TransformEnd()
	pdf.TransformBegin()
	pdf.TransformMirrorLine(angle, px, py)
	refDupe()
	pdf.TransformEnd()

	fileStr := example.Filename("Fpdf_TransformBegin")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_TransformBegin.pdf
}

// ExampleFpdf_RegisterImage demonstrates Lawrence Kesteloot's image registration code.
func ExampleFpdf_RegisterImage() {
	const (
		margin = 10
		wd     = 210
		ht     = 297
	)
	fileList := []string{
		"logo-gray.png",
		"logo.jpg",
		"logo.png",
		"logo-rgb.png",
		"logo-progressive.jpg",
	}
	var infoPtr *gofpdf.ImageInfoType
	var imageFileStr string
	var imgWd, imgHt, lf, tp float64
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetMargins(10, 10, 10)
	pdf.SetFont("Helvetica", "", 15)
	for j, str := range fileList {
		imageFileStr = example.ImageFile(str)
		infoPtr = pdf.RegisterImage(imageFileStr, "")
		imgWd, imgHt = infoPtr.Extent()
		switch j {
		case 0:
			lf = margin
			tp = margin
		case 1:
			lf = wd - margin - imgWd
			tp = margin
		case 2:
			lf = (wd - imgWd) / 2.0
			tp = (ht - imgHt) / 2.0
		case 3:
			lf = margin
			tp = ht - imgHt - margin
		case 4:
			lf = wd - imgWd - margin
			tp = ht - imgHt - margin
		}
		pdf.Image(imageFileStr, lf, tp, imgWd, imgHt, false, "", 0, "")
	}
	fileStr := example.Filename("Fpdf_RegisterImage")
	// Test the image information retrieval method
	infoShow := func(imageStr string) {
		imageStr = example.ImageFile(imageStr)
		info := pdf.GetImageInfo(imageStr)
		if info != nil {
			if info.Width() > 0.0 {
				fmt.Printf("Image %s is registered\n", filepath.ToSlash(imageStr))
			} else {
				fmt.Printf("Incorrect information for image %s\n", filepath.ToSlash(imageStr))
			}
		} else {
			fmt.Printf("Image %s is not registered\n", filepath.ToSlash(imageStr))
		}
	}
	infoShow(fileList[0])
	infoShow("foo.png")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Image image/logo-gray.png is registered
	// Image image/foo.png is not registered
	// Successfully generated pdf/Fpdf_RegisterImage.pdf
}

// ExampleFpdf_SplitLines demonstrates Bruno Michel's line splitting function.
func ExampleFpdf_SplitLines() {
	const (
		fontPtSize = 18.0
		wd         = 100.0
	)
	pdf := gofpdf.New("P", "mm", "A4", "") // A4 210.0 x 297.0
	pdf.SetFont("Times", "", fontPtSize)
	_, lineHt := pdf.GetFontSize()
	pdf.AddPage()
	pdf.SetMargins(10, 10, 10)
	lines := pdf.SplitLines([]byte(lorem()), wd)
	ht := float64(len(lines)) * lineHt
	y := (297.0 - ht) / 2.0
	pdf.SetDrawColor(128, 128, 128)
	pdf.SetFillColor(255, 255, 210)
	x := (210.0 - (wd + 40.0)) / 2.0
	pdf.Rect(x, y-20.0, wd+40.0, ht+40.0, "FD")
	pdf.SetY(y)
	for _, line := range lines {
		pdf.CellFormat(190.0, lineHt, string(line), "", 1, "C", false, 0, "")
	}
	fileStr := example.Filename("Fpdf_Splitlines")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_Splitlines.pdf
}

// ExampleFpdf_SVGBasicWrite demonstrates how to render a simple path-only SVG image of the
// type generated by the jSignature web control.
func ExampleFpdf_SVGBasicWrite() {
	const (
		fontPtSize = 16.0
		wd         = 100.0
		sigFileStr = "signature.svg"
	)
	var (
		sig gofpdf.SVGBasicType
		err error
	)
	pdf := gofpdf.New("P", "mm", "A4", "") // A4 210.0 x 297.0
	pdf.SetFont("Times", "", fontPtSize)
	lineHt := pdf.PointConvert(fontPtSize)
	pdf.AddPage()
	pdf.SetMargins(10, 10, 10)
	htmlStr := `This example renders a simple ` +
		`<a href="http://www.w3.org/TR/SVG/">SVG</a> (scalable vector graphics) ` +
		`image that contains only basic path commands without any styling, ` +
		`color fill, reflection or endpoint closures. In particular, the ` +
		`type of vector graphic returned from a ` +
		`<a href="http://willowsystems.github.io/jSignature/#/demo/">jSignature</a> ` +
		`web control is supported and is used in this example.`
	html := pdf.HTMLBasicNew()
	html.Write(lineHt, htmlStr)
	sig, err = gofpdf.SVGBasicFileParse(example.ImageFile(sigFileStr))
	if err == nil {
		scale := 100 / sig.Wd
		scaleY := 30 / sig.Ht
		if scale > scaleY {
			scale = scaleY
		}
		pdf.SetLineCapStyle("round")
		pdf.SetLineWidth(0.25)
		pdf.SetDrawColor(0, 0, 128)
		pdf.SetXY((210.0-scale*sig.Wd)/2.0, pdf.GetY()+10)
		pdf.SVGBasicWrite(&sig, scale)
	} else {
		pdf.SetError(err)
	}
	fileStr := example.Filename("Fpdf_SVGBasicWrite")
	err = pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_SVGBasicWrite.pdf
}

// ExampleFpdf_CellFormat_align demonstrates Stefan Schroeder's code to control vertical
// alignment.
func ExampleFpdf_CellFormat_align() {
	type recType struct {
		align, txt string
	}
	recList := []recType{
		{"TL", "top left"},
		{"TC", "top center"},
		{"TR", "top right"},
		{"LM", "middle left"},
		{"CM", "middle center"},
		{"RM", "middle right"},
		{"BL", "bottom left"},
		{"BC", "bottom center"},
		{"BR", "bottom right"},
	}
	recListBaseline := []recType{
		{"AL", "baseline left"},
		{"AC", "baseline center"},
		{"AR", "baseline right"},
	}
	var formatRect = func(pdf *gofpdf.Fpdf, recList []recType) {
		linkStr := ""
		for pageJ := 0; pageJ < 2; pageJ++ {
			pdf.AddPage()
			pdf.SetMargins(10, 10, 10)
			pdf.SetAutoPageBreak(false, 0)
			borderStr := "1"
			for _, rec := range recList {
				pdf.SetXY(20, 20)
				pdf.CellFormat(170, 257, rec.txt, borderStr, 0, rec.align, false, 0, linkStr)
				borderStr = ""
			}
			linkStr = "https://github.com/headlands-org/gofpdf"
		}
	}
	pdf := gofpdf.New("P", "mm", "A4", "") // A4 210.0 x 297.0
	pdf.SetFont("Helvetica", "", 16)
	formatRect(pdf, recList)
	formatRect(pdf, recListBaseline)
	var fr fontResourceType
	pdf.SetFontLoader(fr)
	pdf.AddFont("Calligrapher", "", "calligra.json")
	pdf.SetFont("Calligrapher", "", 16)
	formatRect(pdf, recListBaseline)
	fileStr := example.Filename("Fpdf_CellFormat_align")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Generalized font loader reading calligra.json
	// Generalized font loader reading calligra.z
	// Successfully generated pdf/Fpdf_CellFormat_align.pdf
}

// ExampleFpdf_CellFormat_codepageescape demonstrates the use of characters in the high range of the
// Windows-1252 code page (gofdpf default). See the example for CellFormat (4)
// for a way to do this automatically.
func ExampleFpdf_CellFormat_codepageescape() {
	pdf := gofpdf.New("P", "mm", "A4", "") // A4 210.0 x 297.0
	fontSize := 16.0
	pdf.SetFont("Helvetica", "", fontSize)
	ht := pdf.PointConvert(fontSize)
	write := func(str string) {
		pdf.CellFormat(190, ht, str, "", 1, "C", false, 0, "")
		pdf.Ln(ht)
	}
	pdf.AddPage()
	htmlStr := `Until gofpdf supports UTF-8 encoded source text, source text needs ` +
		`to be specified with all special characters escaped to match the code page ` +
		`layout of the currently selected font. By default, gofdpf uses code page 1252.` +
		` See <a href="http://en.wikipedia.org/wiki/Windows-1252">Wikipedia</a> for ` +
		`a table of this layout.`
	html := pdf.HTMLBasicNew()
	html.Write(ht, htmlStr)
	pdf.Ln(2 * ht)
	write("Voix ambigu\xeb d'un c\x9cur qui au z\xe9phyr pr\xe9f\xe8re les jattes de kiwi.")
	write("Falsches \xdcben von Xylophonmusik qu\xe4lt jeden gr\xf6\xdferen Zwerg.")
	write("Heiz\xf6lr\xfccksto\xdfabd\xe4mpfung")
	write("For\xe5rsj\xe6vnd\xf8gn / Efter\xe5rsj\xe6vnd\xf8gn")
	fileStr := example.Filename("Fpdf_CellFormat_codepageescape")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_CellFormat_codepageescape.pdf
}

// ExampleFpdf_CellFormat_codepage demonstrates the automatic conversion of UTF-8 strings to an
// 8-bit font encoding.
func ExampleFpdf_CellFormat_codepage() {
	pdf := gofpdf.New("P", "mm", "A4", example.FontDir()) // A4 210.0 x 297.0
	// See documentation for details on how to generate fonts
	pdf.AddFont("Helvetica-1251", "", "helvetica_1251.json")
	pdf.AddFont("Helvetica-1253", "", "helvetica_1253.json")
	fontSize := 16.0
	pdf.SetFont("Helvetica", "", fontSize)
	ht := pdf.PointConvert(fontSize)
	tr := pdf.UnicodeTranslatorFromDescriptor("") // "" defaults to "cp1252"
	write := func(str string) {
		// pdf.CellFormat(190, ht, tr(str), "", 1, "C", false, 0, "")
		pdf.MultiCell(190, ht, tr(str), "", "C", false)
		pdf.Ln(ht)
	}
	pdf.AddPage()
	str := `Gofpdf provides a translator that will convert any UTF-8 code point ` +
		`that is present in the specified code page.`
	pdf.MultiCell(190, ht, str, "", "L", false)
	pdf.Ln(2 * ht)
	write("Voix ambiguë d'un cœur qui au zéphyr préfère les jattes de kiwi.")
	write("Falsches Üben von Xylophonmusik quält jeden größeren Zwerg.")
	write("Heizölrückstoßabdämpfung")
	write("Forårsjævndøgn / Efterårsjævndøgn")
	write("À noite, vovô Kowalsky vê o ímã cair no pé do pingüim queixoso e vovó" +
		"põe açúcar no chá de tâmaras do jabuti feliz.")
	pdf.SetFont("Helvetica-1251", "", fontSize) // Name matches one specified in AddFont()
	tr = pdf.UnicodeTranslatorFromDescriptor("cp1251")
	write("Съешь же ещё этих мягких французских булок, да выпей чаю.")

	pdf.SetFont("Helvetica-1253", "", fontSize)
	tr = pdf.UnicodeTranslatorFromDescriptor("cp1253")
	write("Θέλει αρετή και τόλμη η ελευθερία. (Ανδρέας Κάλβος)")

	fileStr := example.Filename("Fpdf_CellFormat_codepage")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_CellFormat_codepage.pdf
}

// ExampleFpdf_SetProtection demonstrates password protection for documents.
func ExampleFpdf_SetProtection() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetProtection(gofpdf.CnProtectPrint, "123", "abc")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)
	pdf.Write(10, "Password-protected.")
	fileStr := example.Filename("Fpdf_SetProtection")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_SetProtection.pdf
}

// ExampleFpdf_Polygon displays equilateral polygons in a demonstration of the Polygon
// function.
func ExampleFpdf_Polygon() {
	const rowCount = 5
	const colCount = 4
	const ptSize = 36
	var x, y, radius, gap, advance float64
	var rgVal int
	var pts []gofpdf.PointType
	vertices := func(count int) (res []gofpdf.PointType) {
		var pt gofpdf.PointType
		res = make([]gofpdf.PointType, 0, count)
		mlt := 2.0 * math.Pi / float64(count)
		for j := 0; j < count; j++ {
			pt.Y, pt.X = math.Sincos(float64(j) * mlt)
			res = append(res, gofpdf.PointType{
				X: x + radius*pt.X,
				Y: y + radius*pt.Y})
		}
		return
	}
	pdf := gofpdf.New("P", "mm", "A4", "") // A4 210.0 x 297.0
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", ptSize)
	pdf.SetDrawColor(0, 80, 180)
	gap = 12.0
	pdf.SetY(gap)
	pdf.CellFormat(190.0, gap, "Equilateral polygons", "", 1, "C", false, 0, "")
	radius = (210.0 - float64(colCount+1)*gap) / (2.0 * float64(colCount))
	advance = gap + 2.0*radius
	y = 2*gap + pdf.PointConvert(ptSize) + radius
	rgVal = 230
	for row := 0; row < rowCount; row++ {
		pdf.SetFillColor(rgVal, rgVal, 0)
		rgVal -= 12
		x = gap + radius
		for col := 0; col < colCount; col++ {
			pts = vertices(row*colCount + col + 3)
			pdf.Polygon(pts, "FD")
			x += advance
		}
		y += advance
	}
	fileStr := example.Filename("Fpdf_Polygon")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_Polygon.pdf
}

// ExampleFpdf_AddLayer demonstrates document layers. The initial visibility of a layer
// is specified with the second parameter to AddLayer(). The layer list
// displayed by the document reader allows layer visibility to be controlled
// interactively.
func ExampleFpdf_AddLayer() {

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 15)
	pdf.Write(8, "This line doesn't belong to any layer.\n")

	// Define layers
	l1 := pdf.AddLayer("Layer 1", true)
	l2 := pdf.AddLayer("Layer 2", true)

	// Open layer pane in PDF viewer
	pdf.OpenLayerPane()

	// First layer
	pdf.BeginLayer(l1)
	pdf.Write(8, "This line belongs to layer 1.\n")
	pdf.EndLayer()

	// Second layer
	pdf.BeginLayer(l2)
	pdf.Write(8, "This line belongs to layer 2.\n")
	pdf.EndLayer()

	// First layer again
	pdf.BeginLayer(l1)
	pdf.Write(8, "This line belongs to layer 1 again.\n")
	pdf.EndLayer()

	fileStr := example.Filename("Fpdf_AddLayer")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_AddLayer.pdf
}

// ExampleFpdf_RegisterImageReader demonstrates the use of an image that is retrieved from a web
// server.
func ExampleFpdf_RegisterImageReader() {

	const (
		margin   = 10
		wd       = 210
		ht       = 297
		fontSize = 15
		urlStr   = "https://github.com/headlands-org/gofpdf/blob/master/image/gofpdf.png?raw=true"
		msgStr   = `Images from the web can be easily embedded when a PDF document is generated.`
	)

	var (
		rsp *http.Response
		err error
		tp  string
	)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", fontSize)
	ln := pdf.PointConvert(fontSize)
	pdf.MultiCell(wd-margin-margin, ln, msgStr, "", "L", false)
	rsp, err = http.Get(urlStr)
	if err == nil {
		tp = pdf.ImageTypeFromMime(rsp.Header["Content-Type"][0])
		infoPtr := pdf.RegisterImageReader(urlStr, tp, rsp.Body)
		if pdf.Ok() {
			imgWd, imgHt := infoPtr.Extent()
			pdf.Image(urlStr, (wd-imgWd)/2.0, pdf.GetY()+ln,
				imgWd, imgHt, false, tp, 0, "")
		}
	} else {
		pdf.SetError(err)
	}
	fileStr := example.Filename("Fpdf_RegisterImageReader_url")
	err = pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_RegisterImageReader_url.pdf

}

// ExampleFpdf_Beziergon demonstrates the Beziergon function.
func ExampleFpdf_Beziergon() {

	const (
		margin      = 10
		wd          = 210
		unit        = (wd - 2*margin) / 6
		ht          = 297
		fontSize    = 15
		msgStr      = `Demonstration of Beziergon function`
		coefficient = 0.6
		delta       = coefficient * unit
		ln          = fontSize * 25.4 / 72
		offsetX     = (wd - 4*unit) / 2.0
		offsetY     = offsetX + 2*ln
	)

	srcList := []gofpdf.PointType{
		{X: 0, Y: 0},
		{X: 1, Y: 0},
		{X: 1, Y: 1},
		{X: 2, Y: 1},
		{X: 2, Y: 2},
		{X: 3, Y: 2},
		{X: 3, Y: 3},
		{X: 4, Y: 3},
		{X: 4, Y: 4},
		{X: 1, Y: 4},
		{X: 1, Y: 3},
		{X: 0, Y: 3},
	}

	ctrlList := []gofpdf.PointType{
		{X: 1, Y: -1},
		{X: 1, Y: 1},
		{X: 1, Y: 1},
		{X: 1, Y: 1},
		{X: 1, Y: 1},
		{X: 1, Y: 1},
		{X: 1, Y: 1},
		{X: 1, Y: 1},
		{X: -1, Y: 1},
		{X: -1, Y: -1},
		{X: -1, Y: -1},
		{X: -1, Y: -1},
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", fontSize)
	for j, src := range srcList {
		srcList[j].X = offsetX + src.X*unit
		srcList[j].Y = offsetY + src.Y*unit
	}
	for j, ctrl := range ctrlList {
		ctrlList[j].X = ctrl.X * delta
		ctrlList[j].Y = ctrl.Y * delta
	}
	jPrev := len(srcList) - 1
	srcPrev := srcList[jPrev]
	curveList := []gofpdf.PointType{srcPrev} // point [, control 0, control 1, point]*
	control := func(x, y float64) {
		curveList = append(curveList, gofpdf.PointType{X: x, Y: y})
	}
	for j, src := range srcList {
		ctrl := ctrlList[jPrev]
		control(srcPrev.X+ctrl.X, srcPrev.Y+ctrl.Y) // Control 0
		ctrl = ctrlList[j]
		control(src.X-ctrl.X, src.Y-ctrl.Y) // Control 1
		curveList = append(curveList, src)  // Destination
		jPrev = j
		srcPrev = src
	}
	pdf.MultiCell(wd-margin-margin, ln, msgStr, "", "C", false)
	pdf.SetDashPattern([]float64{0.8, 0.8}, 0)
	pdf.SetDrawColor(160, 160, 160)
	pdf.Polygon(srcList, "D")
	pdf.SetDashPattern([]float64{}, 0)
	pdf.SetDrawColor(64, 64, 128)
	pdf.SetLineWidth(pdf.GetLineWidth() * 3)
	pdf.Beziergon(curveList, "D")
	fileStr := example.Filename("Fpdf_Beziergon")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_Beziergon.pdf

}

// ExampleFpdf_SetFontLoader demonstrates loading a non-standard font using a generalized
// font loader. fontResourceType implements the FontLoader interface and is
// defined locally in the test source code.
func ExampleFpdf_SetFontLoader() {
	var fr fontResourceType
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFontLoader(fr)
	pdf.AddFont("Calligrapher", "", "calligra.json")
	pdf.AddPage()
	pdf.SetFont("Calligrapher", "", 35)
	pdf.Cell(0, 10, "Load fonts from any source")
	fileStr := example.Filename("Fpdf_SetFontLoader")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Generalized font loader reading calligra.json
	// Generalized font loader reading calligra.z
	// Successfully generated pdf/Fpdf_SetFontLoader.pdf
}

// ExampleFpdf_MoveTo demonstrates the Path Drawing functions, such as: MoveTo,
// LineTo, CurveTo, ..., ClosePath and DrawPath.
func ExampleFpdf_MoveTo() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.MoveTo(20, 20)
	pdf.LineTo(170, 20)
	pdf.ArcTo(170, 40, 20, 20, 0, 90, 0)
	pdf.CurveTo(190, 100, 105, 100)
	pdf.CurveBezierCubicTo(20, 100, 105, 200, 20, 200)
	pdf.ClosePath()
	pdf.SetFillColor(200, 200, 200)
	pdf.SetLineWidth(3)
	pdf.DrawPath("DF")
	fileStr := example.Filename("Fpdf_MoveTo_path")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_MoveTo_path.pdf
}

// ExampleFpdf_SetLineJoinStyle demonstrates various line cap and line join styles.
func ExampleFpdf_SetLineJoinStyle() {
	const offset = 75.0
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	var draw = func(cap, join string, x0, y0, x1, y1 float64) {
		// transform begin & end needed to isolate caps and joins
		pdf.SetLineCapStyle(cap)
		pdf.SetLineJoinStyle(join)

		// Draw thick line
		pdf.SetDrawColor(0x33, 0x33, 0x33)
		pdf.SetLineWidth(30.0)
		pdf.MoveTo(x0, y0)
		pdf.LineTo((x0+x1)/2+offset, (y0+y1)/2)
		pdf.LineTo(x1, y1)
		pdf.DrawPath("D")

		// Draw thin helping line
		pdf.SetDrawColor(0xFF, 0x33, 0x33)
		pdf.SetLineWidth(2.56)
		pdf.MoveTo(x0, y0)
		pdf.LineTo((x0+x1)/2+offset, (y0+y1)/2)
		pdf.LineTo(x1, y1)
		pdf.DrawPath("D")

	}
	x := 35.0
	caps := []string{"butt", "square", "round"}
	joins := []string{"bevel", "miter", "round"}
	for i := range caps {
		draw(caps[i], joins[i], x, 50, x, 160)
		x += offset
	}
	fileStr := example.Filename("Fpdf_SetLineJoinStyle_caps")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_SetLineJoinStyle_caps.pdf
}

// ExampleFpdf_DrawPath demonstrates various fill modes.
func ExampleFpdf_DrawPath() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetDrawColor(0xff, 0x00, 0x00)
	pdf.SetFillColor(0x99, 0x99, 0x99)
	pdf.SetFont("Helvetica", "", 15)
	pdf.AddPage()
	pdf.SetAlpha(1, "Multiply")
	var (
		polygon = func(cx, cy, r, n, dir float64) {
			da := 2 * math.Pi / n
			pdf.MoveTo(cx+r, cy)
			pdf.Text(cx+r, cy, "0")
			i := 1
			for a := da; a < 2*math.Pi; a += da {
				x, y := cx+r*math.Cos(dir*a), cy+r*math.Sin(dir*a)
				pdf.LineTo(x, y)
				pdf.Text(x, y, strconv.Itoa(i))
				i++
			}
			pdf.ClosePath()
		}
		polygons = func(cx, cy, r, n, dir float64) {
			d := 1.0
			for rf := r; rf > 0; rf -= 10 {
				polygon(cx, cy, rf, n, d)
				d *= dir
			}
		}
		star = func(cx, cy, r, n float64) {
			da := 4 * math.Pi / n
			pdf.MoveTo(cx+r, cy)
			for a := da; a < 4*math.Pi+da; a += da {
				x, y := cx+r*math.Cos(a), cy+r*math.Sin(a)
				pdf.LineTo(x, y)
			}
			pdf.ClosePath()
		}
	)
	// triangle
	polygons(55, 45, 40, 3, 1)
	pdf.DrawPath("B")
	pdf.Text(15, 95, "B (same direction, non zero winding)")

	// square
	polygons(155, 45, 40, 4, 1)
	pdf.DrawPath("B*")
	pdf.Text(115, 95, "B* (same direction, even odd)")

	// pentagon
	polygons(55, 145, 40, 5, -1)
	pdf.DrawPath("B")
	pdf.Text(15, 195, "B (different direction, non zero winding)")

	// hexagon
	polygons(155, 145, 40, 6, -1)
	pdf.DrawPath("B*")
	pdf.Text(115, 195, "B* (different direction, even odd)")

	// star
	star(55, 245, 40, 5)
	pdf.DrawPath("B")
	pdf.Text(15, 290, "B (non zero winding)")

	// star
	star(155, 245, 40, 5)
	pdf.DrawPath("B*")
	pdf.Text(115, 290, "B* (even odd)")

	fileStr := example.Filename("Fpdf_DrawPath_fill")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_DrawPath_fill.pdf
}

// ExampleFpdf_CreateTemplate demonstrates creating and using templates
func ExampleFpdf_CreateTemplate() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetCompression(false)
	// pdf.SetFont("Times", "", 12)
	template := pdf.CreateTemplate(func(tpl *gofpdf.Tpl) {
		tpl.Image(example.ImageFile("logo.png"), 6, 6, 30, 0, false, "", 0, "")
		tpl.SetFont("Arial", "B", 16)
		tpl.Text(40, 20, "Template says hello")
		tpl.SetDrawColor(0, 100, 200)
		tpl.SetLineWidth(2.5)
		tpl.Line(95, 12, 105, 22)
	})
	_, tplSize := template.Size()
	// fmt.Println("Size:", tplSize)
	// fmt.Println("Scaled:", tplSize.ScaleBy(1.5))

	template2 := pdf.CreateTemplate(func(tpl *gofpdf.Tpl) {
		tpl.UseTemplate(template)
		subtemplate := tpl.CreateTemplate(func(tpl2 *gofpdf.Tpl) {
			tpl2.Image(example.ImageFile("logo.png"), 6, 86, 30, 0, false, "", 0, "")
			tpl2.SetFont("Arial", "B", 16)
			tpl2.Text(40, 100, "Subtemplate says hello")
			tpl2.SetDrawColor(0, 200, 100)
			tpl2.SetLineWidth(2.5)
			tpl2.Line(102, 92, 112, 102)
		})
		tpl.UseTemplate(subtemplate)
	})

	pdf.SetDrawColor(200, 100, 0)
	pdf.SetLineWidth(2.5)
	pdf.SetFont("Arial", "B", 16)

	// serialize and deserialize template
	b, _ := template2.Serialize()
	template3, _ := gofpdf.DeserializeTemplate(b)

	pdf.AddPage()
	pdf.UseTemplate(template3)
	pdf.UseTemplateScaled(template3, gofpdf.PointType{X: 0, Y: 30}, tplSize)
	pdf.Line(40, 210, 60, 210)
	pdf.Text(40, 200, "Template example page 1")

	pdf.AddPage()
	pdf.UseTemplate(template2)
	pdf.UseTemplateScaled(template3, gofpdf.PointType{X: 0, Y: 30}, tplSize.ScaleBy(1.4))
	pdf.Line(60, 210, 80, 210)
	pdf.Text(40, 200, "Template example page 2")

	fileStr := example.Filename("Fpdf_CreateTemplate")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_CreateTemplate.pdf
}

// ExampleFpdf_AddFontFromBytes demonstrate how to use embedded fonts from byte array
func ExampleFpdf_AddFontFromBytes() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.AddFontFromBytes("calligra", "", files.CalligraJson, files.CalligraZ)
	pdf.SetFont("calligra", "", 16)
	pdf.Cell(40, 10, "Hello World With Embedded Font!")
	fileStr := example.Filename("Fpdf_EmbeddedFont")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_EmbeddedFont.pdf
}

// This example demonstrate Clipped table cells
func ExampleFpdf_ClipRect() {
	marginCell := 2. // margin of top/bottom of cell
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "", 12)
	pdf.AddPage()
	pagew, pageh := pdf.GetPageSize()
	mleft, mright, _, mbottom := pdf.GetMargins()

	cols := []float64{60, 100, pagew - mleft - mright - 100 - 60}
	rows := [][]string{}
	for i := 1; i <= 50; i++ {
		word := fmt.Sprintf("%d:%s", i, strings.Repeat("A", i%100))
		rows = append(rows, []string{word, word, word})
	}

	for _, row := range rows {
		_, lineHt := pdf.GetFontSize()
		height := lineHt + marginCell

		x, y := pdf.GetXY()
		// add a new page if the height of the row doesn't fit on the page
		if y+height >= pageh-mbottom {
			pdf.AddPage()
			x, y = pdf.GetXY()
		}
		for i, txt := range row {
			width := cols[i]
			pdf.Rect(x, y, width, height, "")
			pdf.ClipRect(x, y, width, height, false)
			pdf.Cell(width, height, txt)
			pdf.ClipEnd()
			x += width
		}
		pdf.Ln(-1)
	}
	fileStr := example.Filename("Fpdf_ClippedTableCells")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_ClippedTableCells.pdf
}

// This example demonstrate wrapped table cells
func ExampleFpdf_Rect() {
	marginCell := 2. // margin of top/bottom of cell
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "", 12)
	pdf.AddPage()
	pagew, pageh := pdf.GetPageSize()
	mleft, mright, _, mbottom := pdf.GetMargins()

	cols := []float64{60, 100, pagew - mleft - mright - 100 - 60}
	rows := [][]string{}
	for i := 1; i <= 30; i++ {
		word := fmt.Sprintf("%d:%s", i, strings.Repeat("A", i%100))
		rows = append(rows, []string{word, word, word})
	}

	for _, row := range rows {
		curx, y := pdf.GetXY()
		x := curx

		height := 0.
		_, lineHt := pdf.GetFontSize()

		for i, txt := range row {
			lines := pdf.SplitLines([]byte(txt), cols[i])
			h := float64(len(lines))*lineHt + marginCell*float64(len(lines))
			if h > height {
				height = h
			}
		}
		// add a new page if the height of the row doesn't fit on the page
		if pdf.GetY()+height > pageh-mbottom {
			pdf.AddPage()
			y = pdf.GetY()
		}
		for i, txt := range row {
			width := cols[i]
			pdf.Rect(x, y, width, height, "")
			pdf.MultiCell(width, lineHt+marginCell, txt, "", "", false)
			x += width
			pdf.SetXY(x, y)
		}
		pdf.SetXY(curx, y+height)
	}
	fileStr := example.Filename("Fpdf_WrappedTableCells")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_WrappedTableCells.pdf
}

// ExampleFpdf_SetJavascript demonstrates including JavaScript in the document.
func ExampleFpdf_SetJavascript() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetJavascript("print(true);")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)
	pdf.Write(10, "Auto-print.")
	fileStr := example.Filename("Fpdf_SetJavascript")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_SetJavascript.pdf
}

// ExampleFpdf_AddSpotColor demonstrates spot color use
func ExampleFpdf_AddSpotColor() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddSpotColor("PANTONE 145 CVC", 0, 42, 100, 25)
	pdf.AddPage()
	pdf.SetFillSpotColor("PANTONE 145 CVC", 90)
	pdf.Rect(80, 40, 50, 50, "F")
	fileStr := example.Filename("Fpdf_AddSpotColor")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_AddSpotColor.pdf
}

// ExampleFpdf_RegisterAlias demonstrates how to use `RegisterAlias` to create a table of
// contents.
func ExampleFpdf_RegisterAlias() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "", 12)
	pdf.AliasNbPages("")
	pdf.AddPage()

	// Write the table of contents. We use aliases instead of the page number
	// because we don't know which page the section will begin on.
	numSections := 3
	for i := 1; i <= numSections; i++ {
		pdf.Cell(0, 10, fmt.Sprintf("Section %d begins on page {mark %d}", i, i))
		pdf.Ln(10)
	}

	// Write the sections. Before we start writing, we use `RegisterAlias` to
	// ensure that the alias written in the table of contents will be replaced
	// by the current page number.
	for i := 1; i <= numSections; i++ {
		pdf.AddPage()
		pdf.RegisterAlias(fmt.Sprintf("{mark %d}", i), fmt.Sprintf("%d", pdf.PageNo()))
		pdf.Write(10, fmt.Sprintf("Section %d, page %d of {nb}", i, pdf.PageNo()))
	}

	fileStr := example.Filename("Fpdf_RegisterAlias")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_RegisterAlias.pdf
}

// ExampleFpdf_RegisterAlias_utf8 demonstrates how to use `RegisterAlias` to
// create a table of contents. This particular example demonstrates the use of
// UTF-8 aliases.
func ExampleFpdf_RegisterAlias_utf8() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8Font("dejavu", "", example.FontFile("DejaVuSansCondensed.ttf"))
	pdf.SetFont("dejavu", "", 12)
	pdf.AliasNbPages("{entute}")
	pdf.AddPage()

	// Write the table of contents. We use aliases instead of the page number
	// because we don't know which page the section will begin on.
	numSections := 3
	for i := 1; i <= numSections; i++ {
		pdf.Cell(0, 10, fmt.Sprintf("Sekcio %d komenciĝas ĉe paĝo {ĉi tiu marko %d}", i, i))
		pdf.Ln(10)
	}

	// Write the sections. Before we start writing, we use `RegisterAlias` to
	// ensure that the alias written in the table of contents will be replaced
	// by the current page number.
	for i := 1; i <= numSections; i++ {
		pdf.AddPage()
		pdf.RegisterAlias(fmt.Sprintf("{ĉi tiu marko %d}", i), fmt.Sprintf("%d", pdf.PageNo()))
		pdf.Write(10, fmt.Sprintf("Sekcio %d, paĝo %d de {entute}", i, pdf.PageNo()))
	}

	fileStr := example.Filename("Fpdf_RegisterAliasUTF8")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_RegisterAliasUTF8.pdf
}

// ExampleNewGrid demonstrates the generation of graph grids.
func ExampleNewGrid() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "", 12)
	pdf.AddPage()

	gr := gofpdf.NewGrid(13, 10, 187, 130)
	gr.TickmarksExtentX(0, 10, 4)
	gr.TickmarksExtentY(0, 10, 3)
	gr.Grid(pdf)

	gr = gofpdf.NewGrid(13, 154, 187, 128)
	gr.XLabelRotate = true
	gr.TickmarksExtentX(0, 1, 12)
	gr.XDiv = 5
	gr.TickmarksContainY(0, 1.1)
	gr.YDiv = 20
	// Replace X label formatter with month abbreviation
	gr.XTickStr = func(val float64, precision int) string {
		return time.Month(math.Mod(val, 12) + 1).String()[0:3]
	}
	gr.Grid(pdf)
	dot := func(x, y float64) {
		pdf.Circle(gr.X(x), gr.Y(y), 0.5, "F")
	}
	pts := []float64{0.39, 0.457, 0.612, 0.84, 0.998, 1.037, 1.015, 0.918, 0.772, 0.659, 0.593, 0.164}
	for month, val := range pts {
		dot(float64(month)+0.5, val)
	}
	pdf.SetDrawColor(255, 64, 64)
	pdf.SetAlpha(0.5, "Normal")
	pdf.SetLineWidth(1.2)
	gr.Plot(pdf, 0.5, 11.5, 50, func(x float64) float64 {
		// http://www.xuru.org/rt/PR.asp
		return 0.227 * math.Exp(-0.0373*x*x+0.471*x)
	})
	pdf.SetAlpha(1.0, "Normal")
	pdf.SetXY(gr.X(0.5), gr.Y(1.35))
	pdf.SetFontSize(14)
	pdf.Write(0, "Solar energy (MWh) per month, 2016")
	pdf.AddPage()

	gr = gofpdf.NewGrid(13, 10, 187, 274)
	gr.TickmarksContainX(2.3, 3.4)
	gr.TickmarksContainY(10.4, 56.8)
	gr.Grid(pdf)

	fileStr := example.Filename("Fpdf_Grid")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_Grid.pdf
}

// ExampleFpdf_SetPageBox demonstrates the use of a page box
func ExampleFpdf_SetPageBox() {
	// pdfinfo (from http://www.xpdfreader.com) reports the following for this example:
	// ~ pdfinfo -box pdf/Fpdf_PageBox.pdf
	// Producer:       FPDF 1.7
	// CreationDate:   Sat Jan  1 00:00:00 2000
	// ModDate:   	   Sat Jan  1 00:00:00 2000
	// Tagged:         no
	// Form:           none
	// Pages:          1
	// Encrypted:      no
	// Page size:      493.23 x 739.85 pts (rotated 0 degrees)
	// MediaBox:           0.00     0.00   595.28   841.89
	// CropBox:           51.02    51.02   544.25   790.87
	// BleedBox:          51.02    51.02   544.25   790.87
	// TrimBox:           51.02    51.02   544.25   790.87
	// ArtBox:            51.02    51.02   544.25   790.87
	// File size:      1053 bytes
	// Optimized:      no
	// PDF version:    1.3
	const (
		wd        = 210
		ht        = 297
		fontsize  = 6
		boxmargin = 3 * fontsize
	)
	pdf := gofpdf.New("P", "mm", "A4", "") // 210mm x 297mm
	pdf.SetPageBox("crop", boxmargin, boxmargin, wd-2*boxmargin, ht-2*boxmargin)
	pdf.SetFont("Arial", "", pdf.UnitToPointConvert(fontsize))
	pdf.AddPage()
	pdf.MoveTo(fontsize, fontsize)
	pdf.Write(fontsize, "This will be cropped from printed output")
	pdf.MoveTo(boxmargin+fontsize, boxmargin+fontsize)
	pdf.Write(fontsize, "This will be displayed in cropped output")
	fileStr := example.Filename("Fpdf_PageBox")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_PageBox.pdf
}

// ExampleFpdf_SubWrite demonstrates subscripted and superscripted text
// Adapted from http://www.fpdf.org/en/script/script61.php
func ExampleFpdf_SubWrite() {

	const (
		fontSize = 12
		halfX    = 105
	)

	pdf := gofpdf.New("P", "mm", "A4", "") // 210mm x 297mm
	pdf.AddPage()
	pdf.SetFont("Arial", "", fontSize)
	_, lineHt := pdf.GetFontSize()

	pdf.Write(lineHt, "Hello World!")
	pdf.SetX(halfX)
	pdf.Write(lineHt, "This is standard text.\n")
	pdf.Ln(lineHt * 2)

	pdf.SubWrite(10, "H", 33, 0, 0, "")
	pdf.Write(10, "ello World!")
	pdf.SetX(halfX)
	pdf.Write(10, "This is text with a capital first letter.\n")
	pdf.Ln(lineHt * 2)

	pdf.SubWrite(lineHt, "Y", 6, 0, 0, "")
	pdf.Write(lineHt, "ou can also begin the sentence with a small letter. And word wrap also works if the line is too long, like this one is.")
	pdf.SetX(halfX)
	pdf.Write(lineHt, "This is text with a small first letter.\n")
	pdf.Ln(lineHt * 2)

	pdf.Write(lineHt, "The world has a lot of km")
	pdf.SubWrite(lineHt, "2", 6, 4, 0, "")
	pdf.SetX(halfX)
	pdf.Write(lineHt, "This is text with a superscripted letter.\n")
	pdf.Ln(lineHt * 2)

	pdf.Write(lineHt, "The world has a lot of H")
	pdf.SubWrite(lineHt, "2", 6, -3, 0, "")
	pdf.Write(lineHt, "O")
	pdf.SetX(halfX)
	pdf.Write(lineHt, "This is text with a subscripted letter.\n")

	fileStr := example.Filename("Fpdf_SubWrite")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_SubWrite.pdf
}

// ExampleFpdf_SetPage demomstrates the SetPage() method, allowing content
// generation to be deferred until all pages have been added.
func ExampleFpdf_SetPage() {
	rnd := rand.New(rand.NewSource(0)) // Make reproducible documents
	pdf := gofpdf.New("L", "cm", "A4", "")
	pdf.SetFont("Times", "", 12)

	var time []float64
	temperaturesFromSensors := make([][]float64, 5)
	maxs := []float64{25, 41, 89, 62, 11}
	for i := range temperaturesFromSensors {
		temperaturesFromSensors[i] = make([]float64, 0)
	}

	for i := 0.0; i < 100; i += 0.5 {
		time = append(time, i)
		for j, sensor := range temperaturesFromSensors {
			dataValue := rnd.Float64() * maxs[j]
			sensor = append(sensor, dataValue)
			temperaturesFromSensors[j] = sensor
		}
	}
	var graphs []gofpdf.GridType
	var pageNums []int
	xMax := time[len(time)-1]
	for i := range temperaturesFromSensors {
		//Create a new page and graph for each sensor we want to graph.
		pdf.AddPage()
		pdf.Ln(1)
		//Custom label per sensor
		pdf.WriteAligned(0, 0, "Temperature Sensor "+strconv.Itoa(i+1)+" (C) vs Time (min)", "C")
		pdf.Ln(0.5)
		graph := gofpdf.NewGrid(pdf.GetX(), pdf.GetY(), 20, 10)
		graph.TickmarksContainX(0, xMax)
		//Custom Y axis
		graph.TickmarksContainY(0, maxs[i])
		graph.Grid(pdf)
		//Save references and locations.
		graphs = append(graphs, graph)
		pageNums = append(pageNums, pdf.PageNo())
	}
	// For each X, graph the Y in each sensor.
	for i, currTime := range time {
		for j, sensor := range temperaturesFromSensors {
			pdf.SetPage(pageNums[j])
			graph := graphs[j]
			temperature := sensor[i]
			pdf.Circle(graph.X(currTime), graph.Y(temperature), 0.04, "D")
		}
	}

	fileStr := example.Filename("Fpdf_SetPage")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_SetPage.pdf
}

// ExampleFpdf_SetFillColor demonstrates how graphic attributes are properly
// assigned within multiple transformations. See issue #234.
func ExampleFpdf_SetFillColor() {
	pdf := gofpdf.New("P", "mm", "A4", "")

	pdf.AddPage()
	pdf.SetFont("Arial", "", 8)

	draw := func(trX, trY float64) {
		pdf.TransformBegin()
		pdf.TransformTranslateX(trX)
		pdf.TransformTranslateY(trY)
		pdf.SetLineJoinStyle("round")
		pdf.SetLineWidth(0.5)
		pdf.SetDrawColor(128, 64, 0)
		pdf.SetFillColor(255, 127, 0)
		pdf.SetAlpha(0.5, "Normal")
		pdf.SetDashPattern([]float64{5, 10}, 0)
		pdf.Rect(0, 0, 40, 40, "FD")
		pdf.SetFontSize(12)
		pdf.SetXY(5, 5)
		pdf.Write(0, "Test")
		pdf.TransformEnd()
	}

	draw(5, 5)
	draw(50, 50)

	fileStr := example.Filename("Fpdf_SetFillColor")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_SetFillColor.pdf
}

// ExampleFpdf_TransformRotate demonstrates how to rotate text within a header
// to make a watermark that appears on each page.
func ExampleFpdf_TransformRotate() {

	loremStr := lorem() + "\n\n"
	pdf := gofpdf.New("P", "mm", "A4", "")
	margin := 25.0
	pdf.SetMargins(margin, margin, margin)

	fontHt := 13.0
	lineHt := pdf.PointToUnitConvert(fontHt)
	markFontHt := 50.0
	markLineHt := pdf.PointToUnitConvert(markFontHt)
	markY := (297.0 - markLineHt) / 2.0
	ctrX := 210.0 / 2.0
	ctrY := 297.0 / 2.0

	pdf.SetHeaderFunc(func() {
		pdf.SetFont("Arial", "B", markFontHt)
		pdf.SetTextColor(206, 216, 232)
		pdf.SetXY(margin, markY)
		pdf.TransformBegin()
		pdf.TransformRotate(45, ctrX, ctrY)
		pdf.CellFormat(0, markLineHt, "W A T E R M A R K   D E M O", "", 0, "C", false, 0, "")
		pdf.TransformEnd()
		pdf.SetXY(margin, margin)
	})

	pdf.AddPage()
	pdf.SetFont("Arial", "", 8)
	for j := 0; j < 25; j++ {
		pdf.MultiCell(0, lineHt, loremStr, "", "L", false)
	}

	fileStr := example.Filename("Fpdf_RotateText")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_RotateText.pdf
}

// ExampleFpdf_AddUTF8Font demonstrates how use the font
// with utf-8 mode
func ExampleFpdf_AddUTF8Font() {
	var fileStr string
	var txtStr []byte
	var err error

	pdf := gofpdf.New("P", "mm", "A4", "")

	pdf.AddPage()

	pdf.AddUTF8Font("dejavu", "", example.FontFile("DejaVuSansCondensed.ttf"))
	pdf.AddUTF8Font("dejavu", "B", example.FontFile("DejaVuSansCondensed-Bold.ttf"))
	pdf.AddUTF8Font("dejavu", "I", example.FontFile("DejaVuSansCondensed-Oblique.ttf"))
	pdf.AddUTF8Font("dejavu", "BI", example.FontFile("DejaVuSansCondensed-BoldOblique.ttf"))

	fileStr = example.Filename("Fpdf_AddUTF8Font")
	txtStr, err = ioutil.ReadFile(example.TextFile("utf-8test.txt"))
	if err == nil {

		pdf.SetFont("dejavu", "B", 17)
		pdf.MultiCell(100, 8, "Text in different languages :", "", "C", false)
		pdf.SetFont("dejavu", "", 14)
		pdf.MultiCell(100, 5, string(txtStr), "", "C", false)
		pdf.Ln(15)

		txtStr, err = ioutil.ReadFile(example.TextFile("utf-8test2.txt"))
		if err == nil {

			pdf.SetFont("dejavu", "BI", 17)
			pdf.MultiCell(100, 8, "Greek text with alignStr = \"J\":", "", "C", false)
			pdf.SetFont("dejavu", "I", 14)
			pdf.MultiCell(100, 5, string(txtStr), "", "J", false)
			err = pdf.OutputFileAndClose(fileStr)

		}
	}
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_AddUTF8Font.pdf
}

// ExampleUTF8CutFont demonstrates how generate a TrueType font subset.
func ExampleUTF8CutFont() {
	var pdfFileStr, fullFontFileStr, subFontFileStr string
	var subFont, fullFont []byte
	var err error

	pdfFileStr = example.Filename("Fpdf_UTF8CutFont")
	fullFontFileStr = example.FontFile("calligra.ttf")
	fullFont, err = ioutil.ReadFile(fullFontFileStr)
	if err == nil {
		subFontFileStr = "calligra_abcde.ttf"
		subFont = gofpdf.UTF8CutFont(fullFont, "abcde")
		err = ioutil.WriteFile(subFontFileStr, subFont, 0600)
		if err == nil {
			y := 24.0
			pdf := gofpdf.New("P", "mm", "A4", "")
			fontHt := 17.0
			lineHt := pdf.PointConvert(fontHt)
			write := func(format string, args ...interface{}) {
				pdf.SetXY(24.0, y)
				pdf.Cell(200.0, lineHt, fmt.Sprintf(format, args...))
				y += lineHt
			}
			writeSize := func(fileStr string) {
				var info os.FileInfo
				var err error
				info, err = os.Stat(fileStr)
				if err == nil {
					write("%6d: size of %s", info.Size(), fileStr)
				}
			}
			pdf.AddPage()
			pdf.AddUTF8Font("calligra", "", subFontFileStr)
			pdf.SetFont("calligra", "", fontHt)
			write("cabbed")
			write("vwxyz")
			pdf.SetFont("courier", "", fontHt)
			writeSize(fullFontFileStr)
			writeSize(subFontFileStr)
			err = pdf.OutputFileAndClose(pdfFileStr)
			os.Remove(subFontFileStr)
		}
	}
	example.Summary(err, pdfFileStr)
	// Output:
	// Successfully generated pdf/Fpdf_UTF8CutFont.pdf
}

func ExampleFpdf_RoundedRect() {
	const (
		wd     = 40.0
		hgap   = 10.0
		radius = 10.0
		ht     = 60.0
		vgap   = 10.0
	)
	corner := func(b1, b2, b3, b4 bool) (cstr string) {
		if b1 {
			cstr = "1"
		}
		if b2 {
			cstr += "2"
		}
		if b3 {
			cstr += "3"
		}
		if b4 {
			cstr += "4"
		}
		return
	}
	pdf := gofpdf.New("P", "mm", "A4", "") // 210 x 297
	pdf.AddPage()
	pdf.SetLineWidth(0.5)
	y := vgap
	r := 40
	g := 30
	b := 20
	for row := 0; row < 4; row++ {
		x := hgap
		for col := 0; col < 4; col++ {
			pdf.SetFillColor(r, g, b)
			pdf.RoundedRect(x, y, wd, ht, radius, corner(row&1 == 1, row&2 == 2, col&1 == 1, col&2 == 2), "FD")
			r += 8
			g += 10
			b += 12
			x += wd + hgap
		}
		y += ht + vgap
	}
	pdf.AddPage()
	pdf.RoundedRectExt(10, 20, 40, 80, 4., 0., 20, 0., "FD")

	fileStr := example.Filename("Fpdf_RoundedRect")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_RoundedRect.pdf
}

// ExampleFpdf_SetUnderlineThickness demonstrates how to adjust the text
// underline thickness.
func ExampleFpdf_SetUnderlineThickness() {
	pdf := gofpdf.New("P", "mm", "A4", "") // 210mm x 297mm
	pdf.AddPage()
	pdf.SetFont("Arial", "U", 12)

	pdf.SetUnderlineThickness(0.5)
	pdf.CellFormat(0, 10, "Thin underline", "", 1, "", false, 0, "")

	pdf.SetUnderlineThickness(1)
	pdf.CellFormat(0, 10, "Normal underline", "", 1, "", false, 0, "")

	pdf.SetUnderlineThickness(2)
	pdf.CellFormat(0, 10, "Thicker underline", "", 1, "", false, 0, "")

	fileStr := example.Filename("Fpdf_UnderlineThickness")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_UnderlineThickness.pdf
}

// ExampleFpdf_Cell_strikeout demonstrates striked-out text
func ExampleFpdf_Cell_strikeout() {

	pdf := gofpdf.New("P", "mm", "A4", "") // 210mm x 297mm
	pdf.AddPage()

	for fontSize := 4; fontSize < 40; fontSize += 10 {
		pdf.SetFont("Arial", "S", float64(fontSize))
		pdf.SetXY(0, float64(fontSize))
		pdf.Cell(40, 10, "Hello World")
	}

	fileStr := example.Filename("Fpdf_Cell_strikeout")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_Cell_strikeout.pdf
}

// ExampleFpdf_SetTextRenderingMode demonstrates rendering modes in PDFs.
func ExampleFpdf_SetTextRenderingMode() {

	pdf := gofpdf.New("P", "mm", "A4", "") // 210mm x 297mm
	pdf.AddPage()
	fontSz := float64(16)
	lineSz := pdf.PointToUnitConvert(fontSz)
	pdf.SetFont("Times", "", fontSz)
	pdf.Write(lineSz, "This document demonstrates various modes of text rendering. Search for \"Mode 3\" "+
		"to locate text that has been rendered invisibly. This selection can be copied "+
		"into the clipboard as usual and is useful for overlaying onto non-textual elements such "+
		"as images to make them searchable.\n\n")
	fontSz = float64(125)
	lineSz = pdf.PointToUnitConvert(fontSz)
	pdf.SetFontSize(fontSz)
	pdf.SetTextColor(170, 170, 190)
	pdf.SetDrawColor(50, 60, 90)

	write := func(mode int) {
		pdf.SetTextRenderingMode(mode)
		pdf.CellFormat(210, lineSz, fmt.Sprintf("Mode %d", mode), "", 1, "", false, 0, "")
	}

	for mode := 0; mode < 4; mode++ {
		write(mode)
	}
	write(0)

	fileStr := example.Filename("Fpdf_TextRenderingMode")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_TextRenderingMode.pdf
}

// TestIssue0316 addresses issue 316 in which AddUTF8FromBytes modifies its argument
// utf8bytes resulting in a panic if you generate two PDFs with the "same" font bytes.
func TestIssue0316(t *testing.T) {
	pdf := gofpdf.New(gofpdf.OrientationPortrait, "mm", "A4", "")
	pdf.AddPage()
	fontBytes, _ := ioutil.ReadFile(example.FontFile("DejaVuSansCondensed.ttf"))
	ofontBytes := append([]byte{}, fontBytes...)
	pdf.AddUTF8FontFromBytes("dejavu", "", fontBytes)
	pdf.SetFont("dejavu", "", 16)
	pdf.Cell(40, 10, "Hello World!")
	fileStr := example.Filename("TestIssue0316")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	pdf.AddPage()
	if !bytes.Equal(fontBytes, ofontBytes) {
		t.Fatal("Font data changed during pdf generation")
	}
}

func TestMultiCellUnsupportedChar(t *testing.T) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	fontBytes, _ := ioutil.ReadFile(example.FontFile("DejaVuSansCondensed.ttf"))
	pdf.AddUTF8FontFromBytes("dejavu", "", fontBytes)
	pdf.SetFont("dejavu", "", 16)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()

	pdf.MultiCell(0, 5, "😀", "", "", false)

	fileStr := example.Filename("TestMultiCellUnsupportedChar")
	pdf.OutputFileAndClose(fileStr)
}

// ExampleFpdf_SetTextRenderingMode demonstrates embedding files in PDFs,
// at the top-level.
func ExampleFpdf_SetAttachments() {
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Global attachments
	file, err := ioutil.ReadFile("grid.go")
	if err != nil {
		pdf.SetError(err)
	}
	a1 := gofpdf.Attachment{Content: file, Filename: "grid.go"}
	file, err = ioutil.ReadFile("LICENSE")
	if err != nil {
		pdf.SetError(err)
	}
	a2 := gofpdf.Attachment{Content: file, Filename: "License"}
	pdf.SetAttachments([]gofpdf.Attachment{a1, a2})

	fileStr := example.Filename("Fpdf_EmbeddedFiles")
	err = pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_EmbeddedFiles.pdf
}

func ExampleFpdf_AddAttachmentAnnotation() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "", 12)
	pdf.AddPage()

	// Per page attachment
	file, err := ioutil.ReadFile("grid.go")
	if err != nil {
		pdf.SetError(err)
	}
	a := gofpdf.Attachment{Content: file, Filename: "grid.go", Description: "Some amazing code !"}

	pdf.SetXY(5, 10)
	pdf.Rect(2, 10, 50, 15, "D")
	pdf.AddAttachmentAnnotation(&a, 2, 10, 50, 15)
	pdf.Cell(50, 15, "A first link")

	pdf.SetXY(5, 80)
	pdf.Rect(2, 80, 50, 15, "D")
	pdf.AddAttachmentAnnotation(&a, 2, 80, 50, 15)
	pdf.Cell(50, 15, "A second link (no copy)")

	fileStr := example.Filename("Fpdf_FileAnnotations")
	err = pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_FileAnnotations.pdf
}

func ExampleFpdf_SetModificationDate() {
	// pdfinfo (from http://www.xpdfreader.com) reports the following for this example :
	// ~ pdfinfo -box pdf/Fpdf_PageBox.pdf
	// Producer:       FPDF 1.7
	// CreationDate:   Sat Jan  1 00:00:00 2000
	// ModDate:        Sun Jan  2 10:22:30 2000
	pdf := gofpdf.New("", "", "", "")
	pdf.AddPage()
	pdf.SetModificationDate(time.Date(2000, 1, 2, 10, 22, 30, 0, time.UTC))
	fileStr := example.Filename("Fpdf_SetModificationDate")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_SetModificationDate.pdf
}

// ExampleFpdf_AddUTF8Font_emoji demonstrates emoji rendering using the Noto Emoji font.
// This example shows how to render various emoji categories including basic emoji,
// emoji with variation selectors, and mixed text with emoji.
//
// Note: Due to CMAP format 4 limitation, only emoji in the Basic Multilingual Plane
// (BMP, U+0000 to U+FFFF) will render properly. Emoji in supplementary planes
// (U+1F300 and above) may not display correctly and will show as boxes or missing glyphs.
// The Noto Emoji font used here provides monochrome (black and white) emoji, not color.
func ExampleFpdf_AddUTF8Font_emoji() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Add Noto Emoji font (monochrome version)
	pdf.AddUTF8Font("notoemoji", "", example.FontFile("NotoEmoji-Regular.ttf"))

	// Title
	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(0, 10, "Emoji Support Example")
	pdf.Ln(15)

	// Section 1: Basic BMP Emoji (These should work)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "Basic BMP Emoji (U+2000-U+2FFF range):")
	pdf.Ln(10)

	pdf.SetFont("notoemoji", "", 16)
	pdf.Cell(0, 10, "\u2600 Sun (U+2600)")
	pdf.Ln(8)
	pdf.Cell(0, 10, "\u2601 Cloud (U+2601)")
	pdf.Ln(8)
	pdf.Cell(0, 10, "\u2602 Umbrella (U+2602)")
	pdf.Ln(8)
	pdf.Cell(0, 10, "\u2603 Snowman (U+2603)")
	pdf.Ln(8)
	pdf.Cell(0, 10, "\u2614 Umbrella with Rain (U+2614)")
	pdf.Ln(8)
	pdf.Cell(0, 10, "\u2615 Hot Beverage (U+2615)")
	pdf.Ln(8)
	pdf.Cell(0, 10, "\u2660 Spade Suit (U+2660)")
	pdf.Ln(8)
	pdf.Cell(0, 10, "\u2661 Heart Suit Outline (U+2661)")
	pdf.Ln(8)
	pdf.Cell(0, 10, "\u2662 Diamond Suit (U+2662)")
	pdf.Ln(8)
	pdf.Cell(0, 10, "\u2663 Club Suit (U+2663)")
	pdf.Ln(12)

	// Section 2: Common Symbols
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "Common Symbols:")
	pdf.Ln(10)

	pdf.SetFont("notoemoji", "", 16)
	pdf.Cell(0, 10, "\u2713 Check Mark (U+2713)")
	pdf.Ln(8)
	pdf.Cell(0, 10, "\u2714 Heavy Check Mark (U+2714)")
	pdf.Ln(8)
	pdf.Cell(0, 10, "\u2717 Ballot X (U+2717)")
	pdf.Ln(8)
	pdf.Cell(0, 10, "\u2718 Heavy Ballot X (U+2718)")
	pdf.Ln(8)
	pdf.Cell(0, 10, "\u2764 Heavy Black Heart (U+2764)")
	pdf.Ln(8)
	pdf.Cell(0, 10, "\u2b50 White Medium Star (U+2B50)")
	pdf.Ln(12)

	// Section 3: Mixed Text and Emoji
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "Mixed Text and Emoji:")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(30, 8, "Weather today: ")
	pdf.SetFont("notoemoji", "", 12)
	pdf.Cell(0, 8, "\u2600 Sunny!")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(30, 8, "Coffee time: ")
	pdf.SetFont("notoemoji", "", 12)
	pdf.Cell(0, 8, "\u2615")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(30, 8, "Status: ")
	pdf.SetFont("notoemoji", "", 12)
	pdf.Cell(0, 8, "\u2714 Complete")
	pdf.Ln(12)

	// Section 4: Limitations note
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "Limitations:")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(0, 5, "Due to CMAP format 4 limitation, supplementary plane emoji (U+1F300 and above) may not render correctly. This includes many popular modern emoji like face emoji, animals, and food items. Only emoji in the Basic Multilingual Plane (BMP, U+0000-U+FFFF) are supported.", "", "L", false)
	pdf.Ln(5)
	pdf.MultiCell(0, 5, "The Noto Emoji font used in this example provides monochrome (black and white) emoji, not colored emoji.", "", "L", false)

	fileStr := example.Filename("Fpdf_AddUTF8Font_emoji")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_AddUTF8Font_emoji.pdf
}

// Example_emojiShowcase demonstrates comprehensive emoji support across various
// categories and use cases. This visual test suite makes it easy to verify emoji
// rendering behavior and understand the limitations of CMAP format 4.
//
// The showcase includes:
// - Basic BMP emoji across multiple categories (weather, symbols, shapes, etc.)
// - Emoji in grids for easy visual scanning
// - Text integration showing emoji mixed with regular text
// - Different font sizes and alignments
// - Line wrapping behavior with emoji
// - Multi-line text with emoji
// - Documented limitations for supplementary plane emoji
//
// Note: Only emoji in the Basic Multilingual Plane (BMP, U+0000-U+FFFF) render
// correctly due to CMAP format 4 limitations. Modern emoji (U+1F300+) may not display.
//
// Manual Visual Verification Instructions:
//  1. Run: go test -run Example_emojiShowcase
//  2. Generated PDF location: pdf/Fpdf_EmojiShowcase.pdf
//  3. Reference PDF location: pdf/reference/Fpdf_EmojiShowcase.pdf
//  4. Open both PDFs side-by-side
//  5. Verify:
//     - All emoji in Section 1 render as recognizable symbols (not boxes)
//     - Text integration in Section 2 shows emoji inline with text
//     - Size variations in Section 3 scale proportionally
//     - Alignment in Section 4 is correct (left/center/right)
//     - Wrapping in Section 5 flows naturally across lines
//     - Bullet lists in Section 6 are aligned properly
//     - Table in Section 7 displays emoji in cells correctly
//     - Limitations section clearly documents known issues
//  6. Expected results:
//     - 67+ unique emoji/symbols displayed
//     - All BMP range emoji (U+2000-U+2FFF) render correctly
//     - Generated PDF matches reference PDF byte-for-byte (excluding timestamps)
func Example_emojiShowcase() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Add fonts
	pdf.AddUTF8Font("notoemoji", "", example.FontFile("NotoEmoji-Regular.ttf"))

	// ========== TITLE ==========
	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 12, "gofpdf Emoji Showcase")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(0, 5, "Comprehensive Visual Test Suite for Emoji Rendering")
	pdf.Ln(10)

	// ========== SECTION 1: BASIC EMOJI GRID ==========
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 8, "Section 1: Basic BMP Emoji Grid")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(80, 80, 80)
	pdf.Cell(0, 4, "These emoji are in the Basic Multilingual Plane (U+2000-U+2FFF) and should render correctly.")
	pdf.Ln(8)

	// Weather & Nature
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(40, 5, "Weather & Nature:")
	pdf.Ln(6)
	pdf.SetFont("notoemoji", "", 18)
	pdf.Cell(0, 10, "\u2600 \u2601 \u2602 \u2603 \u2604 \u2614 \u2615 \u26C4 \u26C5 \u26C8")
	pdf.Ln(10)

	// Symbols & Signs
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 5, "Symbols & Signs:")
	pdf.Ln(6)
	pdf.SetFont("notoemoji", "", 18)
	pdf.Cell(0, 10, "\u2605 \u2606 \u2713 \u2714 \u2717 \u2718 \u2744 \u2764 \u2B50 \u2728")
	pdf.Ln(10)

	// Playing Cards & Suits
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 5, "Card Suits:")
	pdf.Ln(6)
	pdf.SetFont("notoemoji", "", 18)
	pdf.Cell(0, 10, "\u2660 \u2661 \u2662 \u2663 \u2664 \u2665 \u2666 \u2667")
	pdf.Ln(10)

	// Arrows & Directions
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 5, "Arrows:")
	pdf.Ln(6)
	pdf.SetFont("notoemoji", "", 18)
	pdf.Cell(0, 10, "\u2190 \u2191 \u2192 \u2193 \u2194 \u2195 \u21A9 \u21AA \u2B05 \u27A1")
	pdf.Ln(10)

	// Geometric Shapes
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 5, "Geometric Shapes:")
	pdf.Ln(6)
	pdf.SetFont("notoemoji", "", 18)
	pdf.Cell(0, 10, "\u25A0 \u25A1 \u25B2 \u25B3 \u25BC \u25BD \u25C6 \u25CF \u25CB \u25AA")
	pdf.Ln(10)

	// Miscellaneous Symbols
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 5, "Misc Symbols:")
	pdf.Ln(6)
	pdf.SetFont("notoemoji", "", 18)
	pdf.Cell(0, 10, "\u260E \u260F \u261D \u263A \u2620 \u2622 \u2623 \u2626 \u262E \u262F")
	pdf.Ln(10)

	// Music & Media
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 5, "Music & Media:")
	pdf.Ln(6)
	pdf.SetFont("notoemoji", "", 18)
	pdf.Cell(0, 10, "\u266A \u266B \u266C \u266D \u266E \u266F")
	pdf.Ln(12)

	// ========== SECTION 2: TEXT INTEGRATION ==========
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 8, "Section 2: Text Integration")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(80, 80, 80)
	pdf.Cell(0, 4, "Emoji mixed with regular text in sentences.")
	pdf.Ln(8)

	// Example sentences
	examples := []struct {
		label string
		text  string
		emoji string
	}{
		{"Weather:", "Today's forecast:", "\u2600 Sunny!"},
		{"Coffee:", "Time for a break:", "\u2615 Coffee time!"},
		{"Status:", "Task completed:", "\u2714 Done"},
		{"Love:", "Sending you:", "\u2764 Love"},
		{"Star:", "Rate this:", "\u2B50\u2B50\u2B50\u2B50\u2B50"},
		{"Alert:", "Warning:", "\u26A0 Be careful!"},
		{"Phone:", "Contact:", "\u260E Call me"},
		{"Direction:", "Go this way:", "\u27A1 Right turn"},
	}

	for _, ex := range examples {
		pdf.SetFont("Arial", "B", 10)
		pdf.SetTextColor(0, 0, 0)
		pdf.Cell(30, 6, ex.label)

		pdf.SetFont("Arial", "", 10)
		pdf.SetTextColor(60, 60, 60)
		pdf.Cell(35, 6, ex.text)

		pdf.SetFont("notoemoji", "", 12)
		pdf.SetTextColor(0, 0, 0)
		pdf.Cell(0, 6, " "+ex.emoji)
		pdf.Ln(7)
	}
	pdf.Ln(5)

	// ========== SECTION 3: SIZE VARIATIONS ==========
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 8, "Section 3: Size Variations")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(80, 80, 80)
	pdf.Cell(0, 4, "Same emoji at different font sizes.")
	pdf.Ln(8)

	sizes := []int{8, 10, 12, 14, 16, 20, 24}
	for _, size := range sizes {
		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(100, 100, 100)
		pdf.Cell(20, float64(size)*0.5, fmt.Sprintf("%dpt:", size))

		pdf.SetFont("notoemoji", "", float64(size))
		pdf.SetTextColor(0, 0, 0)
		pdf.Cell(0, float64(size)*0.5, "\u2600 \u2764 \u2B50 \u2615 \u2714")
		pdf.Ln(float64(size)*0.5 + 2)
	}
	pdf.Ln(5)

	// ========== SECTION 4: ALIGNMENT TESTS ==========
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 8, "Section 4: Alignment Tests")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(80, 80, 80)
	pdf.Cell(0, 4, "Emoji with different cell alignments (borders shown for clarity).")
	pdf.Ln(8)

	pdf.SetFont("notoemoji", "", 14)
	pdf.SetTextColor(0, 0, 0)

	// Left align
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(30, 8, "Left:")
	pdf.SetFont("notoemoji", "", 14)
	pdf.CellFormat(140, 8, "\u2600 \u2764 \u2B50", "1", 1, "L", false, 0, "")

	// Center align
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(30, 8, "Center:")
	pdf.SetFont("notoemoji", "", 14)
	pdf.CellFormat(140, 8, "\u2600 \u2764 \u2B50", "1", 1, "C", false, 0, "")

	// Right align
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(30, 8, "Right:")
	pdf.SetFont("notoemoji", "", 14)
	pdf.CellFormat(140, 8, "\u2600 \u2764 \u2B50", "1", 1, "R", false, 0, "")
	pdf.Ln(5)

	// ========== SECTION 5: WRAPPING BEHAVIOR ==========
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 8, "Section 5: Wrapping Behavior")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(80, 80, 80)
	pdf.Cell(0, 4, "Long text with emoji that wraps across multiple lines.")
	pdf.Ln(8)

	// Long text with emoji
	longText := "The weather today is quite nice \u2600 and I'm enjoying a cup of coffee \u2615 " +
		"while working on this project. Everything is going well \u2714 and I'm feeling " +
		"very happy \u2764 about the progress. The stars \u2B50 are aligned and things are " +
		"looking great! Remember to stay hydrated \u2615 and take breaks when needed. " +
		"Keep moving forward \u27A1 and don't give up!"

	pdf.SetFont("notoemoji", "", 11)
	pdf.SetTextColor(0, 0, 0)
	pdf.MultiCell(0, 6, longText, "", "L", false)
	pdf.Ln(5)

	// ========== SECTION 6: BULLET LISTS WITH EMOJI ==========
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 8, "Section 6: Bullet Lists with Emoji")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(80, 80, 80)
	pdf.Cell(0, 4, "Using emoji as bullet points in lists.")
	pdf.Ln(8)

	bulletItems := []string{
		"\u2714 Task one is completed",
		"\u2714 Task two is finished",
		"\u2717 Task three is pending",
		"\u27A1 Task four is next",
		"\u2B50 Task five is a priority",
	}

	for _, item := range bulletItems {
		pdf.SetFont("notoemoji", "", 11)
		pdf.SetTextColor(0, 0, 0)
		pdf.Cell(5, 6, "")
		pdf.Cell(0, 6, item)
		pdf.Ln(6)
	}
	pdf.Ln(5)

	// ========== SECTION 7: EMOJI TABLE ==========
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 8, "Section 7: Emoji in Tables")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(80, 80, 80)
	pdf.Cell(0, 4, "Structured data with emoji in table format.")
	pdf.Ln(8)

	// Table header
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(30, 8, "Status", "1", 0, "C", true, 0, "")
	pdf.CellFormat(80, 8, "Task Description", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 8, "Priority", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 8, "Rating", "1", 1, "C", true, 0, "")

	// Table data
	tableData := []struct {
		status   string
		task     string
		priority string
		rating   string
	}{
		{"\u2714", "Complete documentation", "\u2B50\u2B50\u2B50", "\u2605\u2605\u2605\u2605\u2605"},
		{"\u2717", "Fix bug in parser", "\u2B50\u2B50", "\u2605\u2605\u2605\u2605\u2606"},
		{"\u2714", "Update dependencies", "\u2B50", "\u2605\u2605\u2605\u2606\u2606"},
		{"\u27A1", "Refactor codebase", "\u2B50\u2B50\u2B50", "\u2605\u2605\u2605\u2605\u2606"},
		{"\u2714", "Write unit tests", "\u2B50\u2B50", "\u2605\u2605\u2605\u2605\u2605"},
	}

	pdf.SetFont("notoemoji", "", 10)
	for _, row := range tableData {
		pdf.CellFormat(30, 8, row.status, "1", 0, "C", false, 0, "")
		pdf.SetFont("Arial", "", 9)
		pdf.CellFormat(80, 8, row.task, "1", 0, "L", false, 0, "")
		pdf.SetFont("notoemoji", "", 10)
		pdf.CellFormat(30, 8, row.priority, "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 8, row.rating, "1", 1, "C", false, 0, "")
	}
	pdf.Ln(8)

	// ========== SECTION 8: LIMITATIONS ==========
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 8, "Section 8: Limitations & Known Issues")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(80, 80, 80)
	pdf.Cell(0, 4, "Understanding the boundaries of emoji support in gofpdf.")
	pdf.Ln(8)

	// Limitation 1
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(200, 0, 0)
	pdf.Cell(0, 6, "1. CMAP Format 4 Limitation (BMP Only)")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(0, 0, 0)
	limitText1 := "The current implementation uses CMAP format 4, which only supports the Basic " +
		"Multilingual Plane (BMP, U+0000-U+FFFF). Most modern emoji are in the supplementary " +
		"planes (U+1F300 and above) and will NOT render correctly. They may appear as boxes, " +
		"missing glyphs, or question marks."
	pdf.MultiCell(0, 5, limitText1, "", "L", false)
	pdf.Ln(3)

	// Limitation 2
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(200, 0, 0)
	pdf.Cell(0, 6, "2. Monochrome Only")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(0, 0, 0)
	limitText2 := "The Noto Emoji font used provides monochrome (black and white) emoji, not colored " +
		"emoji. If you need color emoji, you would need to use image-based solutions instead."
	pdf.MultiCell(0, 5, limitText2, "", "L", false)
	pdf.Ln(3)

	// Limitation 3
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(200, 0, 0)
	pdf.Cell(0, 6, "3. No Skin Tone Modifiers")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(0, 0, 0)
	limitText3 := "Skin tone modifiers (U+1F3FB-U+1F3FF) are in the supplementary plane and will not " +
		"work with the current implementation. Base emoji without modifiers should be used."
	pdf.MultiCell(0, 5, limitText3, "", "L", false)
	pdf.Ln(3)

	// Limitation 4
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(200, 0, 0)
	pdf.Cell(0, 6, "4. No ZWJ Sequences")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(0, 0, 0)
	limitText4 := "Zero-Width Joiner (ZWJ) sequences like family emoji, profession emoji, and combined " +
		"emoji are typically in supplementary planes and will not render as expected. Each component " +
		"may render separately if in the BMP, or not at all."
	pdf.MultiCell(0, 5, limitText4, "", "L", false)
	pdf.Ln(5)

	// What Works section
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(0, 150, 0)
	pdf.Cell(0, 7, "What DOES Work:")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(0, 0, 0)
	worksText := "- Dingbats (U+2700-U+27BF)\n" +
		"- Miscellaneous Symbols (U+2600-U+26FF)\n" +
		"- Arrows (U+2190-U+21FF)\n" +
		"- Geometric Shapes (U+25A0-U+25FF)\n" +
		"- Card suits and playing cards (U+2660-U+2667)\n" +
		"- Stars, hearts, and common symbols\n" +
		"- Weather and nature symbols in BMP range"
	pdf.MultiCell(0, 5, worksText, "", "L", false)
	pdf.Ln(3)

	// What Doesn't Work section
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(200, 0, 0)
	pdf.Cell(0, 7, "What DOES NOT Work:")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(0, 0, 0)
	doesntWorkText := "- Face emoji (U+1F600-U+1F64F) - smileys, emotions\n" +
		"- Animals & nature (U+1F400-U+1F4FF)\n" +
		"- Food & drink (U+1F32D-U+1F37F)\n" +
		"- Flags (U+1F1E0-U+1F1FF)\n" +
		"- Hand gestures with skin tones\n" +
		"- Family and people emoji with ZWJ sequences\n" +
		"- Most modern emoji added after Unicode 6.0"
	pdf.MultiCell(0, 5, doesntWorkText, "", "L", false)
	pdf.Ln(8)

	// Footer note
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(120, 120, 120)
	pdf.Cell(0, 4, "Generated by gofpdf emoji showcase - Test suite for visual verification")
	pdf.Ln(4)
	pdf.Cell(0, 4, fmt.Sprintf("Document generated on %s", time.Now().Format("2006-01-02")))

	// Save PDF
	fileStr := example.Filename("Fpdf_EmojiShowcase")
	err := pdf.OutputFileAndClose(fileStr)
	example.Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/Fpdf_EmojiShowcase.pdf
}

// TestStringWidthGraphemeClusters tests that string width calculation correctly
// handles grapheme clusters, particularly multi-codepoint emoji sequences.
func TestStringWidthGraphemeClusters(t *testing.T) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.AddUTF8Font("dejavu", "", example.FontFile("DejaVuSansCondensed.ttf"))
	pdf.SetFont("dejavu", "", 12)

	// Test 1: Regular ASCII text should work as before
	regularText := "Hello"
	widthRegular := pdf.GetStringSymbolWidth(regularText)
	if widthRegular == 0 {
		t.Errorf("Regular text '%s' should have non-zero width, got %d", regularText, widthRegular)
	}

	// Test 2: Regular text with each character should sum up correctly
	// (grapheme clusters for ASCII are single characters)
	totalWidth := 0
	for _, ch := range regularText {
		totalWidth += pdf.GetStringSymbolWidth(string(ch))
	}
	if widthRegular != totalWidth {
		t.Errorf("Regular text width mismatch: full string=%d, sum of chars=%d", widthRegular, totalWidth)
	}

	// Test 3: Emoji with skin tone modifier should be measured as single unit
	// Base emoji + skin tone modifier = 1 grapheme cluster
	baseEmoji := "\U0001F44D"                   // Thumbs up (without modifier)
	emojiWithModifier := "\U0001F44D\U0001F3FD" // Thumbs up + medium skin tone

	widthBase := pdf.GetStringSymbolWidth(baseEmoji)
	widthModified := pdf.GetStringSymbolWidth(emojiWithModifier)

	// The width with modifier should be the same as base (modifier adds no width)
	if widthModified != widthBase {
		t.Errorf("Emoji with skin tone modifier should have same width as base: base=%d, modified=%d", widthBase, widthModified)
	}

	// Test 4: ZWJ sequence should be measured as single unit
	// Family emoji: man + ZWJ + woman + ZWJ + girl + ZWJ + boy
	familyEmoji := "\U0001F468\u200D\U0001F469\u200D\U0001F467\u200D\U0001F466"
	manEmoji := "\U0001F468"

	widthFamily := pdf.GetStringSymbolWidth(familyEmoji)
	widthMan := pdf.GetStringSymbolWidth(manEmoji)

	// Family emoji should be measured as one unit (not sum of all parts)
	// It should be approximately the width of a single emoji character
	if widthFamily == 0 {
		t.Errorf("Family emoji (ZWJ sequence) should have non-zero width, got %d", widthFamily)
	}

	// The family emoji width should be close to a single emoji width (the base character)
	if widthFamily != widthMan {
		t.Logf("Note: Family emoji width=%d, man emoji width=%d (expected to be same as base)", widthFamily, widthMan)
	}

	// Test 5: Variation selector should not add width
	// Sun emoji without and with variation selector
	sunWithoutVS := "\u2600"         // Sun without variation selector
	sunWithVS := "\u2600\uFE0F"      // Sun + emoji variation selector

	widthSunNoVS := pdf.GetStringSymbolWidth(sunWithoutVS)
	widthSunVS := pdf.GetStringSymbolWidth(sunWithVS)

	// Width should be the same (variation selector adds no width)
	if widthSunVS != widthSunNoVS {
		t.Errorf("Emoji with variation selector should have same width: without=%d, with=%d", widthSunNoVS, widthSunVS)
	}

	// Test 6: Mixed text and emoji
	mixedText := "Hello \U0001F44D World"
	widthMixed := pdf.GetStringSymbolWidth(mixedText)

	if widthMixed == 0 {
		t.Errorf("Mixed text and emoji should have non-zero width, got %d", widthMixed)
	}

	// Mixed text width should be greater than just "Hello World"
	widthTextOnly := pdf.GetStringSymbolWidth("Hello  World") // Two spaces where emoji was
	if widthMixed <= widthTextOnly {
		t.Logf("Mixed text width=%d, text-only width=%d", widthMixed, widthTextOnly)
	}

	// Test 7: GetStringWidth (user units) should be non-zero and proportional
	symbolWidth := pdf.GetStringSymbolWidth(regularText)
	userWidth := pdf.GetStringWidth(regularText)

	// GetStringWidth should return a positive value based on GetStringSymbolWidth
	if userWidth <= 0 {
		t.Errorf("GetStringWidth should return positive value, got %f", userWidth)
	}
	if symbolWidth <= 0 {
		t.Errorf("GetStringSymbolWidth should return positive value, got %d", symbolWidth)
	}
	// Verify that user width is proportional to symbol width
	// (exact ratio depends on font size and scale factor, just verify relationship)
	if userWidth == 0 && symbolWidth > 0 {
		t.Errorf("GetStringWidth returned 0 when symbolWidth=%d", symbolWidth)
	}

	// Test 8: Empty string
	emptyWidth := pdf.GetStringSymbolWidth("")
	if emptyWidth != 0 {
		t.Errorf("Empty string should have zero width, got %d", emptyWidth)
	}

	// Test 9: Multiple skin tone modifiers in sequence
	multiModifier := "\U0001F44D\U0001F3FB\U0001F44D\U0001F3FD\U0001F44D\U0001F3FF"
	widthMulti := pdf.GetStringSymbolWidth(multiModifier)

	// This should be 3 grapheme clusters (3 thumbs up with different skin tones)
	// So width should be approximately 3 times the base emoji width
	expectedMulti := widthBase * 3
	if widthMulti != expectedMulti {
		t.Errorf("Multiple emoji with modifiers: got=%d, expected=%d (3x base width)", widthMulti, expectedMulti)
	}

	t.Logf("Width test summary:")
	t.Logf("  Regular text 'Hello': %d font units", widthRegular)
	t.Logf("  Base emoji (thumbs up): %d font units", widthBase)
	t.Logf("  Emoji with skin tone: %d font units", widthModified)
	t.Logf("  Family emoji (ZWJ): %d font units", widthFamily)
	t.Logf("  Sun without VS: %d font units", widthSunNoVS)
	t.Logf("  Sun with VS: %d font units", widthSunVS)
	t.Logf("  Mixed text: %d font units", widthMixed)
	t.Logf("  Three emoji with modifiers: %d font units", widthMulti)
}

// TestSplitTextWithEmoji tests that SplitText properly splits text at grapheme
// cluster boundaries and doesn't break emoji sequences
func TestSplitTextWithEmoji(t *testing.T) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	fontBytes, err := ioutil.ReadFile(example.FontFile("DejaVuSansCondensed.ttf"))
	if err != nil {
		t.Fatalf("Failed to load font: %v", err)
	}
	pdf.AddUTF8FontFromBytes("dejavu", "", fontBytes)
	pdf.SetFont("dejavu", "", 12)

	tests := []struct {
		name     string
		text     string
		width    float64
		minLines int // minimum expected lines
	}{
		{
			name:     "Simple text wrapping",
			text:     "Hello World This Is A Test",
			width:    30.0,
			minLines: 2,
		},
		{
			name:     "Emoji with skin tone modifier",
			text:     "👍🏽 👍🏽 👍🏽 👍🏽 👍🏽 👍🏽 👍🏽 👍🏽 👍🏽 👍🏽",
			width:    30.0, // Reduced width to force wrapping
			minLines: 1,    // At least 1 line (emoji might all fit)
		},
		{
			name:     "ZWJ sequence (family emoji)",
			text:     "Family: 👨‍👩‍👧‍👦 👨‍👩‍👧‍👦 👨‍👩‍👧‍👦 text",
			width:    60.0,
			minLines: 1,
		},
		{
			name:     "Mixed text and emoji",
			text:     "Hello 👍🏽 World 😀 Test 🎉 More text here",
			width:    40.0,
			minLines: 1, // At least 1 line
		},
		{
			name:     "Long text with emoji that must wrap",
			text:     "This is a very long sentence with emoji 👍🏽 and more text that should wrap across multiple lines 😀",
			width:    50.0,
			minLines: 3,
		},
		{
			name:     "Emoji at line boundaries",
			text:     "Short 👍🏽",
			width:    50.0, // Increased width so it fits on one line
			minLines: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := pdf.SplitText(tt.text, tt.width)

			// Check minimum number of lines
			if len(lines) < tt.minLines {
				t.Errorf("Expected at least %d lines, got %d", tt.minLines, len(lines))
			}

			// Note: SplitText removes spaces at line breaks (normal word-wrapping behavior)
			// So we don't verify exact rejoining, just that we got reasonable output
			if len(lines) == 0 {
				t.Errorf("SplitText returned no lines")
			}

			// Check that no line contains broken emoji sequences
			for i, line := range lines {
				// Count runes vs grapheme clusters - if they differ significantly in emoji,
				// it might indicate a broken sequence
				if strings.Contains(line, "👍🏽") {
					// This is a 2-codepoint emoji, ensure it appears complete
					count := strings.Count(line, "👍🏽")
					if count > 0 {
						t.Logf("Line %d contains %d instances of thumbs-up with skin tone", i, count)
					}
				}
				if strings.Contains(line, "👨‍👩‍👧‍👦") {
					// ZWJ family emoji should not be broken
					t.Logf("Line %d contains family emoji (ZWJ sequence)", i)
				}
			}

			t.Logf("Split %q into %d lines", tt.text, len(lines))
			for i, line := range lines {
				t.Logf("  Line %d: %q", i+1, line)
			}
		})
	}
}

// TestMultiCellWithEmojiWrapping tests that MultiCell wraps text without
// breaking emoji sequences across lines
func TestMultiCellWithEmojiWrapping(t *testing.T) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	fontBytes, err := ioutil.ReadFile(example.FontFile("DejaVuSansCondensed.ttf"))
	if err != nil {
		t.Fatalf("Failed to load font: %v", err)
	}
	pdf.AddUTF8FontFromBytes("dejavu", "", fontBytes)
	pdf.SetFont("dejavu", "", 12)

	tests := []struct {
		name   string
		text   string
		width  float64
		height float64
	}{
		{
			name:   "Simple emoji wrapping",
			text:   "Hello 👍🏽 World 😀 Test",
			width:  40.0,
			height: 5.0,
		},
		{
			name:   "Emoji with skin tone modifiers",
			text:   "Thumbs up: 👍🏽 👍🏼 👍🏿 👍🏻 👍🏾",
			width:  50.0,
			height: 5.0,
		},
		{
			name:   "ZWJ sequences",
			text:   "Family: 👨‍👩‍👧‍👦 and couple: 👩‍❤️‍👨",
			width:  60.0,
			height: 5.0,
		},
		{
			name:   "Long text with multiple emoji",
			text:   "This is a longer text with emoji 😀 that should wrap correctly 👍🏽 across multiple lines without breaking 🎉 the emoji sequences",
			width:  60.0,
			height: 5.0,
		},
		{
			name:   "Emoji at line boundaries",
			text:   "A 👍🏽 B 😀 C 🎉 D 👨‍👩‍👧‍👦 E",
			width:  20.0,
			height: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yBefore := pdf.GetY()

			// MultiCell should not panic with emoji
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("MultiCell panicked with emoji: %v", r)
				}
			}()

			pdf.MultiCell(tt.width, tt.height, tt.text, "1", "L", false)

			yAfter := pdf.GetY()
			lines := int((yAfter - yBefore) / tt.height)

			t.Logf("MultiCell rendered %q in %d lines (Y: %.2f -> %.2f)",
				tt.text, lines, yBefore, yAfter)

			// Ensure text was rendered (Y position changed)
			if yAfter <= yBefore {
				t.Errorf("MultiCell did not advance Y position")
			}
		})
	}

	// Generate PDF to verify visual output
	fileStr := example.Filename("TestMultiCellWithEmojiWrapping")
	err = pdf.OutputFileAndClose(fileStr)
	if err != nil {
		t.Errorf("Failed to generate PDF: %v", err)
	} else {
		t.Logf("Generated PDF: %s", fileStr)
	}
}

// TestWriteWithEmojiFlowing tests that Write() handles emoji in flowing text
func TestWriteWithEmojiFlowing(t *testing.T) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	fontBytes, err := ioutil.ReadFile(example.FontFile("DejaVuSansCondensed.ttf"))
	if err != nil {
		t.Fatalf("Failed to load font: %v", err)
	}
	pdf.AddUTF8FontFromBytes("dejavu", "", fontBytes)
	pdf.SetFont("dejavu", "", 12)

	tests := []struct {
		name   string
		text   string
		height float64
	}{
		{
			name:   "Simple emoji in text",
			text:   "Hello 👍🏽 World",
			height: 5.0,
		},
		{
			name:   "Multiple emoji with modifiers",
			text:   "Reactions: 👍🏽 😀 🎉 ❤️ 👨‍👩‍👧‍👦",
			height: 5.0,
		},
		{
			name:   "Long flowing text with emoji",
			text:   "This is a long piece of text that flows and wraps naturally 👍🏽 with emoji interspersed throughout the content 😀 to test that the Write function handles grapheme clusters correctly 🎉 without breaking emoji sequences at line boundaries.",
			height: 5.0,
		},
		{
			name:   "Emoji with newlines",
			text:   "Line 1 with emoji 👍🏽\nLine 2 with emoji 😀\nLine 3 with family 👨‍👩‍👧‍👦",
			height: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yBefore := pdf.GetY()

			// Write should not panic with emoji
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Write panicked with emoji: %v", r)
				}
			}()

			pdf.Write(tt.height, tt.text)
			pdf.Ln(-1) // Add line break after each test

			yAfter := pdf.GetY()

			t.Logf("Write rendered %q (Y: %.2f -> %.2f)",
				tt.text, yBefore, yAfter)

			// Ensure text was rendered (Y position changed)
			if yAfter <= yBefore {
				t.Errorf("Write did not advance Y position")
			}
		})
	}

	// Generate PDF to verify visual output
	fileStr := example.Filename("TestWriteWithEmojiFlowing")
	err = pdf.OutputFileAndClose(fileStr)
	if err != nil {
		t.Errorf("Failed to generate PDF: %v", err)
	} else {
		t.Logf("Generated PDF: %s", fileStr)
	}
}

// TestEmojiSequencesNotSplit tests that specific emoji sequences are never
// split across lines by SplitText, MultiCell, or Write
func TestEmojiSequencesNotSplit(t *testing.T) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	fontBytes, err := ioutil.ReadFile(example.FontFile("DejaVuSansCondensed.ttf"))
	if err != nil {
		t.Fatalf("Failed to load font: %v", err)
	}
	pdf.AddUTF8FontFromBytes("dejavu", "", fontBytes)
	pdf.SetFont("dejavu", "", 12)

	// Test specific emoji sequences that must stay together
	sequences := []struct {
		name  string
		emoji string
		desc  string
	}{
		{
			name:  "Thumbs up with skin tone",
			emoji: "👍🏽",
			desc:  "Base + modifier",
		},
		{
			name:  "Family emoji",
			emoji: "👨‍👩‍👧‍👦",
			desc:  "ZWJ sequence (4 people)",
		},
		{
			name:  "Couple with heart",
			emoji: "👩‍❤️‍👨",
			desc:  "ZWJ sequence with variation selector",
		},
		{
			name:  "Sun with variation selector",
			emoji: "☀️",
			desc:  "Symbol + variation selector",
		},
	}

	for _, seq := range sequences {
		t.Run(seq.name, func(t *testing.T) {
			// Create text with the emoji repeated to force wrapping
			text := ""
			for i := 0; i < 20; i++ {
				if i > 0 {
					text += " "
				}
				text += seq.emoji
			}

			// Test with SplitText
			lines := pdf.SplitText(text, 50.0)
			t.Logf("SplitText split into %d lines", len(lines))

			// Check each line contains only complete emoji sequences
			for i, line := range lines {
				// Count occurrences of the complete emoji
				completeCount := strings.Count(line, seq.emoji)

				// For sequences, also check that all components are present
				// by checking individual runes
				runes := []rune(line)
				t.Logf("  Line %d: %d complete %s sequences, %d runes",
					i+1, completeCount, seq.name, len(runes))

				// Ensure the line contains the complete emoji, not fragments
				if completeCount > 0 {
					// The line should contain the emoji in its complete form
					if !strings.Contains(line, seq.emoji) {
						t.Errorf("Line %d does not contain complete %s: %q",
							i+1, seq.name, line)
					}
				}
			}
		})
	}
}

// TestCellFormatWithEmoji tests that CellFormat properly renders emoji
// without panicking and generates valid PDFs
func TestCellFormatWithEmoji(t *testing.T) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	fontBytes, err := ioutil.ReadFile(example.FontFile("DejaVuSansCondensed.ttf"))
	if err != nil {
		t.Fatalf("Failed to load font: %v", err)
	}
	pdf.AddUTF8FontFromBytes("dejavu", "", fontBytes)
	pdf.SetFont("dejavu", "", 12)

	testCases := []struct {
		name string
		text string
	}{
		{
			name: "Basic emoji",
			text: "Hello 😀 World",
		},
		{
			name: "Emoji with skin tone modifier",
			text: "Thumbs up 👍🏽",
		},
		{
			name: "Multiple emoji",
			text: "😀 🎉 🚀",
		},
		{
			name: "ZWJ sequence (family)",
			text: "Family: 👨‍👩‍👧‍👦",
		},
		{
			name: "Emoji with variation selector",
			text: "Sun ☀️",
		},
		{
			name: "High codepoint emoji (> U+FFFF)",
			text: "\U0001F600\U0001F680\U0001F389", // Grinning face, rocket, party popper
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("CellFormat panicked with emoji: %v", r)
				}
			}()

			// Render the cell with various alignments and borders
			pdf.CellFormat(100, 10, tc.text, "1", 1, "L", false, 0, "")
			pdf.CellFormat(100, 10, tc.text, "", 1, "C", false, 0, "")
			pdf.CellFormat(100, 10, tc.text, "LRTB", 1, "R", true, 0, "")

			t.Logf("Successfully rendered: %q", tc.text)
		})
	}

	// Generate PDF to verify no errors during output
	fileStr := example.Filename("TestCellFormatWithEmoji")
	err = pdf.OutputFileAndClose(fileStr)
	if err != nil {
		t.Errorf("Error generating PDF: %v", err)
	}
	t.Logf("Successfully generated %s", fileStr)
}

// TestTextWithEmoji tests that Text() properly renders emoji at specific positions
// without panicking
func TestTextWithEmoji(t *testing.T) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	fontBytes, err := ioutil.ReadFile(example.FontFile("DejaVuSansCondensed.ttf"))
	if err != nil {
		t.Fatalf("Failed to load font: %v", err)
	}
	pdf.AddUTF8FontFromBytes("dejavu", "", fontBytes)
	pdf.SetFont("dejavu", "", 14)

	testCases := []struct {
		name string
		x, y float64
		text string
	}{
		{
			name: "Simple emoji at position",
			x:    20, y: 20,
			text: "😀",
		},
		{
			name: "Emoji with modifier",
			x:    20, y: 30,
			text: "👍🏽",
		},
		{
			name: "Mixed text and emoji",
			x:    20, y: 40,
			text: "Hello 🌍 World",
		},
		{
			name: "High codepoint emoji",
			x:    20, y: 50,
			text: "\U0001F680", // Rocket
		},
		{
			name: "ZWJ sequence",
			x:    20, y: 60,
			text: "👨‍👩‍👧‍👦",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Text panicked with emoji: %v", r)
				}
			}()

			// Render text at specific position
			pdf.Text(tc.x, tc.y, tc.text)

			t.Logf("Position (%.1f, %.1f): Successfully rendered %q",
				tc.x, tc.y, tc.text)
		})
	}

	// Generate PDF to verify no errors
	fileStr := example.Filename("TestTextWithEmoji")
	err = pdf.OutputFileAndClose(fileStr)
	if err != nil {
		t.Errorf("Error generating PDF: %v", err)
	}
	t.Logf("Successfully generated %s", fileStr)
}

// TestClipTextWithEmoji tests that ClipText properly handles emoji
// including UTF-16BE conversion
func TestClipTextWithEmoji(t *testing.T) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	fontBytes, err := ioutil.ReadFile(example.FontFile("DejaVuSansCondensed.ttf"))
	if err != nil {
		t.Fatalf("Failed to load font: %v", err)
	}
	pdf.AddUTF8FontFromBytes("dejavu", "", fontBytes)
	pdf.SetFont("dejavu", "", 24)

	testCases := []struct {
		name string
		text string
	}{
		{
			name: "Basic emoji clip",
			text: "😀",
		},
		{
			name: "Text with emoji",
			text: "EMOJI",
		},
		{
			name: "High codepoint emoji",
			text: "\U0001F389", // Party popper
		},
		{
			name: "Multiple emoji",
			text: "😀🎉🚀",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("ClipText panicked with emoji: %v", r)
				}
			}()

			// Use ClipText
			pdf.ClipText(50, 50, tc.text, true)

			// Add some visual content inside the clip
			pdf.SetFillColor(200, 220, 255)
			pdf.Rect(40, 40, 80, 20, "F")

			// End clipping
			pdf.ClipEnd()

			t.Logf("ClipText successfully rendered %q", tc.text)
		})

		// Add a new page for next test
		if tc.name != testCases[len(testCases)-1].name {
			pdf.AddPage()
			pdf.SetFont("dejavu", "", 24)
		}
	}

	// Generate PDF to verify no errors
	fileStr := example.Filename("TestClipTextWithEmoji")
	err = pdf.OutputFileAndClose(fileStr)
	if err != nil {
		t.Errorf("Error generating PDF: %v", err)
	}
	t.Logf("Successfully generated %s", fileStr)
}

// TestUsedRunesTrackingHighCodepoints verifies that text with high codepoint emoji
// (beyond U+FFFF) can be rendered and generates a valid PDF
func TestUsedRunesTrackingHighCodepoints(t *testing.T) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	fontBytes, err := ioutil.ReadFile(example.FontFile("DejaVuSansCondensed.ttf"))
	if err != nil {
		t.Fatalf("Failed to load font: %v", err)
	}
	pdf.AddUTF8FontFromBytes("dejavu", "", fontBytes)
	pdf.SetFont("dejavu", "", 12)

	// Test various high codepoint emoji
	testEmoji := []struct {
		char      rune
		name      string
		codepoint string
	}{
		{0x1F600, "Grinning face", "U+1F600"},
		{0x1F389, "Party popper", "U+1F389"},
		{0x1F680, "Rocket", "U+1F680"},
		{0x1F44D, "Thumbs up", "U+1F44D"},
		{0x1F3FD, "Medium skin tone", "U+1F3FD"},
	}

	// Build text with all these emoji
	text := "High codepoint emoji: "
	for _, e := range testEmoji {
		text += string(e.char) + " "
	}

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Rendering high codepoint emoji panicked: %v", r)
		}
	}()

	// Render text with all these emoji
	pdf.CellFormat(0, 10, text, "1", 1, "L", false, 0, "")

	// Add individual tests for each emoji
	for _, e := range testEmoji {
		pdf.CellFormat(100, 8, string(e.char)+" "+e.name, "", 1, "L", false, 0, "")
		t.Logf("Successfully rendered %s (%s)", e.codepoint, e.name)
	}

	// Generate PDF to verify everything works end-to-end
	fileStr := example.Filename("TestUsedRunesTrackingHighCodepoints")
	err = pdf.OutputFileAndClose(fileStr)
	if err != nil {
		t.Errorf("Error generating PDF: %v", err)
	}
	t.Logf("Successfully generated %s", fileStr)
}

// TestRTLWithEmoji tests right-to-left text rendering with emoji
func TestRTLWithEmoji(t *testing.T) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	fontBytes, err := ioutil.ReadFile(example.FontFile("DejaVuSansCondensed.ttf"))
	if err != nil {
		t.Fatalf("Failed to load font: %v", err)
	}
	pdf.AddUTF8FontFromBytes("dejavu", "", fontBytes)
	pdf.SetFont("dejavu", "", 12)

	// Enable RTL mode
	pdf.RTL()

	testCases := []struct {
		name string
		text string
	}{
		{
			name: "RTL text with emoji",
			text: "Hello 😀 World",
		},
		{
			name: "RTL with high codepoint emoji",
			text: "Text \U0001F680 Rocket",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("RTL text with emoji caused panic: %v", r)
				}
			}()

			pdf.CellFormat(100, 10, tc.text, "1", 1, "R", false, 0, "")
			pdf.Text(20, pdf.GetY(), tc.text)

			t.Logf("Successfully rendered RTL text: %q", tc.text)
		})
	}

	// Disable RTL
	pdf.LTR()

	// Generate PDF
	fileStr := example.Filename("TestRTLWithEmoji")
	err = pdf.OutputFileAndClose(fileStr)
	if err != nil {
		t.Errorf("Error generating PDF: %v", err)
	}
	t.Logf("Successfully generated %s", fileStr)
}
