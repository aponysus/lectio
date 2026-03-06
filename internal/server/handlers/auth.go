package handlers

import (
	"log/slog"
	"net/http"

	lectioauth "github.com/aponysus/lectio/internal/auth"
	"github.com/aponysus/lectio/internal/config"
	"github.com/aponysus/lectio/internal/server/httpx"
)

type AuthHandler struct {
	Logger *slog.Logger
	Config config.Config
	Auth   *lectioauth.Manager
}

type loginRequest struct {
	Password string `json:"password"`
}

func (h AuthHandler) Session(w http.ResponseWriter, r *http.Request) {
	csrfToken, err := h.Auth.EnsureCSRFCookie(w, r)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to issue csrf token")
		return
	}

	response := map[string]any{
		"authenticated": false,
		"csrf_token":    csrfToken,
	}

	if session, ok := lectioauth.SessionFromContext(r.Context()); ok {
		response["authenticated"] = true
		response["user_id"] = session.Subject
		response["expires_at"] = session.ExpiresAt.UTC().Format(timeFormat)
	}

	httpx.WriteData(w, http.StatusOK, response)
}

func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_error", "invalid login payload")
		return
	}

	if !lectioauth.ValidatePassword(req.Password, h.Config.BootstrapPassword) {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid credentials")
		return
	}

	token, err := h.Auth.NewSession(lectioauth.DefaultUserID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to create session")
		return
	}

	h.Auth.SetSessionCookie(w, token)
	httpx.WriteJSON(w, http.StatusNoContent, nil)
}

func (h AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.Auth.ClearSessionCookie(w)
	httpx.WriteJSON(w, http.StatusNoContent, nil)
}

const timeFormat = "2006-01-02T15:04:05Z07:00"
