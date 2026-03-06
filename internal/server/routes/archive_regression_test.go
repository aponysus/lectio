package routes

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
	"time"

	lectioauth "github.com/aponysus/lectio/internal/auth"
	"github.com/aponysus/lectio/internal/config"
	"github.com/aponysus/lectio/internal/model"
	"github.com/aponysus/lectio/internal/store"
)

type apiEnvelope[T any] struct {
	Data T `json:"data"`
}

func TestArchiveEndpointsRegression(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client, repo := newArchiveTestServer(t)
	csrfToken := loginArchiveTestClient(t, client)

	source, err := repo.CreateSource(ctx, model.SourceInput{
		Title:  "Archive Regression Source",
		Medium: string(model.SourceMediumBook),
	})
	if err != nil {
		t.Fatalf("CreateSource() error = %v", err)
	}

	engagement, err := repo.CreateEngagement(ctx, model.EngagementInput{
		SourceID:   source.ID,
		EngagedAt:  "2026-03-06T16:00:00Z",
		Reflection: "This engagement exists to exercise archive behavior across the active surfaces.",
	})
	if err != nil {
		t.Fatalf("CreateEngagement() error = %v", err)
	}

	inquiry, err := repo.CreateInquiry(ctx, model.InquiryInput{
		Title:    "Does archive behavior remain coherent across the app?",
		Question: "After archival, do the active routes stop surfacing the record and its dependent prompts?",
		Status:   string(model.InquiryStatusActive),
	})
	if err != nil {
		t.Fatalf("CreateInquiry() error = %v", err)
	}

	if err := repo.ReplaceEngagementInquiries(ctx, engagement.ID, []string{inquiry.ID}); err != nil {
		t.Fatalf("ReplaceEngagementInquiries() error = %v", err)
	}

	for _, engagedAt := range []string{
		"2026-03-05T11:00:00Z",
		"2026-03-04T10:00:00Z",
	} {
		extraEngagement, err := repo.CreateEngagement(ctx, model.EngagementInput{
			SourceID:   source.ID,
			EngagedAt:  engagedAt,
			Reflection: "Additional density for the synthesis eligibility regression.",
		})
		if err != nil {
			t.Fatalf("CreateEngagement(extra) error = %v", err)
		}
		if err := repo.ReplaceEngagementInquiries(ctx, extraEngagement.ID, []string{inquiry.ID}); err != nil {
			t.Fatalf("ReplaceEngagementInquiries(extra) error = %v", err)
		}
	}

	claim, err := repo.CreateClaim(ctx, model.ClaimInput{
		Text:               "Archive is only trustworthy when the record falls out of active work, not just detail pages.",
		ClaimType:          string(model.ClaimTypeInterpretation),
		Status:             string(model.ClaimStatusActive),
		OriginEngagementID: engagement.ID,
	}, []string{inquiry.ID})
	if err != nil {
		t.Fatalf("CreateClaim() error = %v", err)
	}

	note, err := repo.CreateLanguageNote(ctx, model.LanguageNoteInput{
		EngagementID: engagement.ID,
		Term:         "archive",
		Language:     "en",
		NoteType:     string(model.LanguageNoteTypeTranslation),
		Content:      "Used here as a regression marker for note archive behavior.",
	})
	if err != nil {
		t.Fatalf("CreateLanguageNote() error = %v", err)
	}

	synthesis, err := repo.CreateSynthesis(ctx, model.SynthesisInput{
		Title:     "Archive regression checkpoint",
		Body:      "The inquiry currently has an active synthesis, so it should not appear in the synthesis-ready surface until this record is archived.",
		Type:      string(model.SynthesisTypeCheckpoint),
		InquiryID: inquiry.ID,
	})
	if err != nil {
		t.Fatalf("CreateSynthesis() error = %v", err)
	}

	deleteAndExpectNoContent(t, client, "/api/language-notes/"+note.ID, csrfToken)
	assertStatus(t, client, http.MethodDelete, "/api/language-notes/"+note.ID, csrfToken, nil, http.StatusNotFound)

	noteList := getAPIData[[]model.LanguageNote](t, client, "/api/engagements/"+engagement.ID+"/language-notes")
	if len(noteList) != 0 {
		t.Fatalf("expected archived language note to disappear from engagement notes, got %d", len(noteList))
	}

	filteredEngagements := getAPIData[[]model.Engagement](t, client, "/api/engagements?has_language_notes=true")
	if len(filteredEngagements) != 0 {
		t.Fatalf("expected archived language note to remove engagement from has_language_notes filter, got %d", len(filteredEngagements))
	}

	deleteAndExpectNoContent(t, client, "/api/claims/"+claim.ID, csrfToken)
	assertStatus(t, client, http.MethodGet, "/api/claims/"+claim.ID, "", nil, http.StatusNotFound)

	claimResults := getAPIData[[]model.Claim](t, client, "/api/claims?q=archive%20is%20only%20trustworthy")
	if len(claimResults) != 0 {
		t.Fatalf("expected archived claim to disappear from search results, got %d", len(claimResults))
	}

	inquiryClaims := getAPIData[[]model.Claim](t, client, "/api/inquiries/"+inquiry.ID+"/claims")
	if len(inquiryClaims) != 0 {
		t.Fatalf("expected archived claim to disappear from inquiry claims, got %d", len(inquiryClaims))
	}

	deleteAndExpectNoContent(t, client, "/api/syntheses/"+synthesis.ID, csrfToken)
	assertStatus(t, client, http.MethodGet, "/api/syntheses/"+synthesis.ID, "", nil, http.StatusNotFound)

	eligible := getAPIData[[]model.Inquiry](t, client, "/api/inquiries/eligible-for-synthesis")
	if len(eligible) != 1 || eligible[0].ID != inquiry.ID {
		t.Fatalf("expected inquiry %s to reappear as synthesis-eligible after synthesis archive, got %+v", inquiry.ID, eligible)
	}

	deleteAndExpectNoContent(t, client, "/api/inquiries/"+inquiry.ID, csrfToken)
	assertStatus(t, client, http.MethodGet, "/api/inquiries/"+inquiry.ID, "", nil, http.StatusNotFound)

	inquiryLinks := getAPIData[[]model.InquirySummary](t, client, "/api/engagements/"+engagement.ID+"/inquiries")
	if len(inquiryLinks) != 0 {
		t.Fatalf("expected archived inquiry to disappear from engagement inquiry links, got %d", len(inquiryLinks))
	}

	deleteAndExpectNoContent(t, client, "/api/engagements/"+engagement.ID, csrfToken)
	assertStatus(t, client, http.MethodGet, "/api/engagements/"+engagement.ID, "", nil, http.StatusNotFound)

	engagementResults := getAPIData[[]model.Engagement](t, client, "/api/engagements")
	for _, listed := range engagementResults {
		if listed.ID == engagement.ID {
			t.Fatalf("expected archived engagement %s to disappear from engagement list", engagement.ID)
		}
	}

	deleteAndExpectNoContent(t, client, "/api/sources/"+source.ID, csrfToken)
	assertStatus(t, client, http.MethodDelete, "/api/sources/"+source.ID, csrfToken, nil, http.StatusNotFound)
	assertStatus(t, client, http.MethodGet, "/api/sources/"+source.ID, "", nil, http.StatusNotFound)

	sourceResults := getAPIData[[]model.Source](t, client, "/api/sources")
	if len(sourceResults) != 0 {
		t.Fatalf("expected archived source to disappear from source list, got %d", len(sourceResults))
	}
}

