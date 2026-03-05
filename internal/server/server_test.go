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
	"sort"
	"strconv"
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
	sessionCookie, csrfCookie, csrfHeader := login(t, handler, cfg)
	protectedReq := httptest.NewRequest(http.MethodGet, "/api/entries", nil)
	protectedReq.AddCookie(sessionCookie)
	protectedRec := httptest.NewRecorder()
	handler.ServeHTTP(protectedRec, protectedReq)

	if protectedRec.Code != http.StatusOK {
		t.Fatalf("expected authenticated request to reach handler, got %d: %s", protectedRec.Code, protectedRec.Body.String())
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutReq.AddCookie(sessionCookie)
	logoutReq.AddCookie(csrfCookie)
	logoutReq.Header.Set(cfg.CSRFHeaderName, csrfHeader)
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

func TestCreateAndListEntries(t *testing.T) {
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
	sessionCookie, csrfCookie, csrfHeader := login(t, handler, cfg)

	createBody, err := json.Marshal(map[string]any{
		"source": map[string]any{
			"title":     "The Cloud of Unknowing",
			"author":    "Anonymous",
			"tradition": "Christian",
		},
		"passage":    "A short passage",
		"reflection": "A careful reflection",
		"mood":       "focused",
		"energy":     4,
		"tags":       []string{"Kenosis", " kenosis ", "Prayer"},
	})
	if err != nil {
		t.Fatalf("marshal create body: %v", err)
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/entries", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set(cfg.CSRFHeaderName, csrfHeader)
	createReq.AddCookie(sessionCookie)
	createReq.AddCookie(csrfCookie)
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 create entry, got %d: %s", createRec.Code, createRec.Body.String())
	}

	var createPayload struct {
		Data struct {
			ID       int64 `json:"id"`
			SourceID int64 `json:"source_id"`
			Tags     []struct {
				Slug string `json:"slug"`
			} `json:"tags"`
			Energy *int `json:"energy"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createRec.Body.Bytes(), &createPayload); err != nil {
		t.Fatalf("decode create payload: %v", err)
	}
	if createPayload.Data.ID == 0 || createPayload.Data.SourceID == 0 {
		t.Fatalf("expected ids in create response: %s", createRec.Body.String())
	}
	if createPayload.Data.Energy == nil || *createPayload.Data.Energy != 4 {
		t.Fatalf("expected energy in create response: %s", createRec.Body.String())
	}
	gotSlugs := make([]string, 0, len(createPayload.Data.Tags))
	for _, tag := range createPayload.Data.Tags {
		gotSlugs = append(gotSlugs, tag.Slug)
	}
	sort.Strings(gotSlugs)
	if strings.Join(gotSlugs, ",") != "kenosis,prayer" {
		t.Fatalf("expected normalized tags, got %v", gotSlugs)
	}

	sources, err := st.Sources().List(context.Background(), store.SourceListFilter{Query: "The"})
	if err != nil {
		t.Fatalf("list sources: %v", err)
	}
	if len(sources) != 1 || sources[0].EntryCount != 1 {
		t.Fatalf("expected created source with one entry, got %+v", sources)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/entries?tag=KENOSIS&source_id="+strconv.FormatInt(createPayload.Data.SourceID, 10), nil)
	listReq.AddCookie(sessionCookie)
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200 list entries, got %d: %s", listRec.Code, listRec.Body.String())
	}

	var listPayload struct {
		Data []struct {
			ID       int64 `json:"id"`
			SourceID int64 `json:"source_id"`
		} `json:"data"`
		Meta struct {
			Page     int   `json:"page"`
			PageSize int   `json:"page_size"`
			Total    int64 `json:"total"`
			HasNext  bool  `json:"has_next"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("decode list payload: %v", err)
	}
	if len(listPayload.Data) != 1 || listPayload.Data[0].ID != createPayload.Data.ID {
		t.Fatalf("expected created entry in list payload: %s", listRec.Body.String())
	}
	if listPayload.Meta.Total != 1 || listPayload.Meta.Page != 1 || listPayload.Meta.PageSize != 20 || listPayload.Meta.HasNext {
		t.Fatalf("unexpected list meta: %+v", listPayload.Meta)
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

func login(t *testing.T, handler http.Handler, cfg config.Config) (*http.Cookie, *http.Cookie, string) {
	t.Helper()

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

	return findCookie(t, loginRec.Result().Cookies(), cfg.SessionCookieName),
		findCookie(t, loginRec.Result().Cookies(), cfg.CSRFCookieName),
		loginRec.Header().Get(cfg.CSRFHeaderName)
}
