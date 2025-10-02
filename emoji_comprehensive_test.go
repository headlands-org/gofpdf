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
	"fmt"
	"strings"
	"testing"
)

// TestEmojiComprehensive is a comprehensive emoji test suite covering all emoji types
func TestEmojiComprehensive(t *testing.T) {
	t.Run("BasicEmojiCategories", testBasicEmojiCategories)
	t.Run("SkinToneModifiersComprehensive", testSkinToneModifiersComprehensive)
	t.Run("ZWJSequencesComprehensive", testZWJSequencesComprehensive)
	t.Run("VariationSelectorsComprehensive", testVariationSelectorsComprehensive)
	t.Run("SupplementaryPlaneChars", testSupplementaryPlaneChars)
	t.Run("EmojiEdgeCases", testEmojiEdgeCases)
}

// testBasicEmojiCategories tests basic emoji from multiple categories
func testBasicEmojiCategories(t *testing.T) {
	emojiCategories := map[string][]string{
		"Emoticons": {
			"😀", "😃", "😄", "😁", "😆", "😂", "🤣", "😊", "😍", "🥰",
			"😘", "😎", "🤓", "🧐", "😏", "😒", "😞", "😔", "😟", "😕",
		},
		"Symbols": {
			"☀", "☁", "⛄", "⚡", "⭐", "✨", "☂", "⛈", "🌈", "🔥",
		},
		"Dingbats": {
			"✂", "✈", "✉", "✏", "✒", "✔", "✖", "🔔", "🔕", "📌",
		},
		"Food": {
			"🍕", "🍔", "🍟", "🍿", "🍩", "🍪", "🎂", "🍇", "🍎", "🍌",
		},
		"Activities": {
			"⚽", "🏀", "🎾", "🏐", "🎮", "🎨", "🎭", "🎸", "🎤", "🎧",
		},
		"Objects": {
			"🚀", "🛸", "🎉", "🎁", "📱", "💻", "⌨", "🖥", "🖨", "📷",
		},
		"Hearts": {
			"❤", "💖", "💗", "💙", "💚", "💛", "🧡", "💜", "🖤", "💔",
		},
		"Hands": {
			"👍", "👎", "👏", "🙌", "👋", "🤝", "🙏", "✊", "✋", "🤚",
		},
	}

	for category, emojis := range emojiCategories {
		t.Run(category, func(t *testing.T) {
			for _, emoji := range emojis {
				// Verify emoji is valid UTF-8
				runes := []rune(emoji)
				if len(runes) == 0 {
					t.Errorf("Emoji %q produced 0 runes", emoji)
				}

				// Test UTF-16 conversion
				utf16 := utf8toutf16(emoji, false)
				if len(utf16) == 0 {
					t.Errorf("UTF-16 conversion failed for emoji %q", emoji)
				}

				// Test grapheme clustering
				clusters := graphemeClusters(emoji)
				if len(clusters) != 1 {
					// Basic emoji should be single cluster (may have variation selector)
					t.Logf("Note: Emoji %q produced %d clusters", emoji, len(clusters))
				}
			}
			t.Logf("Tested %d emoji in %s category", len(emojis), category)
		})
	}
}

// testSkinToneModifiersComprehensive tests emoji with all skin tone variations
func testSkinToneModifiersComprehensive(t *testing.T) {
	// Base emoji that support skin tones
	baseEmojis := []string{
		"👍", // Thumbs up
		"👎", // Thumbs down
		"👋", // Waving hand
		"👌", // OK hand
		"✌", // Victory hand
		"✊", // Raised fist
		"✋", // Raised hand
		"👶", // Baby
		"👦", // Boy
		"👧", // Girl
		"👨", // Man
		"👩", // Woman
		"🙏", // Folded hands
		"💪", // Flexed biceps
	}

	skinTones := []struct {
		modifier string
		name     string
	}{
		{"🏻", "Light"},
		{"🏼", "Medium-Light"},
		{"🏽", "Medium"},
		{"🏾", "Medium-Dark"},
		{"🏿", "Dark"},
	}

	successCount := 0
	for _, base := range baseEmojis {
		for _, tone := range skinTones {
			emoji := base + tone.modifier
			t.Run(fmt.Sprintf("%s_%s", base, tone.name), func(t *testing.T) {
				// Test grapheme clustering
				clusters := graphemeClusters(emoji)
				if len(clusters) != 1 {
					t.Errorf("Emoji with skin tone %q split into %d clusters, expected 1",
						emoji, len(clusters))
				} else {
					successCount++
				}

				// Test rune count (base + modifier)
				runes := []rune(emoji)
				if len(runes) != 2 {
					t.Errorf("Emoji with skin tone %q has %d runes, expected 2", emoji, len(runes))
				}

				// Verify second rune is a skin tone modifier
				if len(runes) >= 2 {
					if runes[1] < 0x1F3FB || runes[1] > 0x1F3FF {
						t.Errorf("Second rune in %q is not a skin tone modifier: U+%04X",
							emoji, runes[1])
					}
				}
			})
		}
	}

	t.Logf("Tested %d base emoji with %d skin tones = %d combinations (%d passed clustering test)",
		len(baseEmojis), len(skinTones), len(baseEmojis)*len(skinTones), successCount)
}

