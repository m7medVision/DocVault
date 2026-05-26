package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/docvault/backend/internal/config"
	handler "github.com/docvault/backend/internal/transport/http"
)

func newTestHandler() *handler.Handler {
	cfg := &config.Config{Environment: "development"}
	return handler.New(cfg, handler.Dependencies{})
}

// TestHealthCheck tests the basic health endpoint.
func TestHealthCheck(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.HealthCheck(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", response["status"])
	}

	if response["service"] != "docvault-backend" {
		t.Errorf("expected service 'docvault-backend', got %v", response["service"])
	}

	if _, exists := response["timestamp"]; !exists {
		t.Error("expected timestamp in response")
	}
}

// TestHealthReady tests the readiness endpoint.
func TestHealthReady(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	w := httptest.NewRecorder()

	h.HealthReady(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", response["status"])
	}

	components, ok := response["components"].(map[string]interface{})
	if !ok {
		t.Fatal("expected components map in response")
	}

	expectedComponents := []string{"database", "minio", "rabbitmq", "redis", "authz"}
	for _, name := range expectedComponents {
		if _, exists := components[name]; !exists {
			t.Errorf("expected component %s in response", name)
		}
	}
}

// TestUpdateMemberRoleWithoutOwnerRole tests that role changes without owner return 403.
// This test verifies that the handler checks for owner role correctly.
// Note: When called directly without auth context, it returns 403 as expected.
func TestUpdateMemberRoleWithoutOwnerRole(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/members/test-id/role", nil)
	w := httptest.NewRecorder()

	h.UpdateMemberRole(w, req)

	// Without auth context, should be forbidden
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden when called without auth context, got %d", w.Code)
	}
}

// TestGetDocumentWithoutAuth tests that getting a document without auth returns 403.
func TestGetDocumentWithoutAuth(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/test-id", nil)
	w := httptest.NewRecorder()

	h.GetDocument(w, req)

	// Without auth context, should be forbidden
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden when called without auth context, got %d", w.Code)
	}
}

// TestDeleteDocumentWithoutAuth tests that deleting a document without auth returns 403.
func TestDeleteDocumentWithoutAuth(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/documents/test-id", nil)
	w := httptest.NewRecorder()

	h.DeleteDocument(w, req)

	// Without auth context, should be forbidden
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden when called without auth context, got %d", w.Code)
	}
}

// TestUploadDocumentWithoutAuth tests that uploading without auth returns 403.
func TestUploadDocumentWithoutAuth(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/documents/upload", nil)
	w := httptest.NewRecorder()

	h.UploadDocument(w, req)

	// Without auth context, should be forbidden
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden when called without auth context, got %d", w.Code)
	}
}

// TestGetAuditLogWithoutAuth tests that getting audit log without auth returns 403.
func TestGetAuditLogWithoutAuth(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit", nil)
	w := httptest.NewRecorder()

	h.GetAuditLog(w, req)

	// Without auth context, should be forbidden
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden when called without auth context, got %d", w.Code)
	}
}

// TestListMembersWithoutAuth tests that listing members without auth returns 403.
func TestListMembersWithoutAuth(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/members", nil)
	w := httptest.NewRecorder()

	h.ListMembers(w, req)

	// Without auth context, should be forbidden
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden when called without auth context, got %d", w.Code)
	}
}

// TestSearchWithoutAuth tests that searching without auth returns 403.
func TestSearchWithoutAuth(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/search", nil)
	w := httptest.NewRecorder()

	h.Search(w, req)

	// Without auth context, should be forbidden
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden when called without auth context, got %d", w.Code)
	}
}

// TestResponseHeaders tests that responses have correct headers.
func TestResponseHeaders(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.HealthCheck(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", contentType)
	}
}
