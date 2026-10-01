// Package xlsx writes simple multi-sheet spreadsheets (Office Open XML) without
// external dependencies. Each sheet has a styled, frozen header row, optional
// filters, thin borders, per-column kinds (label, wrapped text, centered text,
// integers, one-decimal numbers) and column widths fitted to the content. Text
// goes in as inline strings, so a value that starts with "=" is shown as text and
// is never evaluated as a formula.
package xlsx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Kind decides how a column is formatted.
type Kind int

const (
	// KindText is plain text, left-aligned (the default).
	KindText Kind = iota
	// KindLabel is bold text on a light fill, for a column of field names.
	KindLabel
	// KindWrap is long text that wraps inside the cell.
	KindWrap
	// KindCenter is short text centered in the cell, such as dates and times.
	KindCenter
	// KindInt is a whole number, centered.
	KindInt
	// KindDecimal is a number shown with one decimal place, centered.
	KindDecimal
)

// Sheet is one worksheet. Row values may be string, int, int16, int32, int64,
// float64 or nil (empty cell). Kinds, when set, must have one entry per column.
type Sheet struct {
	Name   string
	Header []string
	Kinds  []Kind
	Rows   [][]any
	// Filter adds filter buttons to the header row.
	Filter bool
}

const (
	maxSheetNameLength = 31
	minColumnWidth     = 10
	maxTextWidth       = 48
	maxWrapWidth       = 60
	maxHeaderWidth     = 18
	lineHeight         = 15.0
)

// Cell format indexes, in the order they appear in cellXfs.
const (
	styleBase = iota
	styleHeader
	styleText
	styleLabel
	styleWrap
	styleCenter
	styleInt
	styleDecimal
	styleNumberLeft
)

// Write encodes the sheets as an .xlsx file.
func Write(w io.Writer, sheets []Sheet) error {
	if len(sheets) == 0 {
		return fmt.Errorf("xlsx: at least one sheet is required")
	}
	seen := map[string]bool{}
	for _, sheet := range sheets {
		if err := validateName(sheet.Name); err != nil {
			return err
		}
		key := strings.ToLower(sheet.Name)
		if seen[key] {
			return fmt.Errorf("xlsx: duplicate sheet name %q", sheet.Name)
		}
		seen[key] = true
		if sheet.Kinds != nil && len(sheet.Kinds) != len(sheet.Header) {
			return fmt.Errorf("xlsx: sheet %q has %d kinds for %d columns", sheet.Name, len(sheet.Kinds), len(sheet.Header))
		}
	}

	type part struct{ name, body string }
	parts := []part{
		{"[Content_Types].xml", contentTypes(len(sheets))},
		{"_rels/.rels", rootRels},
		{"xl/workbook.xml", workbook(sheets)},
		{"xl/_rels/workbook.xml.rels", workbookRels(len(sheets))},
		{"xl/styles.xml", styles},
	}
	for i, sheet := range sheets {
		body, err := worksheet(sheet)
		if err != nil {
			return err
		}
		parts = append(parts, part{fmt.Sprintf("xl/worksheets/sheet%d.xml", i+1), body})
	}

	archive := zip.NewWriter(w)
	for _, p := range parts {
		entry, err := archive.Create(p.name)
		if err != nil {
			return err
		}
		if _, err := io.WriteString(entry, p.body); err != nil {
			return err
		}
	}
	return archive.Close()
}