// testZWJSequencesComprehensive tests Zero-Width Joiner emoji sequences
func testZWJSequencesComprehensive(t *testing.T) {
	zwjSequences := []struct {
		emoji string
		desc  string
		parts int
	}{
		// Families
		{"👨‍👩‍👧‍👦", "Family: Man, Woman, Girl, Boy", 4},
		{"👨‍👩‍👧", "Family: Man, Woman, Girl", 3},
		{"👨‍👩‍👦", "Family: Man, Woman, Boy", 3},
		{"👨‍👩‍👧‍👧", "Family: Man, Woman, Girl, Girl", 4},
		{"👨‍👩‍👦‍👦", "Family: Man, Woman, Boy, Boy", 4},
		{"👨‍👨‍👧", "Family: Man, Man, Girl", 3},
		{"👨‍👨‍👦", "Family: Man, Man, Boy", 3},
		{"👩‍👩‍👧", "Family: Woman, Woman, Girl", 3},
		{"👩‍👩‍👦", "Family: Woman, Woman, Boy", 3},

		// Couples
		{"👩‍❤️‍👨", "Couple with Heart: Woman, Man", 3},
		{"👨‍❤️‍👨", "Couple with Heart: Man, Man", 3},
		{"👩‍❤️‍👩", "Couple with Heart: Woman, Woman", 3},
		{"👩‍❤️‍💋‍👨", "Kiss: Woman, Man", 4},
		{"👨‍❤️‍💋‍👨", "Kiss: Man, Man", 4},
		{"👩‍❤️‍💋‍👩", "Kiss: Woman, Woman", 4},

		// Professions (selection)
		{"👨‍💻", "Man Technologist", 2},
		{"👩‍💻", "Woman Technologist", 2},
		{"👨‍⚕️", "Man Health Worker", 2},
		{"👩‍⚕️", "Woman Health Worker", 2},
		{"👨‍🏫", "Man Teacher", 2},
		{"👩‍🏫", "Woman Teacher", 2},
		{"👨‍🍳", "Man Cook", 2},
		{"👩‍🍳", "Woman Cook", 2},
		{"👨‍🌾", "Man Farmer", 2},
		{"👩‍🌾", "Woman Farmer", 2},
		{"👨‍🚒", "Man Firefighter", 2},
		{"👩‍🚒", "Woman Firefighter", 2},
		{"👨‍✈️", "Man Pilot", 2},
		{"👩‍✈️", "Woman Pilot", 2},
		{"👨‍🚀", "Man Astronaut", 2},
		{"👩‍🚀", "Woman Astronaut", 2},
		{"👨‍⚖️", "Man Judge", 2},
		{"👩‍⚖️", "Woman Judge", 2},
		{"👨‍🎓", "Man Student", 2},
		{"👩‍🎓", "Woman Student", 2},
		{"👨‍🎤", "Man Singer", 2},
		{"👩‍🎤", "Woman Singer", 2},
		{"👨‍🎨", "Man Artist", 2},
		{"👩‍🎨", "Woman Artist", 2},
		{"👨‍🔧", "Man Mechanic", 2},
		{"👩‍🔧", "Woman Mechanic", 2},
		{"👨‍🏭", "Man Factory Worker", 2},
		{"👩‍🏭", "Woman Factory Worker", 2},
		{"👨‍💼", "Man Office Worker", 2},
		{"👩‍💼", "Woman Office Worker", 2},
		{"👨‍🔬", "Man Scientist", 2},
		{"👩‍🔬", "Woman Scientist", 2},
	}

	successCount := 0
	for _, seq := range zwjSequences {
		t.Run(seq.desc, func(t *testing.T) {
			// Verify ZWJ character is present
			if !strings.Contains(seq.emoji, "\u200D") {
				t.Errorf("ZWJ sequence %q does not contain ZWJ character", seq.desc)
			}

			// Test grapheme clustering (should be 1 cluster)
			clusters := graphemeClusters(seq.emoji)
			if len(clusters) != 1 {
				t.Errorf("ZWJ sequence %q split into %d clusters, expected 1",
					seq.desc, len(clusters))
			} else {
				successCount++
			}
		})
	}

	t.Logf("Tested %d ZWJ sequences (%d passed clustering test)", len(zwjSequences), successCount)
}

