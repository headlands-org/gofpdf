package gofpdf

import (
	"testing"
)

// TestBackwardCompatibility_CoreFonts verifies that core fonts work correctly
func TestBackwardCompatibility_CoreFonts(t *testing.T) {
	coreFonts := []string{"Arial", "Times", "Courier", "Helvetica", "Symbol", "ZapfDingbats"}

	for _, fontName := range coreFonts {
		t.Run(fontName, func(t *testing.T) {
			pdf := New("P", "mm", "A4", "")
			pdf.AddPage()
			pdf.SetFont(fontName, "", 12)

			// Test basic text operations
			text := "Hello World 123"
			width := pdf.GetStringWidth(text)
			if width == 0 {
				t.Errorf("GetStringWidth returned 0 for %s", fontName)
			}

			// Test Cell
			pdf.Cell(40, 10, text)

			// Test CellFormat
			pdf.CellFormat(40, 10, text, "1", 1, "L", false, 0, "")

			// Test MultiCell
			pdf.MultiCell(60, 5, text, "", "L", false)

			// Test Write
			pdf.Write(5, text)

			// Test Text
			pdf.Text(10, 50, text)

			// Verify no errors
			if pdf.Error() != nil {
				t.Errorf("Core font %s produced error: %v", fontName, pdf.Error())
			}
		})
	}
}

// TestBackwardCompatibility_UTF8Fonts verifies that UTF-8 fonts work correctly
func TestBackwardCompatibility_UTF8Fonts(t *testing.T) {
	pdf := New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.AddUTF8Font("DejaVuSans", "", "font/DejaVuSansCondensed.ttf")
	pdf.SetFont("DejaVuSans", "", 12)

	// Test various text types
	testCases := []struct {
		name string
		text string
	}{
		{"ASCII", "Hello World 123"},
		{"Latin", "Héllo Wörld"},
		{"Greek", "Γειά σου Κόσμε"},
		{"Cyrillic", "Привет мир"},
		{"CJK", "你好世界"},
		{"Arabic", "مرحبا بالعالم"},
		{"Emoji", "Hello 😀 World 🎉"},
		{"Mixed", "Hello 世界 مرحبا 😀"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test GetStringWidth
			width := pdf.GetStringWidth(tc.text)
			if width == 0 {
				t.Errorf("GetStringWidth returned 0 for %s: %s", tc.name, tc.text)
			}

			// Test Cell
			pdf.SetXY(10, 10)
			pdf.Cell(40, 10, tc.text)

			// Test CellFormat
			pdf.SetXY(10, 20)
			pdf.CellFormat(40, 10, tc.text, "1", 1, "L", false, 0, "")

			// Test MultiCell
			pdf.SetXY(10, 30)
			pdf.MultiCell(60, 5, tc.text, "", "L", false)

			// Test Write
			pdf.SetXY(10, 50)
			pdf.Write(5, tc.text)

			// Test Text
			pdf.Text(10, 70, tc.text)

			// Verify no errors
			if pdf.Error() != nil {
				t.Errorf("UTF8 font test %s produced error: %v", tc.name, pdf.Error())
			}
		})
	}
}

