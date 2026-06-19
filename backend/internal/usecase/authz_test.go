package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/docvault/backend/internal/platform/apperr"
	"github.com/docvault/backend/internal/repository"
)

// fakeAuthzACL is a minimal AuthorizerACL recording calls and returning
// configured results, so the Authorizer's decision logic can be tested without
// a database.
type fakeAuthzACL struct {
	visible       bool
	visErr        error
	writable      bool
	writErr       error
	folderVisible bool
	folderErr     error
	groupErr      error

	groupCalls    int
	docVisCalls   int
	docWriteCalls int
	folderCalls   int
	lastDocParams repository.VisibilityParams
}

func (f *fakeAuthzACL) ListUserGroupIDs(_ context.Context, _, _ string) ([]string, error) {
	f.groupCalls++
	if f.groupErr != nil {
		return nil, f.groupErr
	}
	return []string{"g1"}, nil
}

func (f *fakeAuthzACL) IsDocumentVisible(_ context.Context, p repository.VisibilityParams) (bool, error) {
	f.docVisCalls++
	f.lastDocParams = p
	return f.visible, f.visErr
}

func (f *fakeAuthzACL) IsDocumentWritable(_ context.Context, p repository.VisibilityParams) (bool, error) {
	f.docWriteCalls++
	f.lastDocParams = p
	return f.writable, f.writErr
}

func (f *fakeAuthzACL) IsFolderVisible(_ context.Context, _ repository.FolderVisibilityParams) (bool, error) {
	f.folderCalls++
	return f.folderVisible, f.folderErr
}

func memberPrincipal() Principal {
	return Principal{TenantID: "t1", OrgID: "o1", UserID: "u1", Role: "member", IsAdmin: false}
}

func adminPrincipal() Principal {
	return Principal{TenantID: "t1", OrgID: "o1", UserID: "a1", Role: "admin", IsAdmin: true}
}

func assertNotFound(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected a NotFound error, got nil")
	}
	if apperr.KindOf(err) != apperr.KindNotFound {
		t.Fatalf("expected KindNotFound, got kind %v (err=%v)", apperr.KindOf(err), err)
	}
}

func TestAuthorizer_AdminShortCircuitsWithoutAnyLookup(t *testing.T) {
	acl := &fakeAuthzACL{visible: false, writable: false, folderVisible: false}
	a := NewAuthorizer(acl)
	ctx := context.Background()

	if err := a.RequireDocVisible(ctx, adminPrincipal(), "d1"); err != nil {
		t.Fatalf("admin doc-visible: %v", err)
	}
	if err := a.RequireDocWritable(ctx, adminPrincipal(), "d1"); err != nil {
		t.Fatalf("admin doc-writable: %v", err)
	}
	if err := a.RequireFolderVisible(ctx, adminPrincipal(), "f1"); err != nil {
		t.Fatalf("admin folder-visible: %v", err)
	}
	if acl.groupCalls+acl.docVisCalls+acl.docWriteCalls+acl.folderCalls != 0 {
		t.Fatalf("admin path hit the ACL repository: groups=%d docVis=%d docWrite=%d folder=%d",
			acl.groupCalls, acl.docVisCalls, acl.docWriteCalls, acl.folderCalls)
	}
}

func TestAuthorizer_MemberVisibleAllowsAndSendsIsAdminFalse(t *testing.T) {
	acl := &fakeAuthzACL{visible: true}
	a := NewAuthorizer(acl)

	if err := a.RequireDocVisible(context.Background(), memberPrincipal(), "d1"); err != nil {
		t.Fatalf("visible doc should be allowed: %v", err)
	}
	if acl.docVisCalls != 1 {
		t.Fatalf("IsDocumentVisible calls = %d, want 1", acl.docVisCalls)
	}
	if acl.lastDocParams.IsAdmin {
		t.Fatal("member visibility check sent IsAdmin=true; want false")
	}
	if acl.lastDocParams.DocumentID != "d1" {
		t.Fatalf("document id = %q, want d1", acl.lastDocParams.DocumentID)
	}
}

func TestAuthorizer_InvisibleAndErrorsBothMapToNotFound(t *testing.T) {
	// Not visible.
	assertNotFound(t, NewAuthorizer(&fakeAuthzACL{visible: false}).
		RequireDocVisible(context.Background(), memberPrincipal(), "d1"))
	// Visibility lookup error.
	assertNotFound(t, NewAuthorizer(&fakeAuthzACL{visErr: errors.New("boom")}).
		RequireDocVisible(context.Background(), memberPrincipal(), "d1"))
	// Group lookup error (cannot resolve memberships -> deny as not found).
	assertNotFound(t, NewAuthorizer(&fakeAuthzACL{groupErr: errors.New("boom")}).
		RequireDocVisible(context.Background(), memberPrincipal(), "d1"))
	// Not writable.
	assertNotFound(t, NewAuthorizer(&fakeAuthzACL{writable: false}).
		RequireDocWritable(context.Background(), memberPrincipal(), "d1"))
	// Folder not visible.
	assertNotFound(t, NewAuthorizer(&fakeAuthzACL{folderVisible: false}).
		RequireFolderVisible(context.Background(), memberPrincipal(), "f1"))
}
