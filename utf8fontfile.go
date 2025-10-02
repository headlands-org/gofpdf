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
	"math"
	"sort"
)

// flags
const symbolWords = 1 << 0
const symbolScale = 1 << 3
const symbolContinue = 1 << 5
const symbolAllScale = 1 << 6
const symbol2x2 = 1 << 7

type utf8FontFile struct {
	fileReader           *fileReader
	LastRune             int
	tableDescriptions    map[string]*tableDescription
	outTablesData        map[string][]byte
	symbolPosition       []int
	charSymbolDictionary map[int]int
	Ascent               int
	Descent              int
	fontElementSize      int
	Bbox                 fontBoxType
	CapHeight            int
	StemV                int
	ItalicAngle          int
	Flags                int
	UnderlinePosition    float64
	UnderlineThickness   float64
	CharWidths           map[int]int
	DefaultWidth         float64
	symbolData           map[int]map[string][]int
	CodeSymbolDictionary map[int]int
	ToUnicodeCMap        string // Dynamic ToUnicode CMap for PDF embedding
}

type tableDescription struct {
	name     string
	checksum []int
	position int
	size     int
}

type fileReader struct {
	readerPosition int64
	array          []byte
}

func (fr *fileReader) Read(s int) []byte {
	b := fr.array[fr.readerPosition : fr.readerPosition+int64(s)]
	fr.readerPosition += int64(s)
	return b
}

func (fr *fileReader) seek(shift int64, flag int) (int64, error) {
	if flag == 0 {
		fr.readerPosition = shift
	} else if flag == 1 {
		fr.readerPosition += shift
	} else if flag == 2 {
		fr.readerPosition = int64(len(fr.array)) - shift
	}
	return int64(fr.readerPosition), nil
}

func newUTF8Font(reader *fileReader) *utf8FontFile {
	utf := utf8FontFile{
		fileReader: reader,
	}
	return &utf
}

func (utf *utf8FontFile) parseFile() error {
	utf.fileReader.readerPosition = 0
	utf.symbolPosition = make([]int, 0)
	utf.charSymbolDictionary = make(map[int]int)
	utf.tableDescriptions = make(map[string]*tableDescription)
	utf.outTablesData = make(map[string][]byte)
	utf.Ascent = 0
	utf.Descent = 0
	codeType := uint32(utf.readUint32())
	if codeType == 0x4F54544F {
		return fmt.Errorf("not supported\n ")
	}
	if codeType == 0x74746366 {
		return fmt.Errorf("not supported\n ")
	}
	if codeType != 0x00010000 && codeType != 0x74727565 {
		return fmt.Errorf("Not a TrueType font: codeType=%v\n ", codeType)
	}
	utf.generateTableDescriptions()
	utf.parseTables()
	return nil
}

func (utf *utf8FontFile) generateTableDescriptions() {

	tablesCount := utf.readUint16()
	_ = utf.readUint16()
	_ = utf.readUint16()
	_ = utf.readUint16()
	utf.tableDescriptions = make(map[string]*tableDescription)

	for i := 0; i < tablesCount; i++ {
		record := tableDescription{
			name:     utf.readTableName(),
			checksum: []int{utf.readUint16(), utf.readUint16()},
			position: utf.readUint32(),
			size:     utf.readUint32(),
		}
		utf.tableDescriptions[record.name] = &record
	}
}

func (utf *utf8FontFile) readTableName() string {
	return string(utf.fileReader.Read(4))
}

func (utf *utf8FontFile) readUint16() int {
	s := utf.fileReader.Read(2)
	return (int(s[0]) << 8) + int(s[1])
}

func (utf *utf8FontFile) readUint32() int {
	s := utf.fileReader.Read(4)
	return (int(s[0]) * 16777216) + (int(s[1]) << 16) + (int(s[2]) << 8) + int(s[3]) // 	16777216  = 1<<24
}

func (utf *utf8FontFile) calcInt32(x, y []int) []int {
	answer := make([]int, 2)
	if y[1] > x[1] {
		x[1] += 1 << 16
		x[0]++
	}
	answer[1] = x[1] - y[1]
	if y[0] > x[0] {
		x[0] += 1 << 16
	}
	answer[0] = x[0] - y[0]
	answer[0] = answer[0] & 0xFFFF
	return answer
}

func (utf *utf8FontFile) generateChecksum(data []byte) []int {
	if (len(data) % 4) != 0 {
		for i := 0; (len(data) % 4) != 0; i++ {
			data = append(data, 0)
		}
	}
	answer := []int{0x0000, 0x0000}
	for i := 0; i < len(data); i += 4 {
		answer[0] += (int(data[i]) << 8) + int(data[i+1])
		answer[1] += (int(data[i+2]) << 8) + int(data[i+3])
		answer[0] += answer[1] >> 16
		answer[1] = answer[1] & 0xFFFF
		answer[0] = answer[0] & 0xFFFF
	}
	return answer
}

func (utf *utf8FontFile) seek(shift int) {
	_, _ = utf.fileReader.seek(int64(shift), 0)
}

func (utf *utf8FontFile) skip(delta int) {
	_, _ = utf.fileReader.seek(int64(delta), 1)
}

// SeekTable position
func (utf *utf8FontFile) SeekTable(name string) int {
	return utf.seekTable(name, 0)
}

func (utf *utf8FontFile) seekTable(name string, offsetInTable int) int {
	_, _ = utf.fileReader.seek(int64(utf.tableDescriptions[name].position+offsetInTable), 0)
	return int(utf.fileReader.readerPosition)
}

func (utf *utf8FontFile) readInt16() int16 {
	s := utf.fileReader.Read(2)
	a := (int16(s[0]) << 8) + int16(s[1])
	if (int(a) & (1 << 15)) == 0 {
		a = int16(int(a) - (1 << 16))
	}
	return a
}

func (utf *utf8FontFile) getUint16(pos int) int {
	_, _ = utf.fileReader.seek(int64(pos), 0)
	s := utf.fileReader.Read(2)
	return (int(s[0]) << 8) + int(s[1])
}

