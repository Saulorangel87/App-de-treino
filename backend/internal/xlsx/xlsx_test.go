package xlsx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"
)

func build(t *testing.T, sheets []Sheet) map[string]string {
	t.Helper()
	var buffer bytes.Buffer
	if err := Write(&buffer, sheets); err != nil {
		t.Fatalf("write: %v", err)
	}
	reader, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(buffer.Len()))
	if err != nil {
		t.Fatalf("the output is not a zip archive: %v", err)
	}
	files := map[string]string{}
	for _, file := range reader.File {
		handle, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(handle)
		handle.Close()
		files[file.Name] = string(body)
		decoder := xml.NewDecoder(strings.NewReader(string(body)))
		for {
			if _, err := decoder.Token(); err == io.EOF {
				break
			} else if err != nil {
				t.Fatalf("%s is not well-formed XML: %v", file.Name, err)
			}
		}
	}
	return files
}

func TestWriteProducesAWellFormedWorkbook(t *testing.T) {
	files := build(t, []Sheet{
		{Name: "Conta", Header: []string{"Campo", "Valor"}, Rows: [][]any{{"Nome", "Ana"}}},
		{Name: "Treinos realizados", Header: []string{"Data", "Duração (min)", "Distância (km)"}, Rows: [][]any{{"2026-09-30", 65, 32.5}, {"2026-10-01", nil, int64(7)}}},
	})
	for _, name := range []string{"[Content_Types].xml", "_rels/.rels", "xl/workbook.xml", "xl/_rels/workbook.xml.rels", "xl/styles.xml", "xl/worksheets/sheet1.xml", "xl/worksheets/sheet2.xml"} {
		if _, ok := files[name]; !ok {
			t.Fatalf("missing part %s", name)
		}
	}
	if !strings.Contains(files["xl/workbook.xml"], `name="Treinos realizados"`) {
		t.Fatal("sheet names are missing from the workbook")
	}
	sheet := files["xl/worksheets/sheet2.xml"]
	for _, want := range []string{`<c r="B2"><v>65</v></c>`, `<c r="C2"><v>32.5</v></c>`, `<c r="C3"><v>7</v></c>`, `state="frozen"`} {
		if !strings.Contains(sheet, want) {
			t.Fatalf("sheet is missing %s:\n%s", want, sheet)
		}
	}
	if strings.Contains(sheet, `r="B3"`) {
		t.Fatal("a nil value must leave the cell empty")
	}
}

func TestWriteEscapesTextAndNeverCreatesFormulas(t *testing.T) {
	files := build(t, []Sheet{{Name: "Dados", Header: []string{"Texto"}, Rows: [][]any{
		{`<script>&"x"</script>`}, {`=HYPERLINK("http://x")`}, {"controle\x00\x07 ok"},
	}}})
	sheet := files["xl/worksheets/sheet1.xml"]
	if strings.Contains(sheet, "<script>") || !strings.Contains(sheet, "&lt;script&gt;") {
		t.Fatalf("markup was not escaped:\n%s", sheet)
	}
	if strings.Contains(sheet, "<f>") || !strings.Contains(sheet, `t="inlineStr"`) {
		t.Fatal("text that starts with = must stay text")
	}
	if strings.ContainsRune(sheet, 0) || strings.Contains(sheet, "\x07") || !strings.Contains(sheet, "controle ok") {
		t.Fatal("control characters must be removed")
	}
}

func TestWriteRejectsInvalidInput(t *testing.T) {
	cases := map[string][]Sheet{
		"sem abas":         nil,
		"nome vazio":       {{Name: "", Header: []string{"a"}}},
		"nome inválido":    {{Name: "a/b", Header: []string{"a"}}},
		"nome longo":       {{Name: strings.Repeat("a", 32), Header: []string{"a"}}},
		"nome repetido":    {{Name: "A", Header: []string{"a"}}, {Name: "a", Header: []string{"a"}}},
		"linha torta":      {{Name: "A", Header: []string{"a", "b"}, Rows: [][]any{{"1"}}}},
		"tipo sem suporte": {{Name: "A", Header: []string{"a"}, Rows: [][]any{{struct{}{}}}}},
	}
	for name, sheets := range cases {
		if err := Write(io.Discard, sheets); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestCellRefUsesSpreadsheetColumnNames(t *testing.T) {
	for column, want := range map[int]string{0: "A1", 25: "Z1", 26: "AA1", 27: "AB1", 701: "ZZ1", 702: "AAA1"} {
		if got := cellRef(column, 1); got != want {
			t.Errorf("column %d: got %s, want %s", column, got, want)
		}
	}
}
