# Emoji Migration Guide for gofpdf

## Table of Contents
- [Introduction](#introduction)
- [Prerequisites](#prerequisites)
- [Font Installation](#font-installation)
- [Basic Usage](#basic-usage)
- [Modern Emoji Support](#modern-emoji-support)
- [Advanced Usage](#advanced-usage)
- [Migration Guide](#migration-guide)
- [Understanding Unicode Support](#understanding-unicode-support)
- [Compatible Emoji Fonts](#compatible-emoji-fonts)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [Unicode Reference](#unicode-reference)

---

## Introduction

gofpdf now supports emoji and extended Unicode characters through proper grapheme cluster handling and full 4-byte UTF-8 sequence support. This guide will help you integrate emoji into your existing PDF generation workflows.

### What's Supported
- **Full Unicode emoji support**: U+0000 to U+10FFFF (includes all modern emoji)
- **BMP emoji**: U+2000-U+2FFF range (☀ ☁ ❤ ⭐)
- **Modern emoji**: U+1F300-U+1F9FF range (😀 🎉 🚀 💯)
- **Grapheme clusters**: Emoji with skin tone modifiers, variation selectors
- **ZWJ (Zero-Width Joiner) sequences**: Family emoji, flag sequences
- **Monochrome emoji rendering** via TrueType fonts

### What's Not Supported (Yet)
- Color emoji (requires Type 3 fonts or image embedding)
- Complex script shaping (Arabic ligatures, Indic scripts)

---

## Prerequisites

### System Requirements
- **Go version**: 1.13+ (for module support)
- **gofpdf version**: Latest version with emoji support
- **Dependencies**: `github.com/rivo/uniseg` (automatically installed)

### Font Requirements
You need a TrueType font that includes emoji glyphs **with CMAP Format 12 support** for modern emoji. The recommended fonts are:
- **Noto Emoji** (recommended): Comprehensive, open-source, monochrome, **Format 12 support**
- **Symbola**: Good Unicode coverage, includes many symbols, **Format 12 support**
- **Unifont**: Complete BMP coverage, bitmap style, Format 12 support

**Important**: For modern emoji (😀 🎉 🚀) to work, your font must have a CMAP Format 12 table. Fonts with only Format 4 will be limited to BMP emoji (☀ ☁ ❤).

---

## Font Installation

### Linux

#### Ubuntu/Debian (APT)
```bash
# Install Noto Emoji font
sudo apt update
sudo apt install fonts-noto-color-emoji

# Font will be installed to /usr/share/fonts/truetype/noto/
```

#### Fedora/RHEL (DNF)
```bash
# Install Noto Emoji font
sudo dnf install google-noto-emoji-fonts

# Font will be in /usr/share/fonts/google-noto-emoji/
```

#### Arch Linux (Pacman)
```bash
# Install Noto Emoji font
sudo pacman -S noto-fonts-emoji

# Font location: /usr/share/fonts/noto/
```

### macOS

#### Using Homebrew
```bash
# Install Noto Emoji font via Homebrew Cask
brew tap homebrew/cask-fonts
brew install --cask font-noto-emoji

# Font installed to ~/Library/Fonts/ or /Library/Fonts/
```

#### Manual Installation
1. Download Noto Emoji from [Google Fonts](https://fonts.google.com/noto/specimen/Noto+Emoji)
2. Open the downloaded `.ttf` file
3. Click "Install Font" in Font Book
4. Font will be available system-wide

### Windows

#### Manual Installation
1. Download Noto Emoji from [Google Fonts](https://fonts.google.com/noto/specimen/Noto+Emoji)
2. Extract the `.ttf` file
3. Right-click the font file → "Install" or "Install for all users"
4. Font installed to `C:\Windows\fonts\`

### Download Links

**Noto Emoji (Recommended)**
- GitHub: https://github.com/googlefonts/noto-emoji/
- Google Fonts: https://fonts.google.com/noto/specimen/Noto+Emoji
- Direct: https://github.com/googlefonts/noto-emoji/raw/main/fonts/NotoEmoji-Regular.ttf

**Symbola**
- Font Library: https://fontlibrary.org/en/font/symbola
- Note: Development ceased in 2019, but still widely used

**Unifont**
- Official site: http://unifoundry.com/unifont/
- GNU Unifont: Complete BMP coverage

---

## Basic Usage

### Simple Emoji in PDF

```go
package main

import (
    "github.com/phpdave11/gofpdf"
)

func main() {
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()

    // Add emoji font (ensure path is correct for your system)
    pdf.AddUTF8Font("notoemoji", "", "NotoEmoji-Regular.ttf")

    // Set emoji font
    pdf.SetFont("notoemoji", "", 16)

    // Render BMP emoji (work with CMAP Format 4 and Format 12)
    pdf.Cell(0, 10, "Weather: \u2600 \u2601 \u2614 \u26C4") // ☀ ☁ ☂ ⛄
    pdf.Ln(10)

    pdf.Cell(0, 10, "Symbols: \u2764 \u2B50 \u2714 \u2718") // ❤ ⭐ ✔ ✘
    pdf.Ln(10)

    // Render modern emoji (requires CMAP Format 12)
    pdf.Cell(0, 10, "Modern: \U0001F600 \U0001F389 \U0001F680") // 😀 🎉 🚀
    pdf.Ln(10)

    pdf.Cell(0, 10, "More: \U0001F44D \U0001F4AF \U0001F525") // 👍 💯 🔥
    pdf.Ln(10)

    err := pdf.OutputFileAndClose("emoji_basic.pdf")
    if err != nil {
        panic(err)
    }
}
```

### Mixed Content (Text + Emoji)

```go
func mixedContent() {
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()

    // Add both regular and emoji fonts
    pdf.AddUTF8Font("dejavu", "", "DejaVuSansCondensed.ttf")
    pdf.AddUTF8Font("notoemoji", "", "NotoEmoji-Regular.ttf")

    // Regular text
    pdf.SetFont("dejavu", "", 12)
    pdf.Cell(40, 10, "Status: ")

    // Switch to emoji font for checkmark
    pdf.SetFont("notoemoji", "", 12)
    pdf.Cell(10, 10, "\u2714") // ✔

    // Back to regular font
    pdf.SetFont("dejavu", "", 12)
    pdf.Cell(0, 10, " Complete")

    pdf.OutputFileAndClose("mixed_content.pdf")
}
```

---

## Modern Emoji Support

### What Changed: CMAP Format 12

gofpdf now supports **CMAP Format 12**, enabling the full Unicode range (U+0000 to U+10FFFF). This unlocks modern emoji that were previously unavailable.

#### Technical Overview

**CMAP Format 4 (Legacy)**
- Supports: Basic Multilingual Plane only (U+0000-U+FFFF)
- Character encoding: 16-bit
- Emoji range: Limited to ☀ ☁ ❤ ⭐ and similar BMP symbols
- Still supported for backward compatibility

**CMAP Format 12 (Modern)**
- Supports: Full Unicode range (U+0000-U+10FFFF)
- Character encoding: 32-bit
- Emoji range: All modern emoji including 😀 🎉 🚀 💯 🔥
- Automatically detected and used when available
- More efficient for supplementary plane characters

**How It Works**
1. gofpdf detects the font's CMAP table format during font loading
2. If Format 12 is present (PlatformID=3, EncodingID=10), it's used
3. If only Format 4 is present, falls back to BMP-only mode
4. ToUnicode CMap is dynamically generated based on actual characters used
5. CIDToGIDMap uses identity mapping for supplementary plane efficiency

### Using Modern Emoji

```go
package main

import (
    "github.com/phpdave11/gofpdf"
)

func modernEmoji() {
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()

    // Add emoji font with Format 12 support (e.g., Noto Emoji)
    pdf.AddUTF8Font("notoemoji", "", "NotoEmoji-Regular.ttf")
    pdf.SetFont("notoemoji", "", 16)

    // Modern emoji faces
    pdf.Cell(0, 10, "\U0001F600 \U0001F603 \U0001F604 \U0001F601") // 😀 😃 😄 😁
    pdf.Ln(8)

    // Celebration emoji
    pdf.Cell(0, 10, "\U0001F389 \U0001F38A \U0001F38B \U0001F38D") // 🎉 🎊 🎋 🎍
    pdf.Ln(8)

    // Objects and symbols
    pdf.Cell(0, 10, "\U0001F680 \U0001F6A0 \U0001F697 \U0001F6B2") // 🚀 🚠 🚗 🚲
    pdf.Ln(8)

    // Popular emoji
    pdf.Cell(0, 10, "\U0001F525 \U0001F4AF \U0001F44D \U0001F44F") // 🔥 💯 👍 👏
    pdf.Ln(8)

    pdf.OutputFileAndClose("modern_emoji.pdf")
}
```

### Emoji with Skin Tone Modifiers

Skin tone modifiers (U+1F3FB-U+1F3FF) work with Format 12:

```go
func emojiWithSkinTones() {
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()
    pdf.AddUTF8Font("notoemoji", "", "NotoEmoji-Regular.ttf")
    pdf.SetFont("notoemoji", "", 16)

    // Thumbs up with different skin tones
    pdf.Cell(0, 10, "Thumbs up:")
    pdf.Ln(8)
    pdf.Cell(0, 10, "\U0001F44D") // 👍 (default)
    pdf.Ln(6)
    pdf.Cell(0, 10, "\U0001F44D\U0001F3FB") // 👍🏻 (light)
    pdf.Ln(6)
    pdf.Cell(0, 10, "\U0001F44D\U0001F3FD") // 👍🏽 (medium)
    pdf.Ln(6)
    pdf.Cell(0, 10, "\U0001F44D\U0001F3FF") // 👍🏿 (dark)

    pdf.OutputFileAndClose("emoji_skin_tones.pdf")
}
```

**Note**: Skin tone modifiers are grapheme clusters (multiple codepoints rendered as one character). gofpdf handles these automatically using the `github.com/rivo/uniseg` library.

### Mixed BMP and Modern Emoji

You can freely mix BMP emoji and modern emoji in the same document:

```go
func mixedEmoji() {
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()
    pdf.AddUTF8Font("notoemoji", "", "NotoEmoji-Regular.ttf")
    pdf.SetFont("notoemoji", "", 14)

    // BMP symbols
    pdf.Write(5, "Classic: \u2600 \u2764 \u2B50 ")
    pdf.Ln(8)

    // Modern emoji
    pdf.Write(5, "Modern: \U0001F600 \U0001F389 \U0001F680 ")
    pdf.Ln(8)

    // Mixed in same line
    pdf.Write(5, "Mixed: \u2714 \U0001F44D \u2718 \U0001F44E")

    pdf.OutputFileAndClose("mixed_emoji.pdf")
}
```

### Checking Font Compatibility

Not all fonts support Format 12. Here's how to verify:

```bash
# Linux: Use ttfdump to check CMAP tables
ttfdump -t cmap NotoEmoji-Regular.ttf | grep "Format 12"

# macOS: Use ftdump (from freetype)
ftdump NotoEmoji-Regular.ttf | grep "format 12"

# Or use online tools: https://fontdrop.info/
# Upload font and check "cmap" table for Format 12 (PlatformID 3, EncodingID 10)
```

**Known Fonts with Format 12 Support:**
- Noto Emoji (recommended) ✓
- Symbola ✓
- Noto Sans ✓
- Noto Serif ✓
- Apple Color Emoji ✓ (but color not supported by gofpdf)
- Segoe UI Emoji ✓ (Windows)

**Fonts Limited to Format 4:**
- DejaVu Sans (BMP only)
- Some older emoji fonts

---

## Advanced Usage

### Grapheme Clusters

Grapheme clusters are user-perceived characters that may consist of multiple Unicode codepoints. gofpdf handles these automatically:

```go
func graphemeClusters() {
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()
    pdf.AddUTF8Font("notoemoji", "", "NotoEmoji-Regular.ttf")
    pdf.SetFont("notoemoji", "", 14)

    // Emoji with variation selector (text vs emoji presentation)
    // ☀ (text) vs ☀️ (emoji with U+FE0F variation selector)
    pdf.Cell(0, 10, "\u2600\uFE0F Sun with emoji presentation")
    pdf.Ln(8)

    // Combining characters are treated as single cluster
    pdf.Cell(0, 10, "Single cluster: \u0041\u0301") // Á (A + combining acute)
    pdf.Ln(8)

    pdf.OutputFileAndClose("grapheme_clusters.pdf")
}
```

**Important**: gofpdf uses the `github.com/rivo/uniseg` library to correctly identify and measure grapheme clusters. This ensures emoji with modifiers are not split across lines.

### Font Switching Strategy

For complex documents with mixed content, use a font switching pattern:

```go
func fontSwitching() {
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()

    pdf.AddUTF8Font("dejavu", "", "DejaVuSansCondensed.ttf")
    pdf.AddUTF8Font("dejavu", "B", "DejaVuSansCondensed-Bold.ttf")
    pdf.AddUTF8Font("notoemoji", "", "NotoEmoji-Regular.ttf")

    // Helper function for inline emoji
    writeWithEmoji := func(text string, emoji string) {
        pdf.SetFont("dejavu", "", 12)
        pdf.Write(5, text+" ")

        pdf.SetFont("notoemoji", "", 12)
        pdf.Write(5, emoji+" ")

        pdf.SetFont("dejavu", "", 12)
    }

    writeWithEmoji("Completed task", "\u2714") // ✔
    pdf.Ln(6)
    writeWithEmoji("Priority alert", "\u26A0") // ⚠
    pdf.Ln(6)
    writeWithEmoji("New message", "\u2709") // ✉

    pdf.OutputFileAndClose("font_switching.pdf")
}
```

### Performance Considerations

For documents with many emoji:

```go
func performanceOptimized() {
    pdf := gofpdf.New("P", "mm", "A4", "")

    // Use font subsetting (default behavior)
    // Only emoji actually used will be embedded in PDF
    pdf.AddUTF8Font("notoemoji", "", "NotoEmoji-Regular.ttf")

    // Batch emoji rendering
    pdf.AddPage()
    pdf.SetFont("notoemoji", "", 10)

    // Render multiple emoji at once
    emojiList := []string{
        "\u2600", "\u2601", "\u2614", "\u26C4",
        "\u2764", "\u2B50", "\u2714", "\u2718",
    }

    for _, emoji := range emojiList {
        pdf.Cell(10, 10, emoji)
    }

    pdf.OutputFileAndClose("performance.pdf")
}
```

---

## Migration Guide

### Before: Regular Text Only

**Old code without emoji:**
```go
func statusReport() {
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()
    pdf.SetFont("Arial", "", 12)

    pdf.Cell(0, 10, "Status: Complete")
    pdf.Ln(8)
    pdf.Cell(0, 10, "Priority: High")
    pdf.Ln(8)
    pdf.Cell(0, 10, "New messages: 5")

    pdf.OutputFileAndClose("status_old.pdf")
}
```

### After: With Emoji Support

**New code with emoji:**
```go
func statusReportWithEmoji() {
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()

    // Add fonts
    pdf.AddUTF8Font("dejavu", "", "DejaVuSansCondensed.ttf")
    pdf.AddUTF8Font("notoemoji", "", "NotoEmoji-Regular.ttf")

    // Status with checkmark
    pdf.SetFont("dejavu", "", 12)
    pdf.Cell(30, 10, "Status: ")
    pdf.SetFont("notoemoji", "", 12)
    pdf.Cell(10, 10, "\u2714") // ✔
    pdf.SetFont("dejavu", "", 12)
    pdf.Cell(0, 10, " Complete")
    pdf.Ln(8)

    // Priority with warning symbol
    pdf.SetFont("dejavu", "", 12)
    pdf.Cell(30, 10, "Priority: ")
    pdf.SetFont("notoemoji", "", 12)
    pdf.Cell(10, 10, "\u26A0") // ⚠
    pdf.SetFont("dejavu", "", 12)
    pdf.Cell(0, 10, " High")
    pdf.Ln(8)

    // Messages with envelope
    pdf.SetFont("dejavu", "", 12)
    pdf.Cell(30, 10, "New messages: ")
    pdf.SetFont("notoemoji", "", 12)
    pdf.Cell(10, 10, "\u2709") // ✉
    pdf.SetFont("dejavu", "", 12)
    pdf.Cell(0, 10, " 5")

    pdf.OutputFileAndClose("status_new.pdf")
}
```

### Migration Steps

1. **Install emoji font** (see [Font Installation](#font-installation))

2. **Add font to your code**:
   ```go
   pdf.AddUTF8Font("notoemoji", "", "path/to/NotoEmoji-Regular.ttf")
   ```

3. **Replace ASCII symbols with Unicode emoji**:
   - `[X]` → `\u2714` (✔)
   - `[ ]` → `\u2610` (☐)
   - `!` → `\u26A0` (⚠)
   - `*` → `\u2B50` (⭐)

4. **Add font switching** where needed:
   ```go
   pdf.SetFont("regular", "", 12)  // Regular text
   pdf.Cell(30, 10, "Label: ")

   pdf.SetFont("notoemoji", "", 12)  // Emoji
   pdf.Cell(10, 10, "\u2714")

   pdf.SetFont("regular", "", 12)  // Back to regular
   pdf.Cell(0, 10, " Text continues")
   ```

5. **Test your output** to ensure emoji render correctly

---

## Understanding Unicode Support

### CMAP Format 4 vs Format 12

gofpdf supports **both CMAP Format 4 and Format 12**, providing comprehensive Unicode coverage:

#### BMP Emoji (Format 4 and Format 12)

These emoji work with both Format 4 and Format 12 fonts:

**Weather & Nature**
- ☀ (U+2600) - Sun
- ☁ (U+2601) - Cloud
- ☂ (U+2602) - Umbrella
- ☃ (U+2603) - Snowman
- ⛄ (U+26C4) - Snowman without snow
- ⚡ (U+26A1) - Lightning

**Symbols**
- ❤ (U+2764) - Red heart
- ⭐ (U+2B50) - Star
- ✔ (U+2714) - Checkmark
- ✘ (U+2718) - X mark
- ⚠ (U+26A0) - Warning
- ⛔ (U+26D4) - No entry

**Dingbats**
- ✂ (U+2702) - Scissors
- ✈ (U+2708) - Airplane
- ✉ (U+2709) - Envelope
- ✏ (U+270F) - Pencil
- ✒ (U+2712) - Pen

**Hands & People (simple)**
- ✊ (U+270A) - Fist
- ✋ (U+270B) - Hand
- ☝ (U+261D) - Pointing up
- ✌ (U+270C) - Victory

#### Modern Emoji (Requires Format 12)

These emoji require fonts with CMAP Format 12 support:

**Emoticons (U+1F600-U+1F64F)**
- 😀 (U+1F600) - Grinning face
- 😃 (U+1F603) - Grinning face with big eyes
- 😄 (U+1F604) - Grinning face with smiling eyes
- 😁 (U+1F601) - Beaming face
- 👍 (U+1F44D) - Thumbs up
- 👎 (U+1F44E) - Thumbs down

**Symbols & Objects (U+1F300-U+1F5FF)**
- 🌍 (U+1F30D) - Earth
- 🔥 (U+1F525) - Fire
- 💯 (U+1F4AF) - Hundred points
- 🚀 (U+1F680) - Rocket
- 🎉 (U+1F389) - Party popper
- ⚡ (U+26A1) - Lightning

**Food & Drink (U+1F300-U+1F5FF)**
- 🍕 (U+1F355) - Pizza
- 🍔 (U+1F354) - Hamburger
- 🍰 (U+1F370) - Shortcake

**Note**: If your font only has Format 4, these emoji won't render. Use fonts like Noto Emoji that include Format 12 support.

### Format Detection and Fallback

gofpdf automatically detects which CMAP format your font supports:

```go
// No code changes needed - automatic detection
pdf.AddUTF8Font("notoemoji", "", "NotoEmoji-Regular.ttf")

// If font has Format 12: All emoji work (BMP + modern)
// If font has Format 4 only: Only BMP emoji work
```

**What happens with Format 4-only fonts:**
1. Modern emoji (U+1F600+) will appear as blank spaces or boxes
2. BMP emoji (☀ ☁ ❤) continue to work normally
3. No errors are generated - graceful degradation

### Workarounds for Format 4-Only Fonts

If you must use a font without Format 12 support:

1. **Use BMP alternatives**: Many concepts have BMP equivalents
   - Instead of 😀 (U+1F600), use ☺ (U+263A) - simple smiley
   - Instead of 🔥 (U+1F525), use ✦ (U+2726) - sparkle
   - Instead of 🚀 (U+1F680), use ✈ (U+2708) - airplane

2. **Upgrade to Format 12 font**: Switch to Noto Emoji or Symbola
   ```go
   // Old font (Format 4 only)
   // pdf.AddUTF8Font("dejavu", "", "DejaVuSans.ttf")

   // New font (Format 12 support)
   pdf.AddUTF8Font("notoemoji", "", "NotoEmoji-Regular.ttf")
   ```

3. **Image embedding**: For critical color emoji, embed as images
   ```go
   pdf.ImageOptions("emoji_fire.png", 10, 10, 5, 5, false,
       gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")
   ```

4. **Unicode symbols**: Rich set of symbols in BMP range
   - Arrows: ← → ↑ ↓ (U+2190-U+2193)
   - Mathematical: ∞ ≈ ≠ ± (U+221E, U+2248, U+2260, U+00B1)
   - Currency: € £ ¥ (U+20AC, U+00A3, U+00A5)

---

## Compatible Emoji Fonts

### Recommended Fonts

#### 1. Noto Emoji (Best Overall)
- **Provider**: Google
- **License**: Open Font License (OFL)
- **Coverage**: Comprehensive emoji coverage (BMP + supplementary plane)
- **CMAP Format**: Format 12 ✓ (full Unicode support)
- **Style**: Monochrome (black & white)
- **Size**: ~400KB
- **Download**: https://github.com/googlefonts/noto-emoji/
- **Best for**: Production use, modern emoji, consistent rendering across platforms
- **Modern Emoji**: YES - 😀 🎉 🚀 all work

#### 2. Symbola (Good Unicode Coverage)
- **Provider**: George Douros
- **License**: Freeware
- **Coverage**: Excellent symbol and emoji coverage
- **CMAP Format**: Format 12 ✓ (full Unicode support)
- **Style**: Simple, clean design
- **Size**: ~3MB
- **Download**: https://fontlibrary.org/en/font/symbola
- **Best for**: Documents needing wide Unicode symbol support
- **Modern Emoji**: YES - 😀 🎉 🚀 all work
- **Note**: Development ceased 2019, but still widely used

#### 3. Unifont (Complete Unicode)
- **Provider**: GNU Project
- **License**: GNU GPL v2+ with font embedding exception
- **Coverage**: Complete Unicode (BMP + supplementary plane)
- **CMAP Format**: Format 12 ✓ (full Unicode support)
- **Style**: Bitmap/monospace style
- **Size**: ~12MB
- **Download**: http://unifoundry.com/unifont/
- **Best for**: Technical documents, complete Unicode coverage
- **Modern Emoji**: YES - 😀 🎉 🚀 all work

#### 4. DejaVu Sans (BMP Only)
- **Provider**: DejaVu Fonts Project
- **License**: Free (similar to Bitstream Vera)
- **Coverage**: Limited emoji, good general Unicode
- **CMAP Format**: Format 4 only (BMP only)
- **Style**: Professional sans-serif
- **Size**: ~600KB
- **Download**: https://dejavu-fonts.github.io/
- **Best for**: Main text with occasional BMP symbols
- **Modern Emoji**: NO - Limited to ☀ ☁ ❤ ⭐

### Font Comparison Table

| Font | Emoji Coverage | Size | Format 12 | Modern Emoji | Color | License |
|------|----------------|------|-----------|--------------|-------|---------|
| Noto Emoji | Excellent | 400KB | ✓ | ✓ | ✘ | OFL |
| Symbola | Very Good | 3MB | ✓ | ✓ | ✘ | Freeware |
| Unifont | Complete | 12MB | ✓ | ✓ | ✘ | GPL v2+ |
| DejaVu Sans | Limited | 600KB | ✘ | ✘ | ✘ | Free |

---

## Best Practices

### 1. Font Selection

**Choose the right font for your use case:**
- **For production**: Noto Emoji (reliable, well-maintained)
- **For maximum compatibility**: Unifont (complete BMP coverage)
- **For mixed documents**: DejaVu Sans (main) + Noto Emoji (emoji)

### 2. Modern Emoji Usage

**Use modern emoji with Format 12 fonts:**
```go
// Modern emoji (requires Format 12)
pdf.Cell(10, 10, "\U0001F600") // 😀 grinning face
pdf.Cell(10, 10, "\U0001F389") // 🎉 party popper
pdf.Cell(10, 10, "\U0001F680") // 🚀 rocket

// BMP emoji (works with both Format 4 and 12)
pdf.Cell(10, 10, "\u2764") // ❤ heart
pdf.Cell(10, 10, "\u2B50") // ⭐ star
```

**Check Unicode values:**
- U+0000 - U+FFFF: BMP (works with all fonts)
- U+10000 - U+10FFFF: Supplementary (requires Format 12 font)

### 3. Font Switching Patterns

**Minimize font switches for performance:**
```go
// Less efficient (many switches)
for i := 0; i < 100; i++ {
    pdf.SetFont("regular", "", 12)
    pdf.Cell(30, 10, "Text")
    pdf.SetFont("emoji", "", 12)
    pdf.Cell(10, 10, "\u2714")
}

// Better: Batch operations
pdf.SetFont("regular", "", 12)
for i := 0; i < 100; i++ {
    pdf.Cell(30, 10, "Text")
}
pdf.SetFont("emoji", "", 12)
for i := 0; i < 100; i++ {
    pdf.Cell(10, 10, "\u2714")
}
```

### 4. Testing Emoji Rendering

**Always test your output:**
```go
func TestEmojiRendering(t *testing.T) {
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()
    pdf.AddUTF8Font("notoemoji", "", "NotoEmoji-Regular.ttf")
    pdf.SetFont("notoemoji", "", 12)

    // Test BMP emoji
    bmpEmoji := []string{
        "\u2600", // ☀ Sun
        "\u2764", // ❤ Heart
        "\u2714", // ✔ Checkmark
    }

    for _, emoji := range bmpEmoji {
        pdf.Cell(10, 10, emoji)
    }

    pdf.Ln(10)

    // Test modern emoji (requires Format 12)
    modernEmoji := []string{
        "\U0001F600", // 😀 Grinning face
        "\U0001F389", // 🎉 Party popper
        "\U0001F680", // 🚀 Rocket
    }

    for _, emoji := range modernEmoji {
        pdf.Cell(10, 10, emoji)
    }

    err := pdf.OutputFileAndClose("test_emoji.pdf")
    if err != nil {
        t.Errorf("Failed to generate PDF: %v", err)
    }

    // Manually verify output PDF has emoji
}
```

### 5. Fallback Strategies

**Handle missing glyphs gracefully:**
```go
func renderWithFallback(pdf *gofpdf.Fpdf, text string, emoji string) {
    pdf.SetFont("dejavu", "", 12)
    pdf.Cell(30, 10, text)

    // Try to render emoji
    pdf.SetFont("notoemoji", "", 12)
    pdf.Cell(10, 10, emoji)

    // If emoji font not available, fallback to ASCII
    if pdf.Err() {
        pdf.SetFont("dejavu", "", 12)
        pdf.Cell(10, 10, "[*]") // ASCII fallback
    }
}
```

### 6. Unicode Escape Sequences

**Use consistent Unicode notation:**
```go
// For BMP (4 hex digits): \uXXXX
pdf.Cell(10, 10, "\u2764") // ❤ heart

// For supplementary plane (8 hex digits): \UXXXXXXXX
pdf.Cell(10, 10, "\U0001F600") // 😀 grinning face

// Also valid: Literal Unicode (if editor supports UTF-8)
pdf.Cell(10, 10, "❤")
pdf.Cell(10, 10, "😀")
```

**Format Notes:**
- BMP: `\u` followed by exactly 4 hex digits
- Supplementary: `\U` followed by exactly 8 hex digits (uppercase U)
- Leading zeros required: `\U0001F600` not `\U1F600`

### 7. Document Structure

**Organize font declarations:**
```go
func setupFonts(pdf *gofpdf.Fpdf) {
    // Main text fonts
    pdf.AddUTF8Font("main", "", "DejaVuSansCondensed.ttf")
    pdf.AddUTF8Font("main", "B", "DejaVuSansCondensed-Bold.ttf")
    pdf.AddUTF8Font("main", "I", "DejaVuSansCondensed-Oblique.ttf")

    // Emoji font
    pdf.AddUTF8Font("emoji", "", "NotoEmoji-Regular.ttf")
}

func generateDocument() {
    pdf := gofpdf.New("P", "mm", "A4", "")
    setupFonts(pdf)
    pdf.AddPage()

    // Now use fonts consistently
    pdf.SetFont("main", "", 12)
    // ... document content
}
```

---

## Troubleshooting

### Problem: Emoji Don't Render

**Symptom**: Blank spaces or boxes instead of emoji

**Solutions**:
1. **Check font installation**:
   ```bash
   # Linux
   fc-list | grep -i emoji

   # macOS
   system_profiler SPFontsDataType | grep -i emoji
   ```

2. **Verify font path in code**:
   ```go
   // Use absolute path if relative path fails
   pdf.AddUTF8Font("emoji", "", "/usr/share/fonts/truetype/noto/NotoEmoji-Regular.ttf")
   ```

3. **Check CMAP format for modern emoji**:
   ```go
   // Modern emoji (U+1F600+) require Format 12
   // If your font only has Format 4, use BMP alternatives

   // Instead of:
   // pdf.Cell(10, 10, "\U0001F600") // 😀 (needs Format 12)

   // Use:
   pdf.Cell(10, 10, "\u263A") // ☺ (BMP alternative)
   ```

4. **Verify Unicode range matches font capability**:
   ```go
   // BMP emoji (works with all fonts)
   pdf.Cell(10, 10, "\u2764") // ❤

   // Modern emoji (requires Format 12 font)
   pdf.Cell(10, 10, "\U0001F600") // 😀
   ```

### Problem: Boxes (□) Instead of Emoji

**Symptom**: Square boxes displayed where emoji should be

**Solutions**:
1. **Font not loaded**:
   ```go
   // Make sure you call AddUTF8Font before SetFont
   pdf.AddUTF8Font("emoji", "", "NotoEmoji-Regular.ttf")
   pdf.SetFont("emoji", "", 12) // Now this will work
   ```

2. **Wrong font active**:
   ```go
   // Set emoji font before rendering emoji
   pdf.SetFont("emoji", "", 12) // Must be before Cell()
   pdf.Cell(10, 10, "\u2764")
   ```

3. **Font missing glyphs**:
   ```go
   // Try different font (Symbola has wider coverage)
   pdf.AddUTF8Font("symbola", "", "Symbola.ttf")
   pdf.SetFont("symbola", "", 12)
   ```

### Problem: Error "can't find character"

**Symptom**: Runtime error about missing character

**Solutions**:
1. **Use BMP emoji only**:
   ```go
   // Instead of supplementary plane emoji
   // Use BMP equivalents from U+2600-U+27BF
   ```

2. **Check font supports character**:
   - Open font in font viewer
   - Search for Unicode codepoint
   - Verify glyph exists

### Problem: Emoji Split Across Lines

**Symptom**: Multi-codepoint emoji sequences broken up

**This should not happen** - gofpdf uses grapheme cluster segmentation to prevent this. If you encounter this:

1. **Update to latest version**:
   ```bash
   go get -u github.com/phpdave11/gofpdf
   ```

2. **Report bug** with minimal reproduction case

### Problem: Performance Issues

**Symptom**: Slow PDF generation with many emoji

**Solutions**:
1. **Font subsetting is automatic** (only used glyphs embedded)

2. **Batch operations**:
   ```go
   // Set font once, render many times
   pdf.SetFont("emoji", "", 12)
   for i := 0; i < 1000; i++ {
       pdf.Cell(10, 10, "\u2764")
   }
   ```

3. **Consider image embedding** for complex emoji:
   ```go
   // For small set of repeated emoji
   pdf.RegisterImageOptions("heart", gofpdf.ImageOptions{ImageType: "PNG"})
   pdf.ImageOptions("heart.png", 10, 10, 5, 5, false,
       gofpdf.ImageOptions{}, 0, "heart")
   ```

### Problem: Modern Emoji Don't Work (Format 12 Missing)

**Symptom**: Modern emoji (😀 🎉 🚀) appear as blank spaces or boxes, but BMP emoji (☀ ❤) work fine

**Diagnosis**: Your font doesn't have a CMAP Format 12 table

**Solutions**:

1. **Verify font CMAP format**:
   ```bash
   # Check if font has Format 12 support
   # Linux: Install fonttools
   ttx -t cmap NotoEmoji-Regular.ttf
   grep "platformID=\"3\" platEncID=\"10\"" *.ttx

   # If no output, font lacks Format 12
   ```

2. **Switch to Format 12-compatible font**:
   ```go
   // Old: DejaVu Sans (Format 4 only)
   // pdf.AddUTF8Font("emoji", "", "DejaVuSans.ttf")

   // New: Noto Emoji (Format 12 supported)
   pdf.AddUTF8Font("emoji", "", "NotoEmoji-Regular.ttf")
   ```

3. **Use BMP emoji alternatives**:
   ```go
   // Modern emoji that don't work with Format 4:
   // "\U0001F600" // 😀
   // "\U0001F389" // 🎉
   // "\U0001F680" // 🚀

   // BMP alternatives that work:
   pdf.Cell(10, 10, "\u263A") // ☺ smiley
   pdf.Cell(10, 10, "\u2728") // ✨ sparkles
   pdf.Cell(10, 10, "\u2708") // ✈ airplane
   ```

4. **Mixed font strategy** (if you must keep current font):
   ```go
   // Use current font for text
   pdf.AddUTF8Font("main", "", "DejaVuSans.ttf")

   // Add Format 12 font for modern emoji
   pdf.AddUTF8Font("emoji", "", "NotoEmoji-Regular.ttf")

   // Switch fonts as needed
   pdf.SetFont("main", "", 12)
   pdf.Cell(30, 10, "Status: ")

   pdf.SetFont("emoji", "", 12)
   pdf.Cell(10, 10, "\U0001F680") // 🚀 now works
   ```

5. **Font compatibility matrix**:
   ```
   Format 12 fonts (modern emoji work):
   - Noto Emoji        ✓
   - Symbola           ✓
   - Unifont           ✓
   - Noto Sans/Serif   ✓
   - Apple Color Emoji ✓
   - Segoe UI Emoji    ✓

   Format 4 only (BMP emoji only):
   - DejaVu Sans       ✘
   - Arial             ✘ (most versions)
   - Times New Roman   ✘
   - Helvetica         ✘
   ```

### Problem: "Invalid UTF-8" Error

**Symptom**: Error about invalid UTF-8 sequence

**Solutions**:
1. **Use valid Unicode escapes**:
   ```go
   // Correct
   pdf.Cell(10, 10, "\u2764")

   // Also correct (if editor supports UTF-8)
   pdf.Cell(10, 10, "❤")

   // Wrong - invalid escape
   // pdf.Cell(10, 10, "\uD800") // Invalid surrogate
   ```

2. **Validate input strings**:
   ```go
   import "unicode/utf8"

   if !utf8.ValidString(inputText) {
       // Handle invalid UTF-8
       inputText = strings.ToValidUTF8(inputText, "?")
   }
   ```

### Problem: Font File Not Found

**Symptom**: Error "Can't open font file"

**Solutions**:
1. **Use absolute path**:
   ```go
   pdf.AddUTF8Font("emoji", "", "/full/path/to/NotoEmoji-Regular.ttf")
   ```

2. **Check file permissions**:
   ```bash
   ls -l NotoEmoji-Regular.ttf
   # Should be readable (r-- or rw-)
   ```

3. **Verify file exists**:
   ```bash
   find /usr -name "NotoEmoji*" 2>/dev/null
   ```

---

## Unicode Reference

### Useful Emoji Ranges (BMP)

#### Miscellaneous Symbols (U+2600-U+26FF)
```
☀ U+2600  Sun
☁ U+2601  Cloud
☂ U+2602  Umbrella
☃ U+2603  Snowman
⛄ U+26C4  Snowman without snow
⚡ U+26A1  Lightning
⛈ U+26C8  Cloud with lightning
⭐ U+2B50  Star
✨ U+2728  Sparkles
⚠ U+26A0  Warning
⛔ U+26D4  No entry
```

#### Dingbats (U+2700-U+27BF)
```
✂ U+2702  Scissors
✈ U+2708  Airplane
✉ U+2709  Envelope
✏ U+270F  Pencil
✒ U+2712  Pen
✔ U+2714  Checkmark
✖ U+2716  X mark
✘ U+2718  X mark (heavy)
✓ U+2713  Checkmark (light)
❤ U+2764  Red heart
➤ U+27A4  Arrow right
```

#### Geometric Shapes (U+25A0-U+25FF)
```
■ U+25A0  Black square
□ U+25A1  White square
▲ U+25B2  Black triangle up
△ U+25B3  White triangle up
● U+25CF  Black circle
○ U+25CB  White circle
◆ U+25C6  Black diamond
◇ U+25C7  White diamond
```

#### Playing Cards (U+2660-U+2667)
```
♠ U+2660  Spade
♣ U+2663  Club
♥ U+2665  Heart
♦ U+2666  Diamond
```

#### Arrows (U+2190-U+21FF)
```
← U+2190  Left arrow
→ U+2192  Right arrow
↑ U+2191  Up arrow
↓ U+2193  Down arrow
↔ U+2194  Left-right arrow
⇐ U+21D0  Left double arrow
⇒ U+21D2  Right double arrow
```

#### Mathematical Operators (U+2200-U+22FF)
```
∞ U+221E  Infinity
≈ U+2248  Almost equal
≠ U+2260  Not equal
≤ U+2264  Less than or equal
≥ U+2265  Greater than or equal
± U+00B1  Plus-minus
× U+00D7  Multiplication
÷ U+00F7  Division
```

### Common Unicode Escapes in Go

```go
// 4-digit hex (BMP): \uXXXX
"\u2764"  // ❤ - Heart

// 8-digit hex (supplementary plane): \UXXXXXXXX
// "\U0001F600"  // 😀 - Not yet supported

// Octal: \NNN (not recommended)
"\342\235\244"  // ❤ - Same as above

// Hexadecimal byte: \xXX (not recommended for Unicode)
"\xe2\x9d\xa4"  // ❤ - Same as above
```

### Variation Selectors

Some characters have text vs emoji presentation controlled by variation selectors:

```go
"\u2600"      // ☀ - Sun (text presentation)
"\u2600\uFE0F" // ☀️ - Sun (emoji presentation with U+FE0F)
```

**Note**: Variation selectors are part of grapheme clusters and handled automatically by gofpdf.

### Zero-Width Joiner (ZWJ) Sequences

Multi-person emoji use ZWJ (U+200D) to join individual emoji:

```go
// Family: Man + ZWJ + Woman + ZWJ + Girl + ZWJ + Boy
// "\U0001F468\u200D\U0001F469\u200D\U0001F467\u200D\U0001F466"
// Note: Supplementary plane, not yet supported
```

### Skin Tone Modifiers

Skin tone modifiers (U+1F3FB - U+1F3FF) combine with person emoji:

```go
// Thumbs up with medium skin tone
// "\U0001F44D\U0001F3FD"
// Note: Supplementary plane, not yet supported

// For BMP emoji with skin tones, modifiers work:
"\u270B\U0001F3FD" // ✋ + skin tone (if font supports)
```

---

## Additional Resources

### Documentation
- [gofpdf GoDoc](https://godoc.org/github.com/phpdave11/gofpdf)
- [Unicode Emoji Standard](https://unicode.org/emoji/charts/full-emoji-list.html)
- [Grapheme Clusters (UAX #29)](https://unicode.org/reports/tr29/)
- [TrueType Font Specification](https://docs.microsoft.com/en-us/typography/opentype/spec/)

### Tools
- [Unicode Character Inspector](https://unicode-explorer.com/)
- [Emoji Codepoint Lookup](https://emojipedia.org/)
- [Font Viewer (FontForge)](https://fontforge.org/)

### Support
- GitHub Issues: https://github.com/phpdave11/gofpdf/issues
- Stack Overflow: Tag `gofpdf`

---

## Quick Reference Card

### Installation Checklist
- [ ] Install Go 1.13+
- [ ] Install emoji font (Noto Emoji recommended)
- [ ] Get gofpdf: `go get github.com/phpdave11/gofpdf`
- [ ] Verify font path

### Code Template
```go
package main

import "github.com/phpdave11/gofpdf"

func main() {
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()

    // Add fonts
    pdf.AddUTF8Font("main", "", "DejaVuSansCondensed.ttf")
    pdf.AddUTF8Font("emoji", "", "NotoEmoji-Regular.ttf")

    // Regular text
    pdf.SetFont("main", "", 12)
    pdf.Cell(30, 10, "Status: ")

    // Emoji
    pdf.SetFont("emoji", "", 12)
    pdf.Cell(10, 10, "\u2714") // ✔

    // Save
    pdf.OutputFileAndClose("output.pdf")
}
```

### Emoji Quick Reference

**BMP Emoji (work with all fonts):**
```
Weather: ☀☁☂⛄⚡⭐
Symbols: ❤⭐✔✘⚠⛔
Arrows:  ←→↑↓↔⇒
Math:    ∞≈≠±×÷
Shapes:  ■□●○◆◇
Cards:   ♠♣♥♦
```

**Modern Emoji (require Format 12):**
```
Faces:   😀😃😄😁😅😊
Hands:   👍👎👏🙏🤝✌
Objects: 🚀🎉💯🔥⚡✨
Food:    🍕🍔🍰🍎🍇🍉
Nature:  🌍🌙⭐🔥💧🌈
```

---

**Last Updated**: October 2025
**gofpdf Version**: Latest (CMAP Format 12 support - full Unicode range)
**Maintained by**: gofpdf community

**Recent Changes**:
- Added CMAP Format 12 support for modern emoji (😀 🎉 🚀)
- Full Unicode support (U+0000 - U+10FFFF)
- Automatic format detection and fallback
- Dynamic ToUnicode CMap generation
- Identity CIDToGIDMap for supplementary plane efficiency

For questions or issues, please visit: https://github.com/phpdave11/gofpdf