func (utf *utf8FontFile) splice(stream []byte, offset int, value []byte) []byte {
	stream = append([]byte{}, stream...)
	return append(append(stream[:offset], value...), stream[offset+len(value):]...)
}

func (utf *utf8FontFile) insertUint16(stream []byte, offset int, value int) []byte {
	up := make([]byte, 2)
	binary.BigEndian.PutUint16(up, uint16(value))
	return utf.splice(stream, offset, up)
}

func (utf *utf8FontFile) getRange(pos, length int) []byte {
	_, _ = utf.fileReader.seek(int64(pos), 0)
	if length < 1 {
		return make([]byte, 0)
	}
	s := utf.fileReader.Read(length)
	return s
}

func (utf *utf8FontFile) getTableData(name string) []byte {
	desckrip := utf.tableDescriptions[name]
	if desckrip == nil {
		return nil
	}
	if desckrip.size == 0 {
		return nil
	}
	_, _ = utf.fileReader.seek(int64(desckrip.position), 0)
	s := utf.fileReader.Read(desckrip.size)
	return s
}

func (utf *utf8FontFile) setOutTable(name string, data []byte) {
	if data == nil {
		return
	}
	if name == "head" {
		data = utf.splice(data, 8, []byte{0, 0, 0, 0})
	}
	utf.outTablesData[name] = data
}

func arrayKeys(arr map[int]string) []int {
	answer := make([]int, len(arr))
	i := 0
	for key := range arr {
		answer[i] = key
		i++
	}
	return answer
}

func inArray(s int, arr []int) bool {
	for _, i := range arr {
		if s == i {
			return true
		}
	}
	return false
}

func (utf *utf8FontFile) parseNAMETable() int {
	namePosition := utf.SeekTable("name")
	format := utf.readUint16()
	if format != 0 {
		fmt.Printf("Illegal format %d\n", format)
		return format
	}
	nameCount := utf.readUint16()
	stringDataPosition := namePosition + utf.readUint16()
	names := map[int]string{1: "", 2: "", 3: "", 4: "", 6: ""}
	keys := arrayKeys(names)
	counter := len(names)
	for i := 0; i < nameCount; i++ {
		system := utf.readUint16()
		code := utf.readUint16()
		local := utf.readUint16()
		nameID := utf.readUint16()
		size := utf.readUint16()
		position := utf.readUint16()
		if !inArray(nameID, keys) {
			continue
		}
		currentName := ""
		if system == 3 && code == 1 && local == 0x409 {
			oldPos := utf.fileReader.readerPosition
			utf.seek(stringDataPosition + position)
			if size%2 != 0 {
				fmt.Printf("name is not binar byte format\n")
				return format
			}
			size /= 2
			currentName = ""
			for size > 0 {
				char := utf.readUint16()
				currentName += string(rune(char))
				size--
			}
			utf.fileReader.readerPosition = oldPos
			utf.seek(int(oldPos))
		} else if system == 1 && code == 0 && local == 0 {
			oldPos := utf.fileReader.readerPosition
			currentName = string(utf.getRange(stringDataPosition+position, size))
			utf.fileReader.readerPosition = oldPos
			utf.seek(int(oldPos))
		}
		if currentName != "" && names[nameID] == "" {
			names[nameID] = currentName
			counter--
			if counter == 0 {
				break
			}
		}
	}
	return format
}

func (utf *utf8FontFile) parseHEADTable() {
	utf.SeekTable("head")
	utf.skip(18)
	utf.fontElementSize = utf.readUint16()
	scale := 1000.0 / float64(utf.fontElementSize)
	utf.skip(16)
	xMin := utf.readInt16()
	yMin := utf.readInt16()
	xMax := utf.readInt16()
	yMax := utf.readInt16()
	utf.Bbox = fontBoxType{int(float64(xMin) * scale), int(float64(yMin) * scale), int(float64(xMax) * scale), int(float64(yMax) * scale)}
	utf.skip(3 * 2)
	_ = utf.readUint16()
	symbolDataFormat := utf.readUint16()
	if symbolDataFormat != 0 {
		fmt.Printf("Unknown symbol data format %d\n", symbolDataFormat)
		return
	}
}

func (utf *utf8FontFile) parseHHEATable() int {
	metricsCount := 0
	if _, OK := utf.tableDescriptions["hhea"]; OK {
		scale := 1000.0 / float64(utf.fontElementSize)
		utf.SeekTable("hhea")
		utf.skip(4)
		hheaAscender := utf.readInt16()
		hheaDescender := utf.readInt16()
		utf.Ascent = int(float64(hheaAscender) * scale)
		utf.Descent = int(float64(hheaDescender) * scale)
		utf.skip(24)
		metricDataFormat := utf.readUint16()
		if metricDataFormat != 0 {
			fmt.Printf("Unknown horizontal metric data format %d\n", metricDataFormat)
			return 0
		}
		metricsCount = utf.readUint16()
		if metricsCount == 0 {
			fmt.Printf("Number of horizontal metrics is 0\n")
			return 0
		}
	}
	return metricsCount
}

