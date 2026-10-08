package i18n

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestLanguageComesOnlyFromAnExplicitEnglishTag(t *testing.T) {
	for header, want := range map[string]Language{
		"":                        Portuguese,
		"pt-BR":                   Portuguese,
		"en":                      English,
		"en-US,en;q=0.9,pt;q=0.8": English,
		"EN-gb":                   English,
		"pt-BR,pt;q=0.9,en;q=0.8": Portuguese,
		"fr":                      Portuguese,
		"english":                 Portuguese,
	} {
		if got := FromAcceptLanguage(header); got != want {
			t.Errorf("FromAcceptLanguage(%q) = %q, want %q", header, got, want)
		}
	}
	if FromContext(context.Background()) != Portuguese {
		t.Fatal("a request without a language must stay in Portuguese")
	}
	if FromContext(WithLanguage(context.Background(), English)) != English {
		t.Fatal("the language set on the context must be read back")
	}
}

func TestTranslateCoversTheWaysTheEngineBuildsText(t *testing.T) {
	cases := map[string]string{
		"Giro de base": "Base ride",
		// A phrase reused as a sentence gets a final period.
		"Ritmo confortável e contínuo.": "Comfortable, continuous pace.",
		// Formatted text, with the inner label translated too.
		"Carga compatível com experiência avançada.":                                     "Load suited to advanced experience.",
		"Agendado em um dia com 90 minutos disponíveis.":                                 "Scheduled on a day with 90 minutes available.",
		"Bloco de limiar 2 de 3":                                                         "Threshold block 2 of 3",
		"Preferência por intervalos curtos considerada dentro dos limites de segurança.": "Preference for short intervals considered within the safety limits.",
		"Z2 · Resistência":                                                               "Z2 · Endurance",
		// A summary with an appended sentence.
		"Sessão de base para acumular consistência com baixo custo de recuperação. O volume foi reduzido para a janela pré-prova, sem aumentar a intensidade nem a frequência planejada.": "A base session to build consistency at a low recovery cost. Volume was reduced for the pre-race window, without increasing intensity or the planned frequency.",
	}
	for portuguese, want := range cases {
		if got := T(English, portuguese); got != want {
			t.Errorf("T(%q)\n got %q\nwant %q", portuguese, got, want)
		}
	}
}

func TestTextWithoutEntryStaysAsTheAthleteWroteIt(t *testing.T) {
	for _, text := range []string{"Meu pedal com amigos", "Treino 3 de 5", "Giro de base com chuva", ""} {
		if got := T(English, text); got != text {
			t.Errorf("T(%q) = %q, want it unchanged", text, got)
		}
	}
	if got := T(Portuguese, "Giro de base"); got != "Giro de base" {
		t.Fatalf("Portuguese must never be translated, got %q", got)
	}
}

func TestValueTranslatesStringsButNotKeys(t *testing.T) {
	value := map[string]any{
		"name":         "Pedal longo",
		"status":       "planned",
		"rules":        []any{"Intensidade limitada por uma condição de segurança ativa.", "texto livre"},
		"Giro de base": map[string]any{"duration_minutes": 60.0},
	}
	Value(English, value)
	if value["name"] != "Long ride" || value["status"] != "planned" {
		t.Fatalf("unexpected translation: %#v", value)
	}
	rules := value["rules"].([]any)
	if rules[0] != "Intensity limited by an active safety condition." || rules[1] != "texto livre" {
		t.Fatalf("unexpected rules: %#v", rules)
	}
	if _, ok := value["Giro de base"]; !ok {
		t.Fatal("keys must not be translated")
	}
}

// Every message the API sends through writeError, and every "message" it
// returns, must have an English entry; otherwise an English user would read
// Portuguese.
func TestEveryAPIMessageHasAnEnglishEntry(t *testing.T) {
	files, err := filepath.Glob("../httpapi/*.go")
	if err != nil || len(files) == 0 {
		t.Fatalf("httpapi sources not found: %v", err)
	}
	checked := 0
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		fileSet := token.NewFileSet()
		file, err := parser.ParseFile(fileSet, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			var literal *ast.BasicLit
			switch typed := node.(type) {
			case *ast.CallExpr:
				if name, ok := typed.Fun.(*ast.Ident); ok && name.Name == "writeError" && len(typed.Args) == 4 {
					literal, _ = typed.Args[3].(*ast.BasicLit)
				}
			case *ast.KeyValueExpr:
				if key, ok := typed.Key.(*ast.BasicLit); ok && (key.Value == `"message"` || key.Value == `"warning"`) {
					literal, _ = typed.Value.(*ast.BasicLit)
				}
			}
			if literal == nil || literal.Kind != token.STRING {
				return true
			}
			text, _ := strconv.Unquote(literal.Value)
			checked++
			if !Has(text) {
				t.Errorf("%s: no English entry for %q", fileSet.Position(literal.Pos()), text)
			}
			return true
		})
	}
	if checked < 80 {
		t.Fatalf("only %d messages found; the source scan is probably broken", checked)
	}
}
