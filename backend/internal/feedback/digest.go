package feedback

import (
	"fmt"
	"html"
	"strings"
)

func DigestSubject() string {
	return "Cadência — resumo semanal de feedback"
}

func DigestHTML(entries []DigestEntry) string {
	var body strings.Builder
	body.WriteString("<h1>Resumo semanal de feedback</h1>")
	body.WriteString(fmt.Sprintf("<p>Você recebeu %d novo(s) relato(s) do Cadência.</p>", len(entries)))
	body.WriteString("<hr>")
	for _, entry := range entries {
		body.WriteString("<article>")
		body.WriteString(fmt.Sprintf("<p><strong>%s</strong> · nota %d/5 · %s</p>", html.EscapeString(CategoryLabel(entry.Category)), entry.Rating, html.EscapeString(digestTimestamp(entry))))
		body.WriteString(fmt.Sprintf("<p><strong>%s</strong></p>", html.EscapeString(entry.DisplayName)))
		body.WriteString(fmt.Sprintf("<p>%s</p>", html.EscapeString(entry.Message)))
		body.WriteString("</article>")
	}
	return body.String()
}

func DigestText(entries []DigestEntry) string {
	var body strings.Builder
	body.WriteString(fmt.Sprintf("Resumo semanal de feedback\nVocê recebeu %d novo(s) relato(s) do Cadência.\n\n", len(entries)))
	for index, entry := range entries {
		body.WriteString(fmt.Sprintf("%d. %s · nota %d/5 · %s\n", index+1, CategoryLabel(entry.Category), entry.Rating, digestTimestamp(entry)))
		body.WriteString(fmt.Sprintf("De: %s\n%s\n\n", entry.DisplayName, entry.Message))
	}
	return body.String()
}

func digestTimestamp(entry DigestEntry) string {
	return entry.CreatedAt.UTC().Format("02/01/2006 15:04 UTC")
}