func (utf *utf8FontFile) parseOS2Table() int {
	var weightType int
	scale := 1000.0 / float64(utf.fontElementSize)
	if _, OK := utf.tableDescriptions["OS/2"]; OK {
		utf.SeekTable("OS/2")
		version := utf.readUint16()
		utf.skip(2)
		weightType = utf.readUint16()
		utf.skip(2)
		fsType := utf.readUint16()
		if fsType == 0x0002 || (fsType&0x0300) != 0 {
			fmt.Printf("ERROR - copyright restrictions.\n")
			return 0
		}
		utf.skip(20)
		_ = utf.readInt16()

		utf.skip(36)
		sTypoAscender := utf.readInt16()
		sTypoDescender := utf.readInt16()
		if utf.Ascent == 0 {
			utf.Ascent = int(float64(sTypoAscender) * scale)
		}
		if utf.Descent == 0 {
			utf.Descent = int(float64(sTypoDescender) * scale)
		}
		if version > 1 {
			utf.skip(16)
			sCapHeight := utf.readInt16()
			utf.CapHeight = int(float64(sCapHeight) * scale)
		} else {
			utf.CapHeight = utf.Ascent
		}
	} else {
		weightType = 500
		if utf.Ascent == 0 {
			utf.Ascent = int(float64(utf.Bbox.Ymax) * scale)
		}
		if utf.Descent == 0 {
			utf.Descent = int(float64(utf.Bbox.Ymin) * scale)
		}
		utf.CapHeight = utf.Ascent
	}
	utf.StemV = 50 + int(math.Pow(float64(weightType)/65.0, 2))
	return weightType
}

func (utf *utf8FontFile) parsePOSTTable(weight int) {
	utf.SeekTable("post")
	utf.skip(4)
	utf.ItalicAngle = int(utf.readInt16()) + utf.readUint16()/65536.0
	scale := 1000.0 / float64(utf.fontElementSize)
	utf.UnderlinePosition = float64(utf.readInt16()) * scale
	utf.UnderlineThickness = float64(utf.readInt16()) * scale
	fixed := utf.readUint32()

	utf.Flags = 4

	if utf.ItalicAngle != 0 {
		utf.Flags = utf.Flags | 64
	}
	if weight >= 600 {
		utf.Flags = utf.Flags | 262144
	}
	if fixed != 0 {
		utf.Flags = utf.Flags | 1
	}
}

func (utf *utf8FontFile) parseCMAPTable(format int) int {
	cmapPosition := utf.SeekTable("cmap")
	utf.skip(2)
	cmapTableCount := utf.readUint16()
	cidCMAPPosition := 0
	format12Position := 0
	for i := 0; i < cmapTableCount; i++ {
		system := utf.readUint16()
		coded := utf.readUint16()
		position := utf.readUint32()
		oldReaderPosition := utf.fileReader.readerPosition
		// Prioritize Format 12 (full Unicode) over Format 4 (BMP only)
		if (system == 3 && (coded == 1 || coded == 10)) || system == 0 { // Microsoft, Unicode
			format = utf.getUint16(cmapPosition + position)
			if format == 12 {
				// Format 12 found - use it and stop searching
				format12Position = cmapPosition + position
				break
			} else if format == 4 && cidCMAPPosition == 0 {
				// Format 4 found - save as fallback but keep searching for Format 12
				cidCMAPPosition = cmapPosition + position
			}
		}
		utf.seek(int(oldReaderPosition))
	}
	// Use Format 12 if found, otherwise use Format 4
	if format12Position != 0 {
		return format12Position
	}
	if cidCMAPPosition == 0 {
		fmt.Printf("Font does not have cmap for Unicode\n")
		return cidCMAPPosition
	}
	return cidCMAPPosition
}

func (utf *utf8FontFile) parseTables() {
	f := utf.parseNAMETable()
	utf.parseHEADTable()
	n := utf.parseHHEATable()
	w := utf.parseOS2Table()
	utf.parsePOSTTable(w)
	runeCMAPPosition := utf.parseCMAPTable(f)

	utf.SeekTable("maxp")
	utf.skip(4)
	numSymbols := utf.readUint16()

	symbolCharDictionary := make(map[int][]int)
	charSymbolDictionary := make(map[int]int)
	utf.generateSCCSDictionaries(runeCMAPPosition, symbolCharDictionary, charSymbolDictionary)

	scale := 1000.0 / float64(utf.fontElementSize)
	utf.parseHMTXTable(n, numSymbols, symbolCharDictionary, scale)
}

func (utf *utf8FontFile) generateCMAP() map[int][]int {
	cmapPosition := utf.SeekTable("cmap")
	utf.skip(2)
	cmapTableCount := utf.readUint16()
	runeCmapPosition := 0
	format12Position := 0
	for i := 0; i < cmapTableCount; i++ {
		system := utf.readUint16()
		coder := utf.readUint16()
		position := utf.readUint32()
		oldPosition := utf.fileReader.readerPosition
		// Prioritize Format 12 (full Unicode) over Format 4 (BMP only)
		if (system == 3 && (coder == 1 || coder == 10)) || system == 0 {
			format := utf.getUint16(cmapPosition + position)
			if format == 12 {
				// Format 12 found - use it and stop searching
				format12Position = cmapPosition + position
				break
			} else if format == 4 && runeCmapPosition == 0 {
				// Format 4 found - save as fallback but keep searching for Format 12
				runeCmapPosition = cmapPosition + position
			}
		}
		utf.seek(int(oldPosition))
	}

	// Use Format 12 if found, otherwise use Format 4
	if format12Position != 0 {
		runeCmapPosition = format12Position
	}
	if runeCmapPosition == 0 {
		fmt.Printf("Font does not have cmap for Unicode\n")
		return nil
	}

	symbolCharDictionary := make(map[int][]int)
	charSymbolDictionary := make(map[int]int)
	utf.generateSCCSDictionaries(runeCmapPosition, symbolCharDictionary, charSymbolDictionary)

	utf.charSymbolDictionary = charSymbolDictionary

	return symbolCharDictionary
}

