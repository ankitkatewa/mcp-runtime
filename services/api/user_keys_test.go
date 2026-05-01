package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListUserAPIKeysReturnsUnavailableWhenKubernetesUnavailable(t *testing.T) {
	server := &RuntimeServer{}

	_, err := server.ListUserAPIKeys(context.Background(), "user-123")
	if err == nil {
		t.Fatal("ListUserAPIKeys() error = nil, want kubernetes not available")
	}
	if !strings.Contains(err.Error(), "kubernetes not available") {
		t.Fatalf("ListUserAPIKeys() error = %v, want substring %q", err, "kubernetes not available")
	}
}

type recordingUserAPIKeyStore struct {
	listUserID string
	listKeys   []userAPIKeySummary
}

func (s *recordingUserAPIKeyStore) AuthenticateUserAPIKey(context.Context, string) (principal, bool, error) {
	return principal{}, false, nil
}

func (s *recordingUserAPIKeyStore) ListUserAPIKeys(_ context.Context, userID string) ([]userAPIKeySummary, error) {
	s.listUserID = userID
	return s.listKeys, nil
}

func (s *recordingUserAPIKeyStore) CreateUserAPIKey(context.Context, string, string) (userAPIKeySummary, string, error) {
	return userAPIKeySummary{}, "", nil
}

func (s *recordingUserAPIKeyStore) RevokeUserAPIKey(context.Context, string, string) (userAPIKeySummary, error) {
	return userAPIKeySummary{}, nil
}

func TestHandleUserAPIKeysUsesInternalUserID(t *testing.T) {
	store := &recordingUserAPIKeyStore{
		listKeys: []userAPIKeySummary{{ID: "uk_1", Name: "default", Prefix: "mcpu_123"}},
	}
	server := &apiServer{userKeys: store}
	req := httptest.NewRequest(http.MethodGet, "/api/user/api-keys", nil)
	req = req.WithContext(context.WithValue(req.Context(), principalContextKey{}, principal{
		Role:    roleUser,
		Subject: "google-sub-123",
		UserID:  "6d5d8c5a-4c8d-4e50-9e34-3d6439f1aa55",
	}))
	rec := httptest.NewRecorder()

	server.handleUserAPIKeys(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if store.listUserID != "6d5d8c5a-4c8d-4e50-9e34-3d6439f1aa55" {
		t.Fatalf("ListUserAPIKeys called with %q, want internal user id", store.listUserID)
	}

	var payload struct {
		Keys []userAPIKeySummary `json:"keys"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Keys) != 1 || payload.Keys[0].ID != "uk_1" {
		t.Fatalf("keys = %#v, want one returned key", payload.Keys)
	}
}
