package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	model "github.com/docvault/backend/internal/domain/document"
	"github.com/docvault/backend/internal/middleware"
	"github.com/docvault/backend/internal/repository"
)

// fakeACLRepository records the calls made against it so tests can assert
// whether requireDocVisible hit the database. visible controls the result of
// IsDocumentVisible; visErr is returned ahead of visible when set.
type fakeACLRepository struct {
	visible  bool
	visErr   error
	writable bool
	writeErr error

	isDocumentVisibleCalls  int
	isDocumentWritableCalls int
	listGroupIDsCalls       int
	lastVisibilityParams    repository.VisibilityParams
	lastWritableParams      repository.VisibilityParams
}

func (f *fakeACLRepository) IsDocumentVisible(ctx context.Context, params repository.VisibilityParams) (bool, error) {
	f.isDocumentVisibleCalls++
	f.lastVisibilityParams = params
	if f.visErr != nil {
		return false, f.visErr
	}
	return f.visible, nil
}

func (f *fakeACLRepository) IsDocumentWritable(ctx context.Context, params repository.VisibilityParams) (bool, error) {
	f.isDocumentWritableCalls++
	f.lastWritableParams = params
	if f.writeErr != nil {
		return false, f.writeErr
	}
	return f.writable, nil
}

func (f *fakeACLRepository) ListUserGroupIDs(ctx context.Context, userID, orgID string) ([]string, error) {
	f.listGroupIDsCalls++
	return []string{}, nil
}

func (f *fakeACLRepository) ListVisibleDocuments(ctx context.Context, params repository.ListVisibleParams) ([]model.Document, *string, error) {
	return []model.Document{}, nil, nil
}

func (f *fakeACLRepository) CreateGroup(ctx context.Context, params repository.CreateGroupParams) (repository.Group, error) {
	return repository.Group{}, nil
}

func (f *fakeACLRepository) DeleteGroup(ctx context.Context, tenantID, orgID, groupID string) (int64, error) {
	return 0, nil
}

func (f *fakeACLRepository) ListGroups(ctx context.Context, tenantID, orgID string) ([]repository.Group, error) {
	return nil, nil
}

func (f *fakeACLRepository) AddGroupMember(ctx context.Context, tenantID, orgID, groupID, userID string) error {
	return nil
}

func (f *fakeACLRepository) RemoveGroupMember(ctx context.Context, tenantID, orgID, groupID, userID string) (int64, error) {
	return 0, nil
}

func (f *fakeACLRepository) CreateGrant(ctx context.Context, params repository.CreateGrantParams) (string, error) {
	return "", nil
}

func (f *fakeACLRepository) GrantTargetExists(ctx context.Context, ref repository.ResourceRef) (bool, error) {
	return true, nil
}

func (f *fakeACLRepository) DeleteGrant(ctx context.Context, tenantID, orgID, grantID string) (int64, error) {
	return 0, nil
}

func (f *fakeACLRepository) ListGrantsByResource(ctx context.Context, ref repository.ResourceRef) ([]repository.Grant, error) {
	return nil, nil
}

func (f *fakeACLRepository) DeleteGrantsForResource(ctx context.Context, ref repository.ResourceRef) (int64, error) {
	return 0, nil
}

func (f *fakeACLRepository) SetDocumentRestricted(ctx context.Context, params repository.SetRestrictedParams) (int64, error) {
	return 0, nil
}

func (f *fakeACLRepository) SetFolderRestricted(ctx context.Context, params repository.SetRestrictedParams) (int64, error) {
	return 0, nil
}

var _ repository.ACLRepository = (*fakeACLRepository)(nil)

// requestWithAuth builds a request whose context carries the given auth values,
// matching what the JWT middleware would populate in production.
func requestWithAuth(role, userID string) *http.Request {
	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.UserRoleKey, role)
	ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.TenantIDKey, "tenant-1")
	ctx = context.WithValue(ctx, middleware.OrgIDKey, "org-1")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/doc-1", nil)
	return req.WithContext(ctx)
}

func newHandlerWithACL(fake *fakeACLRepository) *Handler {
	return New(nil, Dependencies{ACLRepo: fake})
}

// An admin must never trigger a database lookup: requireDocVisible short-circuits
// before consulting the ACL repository.
func TestRequireDocVisible_AdminShortCircuitsWithoutRepoCall(t *testing.T) {
	fake := &fakeACLRepository{visible: false} // would deny if consulted
	h := newHandlerWithACL(fake)

	w := httptest.NewRecorder()
	req := requestWithAuth(middleware.RoleAdmin, "admin-1")

	wrote := h.requireDocVisible(w, req, "doc-1")

	if wrote {
		t.Fatal("requireDocVisible wrote a response for an admin; expected short-circuit")
	}
	if fake.isDocumentVisibleCalls != 0 {
		t.Fatalf("IsDocumentVisible called %d times for admin; want 0", fake.isDocumentVisibleCalls)
	}
	if fake.listGroupIDsCalls != 0 {
		t.Fatalf("ListUserGroupIDs called %d times for admin; want 0", fake.listGroupIDsCalls)
	}
}

