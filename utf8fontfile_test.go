/*
 * Copyright (c) 2019 Arteom Korotkiy (Gmail: arteomkorotkiy)
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

package gofpdf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
)

// TestPackUint32 tests the packUint32 function with various values
func TestPackUint32(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected []byte
	}{
		{
			name:     "Zero",
			input:    0,
			expected: []byte{0x00, 0x00, 0x00, 0x00},
		},
		{
			name:     "Small value",
			input:    255,
			expected: []byte{0x00, 0x00, 0x00, 0xFF},
		},
		{
			name:     "BMP limit",
			input:    0xFFFF,
			expected: []byte{0x00, 0x00, 0xFF, 0xFF},
		},
		{
			name:     "Unicode limit",
			input:    0x10FFFF,
			expected: []byte{0x00, 0x10, 0xFF, 0xFF},
		},
		{
			name:     "Modern emoji range",
			input:    0x1F600,
			expected: []byte{0x00, 0x01, 0xF6, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := packUint32(tt.input)

			// Check length
			if len(result) != 4 {
				t.Errorf("packUint32(%d) returned %d bytes, expected 4", tt.input, len(result))
			}

			// Check byte values
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("packUint32(%d) = %v, expected %v", tt.input, result, tt.expected)
			}

			// Verify big-endian encoding by decoding
			decoded := binary.BigEndian.Uint32(result)
			if decoded != uint32(tt.input) {
				t.Errorf("packUint32(%d) did not produce valid big-endian encoding: decoded as %d", tt.input, decoded)
			}
		})
	}
}

// TestBuildCmapGroups tests the buildCmapGroups function
func TestBuildCmapGroups(t *testing.T) {
	tests := []struct {
		name     string
		input    map[int]int
		expected []cmapGroup
	}{
		{
			name:     "Empty map",
			input:    map[int]int{},
			expected: []cmapGroup{},
		},
		{
			name: "Single character",
			input: map[int]int{
				65: 10,
			},
			expected: []cmapGroup{
				{startCharCode: 65, endCharCode: 65, startGlyphID: 10},
			},
		},
		{
			name: "Sequential characters and glyphs",
			input: map[int]int{
				65: 10,
				66: 11,
				67: 12,
			},
			expected: []cmapGroup{
				{startCharCode: 65, endCharCode: 67, startGlyphID: 10},
			},
		},
		{
			name: "Non-sequential characters",
			input: map[int]int{
				65: 10,
				70: 15,
			},
			expected: []cmapGroup{
				{startCharCode: 65, endCharCode: 65, startGlyphID: 10},
				{startCharCode: 70, endCharCode: 70, startGlyphID: 15},
			},
		},
		{
			name: "Sequential characters but non-sequential glyphs",
			input: map[int]int{
				65: 10,
				66: 11,
				67: 20, // Breaks sequence
			},
			expected: []cmapGroup{
				{startCharCode: 65, endCharCode: 66, startGlyphID: 10},
				{startCharCode: 67, endCharCode: 67, startGlyphID: 20},
			},
		},
		{
			name: "Mixed sequential and sparse",
			input: map[int]int{
				65:  10,
				66:  11,
				67:  12,
				100: 50,
				101: 51,
				200: 100,
			},
			expected: []cmapGroup{
				{startCharCode: 65, endCharCode: 67, startGlyphID: 10},
				{startCharCode: 100, endCharCode: 101, startGlyphID: 50},
				{startCharCode: 200, endCharCode: 200, startGlyphID: 100},
			},
		},
		{
			name: "Modern emoji characters",
			input: map[int]int{
				0x1F600: 100,
				0x1F601: 101,
				0x1F602: 102,
				0x1F680: 200,
			},
			expected: []cmapGroup{
				{startCharCode: 0x1F600, endCharCode: 0x1F602, startGlyphID: 100},
				{startCharCode: 0x1F680, endCharCode: 0x1F680, startGlyphID: 200},
			},
		},
		{
			name: "Full Unicode range",
			input: map[int]int{
				0x0041:   10,
				0x0042:   11,
				0x10FFFF: 1000,
			},
			expected: []cmapGroup{
				{startCharCode: 0x0041, endCharCode: 0x0042, startGlyphID: 10},
				{startCharCode: 0x10FFFF, endCharCode: 0x10FFFF, startGlyphID: 1000},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildCmapGroups(tt.input)

			// Check number of groups
			if len(result) != len(tt.expected) {
				t.Errorf("buildCmapGroups() returned %d groups, expected %d", len(result), len(tt.expected))
			}

			// Check each group
			for i, expected := range tt.expected {
				if i >= len(result) {
					t.Errorf("Missing group %d", i)
					continue
				}
				group := result[i]
				if group.startCharCode != expected.startCharCode {
					t.Errorf("Group %d: startCharCode = %d, expected %d", i, group.startCharCode, expected.startCharCode)
				}
				if group.endCharCode != expected.endCharCode {
					t.Errorf("Group %d: endCharCode = %d, expected %d", i, group.endCharCode, expected.endCharCode)
				}
				if group.startGlyphID != expected.startGlyphID {
					t.Errorf("Group %d: startGlyphID = %d, expected %d", i, group.startGlyphID, expected.startGlyphID)
				}
			}
		})
	}
}

// TestWriteCmapFormat12Header tests the writeCmapFormat12Header function
func TestWriteCmapFormat12Header(t *testing.T) {
	tests := []struct {
		name      string
		numGroups uint32
	}{
		{
			name:      "Zero groups",
			numGroups: 0,
		},
		{
			name:      "One group",
			numGroups: 1,
		},
		{
			name:      "Multiple groups",
			numGroups: 10,
		},
		{
			name:      "Large number of groups",
			numGroups: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := writeCmapFormat12Header(tt.numGroups)

			// Check length
			if len(result) != 16 {
				t.Errorf("writeCmapFormat12Header(%d) returned %d bytes, expected 16", tt.numGroups, len(result))
			}

			// Parse header fields
			format := binary.BigEndian.Uint16(result[0:2])
			reserved := binary.BigEndian.Uint16(result[2:4])
			length := binary.BigEndian.Uint32(result[4:8])
			language := binary.BigEndian.Uint32(result[8:12])
			numGroups := binary.BigEndian.Uint32(result[12:16])

			// Verify format
			if format != 12 {
				t.Errorf("Format field = %d, expected 12", format)
			}

			// Verify reserved
			if reserved != 0 {
				t.Errorf("Reserved field = %d, expected 0", reserved)
			}

			// Verify length
			expectedLength := uint32(16 + 12*tt.numGroups)
			if length != expectedLength {
				t.Errorf("Length field = %d, expected %d", length, expectedLength)
			}

			// Verify language
			if language != 0 {
				t.Errorf("Language field = %d, expected 0", language)
			}

			// Verify numGroups
			if numGroups != tt.numGroups {
				t.Errorf("NumGroups field = %d, expected %d", numGroups, tt.numGroups)
			}
		})
	}
}

// TestWriteCmapFormat12HeaderBigEndian verifies big-endian byte order explicitly
func TestWriteCmapFormat12HeaderBigEndian(t *testing.T) {
	result := writeCmapFormat12Header(3)

	expected := []byte{
		0x00, 0x0C, // format = 12
		0x00, 0x00, // reserved = 0
		0x00, 0x00, 0x00, 0x34, // length = 52 (16 + 12*3)
		0x00, 0x00, 0x00, 0x00, // language = 0
		0x00, 0x00, 0x00, 0x03, // numGroups = 3
	}

	if !bytes.Equal(result, expected) {
		t.Errorf("writeCmapFormat12Header(3) byte order incorrect\nGot:      %v\nExpected: %v", result, expected)
	}
}

// TestBuildCmapGroupsOptimization verifies that sequential ranges are optimized
func TestBuildCmapGroupsOptimization(t *testing.T) {
	// Create a large sequential range
	input := make(map[int]int)
	for i := 0; i < 1000; i++ {
		input[i] = i + 100
	}

	result := buildCmapGroups(input)

	// Should be optimized into a single group
	if len(result) != 1 {
		t.Errorf("Sequential range should be optimized to 1 group, got %d groups", len(result))
	}

	if len(result) == 1 {
		if result[0].startCharCode != 0 {
			t.Errorf("Start char code = %d, expected 0", result[0].startCharCode)
		}
		if result[0].endCharCode != 999 {
			t.Errorf("End char code = %d, expected 999", result[0].endCharCode)
		}
		if result[0].startGlyphID != 100 {
			t.Errorf("Start glyph ID = %d, expected 100", result[0].startGlyphID)
		}
	}
}

// TestCmapGroupStruct verifies the cmapGroup struct fields
func TestCmapGroupStruct(t *testing.T) {
	group := cmapGroup{
		startCharCode: 0x1F600,
		endCharCode:   0x1F64F,
		startGlyphID:  100,
	}

	if group.startCharCode != 0x1F600 {
		t.Errorf("startCharCode = %d, expected %d", group.startCharCode, 0x1F600)
	}
	if group.endCharCode != 0x1F64F {
		t.Errorf("endCharCode = %d, expected %d", group.endCharCode, 0x1F64F)
	}
	if group.startGlyphID != 100 {
		t.Errorf("startGlyphID = %d, expected 100", group.startGlyphID)
	}
}

// BenchmarkPackUint32 benchmarks the packUint32 function
func BenchmarkPackUint32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		packUint32(0x1F600)
	}
}

// BenchmarkBuildCmapGroups benchmarks the buildCmapGroups function
func BenchmarkBuildCmapGroups(b *testing.B) {
	input := make(map[int]int)
	for i := 0; i < 1000; i++ {
		input[i] = i + 100
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildCmapGroups(input)
	}
}

// BenchmarkWriteCmapFormat12Header benchmarks the writeCmapFormat12Header function
func BenchmarkWriteCmapFormat12Header(b *testing.B) {
	for i := 0; i < b.N; i++ {
		writeCmapFormat12Header(100)
	}
}

// Helper function to create a mock Format 12 CMAP table for testing
func createMockCmapFormat12(groups []cmapGroup) []byte {
	data := make([]byte, 0)

	// Write header
	header := writeCmapFormat12Header(uint32(len(groups)))
	data = append(data, header...)

	// Write each group
	for _, group := range groups {
		data = append(data, packUint32(int(group.startCharCode))...)
		data = append(data, packUint32(int(group.endCharCode))...)
		data = append(data, packUint32(int(group.startGlyphID))...)
	}

	return data
}

// TestParseCmapFormat12 tests the parseCmapFormat12 function
func TestParseCmapFormat12(t *testing.T) {
	tests := []struct {
		name                  string
		groups                []cmapGroup
		expectedCharToSymbol  map[int]int
		expectedSymbolToChars map[int][]int
	}{
		{
			name:                  "Empty groups",
			groups:                []cmapGroup{},
			expectedCharToSymbol:  map[int]int{},
			expectedSymbolToChars: map[int][]int{},
		},
		{
			name: "Single character",
			groups: []cmapGroup{
				{startCharCode: 65, endCharCode: 65, startGlyphID: 10},
			},
			expectedCharToSymbol: map[int]int{
				65: 10,
			},
			expectedSymbolToChars: map[int][]int{
				10: {65},
			},
		},
		{
			name: "Sequential range",
			groups: []cmapGroup{
				{startCharCode: 65, endCharCode: 67, startGlyphID: 10},
			},
			expectedCharToSymbol: map[int]int{
				65: 10,
				66: 11,
				67: 12,
			},
			expectedSymbolToChars: map[int][]int{
				10: {65},
				11: {66},
				12: {67},
			},
		},
		{
			name: "Multiple groups",
			groups: []cmapGroup{
				{startCharCode: 65, endCharCode: 67, startGlyphID: 10},
				{startCharCode: 100, endCharCode: 101, startGlyphID: 50},
			},
			expectedCharToSymbol: map[int]int{
				65:  10,
				66:  11,
				67:  12,
				100: 50,
				101: 51,
			},
			expectedSymbolToChars: map[int][]int{
				10: {65},
				11: {66},
				12: {67},
				50: {100},
				51: {101},
			},
		},
		{
			name: "Modern emoji range",
			groups: []cmapGroup{
				{startCharCode: 0x1F600, endCharCode: 0x1F602, startGlyphID: 100},
			},
			expectedCharToSymbol: map[int]int{
				0x1F600: 100,
				0x1F601: 101,
				0x1F602: 102,
			},
			expectedSymbolToChars: map[int][]int{
				100: {0x1F600},
				101: {0x1F601},
				102: {0x1F602},
			},
		},
		{
			name: "Full Unicode range",
			groups: []cmapGroup{
				{startCharCode: 0x0041, endCharCode: 0x0042, startGlyphID: 10},
				{startCharCode: 0x10FFFF, endCharCode: 0x10FFFF, startGlyphID: 1000},
			},
			expectedCharToSymbol: map[int]int{
				0x0041:   10,
				0x0042:   11,
				0x10FFFF: 1000,
			},
			expectedSymbolToChars: map[int][]int{
				10:   {0x0041},
				11:   {0x0042},
				1000: {0x10FFFF},
			},
		},
		{
			name: "BMP characters",
			groups: []cmapGroup{
				{startCharCode: 0x0020, endCharCode: 0x007E, startGlyphID: 1}, // ASCII printable
			},
			expectedCharToSymbol: map[int]int{
				0x0020: 1,
				0x0021: 2,
				0x007E: 95,
			},
			expectedSymbolToChars: map[int][]int{
				1:  {0x0020},
				2:  {0x0021},
				95: {0x007E},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock CMAP Format 12 table
			data := createMockCmapFormat12(tt.groups)

			// Create a utf8FontFile with the mock data
			utf := newUTF8Font(&fileReader{
				readerPosition: 0,
				array:          data,
			})

			// Parse the CMAP table
			symbolCharDict, charSymbolDict := utf.parseCmapFormat12(0)

			// Calculate expected total mappings from groups
			expectedTotal := 0
			for _, group := range tt.groups {
				expectedTotal += int(group.endCharCode - group.startCharCode + 1)
			}

			// Verify charSymbolDictionary total count
			if len(charSymbolDict) != expectedTotal {
				t.Errorf("charSymbolDict length = %d, expected %d", len(charSymbolDict), expectedTotal)
			}

			// Verify specific charSymbolDictionary entries
			for char, expectedSymbol := range tt.expectedCharToSymbol {
				if symbol, ok := charSymbolDict[char]; !ok {
					t.Errorf("charSymbolDict missing char %d", char)
				} else if symbol != expectedSymbol {
					t.Errorf("charSymbolDict[%d] = %d, expected %d", char, symbol, expectedSymbol)
				}
			}

			// Verify symbolCharDictionary total count
			if len(symbolCharDict) != expectedTotal {
				t.Errorf("symbolCharDict length = %d, expected %d", len(symbolCharDict), expectedTotal)
			}

			// Verify specific symbolCharDictionary entries
			for symbol, expectedChars := range tt.expectedSymbolToChars {
				if chars, ok := symbolCharDict[symbol]; !ok {
					t.Errorf("symbolCharDict missing symbol %d", symbol)
				} else {
					if len(chars) != len(expectedChars) {
						t.Errorf("symbolCharDict[%d] length = %d, expected %d", symbol, len(chars), len(expectedChars))
					}
					for i, expectedChar := range expectedChars {
						if i >= len(chars) || chars[i] != expectedChar {
							t.Errorf("symbolCharDict[%d][%d] = %v, expected %d", symbol, i, chars, expectedChar)
						}
					}
				}
			}
		})
	}
}

// TestParseCmapFormat12InvalidFormat tests error handling for invalid format
func TestParseCmapFormat12InvalidFormat(t *testing.T) {
	// Create data with format 4 instead of 12
	data := make([]byte, 16)
	binary.BigEndian.PutUint16(data[0:2], 4) // Wrong format

	utf := newUTF8Font(&fileReader{
		readerPosition: 0,
		array:          data,
	})

	symbolCharDict, charSymbolDict := utf.parseCmapFormat12(0)

	// Should return empty dictionaries
	if len(symbolCharDict) != 0 {
		t.Errorf("Expected empty symbolCharDict, got %d entries", len(symbolCharDict))
	}
	if len(charSymbolDict) != 0 {
		t.Errorf("Expected empty charSymbolDict, got %d entries", len(charSymbolDict))
	}
}

// TestParseCmapFormat12InvalidLength tests error handling for invalid length
func TestParseCmapFormat12InvalidLength(t *testing.T) {
	data := make([]byte, 16)
	binary.BigEndian.PutUint16(data[0:2], 12)  // format = 12
	binary.BigEndian.PutUint16(data[2:4], 0)   // reserved = 0
	binary.BigEndian.PutUint32(data[4:8], 999) // Invalid length
	binary.BigEndian.PutUint32(data[8:12], 0)  // language = 0
	binary.BigEndian.PutUint32(data[12:16], 2) // numGroups = 2 (requires length = 40, not 999)

	utf := newUTF8Font(&fileReader{
		readerPosition: 0,
		array:          data,
	})

	symbolCharDict, charSymbolDict := utf.parseCmapFormat12(0)

	// Should return empty dictionaries
	if len(symbolCharDict) != 0 {
		t.Errorf("Expected empty symbolCharDict, got %d entries", len(symbolCharDict))
	}
	if len(charSymbolDict) != 0 {
		t.Errorf("Expected empty charSymbolDict, got %d entries", len(charSymbolDict))
	}
}

// TestParseCmapFormat12LargeRange tests parsing a large character range
func TestParseCmapFormat12LargeRange(t *testing.T) {
	// Create a group with a large range (but not too large to avoid timeout)
	groups := []cmapGroup{
		{startCharCode: 0x1F300, endCharCode: 0x1F320, startGlyphID: 100},
	}

	data := createMockCmapFormat12(groups)

	utf := newUTF8Font(&fileReader{
		readerPosition: 0,
		array:          data,
	})

	symbolCharDict, charSymbolDict := utf.parseCmapFormat12(0)

	// Avoid unused variable warning
	_ = symbolCharDict

	// Verify we have the correct number of mappings
	expectedCount := 0x1F320 - 0x1F300 + 1
	if len(charSymbolDict) != expectedCount {
		t.Errorf("charSymbolDict length = %d, expected %d", len(charSymbolDict), expectedCount)
	}

	// Spot check a few mappings
	if charSymbolDict[0x1F300] != 100 {
		t.Errorf("charSymbolDict[0x1F300] = %d, expected 100", charSymbolDict[0x1F300])
	}
	if charSymbolDict[0x1F310] != 116 {
		t.Errorf("charSymbolDict[0x1F310] = %d, expected 116", charSymbolDict[0x1F310])
	}
	if charSymbolDict[0x1F320] != 132 {
		t.Errorf("charSymbolDict[0x1F320] = %d, expected 132", charSymbolDict[0x1F320])
	}
}

// BenchmarkParseCmapFormat12 benchmarks the parseCmapFormat12 function
func BenchmarkParseCmapFormat12(b *testing.B) {
	groups := []cmapGroup{
		{startCharCode: 0x0041, endCharCode: 0x005A, startGlyphID: 10},    // A-Z
		{startCharCode: 0x0061, endCharCode: 0x007A, startGlyphID: 36},    // a-z
		{startCharCode: 0x1F600, endCharCode: 0x1F64F, startGlyphID: 100}, // Emoji
	}

	data := createMockCmapFormat12(groups)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utf := newUTF8Font(&fileReader{
			readerPosition: 0,
			array:          data,
		})
		utf.parseCmapFormat12(0)
	}
}

// TestGenerateToUnicodeCMap tests the generateToUnicodeCMap function
func TestGenerateToUnicodeCMap(t *testing.T) {
	tests := []struct {
		name            string
		cidToUnicode    map[int]int
		expectCodespace string
		expectEntries   map[int]string
	}{
		{
			name:            "Basic BMP",
			cidToUnicode:    map[int]int{1: 0x41, 2: 0x5A, 3: 0x64},
			expectCodespace: "<0000> <FFFF>",
			expectEntries: map[int]string{
				1: "0041",
				2: "005A",
				3: "0064",
			},
		},
		{
			name:            "Supplementary",
			cidToUnicode:    map[int]int{1: 0x1F600},
			expectCodespace: "<0000> <FFFF>",
			expectEntries: map[int]string{
				1: "D83DDE00",
			},
		},
		{
			name:            "Mixed",
			cidToUnicode:    map[int]int{1: 0x41, 2: 0x1F600, 3: 0x2665},
			expectCodespace: "<0000> <FFFF>",
			expectEntries: map[int]string{
				1: "0041",
				2: "D83DDE00",
				3: "2665",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateToUnicodeCMap(tt.cidToUnicode)

			if len(result) == 0 {
				t.Fatal("Generated CMap should not be empty")
			}

			required := []string{
				"/CIDInit /ProcSet findresource begin",
				"begincmap",
				"/CMapName /Adobe-Identity-UCS def",
				"/CMapType 2 def",
				"beginbfchar",
				"endbfchar",
				"endcmap",
			}
			for _, token := range required {
				if !bytes.Contains([]byte(result), []byte(token)) {
					t.Errorf("expected token %q in CMap", token)
				}
			}

			if !bytes.Contains([]byte(result), []byte(tt.expectCodespace)) {
				t.Errorf("expected codespace %s", tt.expectCodespace)
			}

			for cid, hex := range tt.expectEntries {
				entry := fmt.Sprintf("<%04X> <%s>", cid, hex)
				if !bytes.Contains([]byte(result), []byte(entry)) {
					t.Errorf("missing mapping %s", entry)
				}
			}
		})
	}
}

func TestGenerateToUnicodeCMapBlockSplitting(t *testing.T) {
	cidToUnicode := make(map[int]int)
	for i := 1; i <= 205; i++ {
		cidToUnicode[i] = 0x30 + i
	}

	result := generateToUnicodeCMap(cidToUnicode)
	if bytes.Count([]byte(result), []byte("beginbfchar")) < 3 {
		t.Fatalf("expected multiple beginbfchar blocks, got %d", bytes.Count([]byte(result), []byte("beginbfchar")))
	}
}

func TestGenerateToUnicodeCMapSupplementaryPairs(t *testing.T) {
	cidToUnicode := map[int]int{1: 0x1F468, 2: 0x1F469}
	result := generateToUnicodeCMap(cidToUnicode)
	if !bytes.Contains([]byte(result), []byte("<0001> <D83DDC68>")) {
		t.Error("expected surrogate pair for 0x1F468")
	}
	if !bytes.Contains([]byte(result), []byte("<0002> <D83DDC69>")) {
		t.Error("expected surrogate pair for 0x1F469")
	}
}

// BenchmarkGenerateToUnicodeCMap benchmarks the generateToUnicodeCMap function
func BenchmarkGenerateToUnicodeCMap(b *testing.B) {
	// Create a realistic character set
	usedRunes := make(map[int]int)
	// ASCII printable
	for i := 0x20; i <= 0x7E; i++ {
		usedRunes[len(usedRunes)] = i
	}
	// Some emoji
	for i := 0x1F600; i <= 0x1F64F; i++ {
		usedRunes[len(usedRunes)] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateToUnicodeCMap(usedRunes)
	}
}

// BenchmarkGenerateToUnicodeCMapLarge benchmarks with a large character set
func BenchmarkGenerateToUnicodeCMapLarge(b *testing.B) {
	usedRunes := make(map[int]int)
	// Full BMP
	for i := 0; i < 0xFFFF; i += 10 {
		usedRunes[len(usedRunes)] = i
	}
	// Supplementary planes
	for i := 0x10000; i < 0x10FFFF; i += 100 {
		usedRunes[len(usedRunes)] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateToUnicodeCMap(usedRunes)
	}
}
