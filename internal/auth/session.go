package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	SessionCookieName = "lectio_session"
	CSRFCookieName    = "lectio_csrf"
	DefaultUserID     = "lectio"
)

type contextKey string

const sessionContextKey contextKey = "lectio_session"

type Session struct {
	Subject   string    `json:"sub"`
	ExpiresAt time.Time `json:"exp"`
}

type csrfToken struct {
	Nonce     string    `json:"nonce"`
	ExpiresAt time.Time `json:"exp"`
}

type Manager struct {
	sessionSecret []byte
	csrfSecret    []byte
	sessionTTL    time.Duration
	secureCookies bool
}

func NewManager(sessionSecret, csrfSecret string, sessionTTL time.Duration, secureCookies bool) *Manager {
	return &Manager{
		sessionSecret: []byte(sessionSecret),
		csrfSecret:    []byte(csrfSecret),
		sessionTTL:    sessionTTL,
		secureCookies: secureCookies,
	}
}

func (m *Manager) NewSession(subject string) (string, error) {
	payload := Session{
		Subject:   subject,
		ExpiresAt: time.Now().Add(m.sessionTTL),
	}
	return signToken(payload, m.sessionSecret)
}

func (m *Manager) VerifySession(token string) (Session, error) {
	var session Session
	if err := verifyToken(token, m.sessionSecret, &session); err != nil {
		return Session{}, err
	}
	if time.Now().After(session.ExpiresAt) {
		return Session{}, errors.New("session expired")
	}
	return session, nil
}

func (m *Manager) SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   m.secureCookies,
		MaxAge:   int(m.sessionTTL.Seconds()),
	})
}

func (m *Manager) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   m.secureCookies,
		MaxAge:   -1,
	})
}

func (m *Manager) EnsureCSRFCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	if cookie, err := r.Cookie(CSRFCookieName); err == nil {
		if err := m.ValidateCSRFCookie(cookie.Value); err == nil {
			return cookie.Value, nil
		}
	}

	token, err := m.newCSRFToken()
	if err != nil {
		return "", err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
		Secure:   m.secureCookies,
		MaxAge:   int((24 * time.Hour).Seconds()),
	})

	return token, nil
}

func (m *Manager) ValidateCSRFRequest(r *http.Request) error {
	cookie, err := r.Cookie(CSRFCookieName)
	if err != nil {
		return errors.New("missing csrf cookie")
	}

	headerToken := r.Header.Get("X-CSRF-Token")
	if headerToken == "" {
		return errors.New("missing csrf header")
	}

	if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(headerToken)) != 1 {
		return errors.New("csrf token mismatch")
	}

	return m.ValidateCSRFCookie(cookie.Value)
}

func (m *Manager) ValidateCSRFCookie(token string) error {
	var payload csrfToken
	if err := verifyToken(token, m.csrfSecret, &payload); err != nil {
		return err
	}
	if time.Now().After(payload.ExpiresAt) {
		return errors.New("csrf token expired")
	}
	return nil
}

func ValidatePassword(provided, expected string) bool {
	if len(provided) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

func WithSession(ctx context.Context, session Session) context.Context {
	return context.WithValue(ctx, sessionContextKey, session)
}

func SessionFromContext(ctx context.Context) (Session, bool) {
	session, ok := ctx.Value(sessionContextKey).(Session)
	return session, ok
}

func (m *Manager) newCSRFToken() (string, error) {
	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return "", err
	}
	payload := csrfToken{
		Nonce:     base64.RawURLEncoding.EncodeToString(nonce[:]),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	return signToken(payload, m.csrfSecret)
}

func signToken(payload any, secret []byte) (string, error) {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(rawPayload)
	signature := sign(encodedPayload, secret)
	return fmt.Sprintf("%s.%s", encodedPayload, signature), nil
}

func verifyToken(token string, secret []byte, dst any) error {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return errors.New("invalid token format")
	}

	expectedSignature := sign(parts[0], secret)
	if subtle.ConstantTimeCompare([]byte(parts[1]), []byte(expectedSignature)) != 1 {
		return errors.New("invalid token signature")
	}

	rawPayload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return err
	}

	return json.Unmarshal(rawPayload, dst)
}

func sign(payload string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
