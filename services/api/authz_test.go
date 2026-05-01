package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSplitCSVSet(t *testing.T) {
	got := splitCSVSet(" a, b,,c , ")
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if _, ok := got["a"]; !ok {
		t.Fatal("missing a")
	}
}

func TestAuthenticateRequestStaticKey_DefaultAdmin(t *testing.T) {
	srv := &apiServer{
		apiKeys: map[string]struct{}{"key-1": {}},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	req.Header.Set("x-api-key", "key-1")
	p, ok, err := srv.authenticateRequest(req)
	if err != nil {
		t.Fatalf("authenticateRequest error: %v", err)
	}
	if !ok {
		t.Fatal("expected authenticated")
	}
	if p.Role != roleAdmin {
		t.Fatalf("role = %q, want %q", p.Role, roleAdmin)
	}
}

func TestAuthenticateRequestStaticKey_AdminAllowlist(t *testing.T) {
	srv := &apiServer{
		apiKeys:      map[string]struct{}{"key-user": {}, "key-admin": {}},
		adminAPIKeys: map[string]struct{}{"key-admin": {}},
	}

	userReq := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	userReq.Header.Set("x-api-key", "key-user")
	p, ok, err := srv.authenticateRequest(userReq)
	if err != nil || !ok {
		t.Fatalf("user auth failed: ok=%v err=%v", ok, err)
	}
	if p.Role != roleUser {
		t.Fatalf("user role = %q, want %q", p.Role, roleUser)
	}

	adminReq := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	adminReq.Header.Set("x-api-key", "key-admin")
	p, ok, err = srv.authenticateRequest(adminReq)
	if err != nil || !ok {
		t.Fatalf("admin auth failed: ok=%v err=%v", ok, err)
	}
	if p.Role != roleAdmin {
		t.Fatalf("admin role = %q, want %q", p.Role, roleAdmin)
	}
}

func TestRequireRole(t *testing.T) {
	srv := &apiServer{}
	handler := srv.requireRole(roleAdmin, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	req = req.WithContext(context.WithValue(req.Context(), principalContextKey{}, principal{Role: roleUser}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestPrincipalUserIDPrefersInternalUserID(t *testing.T) {
	p := principal{
		Subject: "google-sub-123",
		UserID:  "6d5d8c5a-4c8d-4e50-9e34-3d6439f1aa55",
	}

	if got := p.userID(); got != p.UserID {
		t.Fatalf("userID() = %q, want %q", got, p.UserID)
	}

	legacy := principal{Subject: "legacy-subject"}
	if got := legacy.userID(); got != "legacy-subject" {
		t.Fatalf("legacy userID() = %q, want legacy-subject", got)
	}
}

func TestHandleAuthMeReturnsPrincipalSubject(t *testing.T) {
	srv := &apiServer{}
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req = req.WithContext(context.WithValue(req.Context(), principalContextKey{}, principal{
		Role:      roleUser,
		Subject:   "google-sub-123",
		UserID:    "6d5d8c5a-4c8d-4e50-9e34-3d6439f1aa55",
		Email:     "user@example.com",
		Namespace: "user-17",
	}))
	rec := httptest.NewRecorder()

	srv.handleAuthMe(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var payload struct {
		Authenticated bool `json:"authenticated"`
		Principal     struct {
			Subject string `json:"subject"`
			Email   string `json:"email"`
		} `json:"principal"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !payload.Authenticated {
		t.Fatal("authenticated = false, want true")
	}
	if payload.Principal.Subject != "google-sub-123" {
		t.Fatalf("principal.subject = %q, want google-sub-123", payload.Principal.Subject)
	}
	if payload.Principal.Email != "user@example.com" {
		t.Fatalf("principal.email = %q, want user@example.com", payload.Principal.Email)
	}
}