func newArchiveTestServer(t *testing.T) (*archiveTestClient, *store.Store) {
	t.Helper()

	db, err := store.Open(filepath.Join(t.TempDir(), "lectio.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	if err := store.ApplyMigrations(context.Background(), db); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}

	repo := store.New(db)
	cfg := config.Config{
		Env:               "test",
		Addr:              ":0",
		DBPath:            filepath.Join(t.TempDir(), "unused.db"),
		WebDistDir:        filepath.Join(t.TempDir(), "web-dist"),
		BootstrapPassword: "test-password",
		SessionSecret:     "test-session-secret",
		CSRFSecret:        "test-csrf-secret",
		SessionTTL:        time.Hour,
	}

	router := New(Dependencies{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Config: cfg,
		Store:  repo,
		Auth:   lectioauth.NewManager(cfg.SessionSecret, cfg.CSRFSecret, cfg.SessionTTL, false),
	})

	return &archiveTestClient{
		t:       t,
		handler: router,
		cookies: map[string]*http.Cookie{},
	}, repo
}

func loginArchiveTestClient(t *testing.T, client *archiveTestClient) string {
	t.Helper()

	session := getAPIData[struct {
		CSRFToken string `json:"csrf_token"`
	}](t, client, "/api/auth/session")

	loginBody := bytes.NewBufferString(`{"password":"test-password"}`)
	assertStatus(t, client, http.MethodPost, "/api/auth/login", session.CSRFToken, loginBody, http.StatusNoContent)

	return session.CSRFToken
}

func deleteAndExpectNoContent(t *testing.T, client *archiveTestClient, path, csrfToken string) {
	t.Helper()
	assertStatus(t, client, http.MethodDelete, path, csrfToken, nil, http.StatusNoContent)
}

func assertStatus(t *testing.T, client *archiveTestClient, method, path, csrfToken string, body io.Reader, wantStatus int) {
	t.Helper()
	resp := client.Do(method, path, csrfToken, body)
	defer resp.Body.Close()

	if resp.StatusCode != wantStatus {
		payload, _ := io.ReadAll(resp.Body)
		t.Fatalf("%s %s returned status %d, want %d, body=%s", method, path, resp.StatusCode, wantStatus, strings.TrimSpace(string(payload)))
	}
}

func getAPIData[T any](t *testing.T, client *archiveTestClient, path string) T {
	t.Helper()
	resp := client.Do(http.MethodGet, path, "", nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(resp.Body)
		t.Fatalf("GET %s returned status %d, want %d, body=%s", path, resp.StatusCode, http.StatusOK, strings.TrimSpace(string(payload)))
	}

	var envelope apiEnvelope[T]
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode GET %s response: %v", path, err)
	}

	return envelope.Data
}

type archiveTestClient struct {
	t       *testing.T
	handler http.Handler
	cookies map[string]*http.Cookie
}

func (c *archiveTestClient) Do(method, path, csrfToken string, body io.Reader) *http.Response {
	c.t.Helper()

	req := httptest.NewRequest(method, path, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if csrfToken != "" {
		req.Header.Set("X-CSRF-Token", csrfToken)
	}
	for _, cookie := range c.cookies {
		req.AddCookie(cookie)
	}

	recorder := httptest.NewRecorder()
	c.handler.ServeHTTP(recorder, req)
	resp := recorder.Result()

	for _, cookie := range resp.Cookies() {
		cookieCopy := *cookie
		c.cookies[cookie.Name] = &cookieCopy
	}

	return resp
}