func (utf *utf8FontFile) parseSymbols(usedRunes map[int]int) (map[int]int, map[int]int, map[int]int, []int) {
	symbolCollection := map[int]int{0: 0}
	charSymbolPairCollection := make(map[int]int)
	for _, char := range usedRunes {
		if _, OK := utf.charSymbolDictionary[char]; OK {
			symbolCollection[utf.charSymbolDictionary[char]] = char
			charSymbolPairCollection[char] = utf.charSymbolDictionary[char]

		}
		utf.LastRune = max(utf.LastRune, char)
	}

	begin := utf.tableDescriptions["glyf"].position

	symbolArray := make(map[int]int)
	symbolCollectionKeys := keySortInt(symbolCollection)

	symbolCounter := 0
	for _, oldSymbolIndex := range symbolCollectionKeys {
		symbolArray[oldSymbolIndex] = symbolCounter
		symbolCounter++
	}
	charSymbolPairCollectionKeys := keySortInt(charSymbolPairCollection)
	runeSymbolPairCollection := make(map[int]int)
	for _, runa := range charSymbolPairCollectionKeys {
		runeSymbolPairCollection[runa] = symbolArray[charSymbolPairCollection[runa]]
	}

	symbolCollectionKeys = keySortInt(symbolCollection)
	for _, oldSymbolIndex := range symbolCollectionKeys {
		_, symbolArray, symbolCollection, symbolCollectionKeys = utf.getSymbols(oldSymbolIndex, &begin, symbolArray, symbolCollection, symbolCollectionKeys)
	}

	return runeSymbolPairCollection, symbolArray, symbolCollection, symbolCollectionKeys
}

func (utf *utf8FontFile) generateCMAPTable(cidSymbolPairCollection map[int]int, numSymbols int) []byte {
	// Build optimized cmap groups using Format 12
	groups := buildCmapGroups(cidSymbolPairCollection)

	// Build CMAP table directory header
	// Version: 0, numTables: 1
	cmap := make([]byte, 0)
	cmap = append(cmap, packUint16(0)...) // version
	cmap = append(cmap, packUint16(1)...) // numTables

	// Platform ID: 3 (Windows), Encoding ID: 10 (Unicode full repertoire)
	cmap = append(cmap, packUint16(3)...)  // platformID
	cmap = append(cmap, packUint16(10)...) // encodingID (changed from 1 to 10 for full Unicode)

	// Offset to subtable: 12 bytes (from start of CMAP table)
	cmap = append(cmap, packUint32(12)...) // offset

	// Generate Format 12 subtable header
	header := writeCmapFormat12Header(uint32(len(groups)))
	cmap = append(cmap, header...)

	// Write each group: startCharCode, endCharCode, startGlyphID (all uint32)
	for _, group := range groups {
		cmap = append(cmap, packUint32(int(group.startCharCode))...)
		cmap = append(cmap, packUint32(int(group.endCharCode))...)
		cmap = append(cmap, packUint32(int(group.startGlyphID))...)
	}

	return cmap
}

// GenerateCutFont fill utf8FontFile from .utf file, only with runes from usedRunes
func (utf *utf8FontFile) GenerateCutFont(usedRunes map[int]int) []byte {
	utf.fileReader.readerPosition = 0
	utf.symbolPosition = make([]int, 0)
	utf.charSymbolDictionary = make(map[int]int)
	utf.tableDescriptions = make(map[string]*tableDescription)
	utf.outTablesData = make(map[string][]byte)
	utf.Ascent = 0
	utf.Descent = 0
	utf.skip(4)
	utf.LastRune = 0
	utf.generateTableDescriptions()

	utf.SeekTable("head")
	utf.skip(50)
	LocaFormat := utf.readUint16()

	utf.SeekTable("hhea")
	utf.skip(34)
	metricsCount := utf.readUint16()
	oldMetrics := metricsCount

	utf.SeekTable("maxp")
	utf.skip(4)
	numSymbols := utf.readUint16()

	symbolCharDictionary := utf.generateCMAP()
	if symbolCharDictionary == nil {
		return nil
	}

	utf.parseHMTXTable(metricsCount, numSymbols, symbolCharDictionary, 1.0)

	utf.parseLOCATable(LocaFormat, numSymbols)

	unicodeGlyphMap, symbolArray, symbolCollection, symbolCollectionKeys := utf.parseSymbols(usedRunes)

	cidToUnicode := make(map[int]int)
	cidToGlyph := make(map[int]int)
	maxCID := 0
	for cid, unicode := range usedRunes {
		if cid == 0 {
			continue
		}
		cidToUnicode[cid] = unicode
		if glyph, ok := unicodeGlyphMap[unicode]; ok {
			cidToGlyph[cid] = glyph
		}
		if cid > maxCID {
			maxCID = cid
		}
	}
	utf.ToUnicodeCMap = generateToUnicodeCMap(cidToUnicode)
	utf.CodeSymbolDictionary = cidToGlyph
	utf.LastRune = maxCID

	metricsCount = len(symbolCollection)
	numSymbols = metricsCount

	utf.setOutTable("name", utf.getTableData("name"))
	utf.setOutTable("cvt ", utf.getTableData("cvt "))
	utf.setOutTable("fpgm", utf.getTableData("fpgm"))
	utf.setOutTable("prep", utf.getTableData("prep"))
	utf.setOutTable("gasp", utf.getTableData("gasp"))

	postTable := utf.getTableData("post")
	postTable = append(append([]byte{0x00, 0x03, 0x00, 0x00}, postTable[4:16]...), []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}...)
	utf.setOutTable("post", postTable)

	delete(unicodeGlyphMap, 0)

	utf.setOutTable("cmap", utf.generateCMAPTable(unicodeGlyphMap, numSymbols))

	symbolData := utf.getTableData("glyf")

	offsets := make([]int, 0)
	glyfData := make([]byte, 0)
	pos := 0
	hmtxData := make([]byte, 0)
	utf.symbolData = make(map[int]map[string][]int, 0)

	for _, originalSymbolIdx := range symbolCollectionKeys {
		hm := utf.getMetrics(oldMetrics, originalSymbolIdx)
		hmtxData = append(hmtxData, hm...)

		offsets = append(offsets, pos)
		symbolPos := utf.symbolPosition[originalSymbolIdx]
		symbolLen := utf.symbolPosition[originalSymbolIdx+1] - symbolPos
		data := symbolData[symbolPos : symbolPos+symbolLen]
		var up int
		if symbolLen > 0 {
			up = unpackUint16(data[0:2])
		}

		if symbolLen > 2 && (up&(1<<15)) != 0 {
			posInSymbol := 10
			flags := symbolContinue
			nComponentElements := 0
			for (flags & symbolContinue) != 0 {
				nComponentElements++
				up = unpackUint16(data[posInSymbol : posInSymbol+2])
				flags = up
				up = unpackUint16(data[posInSymbol+2 : posInSymbol+4])
				symbolIdx := up
				if _, OK := utf.symbolData[originalSymbolIdx]; !OK {
					utf.symbolData[originalSymbolIdx] = make(map[string][]int)
				}
				if _, OK := utf.symbolData[originalSymbolIdx]["compSymbols"]; !OK {
					utf.symbolData[originalSymbolIdx]["compSymbols"] = make([]int, 0)
				}
				utf.symbolData[originalSymbolIdx]["compSymbols"] = append(utf.symbolData[originalSymbolIdx]["compSymbols"], symbolIdx)
				data = utf.insertUint16(data, posInSymbol+2, symbolArray[symbolIdx])
				posInSymbol += 4
				if (flags & symbolWords) != 0 {
					posInSymbol += 4
				} else {
					posInSymbol += 2
				}
				if (flags & symbolScale) != 0 {
					posInSymbol += 2
				} else if (flags & symbolAllScale) != 0 {
					posInSymbol += 4
				} else if (flags & symbol2x2) != 0 {
					posInSymbol += 8
				}
			}
		}

		glyfData = append(glyfData, data...)
		pos += symbolLen
		if pos%4 != 0 {
			padding := 4 - (pos % 4)
			glyfData = append(glyfData, make([]byte, padding)...)
			pos += padding
		}
	}

	offsets = append(offsets, pos)
	utf.setOutTable("glyf", glyfData)

	utf.setOutTable("hmtx", hmtxData)

	locaData := make([]byte, 0)
	if ((pos + 1) >> 1) > 0xFFFF {
		LocaFormat = 1
		for _, offset := range offsets {
			locaData = append(locaData, packUint32(offset)...)
		}
	} else {
		LocaFormat = 0
		for _, offset := range offsets {
			locaData = append(locaData, packUint16(offset/2)...)
		}
	}
	utf.setOutTable("loca", locaData)

	headData := utf.getTableData("head")
	headData = utf.insertUint16(headData, 50, LocaFormat)
	utf.setOutTable("head", headData)

	hheaData := utf.getTableData("hhea")
	hheaData = utf.insertUint16(hheaData, 34, metricsCount)
	utf.setOutTable("hhea", hheaData)

	maxp := utf.getTableData("maxp")
	maxp = utf.insertUint16(maxp, 4, numSymbols)
	utf.setOutTable("maxp", maxp)

	os2Data := utf.getTableData("OS/2")
	utf.setOutTable("OS/2", os2Data)

	return utf.assembleTables()
}

