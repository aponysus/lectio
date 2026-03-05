package server

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/aponysus/lectio/internal/config"
	"github.com/aponysus/lectio/internal/store"
	"github.com/aponysus/lectio/ui"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionTTL         = 14 * 24 * time.Hour
	sessionIdleTimeout = 24 * time.Hour
)

type Server struct {
	cfg          config.Config
	logger       *slog.Logger
	store        *store.Store
	staticFS     fs.FS
	staticServer http.Handler
}

func New(cfg config.Config, logger *slog.Logger, db *store.Store) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	assets := ui.FS()

	return &Server{
		cfg:          cfg,
		logger:       logger,
		store:        db,
		staticFS:     assets,
		staticServer: http.FileServer(http.FS(assets)),
	}
}

func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(s.loggingMiddleware)
	r.Use(s.securityHeadersMiddleware)

	r.Route("/api", func(api chi.Router) {
		api.Get("/health", s.handleHealth)

		api.Route("/auth", func(auth chi.Router) {
			auth.Post("/login", s.handleLogin)
			auth.Post("/logout", s.handleLogout)
		})

		api.Group(func(protected chi.Router) {
			protected.Use(s.requireSession)
			protected.Use(s.csrfMiddleware)

			protected.Get("/entries", s.handleListEntries)
			protected.Post("/entries", s.handleCreateEntry)
			protected.Get("/entries/{id}", s.notImplemented("GET /api/entries/{id}"))
			protected.Put("/entries/{id}", s.notImplemented("PUT /api/entries/{id}"))
			protected.Delete("/entries/{id}", s.notImplemented("DELETE /api/entries/{id}"))
			protected.Get("/entries/{id}/resonances", s.notImplemented("GET /api/entries/{id}/resonances"))

			protected.Get("/sources", s.notImplemented("GET /api/sources"))
			protected.Post("/sources", s.notImplemented("POST /api/sources"))
			protected.Get("/sources/{id}", s.notImplemented("GET /api/sources/{id}"))
			protected.Put("/sources/{id}", s.notImplemented("PUT /api/sources/{id}"))
			protected.Get("/sources/{id}/entries", s.notImplemented("GET /api/sources/{id}/entries"))

			protected.Get("/tags", s.notImplemented("GET /api/tags"))
			protected.Get("/tags/co-occurrence", s.notImplemented("GET /api/tags/co-occurrence"))
			protected.Get("/tags/{slug}", s.notImplemented("GET /api/tags/{slug}"))
			protected.Get("/tags/{slug}/entries", s.notImplemented("GET /api/tags/{slug}/entries"))

			protected.Get("/threads", s.notImplemented("GET /api/threads"))
			protected.Post("/threads", s.notImplemented("POST /api/threads"))
			protected.Get("/threads/{id}", s.notImplemented("GET /api/threads/{id}"))
			protected.Put("/threads/{id}", s.notImplemented("PUT /api/threads/{id}"))

			protected.Get("/timeline", s.notImplemented("GET /api/timeline"))
			protected.Get("/resonance/daily", s.notImplemented("GET /api/resonance/daily"))
		})
	})

	r.NotFound(s.handleNotFoundOrSPA)

	return r
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	statusCode := http.StatusOK
	appStatus := "ok"
	dbStatus := "ok"

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := s.store.Ping(ctx); err != nil {
		statusCode = http.StatusServiceUnavailable
		appStatus = "degraded"
		dbStatus = "unavailable"
		s.logger.Error("health check ping failed", "error", err)
	}

	migrationState, err := s.store.MigrationState(ctx)
	if err != nil {
		statusCode = http.StatusServiceUnavailable
		appStatus = "degraded"
		dbStatus = "unavailable"
		s.logger.Error("health check migration state failed", "error", err)
	}

	s.writeJSON(w, statusCode, map[string]any{
		"data": map[string]any{
			"status":                  appStatus,
			"db_status":               dbStatus,
			"schema_version":          migrationState.CurrentVersion,
			"latest_schema_version":   migrationState.LatestVersion,
			"schema_dirty":            migrationState.Dirty,
			"replication_lag_seconds": nil,
			"timestamp":               time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "validation_error", "Invalid request payload", []map[string]string{{"field": "body", "reason": "invalid_json"}})
		return
	}

	user, err := s.lookupLoginUser(r.Context(), req.Email)
	if errors.Is(err, store.ErrNotFound) {
		s.writeError(w, r, http.StatusUnauthorized, "unauthorized", "Invalid credentials", nil)
		return
	}
	if err != nil {
		s.writeError(w, r, http.StatusInternalServerError, "internal_error", "Could not load user", nil)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		s.writeError(w, r, http.StatusUnauthorized, "unauthorized", "Invalid credentials", nil)
		return
	}

	sessionID, err := randomToken(32)
	if err != nil {
		s.writeError(w, r, http.StatusInternalServerError, "internal_error", "Could not create session", nil)
		return
	}
	csrfToken, err := randomToken(32)
	if err != nil {
		s.writeError(w, r, http.StatusInternalServerError, "internal_error", "Could not create csrf token", nil)
		return
	}

	now := time.Now().UTC()
	if existingSession, err := r.Cookie(s.cfg.SessionCookieName); err == nil && existingSession.Value != "" {
		_ = s.store.Sessions().Delete(r.Context(), existingSession.Value)
	}
	if err := s.store.Sessions().Create(r.Context(), store.Session{
		ID:         sessionID,
		UserID:     user.ID,
		ExpiresAt:  now.Add(sessionTTL),
		LastSeenAt: now,
	}); err != nil {
		s.writeError(w, r, http.StatusInternalServerError, "internal_error", "Could not create session", nil)
		return
	}

	secure := s.cfg.Env == "production"
	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.SessionCookieName,
		Value:    sessionID,
		Path:     "/",
		Expires:  now.Add(sessionTTL),
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.CSRFCookieName,
		Value:    csrfToken,
		Path:     "/",
		Expires:  now.Add(sessionTTL),
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	w.Header().Set(s.cfg.CSRFHeaderName, csrfToken)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(s.cfg.SessionCookieName); err == nil && cookie.Value != "" {
		_ = s.store.Sessions().Delete(r.Context(), cookie.Value)
	}

	secure := s.cfg.Env == "production"
	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     s.cfg.CSRFCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleCreateEntry(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Source struct {
			ID        *int64 `json:"id"`
			Title     string `json:"title"`
			Author    string `json:"author"`
			Year      *int   `json:"year"`
			Tradition string `json:"tradition"`
			Language  string `json:"language"`
		} `json:"source"`
		Passage          string   `json:"passage"`
		Reflection       string   `json:"reflection"`
		Mood             string   `json:"mood"`
		Energy           *int     `json:"energy"`
		Tags             []string `json:"tags"`
		ThreadID         *int64   `json:"thread_id"`
		RevisitOfEntryID *int64   `json:"revisit_of_entry_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "validation_error", "Invalid request payload", []map[string]string{{"field": "body", "reason": "invalid_json"}})
		return
	}

	details := validateCreateEntryRequest(req.Reflection, req.Passage, req.Energy, req.Source.ID, req.Source.Title, req.ThreadID, req.RevisitOfEntryID)
	if len(details) > 0 {
		s.writeError(w, r, http.StatusBadRequest, "validation_error", "Invalid request payload", details)
		return
	}

	entry, err := s.store.Entries().Create(r.Context(), store.CreateEntryInput{
		SourceID: req.Source.ID,
		Source: store.SourceInput{
			ID:        req.Source.ID,
			Title:     req.Source.Title,
			Author:    req.Source.Author,
			Year:      req.Source.Year,
			Tradition: req.Source.Tradition,
			Language:  req.Source.Language,
		},
		Passage:    req.Passage,
		Reflection: strings.TrimSpace(req.Reflection),
		Mood:       req.Mood,
		Energy:     req.Energy,
		Tags:       req.Tags,
	})
	if errors.Is(err, store.ErrNotFound) {
		s.writeError(w, r, http.StatusNotFound, "not_found", "Source not found", nil)
		return
	}
	if err != nil {
		s.logger.Error("create entry failed", "error", err)
		s.writeError(w, r, http.StatusInternalServerError, "internal_error", "Could not create entry", nil)
		return
	}

	s.writeJSON(w, http.StatusCreated, map[string]any{
		"data": entryResponse(entry),
	})
}

func (s *Server) handleListEntries(w http.ResponseWriter, r *http.Request) {
	filter, details := parseEntryListFilter(r)
	if len(details) > 0 {
		s.writeError(w, r, http.StatusBadRequest, "validation_error", "Invalid request payload", details)
		return
	}

	result, err := s.store.Entries().List(r.Context(), filter)
	if err != nil {
		s.logger.Error("list entries failed", "error", err)
		s.writeError(w, r, http.StatusInternalServerError, "internal_error", "Could not list entries", nil)
		return
	}

	hasNext := int64(filter.Page*filter.PageSize) < result.Total
	data := make([]map[string]any, 0, len(result.Entries))
	for _, entry := range result.Entries {
		data = append(data, entryResponse(entry))
	}

	s.writeJSON(w, http.StatusOK, map[string]any{
		"data": data,
		"meta": map[string]any{
			"page":      filter.Page,
			"page_size": filter.PageSize,
			"total":     result.Total,
			"has_next":  hasNext,
		},
	})
}

func (s *Server) requireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(s.cfg.SessionCookieName)
		if err != nil || cookie.Value == "" {
			s.writeError(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", nil)
			return
		}

		now := time.Now().UTC()
		session, err := s.store.Sessions().GetByID(r.Context(), cookie.Value)
		if errors.Is(err, store.ErrNotFound) {
			s.writeError(w, r, http.StatusUnauthorized, "unauthorized", "Session expired", nil)
			return
		}
		if err != nil {
			s.writeError(w, r, http.StatusInternalServerError, "internal_error", "Could not validate session", nil)
			return
		}
		if now.After(session.ExpiresAt) || now.Sub(session.LastSeenAt) > sessionIdleTimeout {
			_ = s.store.Sessions().Delete(r.Context(), cookie.Value)
			s.writeError(w, r, http.StatusUnauthorized, "unauthorized", "Session expired", nil)
			return
		}
		if err := s.store.Sessions().Touch(r.Context(), cookie.Value, now); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				s.writeError(w, r, http.StatusUnauthorized, "unauthorized", "Session expired", nil)
				return
			}
			s.writeError(w, r, http.StatusInternalServerError, "internal_error", "Could not refresh session", nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) lookupLoginUser(ctx context.Context, email string) (store.User, error) {
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))
	if normalizedEmail != "" {
		return s.store.Users().GetByEmail(ctx, normalizedEmail)
	}

	if configuredEmail := strings.TrimSpace(strings.ToLower(s.cfg.BootstrapEmail)); configuredEmail != "" {
		user, err := s.store.Users().GetByEmail(ctx, configuredEmail)
		if err == nil {
			return user, nil
		}
		if !errors.Is(err, store.ErrNotFound) {
			return store.User{}, err
		}
	}

	return s.store.Users().GetFirst(ctx)
}

func validateCreateEntryRequest(reflection, passage string, energy *int, sourceID *int64, sourceTitle string, threadID *int64, revisitID *int64) []map[string]string {
	details := make([]map[string]string, 0, 4)

	reflection = strings.TrimSpace(reflection)
	if reflection == "" {
		details = append(details, map[string]string{"field": "reflection", "reason": "required"})
	} else if len(reflection) > 10000 {
		details = append(details, map[string]string{"field": "reflection", "reason": "too_long"})
	}

	if len(strings.TrimSpace(passage)) > 10000 {
		details = append(details, map[string]string{"field": "passage", "reason": "too_long"})
	}

	if energy != nil && (*energy < 1 || *energy > 5) {
		details = append(details, map[string]string{"field": "energy", "reason": "out_of_range"})
	}

	if sourceID == nil && strings.TrimSpace(sourceTitle) == "" {
		details = append(details, map[string]string{"field": "source", "reason": "required"})
	}
	if threadID != nil {
		details = append(details, map[string]string{"field": "thread_id", "reason": "not_supported_yet"})
	}
	if revisitID != nil {
		details = append(details, map[string]string{"field": "revisit_of_entry_id", "reason": "not_supported_yet"})
	}

	return details
}

func parseEntryListFilter(r *http.Request) (store.EntryListFilter, []map[string]string) {
	values := r.URL.Query()
	filter := store.EntryListFilter{
		Page:     1,
		PageSize: 20,
		Tag:      values.Get("tag"),
	}
	details := make([]map[string]string, 0, 4)

	if raw := strings.TrimSpace(values.Get("source_id")); raw != "" {
		sourceID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || sourceID <= 0 {
			details = append(details, map[string]string{"field": "source_id", "reason": "invalid"})
		} else {
			filter.SourceID = &sourceID
		}
	}

	if raw := strings.TrimSpace(values.Get("page")); raw != "" {
		page, err := strconv.Atoi(raw)
		if err != nil || page <= 0 {
			details = append(details, map[string]string{"field": "page", "reason": "invalid"})
		} else {
			filter.Page = page
		}
	}

	if raw := strings.TrimSpace(values.Get("page_size")); raw != "" {
		pageSize, err := strconv.Atoi(raw)
		if err != nil || pageSize <= 0 || pageSize > 100 {
			details = append(details, map[string]string{"field": "page_size", "reason": "invalid"})
		} else {
			filter.PageSize = pageSize
		}
	}

	if raw := strings.TrimSpace(values.Get("from")); raw != "" {
		from, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			details = append(details, map[string]string{"field": "from", "reason": "invalid"})
		} else {
			filter.From = &from
		}
	}

	if raw := strings.TrimSpace(values.Get("to")); raw != "" {
		to, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			details = append(details, map[string]string{"field": "to", "reason": "invalid"})
		} else {
			filter.To = &to
		}
	}

	if filter.From != nil && filter.To != nil && filter.From.After(*filter.To) {
		details = append(details, map[string]string{"field": "date_range", "reason": "invalid"})
	}

	return filter, details
}

func entryResponse(entry store.Entry) map[string]any {
	tagPayload := make([]map[string]any, 0, len(entry.Tags))
	for _, tag := range entry.Tags {
		tagPayload = append(tagPayload, map[string]any{
			"id":          tag.ID,
			"slug":        tag.Slug,
			"label":       tag.Label,
			"entry_count": tag.Count,
			"created_at":  tag.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	payload := map[string]any{
		"id":         entry.ID,
		"source_id":  entry.SourceID,
		"passage":    entry.Passage,
		"reflection": entry.Reflection,
		"mood":       entry.Mood,
		"tags":       tagPayload,
		"created_at": entry.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at": entry.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if entry.Energy != nil {
		payload["energy"] = *entry.Energy
	} else {
		payload["energy"] = nil
	}
	return payload
}

func (s *Server) csrfMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}

		headerToken := strings.TrimSpace(r.Header.Get(s.cfg.CSRFHeaderName))
		cookie, err := r.Cookie(s.cfg.CSRFCookieName)
		if err != nil || cookie.Value == "" || headerToken == "" {
			s.writeError(w, r, http.StatusForbidden, "forbidden", "CSRF token required", nil)
			return
		}

		if subtle.ConstantTimeCompare([]byte(headerToken), []byte(cookie.Value)) != 1 {
			s.writeError(w, r, http.StatusForbidden, "forbidden", "Invalid CSRF token", nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleNotFoundOrSPA(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/") {
		s.writeError(w, r, http.StatusNotFound, "not_found", "Route not found", nil)
		return
	}
	s.serveSPA(w, r)
}

func (s *Server) serveSPA(w http.ResponseWriter, r *http.Request) {
	clean := path.Clean(strings.TrimPrefix(r.URL.Path, "/"))
	if clean == "." || clean == "" {
		clean = "index.html"
	}

	if strings.Contains(clean, "..") {
		s.writeError(w, r, http.StatusBadRequest, "validation_error", "Invalid path", nil)
		return
	}

	if _, err := fs.Stat(s.staticFS, clean); err == nil {
		clone := r.Clone(r.Context())
		clone.URL.Path = "/" + clean
		s.staticServer.ServeHTTP(w, clone)
		return
	}

	if strings.ContainsRune(clean, '.') {
		http.NotFound(w, r)
		return
	}

	index, err := fs.ReadFile(s.staticFS, "index.html")
	if err != nil {
		s.writeError(w, r, http.StatusInternalServerError, "internal_error", "Static assets unavailable", nil)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(index)
}

func (s *Server) notImplemented(route string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.writeError(w, r, http.StatusNotImplemented, "not_implemented", route+" not implemented", nil)
	}
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (s *Server) writeError(w http.ResponseWriter, r *http.Request, status int, code, message string, details any) {
	reqID := middleware.GetReqID(r.Context())
	s.writeJSON(w, status, map[string]any{
		"error": map[string]any{
			"code":       code,
			"message":    message,
			"details":    details,
			"request_id": reqID,
		},
	})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()
		next.ServeHTTP(wrapped, r)

		routePattern := ""
		if rc := chi.RouteContext(r.Context()); rc != nil {
			routePattern = rc.RoutePattern()
		}
		if routePattern == "" {
			routePattern = r.URL.Path
		}

		s.logger.Info("http_request",
			"request_id", middleware.GetReqID(r.Context()),
			"method", r.Method,
			"route", routePattern,
			"status", wrapped.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"bytes", wrapped.BytesWritten(),
		)
	})
}

func (s *Server) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

func randomToken(numBytes int) (string, error) {
	buf := make([]byte, numBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
