/*
 * Copyright (c) 2013 Kurt Jung (Gmail: kurt.w.jung)
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
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
)

// TestUtf8ToUtf16 tests the utf8toutf16 function with various character types
func TestUtf8ToUtf16(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedHex string
		description string
		codepoint   string
	}{
		// 1-byte sequences (ASCII)
		{
			name:        "ASCII_A",
			input:       "A",
			expectedHex: "feff0041",
			description: "ASCII letter A",
			codepoint:   "U+0041",
		},
		{
			name:        "ASCII_Hello",
			input:       "Hello",
			expectedHex: "feff00480065006c006c006f",
			description: "ASCII word 'Hello'",
			codepoint:   "U+0048,U+0065,U+006C,U+006C,U+006F",
		},

		// 2-byte sequences
		{
			name:        "2byte_euro",
			input:       "\u00A9", // Copyright symbol
			expectedHex: "feff00a9",
			description: "2-byte character: Copyright symbol",
			codepoint:   "U+00A9",
		},
		{
			name:        "2byte_alpha",
			input:       "\u03B1", // Greek alpha
			expectedHex: "feff03b1",
			description: "2-byte character: Greek alpha",
			codepoint:   "U+03B1",
		},

		// 3-byte sequences
		{
			name:        "3byte_hiragana",
			input:       "\u3042", // Hiragana A
			expectedHex: "feff3042",
			description: "3-byte character: Hiragana A",
			codepoint:   "U+3042",
		},
		{
			name:        "3byte_cjk",
			input:       "\u4E2D", // Chinese character
			expectedHex: "feff4e2d",
			description: "3-byte character: Chinese",
			codepoint:   "U+4E2D",
		},

		// 4-byte sequences (emoji and supplementary plane characters)
		{
			name:        "4byte_emoji_party",
			input:       "\U0001F389", // Party popper emoji
			expectedHex: "feffd83cdf89",
			description: "4-byte character: Party popper emoji",
			codepoint:   "U+1F389",
		},
		{
			name:        "4byte_emoji_grinning",
			input:       "\U0001F600", // Grinning face emoji
			expectedHex: "feffd83dde00",
			description: "4-byte character: Grinning face emoji",
			codepoint:   "U+1F600",
		},
		{
			name:        "4byte_emoji_rocket",
			input:       "\U0001F680", // Rocket emoji
			expectedHex: "feffd83dde80",
			description: "4-byte character: Rocket emoji",
			codepoint:   "U+1F680",
		},
		{
			name:        "4byte_math_bold_H",
			input:       "\U0001D573", // Mathematical bold italic capital H
			expectedHex: "feffd835dd73",
			description: "4-byte character: Math bold italic H",
			codepoint:   "U+1D573",
		},
		{
			name:        "4byte_linear_b",
			input:       "\U00010000", // Linear B syllable B008 A
			expectedHex: "feffd800dc00",
			description: "4-byte character: Linear B (first supplementary plane char)",
			codepoint:   "U+10000",
		},

		// Mixed sequences
		{
			name:        "mixed_hello_emoji",
			input:       "Hello \U0001F600",
			expectedHex: "feff00480065006c006c006f0020d83dde00",
			description: "Mixed: ASCII + emoji",
			codepoint:   "Mixed",
		},
		{
			name:        "mixed_all_types",
			input:       "A\u00A9\u3042\U0001F389",
			expectedHex: "feff004100a93042d83cdf89",
			description: "Mixed: 1-byte + 2-byte + 3-byte + 4-byte",
			codepoint:   "Mixed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utf8toutf16(tt.input)
			resultHex := hex.EncodeToString([]byte(result))

			if resultHex != tt.expectedHex {
				t.Errorf("utf8toutf16(%q) failed\n  Description: %s (%s)\n  Expected: %s\n  Got:      %s",
					tt.input, tt.description, tt.codepoint, tt.expectedHex, resultHex)
			} else {
				t.Logf("PASS: %s (%s) -> %s", tt.description, tt.codepoint, resultHex)
			}
		})
	}
}

// TestUtf8ToUtf16WithoutBOM tests the utf8toutf16 function without BOM
func TestUtf8ToUtf16WithoutBOM(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedHex string
		description string
	}{
		{
			name:        "ASCII_without_BOM",
			input:       "A",
			expectedHex: "0041",
			description: "ASCII without BOM",
		},
		{
			name:        "emoji_without_BOM",
			input:       "\U0001F389",
			expectedHex: "d83cdf89",
			description: "Emoji without BOM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utf8toutf16(tt.input, false)
			resultHex := hex.EncodeToString([]byte(result))

			if resultHex != tt.expectedHex {
				t.Errorf("utf8toutf16(%q, false) failed\n  Description: %s\n  Expected: %s\n  Got:      %s",
					tt.input, tt.description, tt.expectedHex, resultHex)
			}
		})
	}
}

// TestUtf8ToUtf16SurrogatePairs tests that surrogate pairs are calculated correctly
func TestUtf8ToUtf16SurrogatePairs(t *testing.T) {
	// Test specific surrogate pair calculations
	tests := []struct {
		codepoint    uint32
		expectedHigh uint16
		expectedLow  uint16
		char         string
		description  string
	}{
		{
			codepoint:    0x10000,
			expectedHigh: 0xD800,
			expectedLow:  0xDC00,
			char:         "\U00010000",
			description:  "U+10000 (first supplementary plane)",
		},
		{
			codepoint:    0x1F389,
			expectedHigh: 0xD83C,
			expectedLow:  0xDF89,
			char:         "\U0001F389",
			description:  "U+1F389 (party popper emoji)",
		},
		{
			codepoint:    0x1F600,
			expectedHigh: 0xD83D,
			expectedLow:  0xDE00,
			char:         "\U0001F600",
			description:  "U+1F600 (grinning face emoji)",
		},
		{
			codepoint:    0x1F680,
			expectedHigh: 0xD83D,
			expectedLow:  0xDE80,
			char:         "\U0001F680",
			description:  "U+1F680 (rocket emoji)",
		},
		{
			codepoint:    0x1D573,
			expectedHigh: 0xD835,
			expectedLow:  0xDD73,
			char:         "\U0001D573",
			description:  "U+1D573 (math bold italic H)",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("U+%04X", tt.codepoint), func(t *testing.T) {
			result := utf8toutf16(tt.char, false)

			if len(result) != 4 {
				t.Fatalf("Expected 4 bytes for surrogate pair, got %d", len(result))
			}

			highSurrogate := uint16(result[0])<<8 | uint16(result[1])
			lowSurrogate := uint16(result[2])<<8 | uint16(result[3])

			if highSurrogate != tt.expectedHigh {
				t.Errorf("%s: High surrogate mismatch\n  Expected: 0x%04X\n  Got:      0x%04X",
					tt.description, tt.expectedHigh, highSurrogate)
			}

			if lowSurrogate != tt.expectedLow {
				t.Errorf("%s: Low surrogate mismatch\n  Expected: 0x%04X\n  Got:      0x%04X",
					tt.description, tt.expectedLow, lowSurrogate)
			}

			if highSurrogate == tt.expectedHigh && lowSurrogate == tt.expectedLow {
				t.Logf("PASS: %s -> High: 0x%04X, Low: 0x%04X",
					tt.description, highSurrogate, lowSurrogate)
			}
		})
	}
}

// TestUtf8ToUtf16EmojiRange tests various emojis in the range U+1F300-U+1F9FF
func TestUtf8ToUtf16EmojiRange(t *testing.T) {
	emojis := []struct {
		emoji       string
		codepoint   uint32
		description string
	}{
		{"\U0001F300", 0x1F300, "Cyclone"},
		{"\U0001F389", 0x1F389, "Party popper"},
		{"\U0001F600", 0x1F600, "Grinning face"},
		{"\U0001F680", 0x1F680, "Rocket"},
		{"\U0001F9FF", 0x1F9FF, "Nazar amulet"},
	}

	for _, e := range emojis {
		t.Run(fmt.Sprintf("U+%04X_%s", e.codepoint, e.description), func(t *testing.T) {
			result := utf8toutf16(e.emoji, false)

			if len(result) != 4 {
				t.Errorf("Expected 4 bytes for emoji %s (U+%04X), got %d",
					e.description, e.codepoint, len(result))
				return
			}

			highSurrogate := uint16(result[0])<<8 | uint16(result[1])
			lowSurrogate := uint16(result[2])<<8 | uint16(result[3])

			// Verify surrogate range
			if highSurrogate < 0xD800 || highSurrogate > 0xDBFF {
				t.Errorf("High surrogate out of range for %s: 0x%04X",
					e.description, highSurrogate)
			}

			if lowSurrogate < 0xDC00 || lowSurrogate > 0xDFFF {
				t.Errorf("Low surrogate out of range for %s: 0x%04X",
					e.description, lowSurrogate)
			}

			// Verify we can reconstruct the codepoint
			reconstructed := ((uint32(highSurrogate) - 0xD800) << 10) +
				(uint32(lowSurrogate) - 0xDC00) + 0x10000

			if reconstructed != e.codepoint {
				t.Errorf("Codepoint reconstruction failed for %s\n  Expected: U+%04X\n  Got:      U+%04X",
					e.description, e.codepoint, reconstructed)
			} else {
				t.Logf("PASS: %s (U+%04X) -> High: 0x%04X, Low: 0x%04X",
					e.description, e.codepoint, highSurrogate, lowSurrogate)
			}
		})
	}
}

// TestDynamicToUnicodeCMap verifies that generateToUnicodeCMap correctly generates
// ToUnicode CMaps for both BMP and supplementary plane characters
func TestDynamicToUnicodeCMap(t *testing.T) {
	tests := []struct {
		name          string
		usedRunes     map[int]int
		expectEntries map[int]string
	}{
		{
			name:      "BMP_only",
			usedRunes: map[int]int{1: 0x0041, 2: 0x0042, 3: 0x0043},
			expectEntries: map[int]string{
				1: "0041",
				2: "0042",
				3: "0043",
			},
		},
		{
			name:      "supplementary_plane_only",
			usedRunes: map[int]int{1: 0x1F600, 2: 0x1F601, 3: 0x1F602},
			expectEntries: map[int]string{
				1: "D83DDE00",
				2: "D83DDE01",
				3: "D83DDE02",
			},
		},
		{
			name:      "mixed",
			usedRunes: map[int]int{1: 0x0041, 2: 0x0042, 3: 0x1F600, 4: 0x1F680},
			expectEntries: map[int]string{
				1: "0041",
				2: "0042",
				3: "D83DDE00",
				4: "D83DDE80",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmap := generateToUnicodeCMap(tt.usedRunes)
			if len(cmap) == 0 {
				t.Fatal("Generated CMap is empty")
			}
			if !strings.Contains(cmap, "<0000> <FFFF>") {
				t.Errorf("expected 2-byte codespace range in CMap: %s", cmap)
			}
			if !strings.Contains(cmap, "beginbfchar") {
				t.Errorf("expected bfchar entries in CMap: %s", cmap)
			}
			for cid, hex := range tt.expectEntries {
				pattern := fmt.Sprintf("<%04X> <%s>", cid, hex)
				if !strings.Contains(cmap, pattern) {
					t.Errorf("missing mapping %s", pattern)
				}
			}
		})
	}
}