func (utf *utf8FontFile) getSymbols(originalSymbolIdx int, start *int, symbolSet map[int]int, SymbolsCollection map[int]int, SymbolsCollectionKeys []int) (*int, map[int]int, map[int]int, []int) {
	symbolPos := utf.symbolPosition[originalSymbolIdx]
	symbolSize := utf.symbolPosition[originalSymbolIdx+1] - symbolPos
	if symbolSize == 0 {
		return start, symbolSet, SymbolsCollection, SymbolsCollectionKeys
	}
	utf.seek(*start + symbolPos)

	lineCount := utf.readInt16()

	if lineCount < 0 {
		utf.skip(8)
		flags := symbolContinue
		for flags&symbolContinue != 0 {
			flags = utf.readUint16()
			symbolIndex := utf.readUint16()
			if _, OK := symbolSet[symbolIndex]; !OK {
				symbolSet[symbolIndex] = len(SymbolsCollection)
				SymbolsCollection[symbolIndex] = 1
				SymbolsCollectionKeys = append(SymbolsCollectionKeys, symbolIndex)
			}
			oldPosition, _ := utf.fileReader.seek(0, 1)
			_, _, _, SymbolsCollectionKeys = utf.getSymbols(symbolIndex, start, symbolSet, SymbolsCollection, SymbolsCollectionKeys)
			utf.seek(int(oldPosition))
			if flags&symbolWords != 0 {
				utf.skip(4)
			} else {
				utf.skip(2)
			}
			if flags&symbolScale != 0 {
				utf.skip(2)
			} else if flags&symbolAllScale != 0 {
				utf.skip(4)
			} else if flags&symbol2x2 != 0 {
				utf.skip(8)
			}
		}
	}
	return start, symbolSet, SymbolsCollection, SymbolsCollectionKeys
}

func (utf *utf8FontFile) parseHMTXTable(numberOfHMetrics, numSymbols int, symbolToChar map[int][]int, scale float64) {
	var widths int
	start := utf.SeekTable("hmtx")
	arrayWidths := 0
	var arr []int
	utf.CharWidths = make(map[int]int)
	charCount := 0
	arr = unpackUint16Array(utf.getRange(start, numberOfHMetrics*4))
	for symbol := 0; symbol < numberOfHMetrics; symbol++ {
		arrayWidths = arr[(symbol*2)+1]
		if _, OK := symbolToChar[symbol]; OK || symbol == 0 {

			if arrayWidths >= (1 << 15) {
				arrayWidths = 0
			}
			if symbol == 0 {
				utf.DefaultWidth = scale * float64(arrayWidths)
				continue
			}
			for _, char := range symbolToChar[symbol] {
				if char != 0 && char != 65535 {
					widths = int(math.Round(scale * float64(arrayWidths)))
					if widths == 0 {
						widths = 65535
					}
					utf.CharWidths[char] = widths
					charCount++
				}
			}
		}
	}
	diff := numSymbols - numberOfHMetrics
	for pos := 0; pos < diff; pos++ {
		symbol := pos + numberOfHMetrics
		if _, OK := symbolToChar[symbol]; OK {
			for _, char := range symbolToChar[symbol] {
				if char != 0 && char != 65535 {
					widths = int(math.Round(scale * float64(arrayWidths)))
					if widths == 0 {
						widths = 65535
					}
					utf.CharWidths[char] = widths
					charCount++
				}
			}
		}
	}
	utf.CharWidths[0] = charCount
}