// TestBackwardCompatibility_LegacyCode simulates typical legacy code patterns
func TestBackwardCompatibility_LegacyCode(t *testing.T) {
	// Pattern 1: Simple document with core font
	t.Run("SimpleCoreFont", func(t *testing.T) {
		pdf := New("P", "mm", "A4", "")
		pdf.AddPage()
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(40, 10, "Hello World")
		pdf.Ln(10)
		pdf.MultiCell(60, 5, "This is a longer text that wraps", "", "L", false)

		if pdf.Error() != nil {
			t.Errorf("Simple core font pattern failed: %v", pdf.Error())
		}
	})

	// Pattern 2: UTF-8 font with non-emoji text
	t.Run("UTF8NoEmoji", func(t *testing.T) {
		pdf := New("P", "mm", "A4", "")
		pdf.AddPage()
		pdf.AddUTF8Font("DejaVuSans", "", "font/DejaVuSansCondensed.ttf")
		pdf.SetFont("DejaVuSans", "", 12)

		text := "Hello World with special chars: àéîöü"
		width := pdf.GetStringWidth(text)
		if width == 0 {
			t.Error("GetStringWidth returned 0")
		}

		pdf.Cell(40, 10, text)
		pdf.Ln(10)
		pdf.MultiCell(60, 5, text, "", "L", false)

		if pdf.Error() != nil {
			t.Errorf("UTF8 no emoji pattern failed: %v", pdf.Error())
		}
	})

	// Pattern 3: Table generation
	t.Run("TableGeneration", func(t *testing.T) {
		pdf := New("P", "mm", "A4", "")
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 12)

		// Header
		headers := []string{"Name", "Age", "City"}
		for _, header := range headers {
			pdf.CellFormat(60, 10, header, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(-1)

		// Data rows
		pdf.SetFont("Arial", "", 10)
		data := [][]string{
			{"John Doe", "30", "New York"},
			{"Jane Smith", "25", "London"},
			{"Bob Johnson", "35", "Paris"},
		}

		for _, row := range data {
			for _, cell := range row {
				pdf.CellFormat(60, 8, cell, "1", 0, "L", false, 0, "")
			}
			pdf.Ln(-1)
		}

		if pdf.Error() != nil {
			t.Errorf("Table generation pattern failed: %v", pdf.Error())
		}
	})

	// Pattern 4: Mixed fonts and styles
	t.Run("MixedFontsStyles", func(t *testing.T) {
		pdf := New("P", "mm", "A4", "")
		pdf.AddPage()

		// Regular text
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(40, 10, "Regular text")
		pdf.Ln(10)

		// Bold text
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(40, 10, "Bold text")
		pdf.Ln(10)

		// Italic text
		pdf.SetFont("Arial", "I", 12)
		pdf.Cell(40, 10, "Italic text")
		pdf.Ln(10)

		// Bold italic
		pdf.SetFont("Arial", "BI", 12)
		pdf.Cell(40, 10, "Bold italic text")

		if pdf.Error() != nil {
			t.Errorf("Mixed fonts/styles pattern failed: %v", pdf.Error())
		}
	})
}

// TestBackwardCompatibility_StringWidthConsistency verifies that string width
// calculations remain consistent for non-emoji text
func TestBackwardCompatibility_StringWidthConsistency(t *testing.T) {
	pdf := New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)

	testCases := []struct {
		text string
	}{
		{"Hello"},
		{"Hello World"},
		{"The quick brown fox"},
		{"123456789"},
		{"!@#$%^&*()"},
	}

	for _, tc := range testCases {
		t.Run(tc.text, func(t *testing.T) {
			// Calculate width multiple times - should be consistent
			width1 := pdf.GetStringWidth(tc.text)
			width2 := pdf.GetStringWidth(tc.text)
			width3 := pdf.GetStringWidth(tc.text)

			if width1 != width2 || width2 != width3 {
				t.Errorf("GetStringWidth inconsistent for %q: %f, %f, %f",
					tc.text, width1, width2, width3)
			}

			if width1 == 0 {
				t.Errorf("GetStringWidth returned 0 for %q", tc.text)
			}
		})
	}
}

// TestBackwardCompatibility_NoAPIChanges verifies that the public API hasn't changed
func TestBackwardCompatibility_NoAPIChanges(t *testing.T) {
	// This test verifies that common API patterns still compile and work
	pdf := New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)

	// Test all key methods exist and have correct signatures
	_ = pdf.GetStringWidth("test")
	pdf.Cell(40, 10, "test")
	pdf.CellFormat(40, 10, "test", "1", 0, "L", false, 0, "")
	pdf.MultiCell(60, 5, "test", "", "L", false)
	pdf.Write(5, "test")
	pdf.Text(10, 10, "test")
	_ = pdf.SplitText("test", 60)

	// Verify no errors
	if pdf.Error() != nil {
		t.Errorf("API compatibility test failed: %v", pdf.Error())
	}
}

// TestBackwardCompatibility_NonUTF8Encoding verifies that non-UTF-8 encodings still work
func TestBackwardCompatibility_NonUTF8Encoding(t *testing.T) {
	pdf := New("P", "mm", "A4", "")
	pdf.AddPage()

	// Core fonts use cp1252 encoding by default
	pdf.SetFont("Arial", "", 12)

	// Test standard cp1252 characters
	text := "Standard ASCII: Hello World 123"
	width := pdf.GetStringWidth(text)
	if width == 0 {
		t.Error("GetStringWidth returned 0 for ASCII text")
	}

	pdf.Cell(40, 10, text)
	pdf.Ln(10)
	pdf.MultiCell(60, 5, text, "", "L", false)

	if pdf.Error() != nil {
		t.Errorf("Non-UTF-8 encoding test failed: %v", pdf.Error())
	}
}

// TestBackwardCompatibility_MemoryEfficiency verifies that the map-based approach
// works correctly by testing repeated width calculations
func TestBackwardCompatibility_MemoryEfficiency(t *testing.T) {
	pdf := New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.AddUTF8Font("DejaVuSans", "", "font/DejaVuSansCondensed.ttf")
	pdf.SetFont("DejaVuSans", "", 12)

	// Use only a small set of characters
	text := "Hello"

	// Verify that getting width for the same text multiple times is consistent
	// This tests that the map-based Cw storage works correctly
	width1 := pdf.GetStringWidth(text)
	width2 := pdf.GetStringWidth(text)
	width3 := pdf.GetStringWidth(text)

	if width1 != width2 || width2 != width3 {
		t.Errorf("GetStringWidth inconsistent for repeated calls: %f, %f, %f",
			width1, width2, width3)
	}

	if width1 == 0 {
		t.Error("GetStringWidth returned 0")
	}

	// Test with different text to verify the mechanism works for new characters
	text2 := "World"
	width4 := pdf.GetStringWidth(text2)
	if width4 == 0 {
		t.Error("GetStringWidth returned 0 for second text")
	}

	// Verify original text still works
	width5 := pdf.GetStringWidth(text)
	if width5 != width1 {
		t.Errorf("GetStringWidth changed for original text: %f != %f", width5, width1)
	}
}
