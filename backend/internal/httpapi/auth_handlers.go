package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Saulorangel87/App-de-treino/backend/internal/athlete"
	"github.com/Saulorangel87/App-de-treino/backend/internal/auth"
	"github.com/Saulorangel87/App-de-treino/backend/internal/planning"
)

const sessionCookieName = "cadencia_session"

type Server struct {
	auth          *auth.Service
	athlete       *athlete.Service
	onboarding    *athlete.OnboardingService
	planning      *planning.Service
	secureCookies bool
	sessionTTL    time.Duration
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
	writeJSON(w, http.StatusCreated, map[string]any{"user": user})
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