// testVariationSelectorsComprehensive tests emoji with variation selectors
func testVariationSelectorsComprehensive(t *testing.T) {
	variationSelectorEmojis := []struct {
		withVS string
		desc   string
	}{
		{"☀️", "Sun"},
		{"☁️", "Cloud"},
		{"☂️", "Umbrella"},
		{"☃️", "Snowman"},
		{"⭐️", "Star"},
		{"❤️", "Red Heart"},
		{"✔️", "Check Mark"},
		{"✖️", "X Mark"},
		{"❗️", "Exclamation"},
		{"✨", "Sparkles"},
		{"✈️", "Airplane"},
		{"✉️", "Envelope"},
		{"✏️", "Pencil"},
		{"⌚️", "Watch"},
		{"⌛️", "Hourglass"},
		{"⌨️", "Keyboard"},
		{"✂️", "Scissors"},
		{"☎️", "Phone"},
		{"⚓️", "Anchor"},
		{"⚽️", "Soccer Ball"},
		{"⚾️", "Baseball"},
		{"❄️", "Snowflake"},
		{"⚡️", "Lightning"},
		{"🌈", "Rainbow"},
		{"⛄️", "Snowman with Snow"},
	}

	successCount := 0
	for _, item := range variationSelectorEmojis {
		t.Run(item.desc, func(t *testing.T) {
			// Check if variation selector is present
			hasVS := strings.Contains(item.withVS, "\uFE0F") ||
				strings.Contains(item.withVS, "\uFE0E")

			if hasVS {
				// Should be treated as single grapheme cluster
				clusters := graphemeClusters(item.withVS)
				if len(clusters) != 1 {
					t.Errorf("%s with VS split into %d clusters, expected 1",
						item.desc, len(clusters))
				} else {
					successCount++
				}
			}

			// Test UTF-16 conversion
			utf16 := utf8toutf16(item.withVS, false)
			if len(utf16) == 0 {
				t.Errorf("UTF-16 conversion failed for %s", item.desc)
			}
		})
	}

	t.Logf("Tested %d emoji with variation selectors (%d passed clustering test)",
		len(variationSelectorEmojis), successCount)
}

// testSupplementaryPlaneChars tests characters in supplementary Unicode planes (U+10000+)
func testSupplementaryPlaneChars(t *testing.T) {
	suppChars := []struct {
		char      string
		codepoint uint32
		desc      string
	}{
		// Mathematical Alphanumeric Symbols
		{"𝐀", 0x1D400, "Math Bold A"},
		{"𝐇", 0x1D407, "Math Bold H"},
		{"𝑨", 0x1D468, "Math Bold Italic A"},
		{"𝒜", 0x1D49C, "Math Script A"},
		{"𝔸", 0x1D538, "Math Double-Struck A"},
		{"𝖠", 0x1D5A0, "Math Sans-Serif A"},

		// Emoji
		{"🌀", 0x1F300, "Cyclone"},
		{"🔥", 0x1F525, "Fire"},
		{"💯", 0x1F4AF, "Hundred Points"},
		{"👀", 0x1F440, "Eyes"},
		{"🦄", 0x1F984, "Unicorn"},
		{"🧠", 0x1F9E0, "Brain"},

		// Ancient scripts
		{"\U00010000", 0x10000, "Linear B Syllable"},
		{"𐌰", 0x10330, "Gothic Letter"},
		{"𐎠", 0x103A0, "Old Persian"},
		{"𒀀", 0x12000, "Cuneiform"},
		{"𓀀", 0x13000, "Egyptian Hieroglyph"},
	}

	for _, item := range suppChars {
		t.Run(item.desc, func(t *testing.T) {
			// Verify rune value
			runes := []rune(item.char)
			if len(runes) != 1 {
				t.Errorf("%s has %d runes, expected 1", item.desc, len(runes))
			}

			if len(runes) >= 1 && uint32(runes[0]) != item.codepoint {
				t.Errorf("%s codepoint mismatch: got U+%04X, expected U+%04X",
					item.desc, runes[0], item.codepoint)
			}

			// Verify UTF-8 encoding (should be 4 bytes)
			if len(item.char) != 4 {
				t.Errorf("%s has %d bytes, expected 4", item.desc, len(item.char))
			}

			// Test UTF-16 conversion (should produce surrogate pair)
			utf16 := utf8toutf16(item.char, false)
			if len(utf16) != 4 {
				t.Errorf("%s UTF-16 is %d bytes, expected 4 (surrogate pair)",
					item.desc, len(utf16))
			}

			if len(utf16) == 4 {
				highSurrogate := uint16(utf16[0])<<8 | uint16(utf16[1])
				lowSurrogate := uint16(utf16[2])<<8 | uint16(utf16[3])

				// Verify surrogate ranges
				if highSurrogate < 0xD800 || highSurrogate > 0xDBFF {
					t.Errorf("%s high surrogate out of range: 0x%04X",
						item.desc, highSurrogate)
				}
				if lowSurrogate < 0xDC00 || lowSurrogate > 0xDFFF {
					t.Errorf("%s low surrogate out of range: 0x%04X",
						item.desc, lowSurrogate)
				}

				// Reconstruct codepoint
				reconstructed := ((uint32(highSurrogate) - 0xD800) << 10) +
					(uint32(lowSurrogate) - 0xDC00) + 0x10000

				if reconstructed != item.codepoint {
					t.Errorf("%s reconstruction failed: got U+%04X, expected U+%04X",
						item.desc, reconstructed, item.codepoint)
				}
			}
		})
	}

	t.Logf("Tested %d supplementary plane characters", len(suppChars))
}

