package middleware

import (
	"net/http"

	lectioauth "github.com/aponysus/lectio/internal/auth"
	"github.com/aponysus/lectio/internal/server/httpx"
)

func EnsureCSRF(manager *lectioauth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := manager.EnsureCSRFCookie(w, r); err != nil {
				httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to issue csrf token")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func OptionalSession(manager *lectioauth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(lectioauth.SessionCookieName)
			if err == nil {
				session, verifyErr := manager.VerifySession(cookie.Value)
				if verifyErr == nil {
					r = r.WithContext(lectioauth.WithSession(r.Context(), session))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := lectioauth.SessionFromContext(r.Context()); !ok {
				httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireCSRF(manager *lectioauth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := manager.ValidateCSRFRequest(r); err != nil {
				httpx.WriteError(w, http.StatusForbidden, "forbidden", err.Error())
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
