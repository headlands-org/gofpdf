package gofpdf

import (
	"testing"
)

// TestGraphemeClusters_BasicEmoji tests grapheme cluster splitting with basic emoji
func TestGraphemeClusters_BasicEmoji(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single grinning face emoji",
			input:    "😀",
			expected: []string{"😀"},
		},
		{
			name:     "multiple basic emoji",
			input:    "😀😃😄",
			expected: []string{"😀", "😃", "😄"},
		},
		{
			name:     "text and emoji mixed",
			input:    "Hello😀World",
			expected: []string{"H", "e", "l", "l", "o", "😀", "W", "o", "r", "l", "d"},
		},
		{
			name:     "plain ASCII text",
			input:    "Hello",
			expected: []string{"H", "e", "l", "l", "o"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := graphemeClusters(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("graphemeClusters() returned %d clusters, expected %d", len(result), len(tt.expected))
				return
			}
			for i, cluster := range result {
				if cluster != tt.expected[i] {
					t.Errorf("graphemeClusters()[%d] = %q, expected %q", i, cluster, tt.expected[i])
				}
			}
		})
	}
}

// TestGraphemeClusters_SkinToneModifiers tests emoji with skin tone modifiers
func TestGraphemeClusters_SkinToneModifiers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "thumbs up with medium skin tone",
			input:    "👍🏽",
			expected: []string{"👍🏽"},
		},
		{
			name:     "thumbs up with light skin tone",
			input:    "👍🏻",
			expected: []string{"👍🏻"},
		},
		{
			name:     "thumbs up with dark skin tone",
			input:    "👍🏿",
			expected: []string{"👍🏿"},
		},
		{
			name:     "waving hand with medium-light skin tone",
			input:    "👋🏼",
			expected: []string{"👋🏼"},
		},
		{
			name:     "multiple emoji with skin tones",
			input:    "👍🏽👋🏼",
			expected: []string{"👍🏽", "👋🏼"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := graphemeClusters(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("graphemeClusters() returned %d clusters, expected %d", len(result), len(tt.expected))
				t.Logf("Result: %v", result)
				t.Logf("Expected: %v", tt.expected)
				return
			}
			for i, cluster := range result {
				if cluster != tt.expected[i] {
					t.Errorf("graphemeClusters()[%d] = %q, expected %q", i, cluster, tt.expected[i])
				}
			}
		})
	}
}

// TestGraphemeClusters_ZWJSequences tests zero-width joiner sequences
func TestGraphemeClusters_ZWJSequences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int // expected number of clusters
	}{
		{
			name:     "family emoji (man-woman-girl-boy)",
			input:    "👨‍👩‍👧‍👦",
			expected: 1, // Should be treated as single cluster
		},
		{
			name:     "woman health worker",
			input:    "👩‍⚕️",
			expected: 1, // Should be treated as single cluster
		},
		{
			name:     "man technologist",
			input:    "👨‍💻",
			expected: 1, // Should be treated as single cluster
		},
		{
			name:     "multiple ZWJ sequences",
			input:    "👨‍👩‍👧‍👦👨‍💻",
			expected: 2, // Two separate clusters
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := graphemeClusters(tt.input)
			if len(result) != tt.expected {
				t.Errorf("graphemeClusters() returned %d clusters, expected %d", len(result), tt.expected)
				t.Logf("Result: %v", result)
			}
		})
	}
}

// TestGraphemeClusters_VariationSelectors tests emoji with variation selectors
func TestGraphemeClusters_VariationSelectors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int // expected number of clusters
	}{
		{
			name:     "sun with variation selector",
			input:    "☀️",
			expected: 1, // Should be treated as single cluster
		},
		{
			name:     "heart with variation selector",
			input:    "❤️",
			expected: 1, // Should be treated as single cluster
		},
		{
			name:     "checkmark with variation selector",
			input:    "✔️",
			expected: 1, // Should be treated as single cluster
		},
		{
			name:     "multiple with variation selectors",
			input:    "☀️❤️✔️",
			expected: 3, // Three separate clusters
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := graphemeClusters(tt.input)
			if len(result) != tt.expected {
				t.Errorf("graphemeClusters() returned %d clusters, expected %d", len(result), tt.expected)
				t.Logf("Result: %v", result)
			}
		})
	}
}

