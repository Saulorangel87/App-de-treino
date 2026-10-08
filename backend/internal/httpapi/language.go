package httpapi

import (
	"fmt"
	"html"
	"net/http"

	"github.com/Saulorangel87/App-de-treino/backend/internal/email"
	"github.com/Saulorangel87/App-de-treino/backend/internal/i18n"
)

// languageWriter carries the request language down to writeJSON, which only
// receives the ResponseWriter. Handlers keep writing Portuguese; English
// responses are translated on the way out.
type languageWriter struct {
	http.ResponseWriter
	language i18n.Language
}

func (w *languageWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

// withLanguage reads Accept-Language once per request. The frontend always
// sends the language chosen in the app, so the browser's own language does not
// switch the API.
func withLanguage(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		language := i18n.FromAcceptLanguage(r.Header.Get("Accept-Language"))
		w.Header().Add("Vary", "Accept-Language")
		r = r.WithContext(i18n.WithLanguage(r.Context(), language))
		if language != i18n.Portuguese {
			w = &languageWriter{ResponseWriter: w, language: language}
		}
		next.ServeHTTP(w, r)
	})
}

// languageOf finds the language through the writers wrapped around this one
// (the access log wraps every response).
func languageOf(w http.ResponseWriter) i18n.Language {
	for w != nil {
		if typed, ok := w.(*languageWriter); ok {
			return typed.language
		}
		unwrapper, ok := w.(interface{ Unwrap() http.ResponseWriter })
		if !ok {
			break
		}
		w = unwrapper.Unwrap()
	}
	return i18n.Portuguese
}

// The account deletion confirmation phrase is typed by the athlete in the
// language of the screen; both are accepted.
var deleteConfirmations = []string{"ENCERRAR CONTA", "DELETE ACCOUNT"}

func verificationMessage(language i18n.Language, to, name, actionURL string) email.Message {
	name, link := html.EscapeString(name), html.EscapeString(actionURL)
	if language == i18n.English {
		return email.Message{
			To:      to,
			Subject: "Confirm your email on Cadência",
			HTML:    fmt.Sprintf("<p>Hi, %s.</p><p>Confirm your email to start using Cadência.</p><p><a href=\"%s\">Confirm email</a></p><p>This link expires in 24 hours.</p>", name, link),
			Text:    "Confirm your email: " + actionURL,
		}
	}
	return email.Message{
		To:      to,
		Subject: "Confirme seu e-mail no Cadência",
		HTML:    fmt.Sprintf("<p>Olá, %s.</p><p>Confirme seu e-mail para começar a usar o Cadência.</p><p><a href=\"%s\">Confirmar e-mail</a></p><p>Este link expira em 24 horas.</p>", name, link),
		Text:    "Confirme seu e-mail: " + actionURL,
	}
}

func passwordResetMessage(language i18n.Language, to, name, actionURL string) email.Message {
	name, link := html.EscapeString(name), html.EscapeString(actionURL)
	if language == i18n.English {
		return email.Message{
			To:      to,
			Subject: "Reset your password on Cadência",
			HTML:    fmt.Sprintf("<p>Hi, %s.</p><p>We received a request to reset your password on Cadência.</p><p><a href=\"%s\">Reset password</a></p><p>If it wasn't you, ignore this message. The link expires in 24 hours.</p>", name, link),
			Text:    "Reset your password: " + actionURL,
		}
	}
	return email.Message{
		To:      to,
		Subject: "Redefina sua senha no Cadência",
		HTML:    fmt.Sprintf("<p>Olá, %s.</p><p>Recebemos um pedido para redefinir sua senha no Cadência.</p><p><a href=\"%s\">Redefinir senha</a></p><p>Se não foi você, ignore esta mensagem. O link expira em 24 horas.</p>", name, link),
		Text:    "Redefina sua senha: " + actionURL,
	}
}
