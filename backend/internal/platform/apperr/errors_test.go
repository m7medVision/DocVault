package apperr

import (
	"errors"
	"fmt"
	"testing"
)

func TestKindOf(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want Kind
	}{
		{"nil is internal", nil, KindInternal},
		{"plain error is internal", errors.New("boom"), KindInternal},
		{"not found", NewNotFound("DOC_NOT_FOUND", "document not found"), KindNotFound},
		{"conflict", NewConflict("TITLE_EXISTS", "title exists"), KindConflict},
		{"invalid", NewInvalid("BAD_INPUT", "bad input"), KindInvalid},
		{"wrapped not found stays not found", fmt.Errorf("loading: %w", NewNotFound("X", "x")), KindNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := KindOf(tc.err); got != tc.want {
				t.Fatalf("KindOf = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestWrapPreservesCauseAndKind(t *testing.T) {
	cause := errors.New("pgx: no rows")
	err := NewNotFound("DOC_NOT_FOUND", "document not found").Wrap(cause)

	if !errors.Is(err, cause) {
		t.Fatal("errors.Is should find the wrapped cause")
	}
	if got := KindOf(err); got != KindNotFound {
		t.Fatalf("KindOf = %v, want KindNotFound", got)
	}

	ae, ok := As(err)
	if !ok {
		t.Fatal("As should extract *Error")
	}
	if ae.Message != "document not found" {
		t.Fatalf("Message = %q, want client-safe message", ae.Message)
	}
	// The cause detail must be present in Error() (for logs) but the client
	// message must not leak it.
	if got := err.Error(); got == ae.Message {
		t.Fatal("Error() should include the cause for server-side logging")
	}
}

func TestInternalHidesDetail(t *testing.T) {
	err := Internal(errors.New("connection refused"))
	if err.Kind != KindInternal {
		t.Fatalf("Kind = %v, want KindInternal", err.Kind)
	}
	if err.Message == "connection refused" {
		t.Fatal("internal error message must not leak the cause")
	}
}