func (utf *utf8FontFile) getMetrics(metricCount, gid int) []byte {
	start := utf.SeekTable("hmtx")
	var metrics []byte
	if gid < metricCount {
		utf.seek(start + (gid * 4))
		metrics = utf.fileReader.Read(4)
	} else {
		utf.seek(start + ((metricCount - 1) * 4))
		metrics = utf.fileReader.Read(2)
		utf.seek(start + (metricCount * 2) + (gid * 2))
		metrics = append(metrics, utf.fileReader.Read(2)...)
	}
	return metrics
}

func (utf *utf8FontFile) parseLOCATable(format, numSymbols int) {
	start := utf.SeekTable("loca")
	utf.symbolPosition = make([]int, 0)
	if format == 0 {
		data := utf.getRange(start, (numSymbols*2)+2)
		arr := unpackUint16Array(data)
		for n := 0; n <= numSymbols; n++ {
			utf.symbolPosition = append(utf.symbolPosition, arr[n+1]*2)
		}
	} else if format == 1 {
		data := utf.getRange(start, (numSymbols*4)+4)
		arr := unpackUint32Array(data)
		for n := 0; n <= numSymbols; n++ {
			utf.symbolPosition = append(utf.symbolPosition, arr[n+1])
		}
	} else {
		fmt.Printf("Unknown loca table format %d\n", format)
		return
	}
}

func (utf *utf8FontFile) parseCmapFormat12(offset int) (map[int][]int, map[int]int) {
	symbolCharDictionary := make(map[int][]int)
	charSymbolDictionary := make(map[int]int)

	// Seek to the start of Format 12 table
	utf.seek(offset)

	// Read Format 12 header
	format := utf.readUint16()
	if format != 12 {
		fmt.Printf("Expected CMAP format 12, got %d\n", format)
		return symbolCharDictionary, charSymbolDictionary
	}

	// Skip reserved field (uint16)
	utf.skip(2)

	// Read length (uint32)
	length := utf.readUint32()

	// Read language (uint32)
	_ = utf.readUint32()

	// Read numGroups (uint32)
	numGroups := utf.readUint32()

	// Validate length
	expectedLength := 16 + 12*numGroups
	if length != expectedLength {
		fmt.Printf("Invalid CMAP Format 12 length: got %d, expected %d\n", length, expectedLength)
		return symbolCharDictionary, charSymbolDictionary
	}

	// Parse each group
	for i := 0; i < numGroups; i++ {
		startCharCode := utf.readUint32()
		endCharCode := utf.readUint32()
		startGlyphID := utf.readUint32()

		// For each character in the range
		for charCode := startCharCode; charCode <= endCharCode; charCode++ {
			// Calculate glyph ID
			glyphID := startGlyphID + (charCode - startCharCode)

			// Add to dictionaries
			charSymbolDictionary[charCode] = glyphID
			symbolCharDictionary[glyphID] = append(symbolCharDictionary[glyphID], charCode)
		}
	}

	return symbolCharDictionary, charSymbolDictionary
}

func (utf *utf8FontFile) generateSCCSDictionaries(runeCmapPosition int, symbolCharDictionary map[int][]int, charSymbolDictionary map[int]int) {
	// Detect CMAP format (4 or 12) from table header
	format := utf.getUint16(runeCmapPosition)

	if format == 12 {
		// Format 12: Call parseCmapFormat12 and merge results
		symbolCharDict, charSymbolDict := utf.parseCmapFormat12(runeCmapPosition)

		// Merge the results into the passed-in dictionaries
		for char, symbol := range charSymbolDict {
			charSymbolDictionary[char] = symbol
		}
		for symbol, chars := range symbolCharDict {
			symbolCharDictionary[symbol] = append(symbolCharDictionary[symbol], chars...)
		}
	} else if format == 4 {
		// Format 4: Use existing parsing logic
		maxRune := 0
		utf.seek(runeCmapPosition + 2)
		size := utf.readUint16()
		rim := runeCmapPosition + size
		utf.skip(2)

		segmentSize := utf.readUint16() / 2
		utf.skip(6)
		completers := make([]int, 0)
		for i := 0; i < segmentSize; i++ {
			completers = append(completers, utf.readUint16())
		}
		utf.skip(2)
		beginners := make([]int, 0)
		for i := 0; i < segmentSize; i++ {
			beginners = append(beginners, utf.readUint16())
		}
		sizes := make([]int, 0)
		for i := 0; i < segmentSize; i++ {
			sizes = append(sizes, int(utf.readInt16()))
		}
		readerPositionStart := utf.fileReader.readerPosition
		positions := make([]int, 0)
		for i := 0; i < segmentSize; i++ {
			positions = append(positions, utf.readUint16())
		}
		var symbol int
		for n := 0; n < segmentSize; n++ {
			completePosition := completers[n] + 1
			for char := beginners[n]; char < completePosition; char++ {
				if positions[n] == 0 {
					symbol = (char + sizes[n]) & 0xFFFF
				} else {
					position := (char-beginners[n])*2 + positions[n]
					position = int(readerPositionStart) + 2*n + position
					if position >= rim {
						symbol = 0
					} else {
						symbol = utf.getUint16(position)
						if symbol != 0 {
							symbol = (symbol + sizes[n]) & 0xFFFF
						}
					}
				}
				charSymbolDictionary[char] = symbol
				maxRune = max(char, maxRune)
				symbolCharDictionary[symbol] = append(symbolCharDictionary[symbol], char)
			}
		}
	} else {
		fmt.Printf("Unsupported CMAP format: %d\n", format)
	}
}

func max(i, n int) int {
	if n > i {
		return n
	}
	return i
}