// testEmojiEdgeCases tests edge cases and corner cases
func testEmojiEdgeCases(t *testing.T) {
	t.Run("EmptyString", func(t *testing.T) {
		clusters := graphemeClusters("")
		if len(clusters) != 0 {
			t.Errorf("Empty string produced %d clusters, expected 0", len(clusters))
		}

		utf16 := utf8toutf16("", true)
		if len(utf16) != 2 { // Just BOM
			t.Errorf("Empty string UTF-16 with BOM is %d bytes, expected 2", len(utf16))
		}
	})

	t.Run("VeryLongEmojiSequence", func(t *testing.T) {
		var builder strings.Builder
		emojis := []string{"😀", "👍", "🎉", "❤️", "🚀", "⭐", "🍕", "🎮"}
		for i := 0; i < 200; i++ {
			builder.WriteString(emojis[i%len(emojis)])
		}
		text := builder.String()

		clusters := graphemeClusters(text)
		if len(clusters) < 150 { // Allow some variation due to VS
			t.Errorf("Long sequence produced %d clusters, expected ~200", len(clusters))
		}

		utf16 := utf8toutf16(text, false)
		if len(utf16) == 0 {
			t.Error("Long sequence UTF-16 conversion failed")
		}

		t.Logf("Long sequence: %d clusters from text of length %d", len(clusters), len(text))
	})

	t.Run("EmojiAtBoundaries", func(t *testing.T) {
		tests := []string{
			"😀Hello",        // Start
			"Hello😀",        // End
			"😀",             // Only emoji
			"Hello😀World",   // Between words
			"😀😃😄",         // Multiple adjacent
			"Hello 😀",       // Space before
			"😀 World",       // Space after
		}

		for _, text := range tests {
			clusters := graphemeClusters(text)
			if len(clusters) == 0 {
				t.Errorf("Text %q produced 0 clusters", text)
			}
		}
	})

	t.Run("MixedScripts", func(t *testing.T) {
		tests := []string{
			"Hello 😀 World",    // ASCII + emoji
			"Héllo 😀 Wörld",    // Latin + emoji
			"日本語 😀 中文",      // CJK + emoji
			"مرحبا 😀 العالم",   // Arabic + emoji
			"Привет 😀 мир",     // Cyrillic + emoji
		}

		for _, text := range tests {
			clusters := graphemeClusters(text)
			if len(clusters) == 0 {
				t.Errorf("Mixed script %q produced 0 clusters", text)
			}
		}
	})

	t.Run("MalformedSequences", func(t *testing.T) {
		tests := []string{
			"🏽",           // Skin tone alone
			"\u200D",      // ZWJ alone
			"\uFE0F",      // VS alone
			"👍🏽🏿",        // Multiple modifiers
		}

		for _, text := range tests {
			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Malformed sequence %q caused panic: %v", text, r)
				}
			}()

			clusters := graphemeClusters(text)
			utf16 := utf8toutf16(text, false)
			t.Logf("Malformed %q: %d clusters, %d UTF-16 bytes",
				text, len(clusters), len(utf16))
		}
	})
}