// The owner role is below admin, so a member-level principal must be evaluated
// against the repository.
func TestRequireDocVisible_MemberConsultsRepoAndAllows(t *testing.T) {
	fake := &fakeACLRepository{visible: true}
	h := newHandlerWithACL(fake)

	w := httptest.NewRecorder()
	req := requestWithAuth(middleware.RoleMember, "member-1")

	wrote := h.requireDocVisible(w, req, "doc-1")

	if wrote {
		t.Fatalf("requireDocVisible wrote a response for a visible doc; status=%d", w.Code)
	}
	if fake.isDocumentVisibleCalls != 1 {
		t.Fatalf("IsDocumentVisible called %d times for member; want 1", fake.isDocumentVisibleCalls)
	}
	if fake.lastVisibilityParams.IsAdmin {
		t.Fatal("member visibility check sent IsAdmin=true; want false")
	}
	if fake.lastVisibilityParams.DocumentID != "doc-1" {
		t.Fatalf("visibility check document id = %q, want doc-1", fake.lastVisibilityParams.DocumentID)
	}
}

// When the repository reports the document is not visible, the response must be
// 404 (not 403) so a member cannot probe for the existence of documents they
// are not allowed to see.
func TestRequireDocVisible_NotVisibleReturns404NotForbidden(t *testing.T) {
	fake := &fakeACLRepository{visible: false}
	h := newHandlerWithACL(fake)

	w := httptest.NewRecorder()
	req := requestWithAuth(middleware.RoleMember, "member-1")

	wrote := h.requireDocVisible(w, req, "doc-1")

	if !wrote {
		t.Fatal("requireDocVisible returned false for an invisible doc; expected it to write an error")
	}
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (not 403) to avoid leaking existence", w.Code)
	}
	if fake.isDocumentVisibleCalls != 1 {
		t.Fatalf("IsDocumentVisible called %d times; want 1", fake.isDocumentVisibleCalls)
	}
}

// An admin must never trigger a database lookup for write checks either:
// requireDocWritable short-circuits before consulting the ACL repository.
func TestRequireDocWritable_AdminShortCircuitsWithoutRepoCall(t *testing.T) {
	fake := &fakeACLRepository{writable: false} // would deny if consulted
	h := newHandlerWithACL(fake)

	w := httptest.NewRecorder()
	req := requestWithAuth(middleware.RoleAdmin, "admin-1")

	wrote := h.requireDocWritable(w, req, "doc-1")

	if wrote {
		t.Fatal("requireDocWritable wrote a response for an admin; expected short-circuit")
	}
	if fake.isDocumentWritableCalls != 0 {
		t.Fatalf("IsDocumentWritable called %d times for admin; want 0", fake.isDocumentWritableCalls)
	}
	if fake.listGroupIDsCalls != 0 {
		t.Fatalf("ListUserGroupIDs called %d times for admin; want 0", fake.listGroupIDsCalls)
	}
}

// When the repository reports the document is not writable for a member, the
// response must be 404 (not 403) so a member cannot probe for existence and a
// read-only grant cannot be used to mutate a restricted document.
func TestRequireDocWritable_NotWritableReturns404(t *testing.T) {
	fake := &fakeACLRepository{writable: false}
	h := newHandlerWithACL(fake)

	w := httptest.NewRecorder()
	req := requestWithAuth(middleware.RoleMember, "member-1")

	wrote := h.requireDocWritable(w, req, "doc-1")

	if !wrote {
		t.Fatal("requireDocWritable returned false for a non-writable doc; expected it to write an error")
	}
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (not 403) to avoid leaking existence", w.Code)
	}
	if !strings.Contains(w.Body.String(), "NOT_FOUND") {
		t.Fatalf("body = %q, want NOT_FOUND code", w.Body.String())
	}
	if fake.isDocumentWritableCalls != 1 {
		t.Fatalf("IsDocumentWritable called %d times; want 1", fake.isDocumentWritableCalls)
	}
	if fake.lastWritableParams.IsAdmin {
		t.Fatal("member writability check sent IsAdmin=true; want false")
	}
}

// A repository error must also be reported as 404, not surfaced as a 5xx, so
// existence is never leaked through error behaviour.
func TestRequireDocVisible_RepoErrorReturns404(t *testing.T) {
	fake := &fakeACLRepository{visErr: context.DeadlineExceeded}
	h := newHandlerWithACL(fake)

	w := httptest.NewRecorder()
	req := requestWithAuth(middleware.RoleMember, "member-1")

	wrote := h.requireDocVisible(w, req, "doc-1")

	if !wrote {
		t.Fatal("requireDocVisible returned false on repo error; expected it to write an error")
	}
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 on repo error", w.Code)
	}
}