func (utf *utf8FontFile) assembleTables() []byte {
	answer := make([]byte, 0)
	tablesCount := len(utf.outTablesData)
	findSize := 1
	writer := 0
	for findSize*2 <= tablesCount {
		findSize = findSize * 2
		writer = writer + 1
	}
	findSize = findSize * 16
	rOffset := tablesCount*16 - findSize

	answer = append(answer, packHeader(0x00010000, tablesCount, findSize, writer, rOffset)...)

	tables := utf.outTablesData
	tablesNames := keySortStrings(tables)

	offset := 12 + tablesCount*16
	begin := 0

	for _, name := range tablesNames {
		if name == "head" {
			begin = offset
		}
		answer = append(answer, []byte(name)...)
		checksum := utf.generateChecksum(tables[name])
		answer = append(answer, pack2Uint16(checksum[0], checksum[1])...)
		answer = append(answer, pack2Uint32(offset, len(tables[name]))...)
		paddedLength := (len(tables[name]) + 3) &^ 3
		offset = offset + paddedLength
	}

	for _, key := range tablesNames {
		data := append([]byte{}, tables[key]...)
		data = append(data, []byte{0, 0, 0}...)
		answer = append(answer, data[:(len(data)&^3)]...)
	}

	checksum := utf.generateChecksum([]byte(answer))
	checksum = utf.calcInt32([]int{0xB1B0, 0xAFBA}, checksum)
	answer = utf.splice(answer, (begin + 8), pack2Uint16(checksum[0], checksum[1]))
	return answer
}

func unpackUint16Array(data []byte) []int {
	answer := make([]int, 1)
	r := bytes.NewReader(data)
	bs := make([]byte, 2)
	var e error
	var c int
	c, e = r.Read(bs)
	for e == nil && c > 0 {
		answer = append(answer, int(binary.BigEndian.Uint16(bs)))
		c, e = r.Read(bs)
	}
	return answer
}

func unpackUint32Array(data []byte) []int {
	answer := make([]int, 1)
	r := bytes.NewReader(data)
	bs := make([]byte, 4)
	var e error
	var c int
	c, e = r.Read(bs)
	for e == nil && c > 0 {
		answer = append(answer, int(binary.BigEndian.Uint32(bs)))
		c, e = r.Read(bs)
	}
	return answer
}

func unpackUint16(data []byte) int {
	return int(binary.BigEndian.Uint16(data))
}

func packHeader(N uint32, n1, n2, n3, n4 int) []byte {
	answer := make([]byte, 0)
	bs4 := make([]byte, 4)
	binary.BigEndian.PutUint32(bs4, N)
	answer = append(answer, bs4...)
	bs := make([]byte, 2)
	binary.BigEndian.PutUint16(bs, uint16(n1))
	answer = append(answer, bs...)
	binary.BigEndian.PutUint16(bs, uint16(n2))
	answer = append(answer, bs...)
	binary.BigEndian.PutUint16(bs, uint16(n3))
	answer = append(answer, bs...)
	binary.BigEndian.PutUint16(bs, uint16(n4))
	answer = append(answer, bs...)
	return answer
}

func pack2Uint16(n1, n2 int) []byte {
	answer := make([]byte, 0)
	bs := make([]byte, 2)
	binary.BigEndian.PutUint16(bs, uint16(n1))
	answer = append(answer, bs...)
	binary.BigEndian.PutUint16(bs, uint16(n2))
	answer = append(answer, bs...)
	return answer
}

func pack2Uint32(n1, n2 int) []byte {
	answer := make([]byte, 0)
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, uint32(n1))
	answer = append(answer, bs...)
	binary.BigEndian.PutUint32(bs, uint32(n2))
	answer = append(answer, bs...)
	return answer
}

func packUint32(n1 int) []byte {
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, uint32(n1))
	return bs
}

func packUint16(n1 int) []byte {
	bs := make([]byte, 2)
	binary.BigEndian.PutUint16(bs, uint16(n1))
	return bs
}

