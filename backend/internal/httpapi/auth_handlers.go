package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/ai"
	"github.com/Saulorangel87/App-de-treino/backend/internal/athlete"
	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
	"github.com/Saulorangel87/App-de-treino/backend/internal/email"
	"github.com/Saulorangel87/App-de-treino/backend/internal/evolution"
	"github.com/Saulorangel87/App-de-treino/backend/internal/feedback"
	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
)

const sessionCookieName = "cadencia_session"

type Server struct {
	auth          *auth.Service
	athlete       *athlete.Service
	onboarding    *athlete.OnboardingService
	assessments   *athlete.AssessmentService
	recovery      *athlete.RecoveryService
	evolution     *evolution.Service
	feedback      *feedback.Service
	planning      *planning.Service
	ai            *ai.Service
	secureCookies bool
	sessionTTL    time.Duration
	emailSender   email.Sender
	appBaseURL    string
	emailTokenTTL time.Duration
	development   bool
}

type tokenRequest struct {
	Token string `json:"token"`
}
type emailRequest struct {
	Email string `json:"email"`
}
type resetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type credentialsRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var input credentialsRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	user, token, err := s.auth.Register(r.Context(), input.Email, input.Password, input.DisplayName)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid_registration", "Informe nome, e-mail válido e senha entre 10 e 72 caracteres.")
		case errors.Is(err, auth.ErrEmailExists):
			writeError(w, http.StatusConflict, "email_exists", "Este e-mail já está cadastrado.")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível criar a conta.")
		}
		return
	}
	s.setSessionCookie(w, token)
	verificationURL, err := s.sendVerification(r.Context(), user)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "email_unavailable", "Sua conta foi criada, mas não foi possível enviar a confirmação. Entre novamente para reenviar o link.")
		return
	}
	response := map[string]any{"user": user, "message": "Enviamos um link para confirmar seu e-mail."}
	if s.development {
		response["development_verification_url"] = verificationURL
	}
	writeJSON(w, http.StatusCreated, response)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var input credentialsRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	user, token, err := s.auth.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "E-mail ou senha inválidos.")
		return
	}
	s.setSessionCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	token := sessionToken(r)
	if err := s.auth.Logout(r.Context(), token); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível encerrar a sessão.")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: s.secureCookies, SameSite: http.SameSiteLaxMode})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (s *Server) resendVerification(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	if user.EmailVerified {
		writeJSON(w, http.StatusOK, map[string]string{"message": "Seu e-mail já está confirmado."})
		return
	}
	verificationURL, err := s.sendVerification(r.Context(), user)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "email_unavailable", "Não foi possível enviar a confirmação agora. Tente novamente em instantes.")
		return
	}
	response := map[string]any{"message": "Enviamos um novo link de confirmação."}
	if s.development {
		response["development_verification_url"] = verificationURL
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) verifyEmail(w http.ResponseWriter, r *http.Request) {
	var input tokenRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	user, err := s.auth.VerifyEmail(r.Context(), input.Token)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_email_token", "Este link é inválido, expirou ou já foi usado.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user, "message": "E-mail confirmado com sucesso."})
}

func (s *Server) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var input emailRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	user, token, err := s.auth.CreatePasswordResetToken(r.Context(), input.Email, s.emailTokenTTL)
	resetURL := ""
	if err == nil && token != "" {
		resetURL = s.actionURL("/redefinir-senha", token)
		err = s.emailSender.Send(r.Context(), email.Message{To: user.Email, Subject: "Redefina sua senha no Cadência", HTML: passwordResetHTML(user.DisplayName, resetURL), Text: "Redefina sua senha: " + resetURL})
	}
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "email_unavailable", "Não foi possível enviar o e-mail agora. Tente novamente em instantes.")
		return
	}
	// A resposta não confirma se o endereço existe, evitando enumeração de contas.
	response := map[string]any{"message": "Se houver uma conta com este e-mail, enviaremos um link para redefinir a senha."}
	if s.development && resetURL != "" {
		response["development_reset_url"] = resetURL
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) resetPassword(w http.ResponseWriter, r *http.Request) {
	var input resetPasswordRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := s.auth.ResetPassword(r.Context(), input.Token, input.Password); err != nil {
		if errors.Is(err, auth.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "invalid_password", "Use uma senha entre 10 e 72 caracteres.")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid_email_token", "Este link é inválido, expirou ou já foi usado.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Senha redefinida. Entre novamente para continuar."})
}

func (s *Server) getProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	profile, err := s.athlete.Get(r.Context(), user.ID)
	if errors.Is(err, athlete.ErrProfileMissing) {
		writeJSON(w, http.StatusOK, map[string]any{"profile": nil})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível carregar o perfil.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"profile": profile})
}

func (s *Server) putProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var profile athlete.Profile
	if !decodeJSON(w, r, &profile) {
		return
	}
	profile.UserID = user.ID
	saved, err := s.athlete.Save(r.Context(), profile)
	if errors.Is(err, athlete.ErrInvalidProfile) {
		writeError(w, http.StatusBadRequest, "invalid_profile", "Revise os dados do perfil informados.")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível salvar o perfil.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"profile": saved})
}

func (s *Server) requireUser(w http.ResponseWriter, r *http.Request) (auth.User, bool) {
	user, err := s.auth.Authenticate(r.Context(), sessionToken(r))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Faça login para continuar.")
		return auth.User{}, false
	}
	return user, true
}

func (s *Server) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookieName, Value: token, Path: "/", HttpOnly: true,
		Secure: s.secureCookies, SameSite: http.SameSiteLaxMode,
		MaxAge: int(s.sessionTTL.Seconds()), Expires: time.Now().Add(s.sessionTTL),
	})
}

func sessionToken(r *http.Request) string {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (s *Server) sendVerification(ctx context.Context, user auth.User) (string, error) {
	token, err := s.auth.CreateEmailVerificationToken(ctx, user.ID, s.emailTokenTTL)
	if err != nil {
		return "", err
	}
	verificationURL := s.actionURL("/verificar-email", token)
	err = s.emailSender.Send(ctx, email.Message{To: user.Email, Subject: "Confirme seu e-mail no Cadência", HTML: verificationHTML(user.DisplayName, verificationURL), Text: "Confirme seu e-mail: " + verificationURL})
	return verificationURL, err
}

func (s *Server) actionURL(path, token string) string {
	base, err := url.Parse(s.appBaseURL)
	if err != nil {
		return ""
	}
	base.Path = strings.TrimRight(base.Path, "/") + path
	query := base.Query()
	query.Set("token", token)
	base.RawQuery = query.Encode()
	return base.String()
}

func verificationHTML(name, actionURL string) string {
	return fmt.Sprintf("<p>Olá, %s.</p><p>Confirme seu e-mail para começar a usar o Cadência.</p><p><a href=\"%s\">Confirmar e-mail</a></p><p>Este link expira em 24 horas.</p>", html.EscapeString(name), html.EscapeString(actionURL))
}

func passwordResetHTML(name, actionURL string) string {
	return fmt.Sprintf("<p>Olá, %s.</p><p>Recebemos um pedido para redefinir sua senha no Cadência.</p><p><a href=\"%s\">Redefinir senha</a></p><p>Se não foi você, ignore esta mensagem. O link expira em 24 horas.</p>", html.EscapeString(name), html.EscapeString(actionURL))
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "O conteúdo enviado é inválido.")
		return false
	}
	return true
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
