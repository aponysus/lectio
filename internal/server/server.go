package server

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/aponysus/lectio/internal/config"
	"github.com/aponysus/lectio/ui"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	sessionTTL         = 14 * 24 * time.Hour
	sessionIdleTimeout = 24 * time.Hour
)

type sessionState struct {
	ExpiresAt time.Time
	LastSeen  time.Time
}

type Server struct {
	cfg          config.Config
	logger       *slog.Logger
	staticFS     fs.FS
	staticServer http.Handler

	mu       sync.RWMutex
	sessions map[string]sessionState
}

func New(cfg config.Config, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	assets := ui.FS()

	return &Server{
		cfg:          cfg,
		logger:       logger,
		staticFS:     assets,
		staticServer: http.FileServer(http.FS(assets)),
		sessions:     make(map[string]sessionState),
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

			protected.Get("/entries", s.notImplemented("GET /api/entries"))
			protected.Post("/entries", s.notImplemented("POST /api/entries"))
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
	s.writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"status":                  "ok",
			"db_status":               "not_connected",
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

	if subtle.ConstantTimeCompare([]byte(req.Password), []byte(s.cfg.BootstrapPassword)) != 1 {
		s.writeError(w, r, http.StatusUnauthorized, "unauthorized", "Invalid credentials", nil)
		return
	}
	if s.cfg.BootstrapEmail != "" && !strings.EqualFold(strings.TrimSpace(req.Email), s.cfg.BootstrapEmail) {
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
	s.mu.Lock()
	s.sessions[sessionID] = sessionState{ExpiresAt: now.Add(sessionTTL), LastSeen: now}
	s.mu.Unlock()

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
		s.mu.Lock()
		delete(s.sessions, cookie.Value)
		s.mu.Unlock()
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

func (s *Server) requireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(s.cfg.SessionCookieName)
		if err != nil || cookie.Value == "" {
			s.writeError(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", nil)
			return
		}

		now := time.Now().UTC()
		s.mu.Lock()
		state, ok := s.sessions[cookie.Value]
		if !ok || now.After(state.ExpiresAt) || now.Sub(state.LastSeen) > sessionIdleTimeout {
			delete(s.sessions, cookie.Value)
			s.mu.Unlock()
			s.writeError(w, r, http.StatusUnauthorized, "unauthorized", "Session expired", nil)
			return
		}
		state.LastSeen = now
		s.sessions[cookie.Value] = state
		s.mu.Unlock()

		next.ServeHTTP(w, r)
	})
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
