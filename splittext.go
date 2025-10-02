package gofpdf

import (
	"math"
	//	"strings"
	"unicode"
)

// SplitText splits UTF-8 encoded text into several lines using the current
// font. Each line has its length limited to a maximum width given by w. This
// function can be used to determine the total height of wrapped text for
// vertical placement purposes.
//
// This function is grapheme-cluster aware, meaning it will not split emoji
// sequences (e.g., "👍🏽" or "👨‍👩‍👧‍👦") across lines. Text is split at grapheme
// cluster boundaries, ensuring that user-perceived characters remain intact.
func (f *Fpdf) SplitText(txt string, w float64) (lines []string) {
	cw := f.currentFont.Cw
	wmax := int(math.Ceil((w - 2*f.cMargin) * 1000 / f.fontSize))

	// Split into grapheme clusters instead of runes
	clusters := graphemeClusters(txt)
	nb := len(clusters)

	// Remove trailing newline clusters
	for nb > 0 && clusters[nb-1] == "\n" {
		nb--
	}
	clusters = clusters[0:nb]

	sep := -1
	i := 0
	j := 0
	l := 0

	for i < nb {
		cluster := clusters[i]

		// Calculate cluster width
		clusterWidth := 0
		for _, r := range cluster {
			clusterWidth += cw[int(r)]
		}
		l += clusterWidth

		// Check if we can break at this position
		// We can break at spaces or after Chinese characters
		if len(cluster) == 1 {
			r := []rune(cluster)[0]
			if unicode.IsSpace(r) || isChinese(r) {
				sep = i
			}
		}

		// Check for explicit newline or width limit
		if cluster == "\n" || l > wmax {
			if sep == -1 {
				if i == j {
					i++
				}
				sep = i
			} else {
				i = sep + 1
			}
			// Join clusters back into a string for this line
			var lineBuilder []string
			for k := j; k < sep; k++ {
				lineBuilder = append(lineBuilder, clusters[k])
			}
			lines = append(lines, joinClusters(lineBuilder))
			sep = -1
			j = i
			l = 0
		} else {
			i++
		}
	}

	// Add remaining text as final line
	if i != j {
		var lineBuilder []string
		for k := j; k < i; k++ {
			lineBuilder = append(lineBuilder, clusters[k])
		}
		lines = append(lines, joinClusters(lineBuilder))
	}

	return lines
}

// joinClusters joins a slice of grapheme clusters into a single string
func joinClusters(clusters []string) string {
	result := ""
	for _, cluster := range clusters {
		result += cluster
	}
	return result
}
