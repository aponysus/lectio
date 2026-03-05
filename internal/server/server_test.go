package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aponysus/lectio/internal/config"
	"github.com/aponysus/lectio/internal/server"
	"github.com/aponysus/lectio/internal/store"
)

func TestLoginAndProtectedRouteUseDBSessions(t *testing.T) {
	t.Parallel()

	st := openServerTestStore(t)
	defer st.Close()

	cfg := config.Config{
		Env:               "development",
		BootstrapEmail:    "reader@example.com",
		BootstrapPassword: "sufficiently-secret",
		SessionCookieName: "lectio_session",
		CSRFCookieName:    "lectio_csrf",
		CSRFHeaderName:    "X-CSRF-Token",
	}
	if _, err := st.Users().EnsureBootstrapUser(context.Background(), cfg.BootstrapEmail, cfg.BootstrapPassword); err != nil {
		t.Fatalf("ensure bootstrap user: %v", err)
	}

	handler := server.New(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), st).Handler()

	loginBody, err := json.Marshal(map[string]string{
		"email":    cfg.BootstrapEmail,
		"password": cfg.BootstrapPassword,
	})
	if err != nil {
		t.Fatalf("marshal login body: %v", err)
	}

	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 login, got %d: %s", loginRec.Code, loginRec.Body.String())
	}

	sessionCookie := findCookie(t, loginRec.Result().Cookies(), cfg.SessionCookieName)
	protectedReq := httptest.NewRequest(http.MethodGet, "/api/entries", nil)
	protectedReq.AddCookie(sessionCookie)
	protectedRec := httptest.NewRecorder()
	handler.ServeHTTP(protectedRec, protectedReq)

	if protectedRec.Code != http.StatusNotImplemented {
		t.Fatalf("expected authenticated request to reach handler, got %d: %s", protectedRec.Code, protectedRec.Body.String())
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutReq.AddCookie(sessionCookie)
	logoutReq.AddCookie(findCookie(t, loginRec.Result().Cookies(), cfg.CSRFCookieName))
	logoutReq.Header.Set(cfg.CSRFHeaderName, loginRec.Header().Get(cfg.CSRFHeaderName))
	logoutRec := httptest.NewRecorder()
	handler.ServeHTTP(logoutRec, logoutReq)

	if logoutRec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 logout, got %d: %s", logoutRec.Code, logoutRec.Body.String())
	}

	protectedReq = httptest.NewRequest(http.MethodGet, "/api/entries", nil)
	protectedReq.AddCookie(sessionCookie)
	protectedRec = httptest.NewRecorder()
	handler.ServeHTTP(protectedRec, protectedReq)

	if protectedRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after logout, got %d: %s", protectedRec.Code, protectedRec.Body.String())
	}
}

func openServerTestStore(t *testing.T) *store.Store {
	t.Helper()

	dir := t.TempDir()
	st, err := store.Open(context.Background(), store.OpenConfig{
		Path:          filepath.Join(dir, "lectio.db"),
		MigrationsDir: filepath.Join("..", "..", "migrations"),
		AutoMigrate:   true,
	})
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	return st
}

func findCookie(t *testing.T, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()

	for _, cookie := range cookies {
		if strings.EqualFold(cookie.Name, name) {
			return cookie
		}
	}
	t.Fatalf("cookie %q not found", name)
	return nil
}
