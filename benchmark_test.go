package gofpdf

import (
	"testing"
)

// BenchmarkGetStringWidth benchmarks the GetStringWidth function with various text types
func BenchmarkGetStringWidth(b *testing.B) {
	pdf := New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.AddUTF8Font("DejaVuSans", "", "font/DejaVuSansCondensed.ttf")
	pdf.SetFont("DejaVuSans", "", 12)
	text := "Hello World with some emoji 😀🎉"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pdf.GetStringWidth(text)
	}
}

// BenchmarkGetStringWidthASCII benchmarks GetStringWidth with ASCII-only text
func BenchmarkGetStringWidthASCII(b *testing.B) {
	pdf := New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)
	text := "Hello World this is a test of ASCII text performance"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pdf.GetStringWidth(text)
	}
}

// BenchmarkGetStringWidthUTF8 benchmarks GetStringWidth with UTF-8 text
func BenchmarkGetStringWidthUTF8(b *testing.B) {
	pdf := New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.AddUTF8Font("DejaVuSans", "", "font/DejaVuSansCondensed.ttf")
	pdf.SetFont("DejaVuSans", "", 12)
	text := "Hello World 你好 مرحبا こんにちは"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pdf.GetStringWidth(text)
	}
}

// BenchmarkGetStringWidthEmoji benchmarks GetStringWidth with emoji
func BenchmarkGetStringWidthEmoji(b *testing.B) {
	pdf := New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.AddUTF8Font("DejaVuSans", "", "font/DejaVuSansCondensed.ttf")
	pdf.SetFont("DejaVuSans", "", 12)
	text := "😀 🎉 🚀 👍🏽 👨‍👩‍👧‍👦 ☀️"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pdf.GetStringWidth(text)
	}
}

// BenchmarkCellFormat benchmarks the CellFormat function
func BenchmarkCellFormat(b *testing.B) {
	pdf := New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.AddUTF8Font("DejaVuSans", "", "font/DejaVuSansCondensed.ttf")
	pdf.SetFont("DejaVuSans", "", 12)
	text := "Hello World with some emoji 😀🎉"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pdf.CellFormat(40, 10, text, "1", 0, "L", false, 0, "")
	}
}

// BenchmarkCellFormatASCII benchmarks CellFormat with ASCII-only text
func BenchmarkCellFormatASCII(b *testing.B) {
	pdf := New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)
	text := "Hello World this is a test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pdf.CellFormat(40, 10, text, "1", 0, "L", false, 0, "")
	}
}

// BenchmarkMultiCell benchmarks the MultiCell function
func BenchmarkMultiCell(b *testing.B) {
	pdf := New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.AddUTF8Font("DejaVuSans", "", "font/DejaVuSansCondensed.ttf")
	pdf.SetFont("DejaVuSans", "", 12)
	text := "This is a longer text with emoji 😀 that should wrap correctly 👍🏽 across multiple lines without breaking 🎉 the emoji sequences"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pdf.SetXY(10, 10) // Reset position for each iteration
		pdf.MultiCell(60, 5, text, "", "L", false)
	}
}

// BenchmarkMultiCellASCII benchmarks MultiCell with ASCII-only text
func BenchmarkMultiCellASCII(b *testing.B) {
	pdf := New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)
	text := "This is a longer text that should wrap correctly across multiple lines without breaking the text sequences"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pdf.SetXY(10, 10) // Reset position for each iteration
		pdf.MultiCell(60, 5, text, "", "L", false)
	}
}

// BenchmarkWrite benchmarks the Write function
func BenchmarkWrite(b *testing.B) {
	pdf := New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.AddUTF8Font("DejaVuSans", "", "font/DejaVuSansCondensed.ttf")
	pdf.SetFont("DejaVuSans", "", 12)
	text := "This is flowing text with emoji 😀 that wraps naturally 🎉"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pdf.SetXY(10, 10) // Reset position for each iteration
		pdf.Write(5, text)
	}
}

// BenchmarkWriteASCII benchmarks Write with ASCII-only text
func BenchmarkWriteASCII(b *testing.B) {
	pdf := New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)
	text := "This is flowing text that wraps naturally without emoji"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pdf.SetXY(10, 10) // Reset position for each iteration
		pdf.Write(5, text)
	}
}

// BenchmarkText benchmarks the Text function
func BenchmarkText(b *testing.B) {
	pdf := New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.AddUTF8Font("DejaVuSans", "", "font/DejaVuSansCondensed.ttf")
	pdf.SetFont("DejaVuSans", "", 12)
	text := "Hello World with emoji 😀🎉"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pdf.Text(10, 10, text)
	}
}

// BenchmarkTextASCII benchmarks Text with ASCII-only text
func BenchmarkTextASCII(b *testing.B) {
	pdf := New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)
	text := "Hello World without emoji"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pdf.Text(10, 10, text)
	}
}

// BenchmarkSplitText benchmarks the SplitText function
func BenchmarkSplitText(b *testing.B) {
	pdf := New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.AddUTF8Font("DejaVuSans", "", "font/DejaVuSansCondensed.ttf")
	pdf.SetFont("DejaVuSans", "", 12)
	text := "This is a longer text with emoji 😀 that should wrap correctly 👍🏽 across multiple lines"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pdf.SplitText(text, 60)
	}
}

// BenchmarkSplitTextASCII benchmarks SplitText with ASCII-only text
func BenchmarkSplitTextASCII(b *testing.B) {
	pdf := New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)
	text := "This is a longer text that should wrap correctly across multiple lines without emoji"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pdf.SplitText(text, 60)
	}
}