// TestIsEmoji tests the emoji detection function
func TestIsEmoji(t *testing.T) {
	tests := []struct {
		name     string
		input    rune
		expected bool
	}{
		// Main emoji range (U+1F300 - U+1F9FF)
		{name: "grinning face", input: '😀', expected: true},     // U+1F600
		{name: "thumbs up", input: '👍', expected: true},         // U+1F44D
		{name: "red heart", input: '❤', expected: true},          // U+2764 (in misc symbols range)
		{name: "pizza", input: '🍕', expected: true},             // U+1F355
		{name: "rocket", input: '🚀', expected: true},            // U+1F680
		{name: "smiling cat", input: '😺', expected: true},       // U+1F63A

		// Miscellaneous Symbols (U+2600 - U+26FF)
		{name: "sun", input: '☀', expected: true},                // U+2600
		{name: "cloud", input: '☁', expected: true},              // U+2601
		{name: "umbrella", input: '☂', expected: true},           // U+2602
		{name: "star", input: '★', expected: true},               // U+2605

		// Dingbats (U+2700 - U+27BF)
		{name: "scissors", input: '✂', expected: true},           // U+2702
		{name: "checkmark", input: '✓', expected: true},          // U+2713
		{name: "cross mark", input: '✗', expected: true},         // U+2717

		// Non-emoji characters
		{name: "Latin A", input: 'A', expected: false},           // U+0041
		{name: "digit 5", input: '5', expected: false},           // U+0035
		{name: "space", input: ' ', expected: false},             // U+0020
		{name: "Chinese char", input: '中', expected: false},     // U+4E2D
		{name: "Greek alpha", input: 'α', expected: false},       // U+03B1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEmoji(tt.input)
			if result != tt.expected {
				t.Errorf("isEmoji(%U) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestGraphemeClusterWidth tests width calculation for grapheme clusters
func TestGraphemeClusterWidth(t *testing.T) {
	// Create a mock font with character width map
	mockFont := &fontDefType{
		Cw: map[int]int{
			int('A'):    722,  // Width for 'A'
			int('H'):    722,  // Width for 'H'
			int('e'):    444,  // Width for 'e'
			int('l'):    278,  // Width for 'l'
			int('o'):    500,  // Width for 'o'
			0x1F600:     1000, // Grinning face emoji
			0x1F44D:     1000, // Thumbs up
			0x1F468:     1000, // Man
			0x2600:      800,  // Sun
		},
	}

	tests := []struct {
		name     string
		cluster  string
		expected int
	}{
		{
			name:     "single ASCII character",
			cluster:  "A",
			expected: 722,
		},
		{
			name:     "basic emoji",
			cluster:  "😀",
			expected: 1000,
		},
		{
			name:     "emoji with skin tone modifier",
			cluster:  "👍🏽",
			expected: 1000, // Width of base character only
		},
		{
			name:     "ZWJ sequence (family)",
			cluster:  "👨‍👩‍👧‍👦",
			expected: 1000, // Width of first character (man)
		},
		{
			name:     "sun with variation selector",
			cluster:  "☀️",
			expected: 800, // Width of base character
		},
		{
			name:     "empty cluster",
			cluster:  "",
			expected: 0,
		},
		{
			name:     "character not in font",
			cluster:  "中",
			expected: 0, // Not in mock font's Cw map
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := graphemeClusterWidth(tt.cluster, mockFont)
			if result != tt.expected {
				t.Errorf("graphemeClusterWidth(%q, font) = %d, expected %d", tt.cluster, result, tt.expected)
			}
		})
	}
}

// TestGraphemeClusterWidth_NilFont tests that the function handles edge cases gracefully
func TestGraphemeClusterWidth_EdgeCases(t *testing.T) {
	emptyFont := &fontDefType{
		Cw: map[int]int{},
	}

	tests := []struct {
		name     string
		cluster  string
		font     *fontDefType
		expected int
	}{
		{
			name:     "empty font map",
			cluster:  "A",
			font:     emptyFont,
			expected: 0,
		},
		{
			name:     "empty cluster with empty font",
			cluster:  "",
			font:     emptyFont,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := graphemeClusterWidth(tt.cluster, tt.font)
			if result != tt.expected {
				t.Errorf("graphemeClusterWidth(%q, font) = %d, expected %d", tt.cluster, result, tt.expected)
			}
		})
	}
}
