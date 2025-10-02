// Package gofpdf provides PDF document generation capabilities.
// This file contains grapheme cluster utilities for proper handling of
// multi-codepoint Unicode sequences, including emoji with modifiers and
// zero-width joiner (ZWJ) sequences, ensuring correct text rendering and
// width calculations in PDF documents.
package gofpdf

import (
	"github.com/rivo/uniseg"
)

// graphemeClusters splits a string into grapheme clusters using the uniseg library.
// A grapheme cluster is a user-perceived character, which may consist of multiple
// Unicode codepoints. This is particularly important for handling:
//   - Emoji with skin tone modifiers (e.g., "👍🏽" = base + modifier)
//   - Zero-width joiner sequences (e.g., "👨‍👩‍👧‍👦" = family emoji)
//   - Variation selectors (e.g., "☀️" = sun + variation selector)
//   - Combining characters and diacritics
//
// Examples:
//   - "Hello" → ["H", "e", "l", "l", "o"] (5 clusters)
//   - "👍🏽" → ["👍🏽"] (1 cluster: base + modifier)
//   - "👨‍👩‍👧‍👦" → ["👨‍👩‍👧‍👦"] (1 cluster: 4 people joined with ZWJ)
//
// Parameters:
//   s - The input string to split into grapheme clusters
//
// Returns:
//   A slice of strings, where each string represents one grapheme cluster
func graphemeClusters(s string) []string {
	var clusters []string
	gr := uniseg.NewGraphemes(s)

	for gr.Next() {
		clusters = append(clusters, gr.Str())
	}

	return clusters
}

// graphemeClusterWidth calculates the display width of a grapheme cluster
// in font units (typically 1/1000th of the font size).
//
// The width is determined by the base character of the cluster. Modifiers,
// zero-width joiners, and variation selectors do not add additional width
// as they are rendered as part of the cluster.
//
// Parameters:
//   cluster - A single grapheme cluster (string that may contain multiple codepoints)
//   font    - The font definition containing character width information
//
// Returns:
//   The width in font units (typically in 1000ths). Returns 0 if the cluster
//   is empty or if the base character is not found in the font's width map.
//
// Implementation note:
//   The function uses the first rune of the cluster as the base character
//   for width lookup in the font's Cw (character width) map.
func graphemeClusterWidth(cluster string, font *fontDefType) int {
	if len(cluster) == 0 {
		return 0
	}

	// Get the first rune (base character) of the cluster
	baseRune := []rune(cluster)[0]

	// Look up the width in the font's character width map
	if width, ok := font.Cw[int(baseRune)]; ok {
		return width
	}

	// Return 0 if the character is not found in the font
	return 0
}

// isEmoji checks if a rune represents an emoji codepoint.
// This function covers the most common emoji Unicode ranges:
//
//   - U+1F300 - U+1F9FF: Main emoji range (includes faces, animals, food, etc.)
//   - U+2600  - U+26FF:  Miscellaneous Symbols (sun, moon, stars, etc.)
//   - U+2700  - U+27BF:  Dingbats (scissors, checkmarks, etc.)
//
// Note: This is a simplified check that covers most common emoji.
// Some emoji may use variation selectors (U+FE0F) or skin tone modifiers
// (U+1F3FB - U+1F3FF) which are typically part of a grapheme cluster.
//
// Parameters:
//   r - The rune to check
//
// Returns:
//   true if the rune is in one of the emoji Unicode ranges, false otherwise
//
// Examples:
//   isEmoji('😀') → true  (U+1F600, grinning face)
//   isEmoji('☀')  → true  (U+2600, sun)
//   isEmoji('✂')  → true  (U+2702, scissors)
//   isEmoji('A')  → false (U+0041, Latin letter)
func isEmoji(r rune) bool {
	return (r >= 0x1F300 && r <= 0x1F9FF) || // Main emoji range
		(r >= 0x2600 && r <= 0x26FF) || // Miscellaneous Symbols
		(r >= 0x2700 && r <= 0x27BF) // Dingbats
}
