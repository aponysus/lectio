package auth

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestSessionRoundTrip(t *testing.T) {
	manager := NewManager("session-secret", "csrf-secret", 2*time.Hour, false)

	token, err := manager.NewSession(DefaultUserID)
	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}

	session, err := manager.VerifySession(token)
	if err != nil {
		t.Fatalf("VerifySession() error = %v", err)
	}

	if session.Subject != DefaultUserID {
		t.Fatalf("expected subject %q, got %q", DefaultUserID, session.Subject)
	}
}

func TestCSRFRoundTrip(t *testing.T) {
	manager := NewManager("session-secret", "csrf-secret", 2*time.Hour, false)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/api/auth/session", nil)

	token, err := manager.EnsureCSRFCookie(recorder, request)
	if err != nil {
		t.Fatalf("EnsureCSRFCookie() error = %v", err)
	}

	request.Header.Set("X-CSRF-Token", token)
	for _, cookie := range recorder.Result().Cookies() {
		request.AddCookie(cookie)
	}

	if err := manager.ValidateCSRFRequest(request); err != nil {
		t.Fatalf("ValidateCSRFRequest() error = %v", err)
	}
}
