// Package xlsx writes simple multi-sheet spreadsheets (Office Open XML) without
// external dependencies: a bold, frozen header row, text and numeric cells and
// column widths fitted to the content. Text goes in as inline strings, so a value
// that starts with "=" is shown as text and is never evaluated as a formula.
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

// Sheet is one worksheet. Row values may be string, int, int16, int32, int64,
// float64 or nil (empty cell).
type Sheet struct {
	Name   string
	Header []string
	Rows   [][]any
}

const (
	maxSheetNameLength = 31
	maxColumnWidth     = 60
	minColumnWidth     = 8
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

// Two fonts (regular, bold), the two mandatory fills plus a light one, one border,
// and two cell formats: 0 regular, 1 bold header on a light fill.
const styles = xmlHeader + `<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">` +
	`<fonts count="2"><font><sz val="11"/><name val="Calibri"/></font><font><b/><sz val="11"/><name val="Calibri"/></font></fonts>` +
	`<fills count="3"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill><fill><patternFill patternType="solid"><fgColor rgb="FFE2E8DB"/><bgColor indexed="64"/></patternFill></fill></fills>` +
	`<borders count="1"><border><left/><right/><top/><bottom/><diagonal/></border></borders>` +
	`<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>` +
	`<cellXfs count="2"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/><xf numFmtId="0" fontId="1" fillId="2" borderId="0" xfId="0" applyFont="1" applyFill="1"/></cellXfs>` +
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

func worksheet(sheet Sheet) (string, error) {
	columns := len(sheet.Header)
	widths := make([]int, columns)
	for i, title := range sheet.Header {
		widths[i] = utf8.RuneCountInString(title)
	}
	for _, row := range sheet.Rows {
		if len(row) != columns {
			return "", fmt.Errorf("xlsx: sheet %q has a row with %d cells, want %d", sheet.Name, len(row), columns)
		}
		for i, value := range row {
			switch v := value.(type) {
			case nil:
			case string:
				if length := longestLine(v); length > widths[i] {
					widths[i] = length
				}
			default:
				if length := len(fmt.Sprint(v)); length > widths[i] {
					widths[i] = length
				}
			}
		}
	}

	var b strings.Builder
	b.WriteString(xmlHeader)
	b.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">`)
	if columns > 0 {
		b.WriteString(`<sheetViews><sheetView workbookViewId="0"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/></sheetView></sheetViews>`)
		b.WriteString(`<cols>`)
		for i, width := range widths {
			fitted := min(max(width+2, minColumnWidth), maxColumnWidth)
			fmt.Fprintf(&b, `<col min="%d" max="%d" width="%d" customWidth="1"/>`, i+1, i+1, fitted)
		}
		b.WriteString(`</cols>`)
	}
	b.WriteString(`<sheetData>`)
	if columns > 0 {
		b.WriteString(`<row r="1">`)
		for i, title := range sheet.Header {
			writeText(&b, cellRef(i, 1), title, 1)
		}
		b.WriteString(`</row>`)
	}
	for r, row := range sheet.Rows {
		fmt.Fprintf(&b, `<row r="%d">`, r+2)
		for c, value := range row {
			ref := cellRef(c, r+2)
			switch v := value.(type) {
			case nil:
			case string:
				if v != "" {
					writeText(&b, ref, v, 0)
				}
			case int:
				writeNumber(&b, ref, strconv.FormatInt(int64(v), 10))
			case int16:
				writeNumber(&b, ref, strconv.FormatInt(int64(v), 10))
			case int32:
				writeNumber(&b, ref, strconv.FormatInt(int64(v), 10))
			case int64:
				writeNumber(&b, ref, strconv.FormatInt(v, 10))
			case float64:
				writeNumber(&b, ref, strconv.FormatFloat(v, 'f', -1, 64))
			default:
				return "", fmt.Errorf("xlsx: unsupported cell type %T in sheet %q", value, sheet.Name)
			}
		}
		b.WriteString(`</row>`)
	}
	b.WriteString(`</sheetData></worksheet>`)
	return b.String(), nil
}

func writeText(b *strings.Builder, ref, text string, style int) {
	fmt.Fprintf(b, `<c r="%s" t="inlineStr"`, ref)
	if style != 0 {
		fmt.Fprintf(b, ` s="%d"`, style)
	}
	b.WriteString(`><is><t xml:space="preserve">`)
	b.WriteString(escapeText(text))
	b.WriteString(`</t></is></c>`)
}

func writeNumber(b *strings.Builder, ref, number string) {
	fmt.Fprintf(b, `<c r="%s"><v>%s</v></c>`, ref, number)
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