func keySortStrings(s map[string][]byte) []string {
	keys := make([]string, len(s))
	i := 0
	for key := range s {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return keys
}

func keySortInt(s map[int]int) []int {
	keys := make([]int, len(s))
	i := 0
	for key := range s {
		keys[i] = key
		i++
	}
	sort.Ints(keys)
	return keys
}

func keySortArrayRangeMap(s map[int][]int) []int {
	keys := make([]int, len(s))
	i := 0
	for key := range s {
		keys[i] = key
		i++
	}
	sort.Ints(keys)
	return keys
}

// UTF8CutFont is a utility function that generates a TrueType font composed
// only of the runes included in cutset. The rune glyphs are copied from This
// function is demonstrated in ExampleUTF8CutFont().
func UTF8CutFont(inBuf []byte, cutset string) (outBuf []byte) {
	f := newUTF8Font(&fileReader{readerPosition: 0, array: inBuf})
	runes := map[int]int{}
	for i, r := range cutset {
		runes[i] = int(r)
	}
	outBuf = f.GenerateCutFont(runes)
	return
}

// cmapGroup represents a sequential character range mapping in CMAP Format 12.
// Format 12 uses groups to efficiently map character codes to glyph IDs.
type cmapGroup struct {
	startCharCode uint32 // First character code in this group
	endCharCode   uint32 // Last character code in this group
	startGlyphID  uint32 // Glyph ID for startCharCode (subsequent IDs are sequential)
}

// buildCmapGroups creates optimized sequential groups from a character-to-glyph mapping.
// It combines consecutive character codes with consecutive glyph IDs into single groups
// to minimize the number of groups in the CMAP table.
func buildCmapGroups(cidToGlyph map[int]int) []cmapGroup {
	if len(cidToGlyph) == 0 {
		return []cmapGroup{}
	}

	// Get sorted list of character codes
	cids := make([]int, 0, len(cidToGlyph))
	for cid := range cidToGlyph {
		cids = append(cids, cid)
	}
	sort.Ints(cids)

	groups := make([]cmapGroup, 0)

	// Start first group
	startCid := cids[0]
	endCid := cids[0]
	startGid := cidToGlyph[cids[0]]

	for i := 1; i < len(cids); i++ {
		cid := cids[i]
		gid := cidToGlyph[cid]

		// Check if this character continues the current group
		// (consecutive character codes with consecutive glyph IDs)
		if cid == endCid+1 && gid == cidToGlyph[endCid]+1 {
			endCid = cid
		} else {
			// End current group and start a new one
			groups = append(groups, cmapGroup{
				startCharCode: uint32(startCid),
				endCharCode:   uint32(endCid),
				startGlyphID:  uint32(startGid),
			})
			startCid = cid
			endCid = cid
			startGid = gid
		}
	}

	// Add the final group
	groups = append(groups, cmapGroup{
		startCharCode: uint32(startCid),
		endCharCode:   uint32(endCid),
		startGlyphID:  uint32(startGid),
	})

	return groups
}

// writeCmapFormat12Header generates the 16-byte header for a CMAP Format 12 table.
// Format 12 header structure (all big-endian):
//   - format (uint16): 12
//   - reserved (uint16): 0
//   - length (uint32): total table size in bytes (16 + 12*numGroups)
//   - language (uint32): 0 (language independent)
//   - numGroups (uint32): number of sequential map groups
func writeCmapFormat12Header(numGroups uint32) []byte {
	header := make([]byte, 0, 16)

	// Format: 12 (uint16)
	header = append(header, packUint16(12)...)

	// Reserved: 0 (uint16)
	header = append(header, packUint16(0)...)

	// Length: 16 + 12*numGroups (uint32)
	length := int(16 + 12*numGroups)
	header = append(header, packUint32(length)...)

	// Language: 0 (uint32)
	header = append(header, packUint32(0)...)

	// NumGroups (uint32)
	header = append(header, packUint32(int(numGroups))...)

	return header
}

// generateToUnicodeCMap creates a dynamic ToUnicode CMap based on the actual used runes.
// This function generates a CMap that maps character codes (CIDs) to Unicode values.
// It supports both 2-byte (BMP) and 4-byte (supplementary plane) character codes.
//
// The function detects the maximum character code to determine the codespace range:
// - If max <= 0xFFFF: uses 2-byte codespace <0000> <FFFF>
// - If max > 0xFFFF: uses 4-byte codespace <00000000> <0010FFFF>
//
// For efficiency, consecutive character codes are combined into ranges using bfrange.
// Identity mapping is used (CID == Unicode value).
func generateToUnicodeCMap(cidToUnicode map[int]int) string {
	if len(cidToUnicode) == 0 {
		return `/CIDInit /ProcSet findresource begin
12 dict begin
begincmap
/CIDSystemInfo
<</Registry (Adobe)
/Ordering (UCS)
/Supplement 0
>> def
/CMapName /Adobe-Identity-UCS def
/CMapType 2 def
1 begincodespacerange
<0000> <FFFF>
endcodespacerange
0 beginbfchar
endbfchar
endcmap
CMapName currentdict /CMap defineresource pop
end
end`
	}

	maxCID := 0
	for cid := range cidToUnicode {
		if cid > maxCID {
			maxCID = cid
		}
	}

	use4Byte := maxCID > 0xFFFF

	cids := make([]int, 0, len(cidToUnicode))
	for cid := range cidToUnicode {
		if cid == 0 {
			continue
		}
		cids = append(cids, cid)
	}
	sort.Ints(cids)

	total := len(cids)
	if total == 0 {
		return `/CIDInit /ProcSet findresource begin
12 dict begin
begincmap
/CIDSystemInfo
<</Registry (Adobe)
/Ordering (UCS)
/Supplement 0
>> def
/CMapName /Adobe-Identity-UCS def
/CMapType 2 def
1 begincodespacerange
<0000> <FFFF>
endcodespacerange
0 beginbfchar
endbfchar
endcmap
CMapName currentdict /CMap defineresource pop
end
end`
	}

	var cmap bytes.Buffer
	cmap.WriteString(`/CIDInit /ProcSet findresource begin
12 dict begin
begincmap
/CIDSystemInfo
<</Registry (Adobe)
/Ordering (UCS)
/Supplement 0
>> def
/CMapName /Adobe-Identity-UCS def
/CMapType 2 def
`)

	cmap.WriteString("1 begincodespacerange\n")
	if use4Byte {
		cmap.WriteString("<00000000> <0010FFFF>\n")
	} else {
		cmap.WriteString("<0000> <FFFF>\n")
	}
	cmap.WriteString("endcodespacerange\n")

	blockSize := 100
	for i := 0; i < total; i += blockSize {
		end := i + blockSize
		if end > total {
			end = total
		}
		cmap.WriteString(fmt.Sprintf("%d beginbfchar\n", end-i))
		for _, cid := range cids[i:end] {
			unicode := cidToUnicode[cid]
			cmap.WriteString(fmt.Sprintf("<%s> <%s>\n", formatCIDHex(cid, use4Byte), formatUnicodeHex(unicode)))
		}
		cmap.WriteString("endbfchar\n")
	}

	cmap.WriteString(`endcmap
CMapName currentdict /CMap defineresource pop
end
end`)

	return cmap.String()
}

func formatCIDHex(cid int, use4Byte bool) string {
	if use4Byte {
		return fmt.Sprintf("%08X", cid)
	}
	return fmt.Sprintf("%04X", cid)
}

func formatUnicodeHex(codepoint int) string {
	if codepoint <= 0xFFFF {
		return fmt.Sprintf("%04X", codepoint)
	}
	if codepoint > 0x10FFFF {
		return "0000"
	}
	codepoint -= 0x10000
	high := 0xD800 + ((codepoint >> 10) & 0x3FF)
	low := 0xDC00 + (codepoint & 0x3FF)
	return fmt.Sprintf("%04X%04X", high, low)
}
