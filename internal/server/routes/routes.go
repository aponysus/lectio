package routes

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	lectioauth "github.com/aponysus/lectio/internal/auth"
	"github.com/aponysus/lectio/internal/config"
	"github.com/aponysus/lectio/internal/server/handlers"
	appmiddleware "github.com/aponysus/lectio/internal/server/middleware"
	"github.com/aponysus/lectio/internal/store"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type Dependencies struct {
	Logger *slog.Logger
	Config config.Config
	Store  *store.Store
	Auth   *lectioauth.Manager
}

func New(deps Dependencies) http.Handler {
	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.Recoverer)
	router.Use(chimiddleware.Timeout(30 * time.Second))
	router.Use(appmiddleware.RequestLogger(deps.Logger))

	healthHandler := handlers.HealthHandler{}
	authHandler := handlers.AuthHandler{
		Logger: deps.Logger,
		Config: deps.Config,
		Auth:   deps.Auth,
	}
	systemHandler := handlers.SystemHandler{
		Config: deps.Config,
		Store:  deps.Store,
	}
	sourceHandler := handlers.SourceHandler{
		Store: deps.Store,
	}
	engagementHandler := handlers.EngagementHandler{
		Store: deps.Store,
	}
	inquiryHandler := handlers.InquiryHandler{
		Store: deps.Store,
	}
	engagementInquiryHandler := handlers.EngagementInquiryHandler{
		Store: deps.Store,
	}

	router.Route("/api", func(r chi.Router) {
		r.Use(appmiddleware.EnsureCSRF(deps.Auth))
		r.Use(appmiddleware.OptionalSession(deps.Auth))

		r.Get("/health", healthHandler.Get)
		r.Get("/auth/session", authHandler.Session)
		r.With(appmiddleware.RequireCSRF(deps.Auth)).Post("/auth/login", authHandler.Login)
		r.With(appmiddleware.RequireCSRF(deps.Auth)).Post("/auth/logout", authHandler.Logout)

		r.Group(func(protected chi.Router) {
			protected.Use(appmiddleware.RequireAuth())
			protected.Get("/system/status", systemHandler.Status)
			protected.Get("/sources", sourceHandler.List)
			protected.Get("/sources/{sourceID}", sourceHandler.Get)
			protected.Get("/engagements", engagementHandler.List)
			protected.Get("/engagements/{engagementID}", engagementHandler.Get)
			protected.Get("/engagements/{engagementID}/inquiries", engagementInquiryHandler.List)
			protected.Get("/inquiries", inquiryHandler.List)
			protected.Get("/inquiries/{inquiryID}", inquiryHandler.Get)
			protected.Get("/inquiries/{inquiryID}/engagements", inquiryHandler.ListEngagements)
			protected.With(appmiddleware.RequireCSRF(deps.Auth)).Post("/sources", sourceHandler.Create)
			protected.With(appmiddleware.RequireCSRF(deps.Auth)).Put("/sources/{sourceID}", sourceHandler.Update)
			protected.With(appmiddleware.RequireCSRF(deps.Auth)).Delete("/sources/{sourceID}", sourceHandler.Archive)
			protected.With(appmiddleware.RequireCSRF(deps.Auth)).Post("/engagements", engagementHandler.Create)
			protected.With(appmiddleware.RequireCSRF(deps.Auth)).Put("/engagements/{engagementID}", engagementHandler.Update)
			protected.With(appmiddleware.RequireCSRF(deps.Auth)).Delete("/engagements/{engagementID}", engagementHandler.Archive)
			protected.With(appmiddleware.RequireCSRF(deps.Auth)).Put("/engagements/{engagementID}/inquiries", engagementInquiryHandler.Replace)
			protected.With(appmiddleware.RequireCSRF(deps.Auth)).Post("/inquiries", inquiryHandler.Create)
			protected.With(appmiddleware.RequireCSRF(deps.Auth)).Put("/inquiries/{inquiryID}", inquiryHandler.Update)
			protected.With(appmiddleware.RequireCSRF(deps.Auth)).Delete("/inquiries/{inquiryID}", inquiryHandler.Archive)
		})
	})

	if spa := spaHandler(deps.Config.WebDistDir); spa != nil {
		router.Handle("/*", spa)
	} else {
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<!doctype html><html><body style="font-family: sans-serif; padding: 2rem;"><h1>Lectio API</h1><p>The Vite frontend is not built yet. Run <code>make dev-web</code> during development.</p></body></html>`))
		})
	}

	return router
}

func spaHandler(distDir string) http.Handler {
	indexPath := filepath.Join(distDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		return nil
	}

	fileServer := http.FileServer(http.Dir(distDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		trimmedPath := strings.TrimPrefix(filepath.Clean(r.URL.Path), "/")
		requested := filepath.Join(distDir, trimmedPath)
		if info, err := os.Stat(requested); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, indexPath)
	})
}
