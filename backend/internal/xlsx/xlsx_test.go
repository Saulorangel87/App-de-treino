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
	for _, want := range []string{`<c r="B2" s="8"><v>65</v></c>`, `<c r="C2" s="8"><v>32.5</v></c>`, `<c r="C3" s="8"><v>7</v></c>`, `state="frozen"`} {
		if !strings.Contains(sheet, want) {
			t.Fatalf("sheet is missing %s:\n%s", want, sheet)
		}
	}
	if strings.Contains(sheet, `r="B3"><v>`) || !strings.Contains(sheet, `<c r="B3" s="2"/>`) {
		t.Fatal("a nil value must leave a bordered, empty cell")
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
		"tipos a menos":    {{Name: "A", Header: []string{"a", "b"}, Kinds: []Kind{KindText}}},
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

func TestKindsControlStyleWrappingAndFilter(t *testing.T) {
	long := strings.Repeat("palavra ", 30)
	files := build(t, []Sheet{{
		Name:   "Dados",
		Header: []string{"Campo", "Valor", "Quantidade", "Distância (km)", "Data"},
		Kinds:  []Kind{KindLabel, KindWrap, KindInt, KindDecimal, KindCenter},
		Filter: true,
		Rows:   [][]any{{"Nome", long, 3, 32.5, "30/09/2026"}, {"Outro", "curto", nil, nil, ""}},
	}})
	sheet := files["xl/worksheets/sheet1.xml"]
	for name, want := range map[string]string{
		"cabeçalho estilizado": `<c r="A1" t="inlineStr" s="1">`,
		"rótulo":               `<c r="A2" t="inlineStr" s="3">`,
		"texto longo quebra":   `<c r="B2" t="inlineStr" s="4">`,
		"inteiro centralizado": `<c r="C2" s="6"><v>3</v></c>`,
		"decimal":              `<c r="D2" s="7"><v>32.5</v></c>`,
		"data centralizada":    `<c r="E2" t="inlineStr" s="5">`,
		"filtro":               `<autoFilter ref="A1:E3"/>`,
		"sem linhas de grade":  `showGridLines="0"`,
		"vazio com borda":      `<c r="E3" s="5"/>`,
	} {
		if !strings.Contains(sheet, want) {
			t.Errorf("%s: missing %s\n%s", name, want, sheet)
		}
	}
	if !strings.Contains(sheet, `<row r="2" ht=`) {
		t.Error("a row with long wrapped text needs an explicit height")
	}
	if strings.Contains(sheet, `<row r="3" ht=`) {
		t.Error("a short row keeps the default height")
	}
	if !strings.Contains(files["xl/styles.xml"], `formatCode="0.0"`) {
		t.Error("the one-decimal number format is missing")
	}
}