func validateName(name string) error {
	if name == "" || utf8.RuneCountInString(name) > maxSheetNameLength {
		return fmt.Errorf("xlsx: sheet name %q must have 1 to %d characters", name, maxSheetNameLength)
	}
	if strings.ContainsAny(name, `[]:*?/\`) || strings.HasPrefix(name, "'") || strings.HasSuffix(name, "'") {
		return fmt.Errorf("xlsx: sheet name %q has characters Excel does not accept", name)
	}
	return nil
}

const xmlHeader = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n"

const rootRels = xmlHeader + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`

// Fonts: regular, bold, bold white. Fills: none, gray125 (both mandatory), dark
// header, light label. Border 1 is a thin light-gray box. Custom format 164 is "0.0".
const styles = xmlHeader + `<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">` +
	`<numFmts count="1"><numFmt numFmtId="164" formatCode="0.0"/></numFmts>` +
	`<fonts count="3"><font><sz val="11"/><color rgb="FF1F231C"/><name val="Calibri"/></font>` +
	`<font><b/><sz val="11"/><color rgb="FF1F231C"/><name val="Calibri"/></font>` +
	`<font><b/><sz val="11"/><color rgb="FFFFFFFF"/><name val="Calibri"/></font></fonts>` +
	`<fills count="4"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill>` +
	`<fill><patternFill patternType="solid"><fgColor rgb="FF3D4A36"/><bgColor indexed="64"/></patternFill></fill>` +
	`<fill><patternFill patternType="solid"><fgColor rgb="FFE9EEE2"/><bgColor indexed="64"/></patternFill></fill></fills>` +
	`<borders count="2"><border><left/><right/><top/><bottom/><diagonal/></border>` +
	`<border><left style="thin"><color rgb="FFC3CBBA"/></left><right style="thin"><color rgb="FFC3CBBA"/></right><top style="thin"><color rgb="FFC3CBBA"/></top><bottom style="thin"><color rgb="FFC3CBBA"/></bottom><diagonal/></border></borders>` +
	`<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>` +
	`<cellXfs count="9">` +
	`<xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>` +
	`<xf numFmtId="0" fontId="2" fillId="2" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1" applyAlignment="1"><alignment horizontal="center" vertical="center" wrapText="1"/></xf>` +
	`<xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyBorder="1" applyAlignment="1"><alignment horizontal="left" vertical="center"/></xf>` +
	`<xf numFmtId="0" fontId="1" fillId="3" borderId="1" xfId="0" applyFont="1" applyFill="1" applyBorder="1" applyAlignment="1"><alignment horizontal="left" vertical="center"/></xf>` +
	`<xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyBorder="1" applyAlignment="1"><alignment horizontal="left" vertical="center" wrapText="1"/></xf>` +
	`<xf numFmtId="0" fontId="0" fillId="0" borderId="1" xfId="0" applyBorder="1" applyAlignment="1"><alignment horizontal="center" vertical="center"/></xf>` +
	`<xf numFmtId="1" fontId="0" fillId="0" borderId="1" xfId="0" applyNumberFormat="1" applyBorder="1" applyAlignment="1"><alignment horizontal="center" vertical="center"/></xf>` +
	`<xf numFmtId="164" fontId="0" fillId="0" borderId="1" xfId="0" applyNumberFormat="1" applyBorder="1" applyAlignment="1"><alignment horizontal="center" vertical="center"/></xf>` +
	`<xf numFmtId="1" fontId="0" fillId="0" borderId="1" xfId="0" applyNumberFormat="1" applyBorder="1" applyAlignment="1"><alignment horizontal="left" vertical="center"/></xf>` +
	`</cellXfs>` +
	`<cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles></styleSheet>`

func contentTypes(sheets int) string {
	var b strings.Builder
	b.WriteString(xmlHeader)
	b.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/>`)
	b.WriteString(`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>`)
	b.WriteString(`<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>`)
	for i := 1; i <= sheets; i++ {
		fmt.Fprintf(&b, `<Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`, i)
	}
	b.WriteString(`</Types>`)
	return b.String()
}

func workbook(sheets []Sheet) string {
	var b strings.Builder
	b.WriteString(xmlHeader)
	b.WriteString(`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>`)
	for i, sheet := range sheets {
		fmt.Fprintf(&b, `<sheet name="%s" sheetId="%d" r:id="rId%d"/>`, escapeText(sheet.Name), i+1, i+1)
	}
	b.WriteString(`</sheets></workbook>`)
	return b.String()
}

func workbookRels(sheets int) string {
	var b strings.Builder
	b.WriteString(xmlHeader)
	b.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for i := 1; i <= sheets; i++ {
		fmt.Fprintf(&b, `<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet%d.xml"/>`, i, i)
	}
	fmt.Fprintf(&b, `<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`, sheets+1)
	b.WriteString(`</Relationships>`)
	return b.String()
}

func kindOf(sheet Sheet, column int) Kind {
	if sheet.Kinds == nil {
		return KindText
	}
	return sheet.Kinds[column]
}

func textStyle(kind Kind) int {
	switch kind {
	case KindLabel:
		return styleLabel
	case KindWrap:
		return styleWrap
	case KindCenter:
		return styleCenter
	}
	return styleText
}

func numberStyle(kind Kind) int {
	switch kind {
	case KindInt:
		return styleInt
	case KindDecimal:
		return styleDecimal
	case KindCenter:
		return styleInt
	}
	return styleNumberLeft
}

func columnWidths(sheet Sheet) []float64 {
	widths := make([]float64, len(sheet.Header))
	for i, title := range sheet.Header {
		// Headers wrap onto a second line, so they only need about half their length.
		widths[i] = float64(min(utf8.RuneCountInString(title), maxHeaderWidth))
	}
	for _, row := range sheet.Rows {
		for i, value := range row {
			length := 0
			switch v := value.(type) {
			case nil:
			case string:
				length = longestLine(v)
			default:
				length = len(fmt.Sprint(v))
			}
			if float64(length) > widths[i] {
				widths[i] = float64(length)
			}
		}
	}
	for i := range widths {
		limit := float64(maxTextWidth)
		if kindOf(sheet, i) == KindWrap {
			limit = maxWrapWidth
		}
		widths[i] = min(max(widths[i]+3, minColumnWidth), limit)
	}
	return widths
}

// wrappedLines estimates how many lines a wrapped cell needs at the given width.
func wrappedLines(text string, width float64) int {
	usable := max(int(width)-2, 1)
	lines := 0
	for _, line := range strings.Split(text, "\n") {
		lines += max((utf8.RuneCountInString(line)+usable-1)/usable, 1)
	}
	return lines
}

func worksheet(sheet Sheet) (string, error) {
	columns := len(sheet.Header)
	for _, row := range sheet.Rows {
		if len(row) != columns {
			return "", fmt.Errorf("xlsx: sheet %q has a row with %d cells, want %d", sheet.Name, len(row), columns)
		}
	}
	widths := columnWidths(sheet)

	var b strings.Builder
	b.WriteString(xmlHeader)
	b.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">`)
	if columns > 0 {
		b.WriteString(`<sheetViews><sheetView showGridLines="0" workbookViewId="0"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews>`)
		b.WriteString(`<sheetFormatPr defaultRowHeight="18" customHeight="1"/>`)
		b.WriteString(`<cols>`)
		for i, width := range widths {
			fmt.Fprintf(&b, `<col min="%d" max="%d" width="%s" customWidth="1"/>`, i+1, i+1, strconv.FormatFloat(width, 'f', 1, 64))
		}
		b.WriteString(`</cols>`)
	}
	b.WriteString(`<sheetData>`)
	if columns > 0 {
		headerLines := 1
		for i, title := range sheet.Header {
			headerLines = max(headerLines, wrappedLines(title, widths[i]))
		}
		fmt.Fprintf(&b, `<row r="1" ht="%s" customHeight="1">`, strconv.FormatFloat(max(24, float64(headerLines)*lineHeight+9), 'f', 1, 64))
		for i, title := range sheet.Header {
			writeText(&b, cellRef(i, 1), title, styleHeader)
		}
		b.WriteString(`</row>`)
	}
	for r, row := range sheet.Rows {
		rowLines := 1
		for c, value := range row {
			if text, ok := value.(string); ok && kindOf(sheet, c) == KindWrap {
				rowLines = max(rowLines, wrappedLines(text, widths[c]))
			}
		}
		if rowLines > 1 {
			fmt.Fprintf(&b, `<row r="%d" ht="%s" customHeight="1">`, r+2, strconv.FormatFloat(float64(rowLines)*lineHeight+5, 'f', 1, 64))
		} else {
			fmt.Fprintf(&b, `<row r="%d">`, r+2)
		}
		for c, value := range row {
			ref := cellRef(c, r+2)
			kind := kindOf(sheet, c)
			switch v := value.(type) {
			case nil:
				fmt.Fprintf(&b, `<c r="%s" s="%d"/>`, ref, textStyle(kind))
			case string:
				if v == "" {
					fmt.Fprintf(&b, `<c r="%s" s="%d"/>`, ref, textStyle(kind))
				} else {
					writeText(&b, ref, v, textStyle(kind))
				}
			case int:
				writeNumber(&b, ref, strconv.FormatInt(int64(v), 10), numberStyle(kind))
			case int16:
				writeNumber(&b, ref, strconv.FormatInt(int64(v), 10), numberStyle(kind))
			case int32:
				writeNumber(&b, ref, strconv.FormatInt(int64(v), 10), numberStyle(kind))
			case int64:
				writeNumber(&b, ref, strconv.FormatInt(v, 10), numberStyle(kind))
			case float64:
				writeNumber(&b, ref, strconv.FormatFloat(v, 'f', -1, 64), numberStyle(kind))
			default:
				return "", fmt.Errorf("xlsx: unsupported cell type %T in sheet %q", value, sheet.Name)
			}
		}
		b.WriteString(`</row>`)
	}
	b.WriteString(`</sheetData>`)
	if sheet.Filter && columns > 0 {
		fmt.Fprintf(&b, `<autoFilter ref="A1:%s"/>`, cellRef(columns-1, max(len(sheet.Rows)+1, 1)))
	}
	b.WriteString(`</worksheet>`)
	return b.String(), nil
}

func writeText(b *strings.Builder, ref, text string, style int) {
	fmt.Fprintf(b, `<c r="%s" t="inlineStr" s="%d"><is><t xml:space="preserve">`, ref, style)
	b.WriteString(escapeText(text))
	b.WriteString(`</t></is></c>`)
}

func writeNumber(b *strings.Builder, ref, number string, style int) {
	fmt.Fprintf(b, `<c r="%s" s="%d"><v>%s</v></c>`, ref, style, number)
}

// cellRef turns a zero-based column and a one-based row into "A1" notation.
func cellRef(column, row int) string {
	name := ""
	for n := column + 1; n > 0; n = (n - 1) / 26 {
		name = string(rune('A'+(n-1)%26)) + name
	}
	return name + strconv.Itoa(row)
}

func longestLine(text string) int {
	longest := 0
	for _, line := range strings.Split(text, "\n") {
		if length := utf8.RuneCountInString(line); length > longest {
			longest = length
		}
	}
	return longest
}

// escapeText drops characters XML 1.0 cannot carry and escapes the rest.
func escapeText(text string) string {
	cleaned := strings.Map(func(r rune) rune {
		switch {
		case r == '\t' || r == '\n' || r == '\r':
			return r
		case r < 0x20 || r == 0xFFFE || r == 0xFFFF || (r >= 0xD800 && r <= 0xDFFF):
			return -1
		}
		return r
	}, text)
	var buffer bytes.Buffer
	_ = xml.EscapeText(&buffer, []byte(cleaned))
	return buffer.String()
}
