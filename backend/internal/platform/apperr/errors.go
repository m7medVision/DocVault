// Package apperr defines the application's domain-error taxonomy: a small set
// of error Kinds, a machine-readable Code, and a safe client-facing Message.
//
// Domain and application code returns *Error values (optionally wrapping an
// underlying cause with Wrap) instead of signalling failure modes through
// error-string or Postgres-index-name matching. The transport layer owns the
// single Kind -> HTTP status mapping, so this package stays free of HTTP and
// any other delivery concern.
package apperr

import (
	"errors"
	"fmt"
)

// Kind classifies a failure independently of any transport. The transport layer
// maps each Kind to a concrete status (e.g. HTTP) in exactly one place.
type Kind int

const (
	// KindInternal is an unexpected server-side failure. It is the zero value
	// so that an unclassified error is treated as internal by default.
	KindInternal Kind = iota
	// KindInvalid is a malformed or semantically invalid request.
	KindInvalid
	// KindUnauthorized is a missing or invalid authentication.
	KindUnauthorized
	// KindForbidden is an authenticated principal lacking permission.
	KindForbidden
	// KindNotFound is a missing resource. For visibility-restricted resources
	// it is also returned in place of Forbidden so existence is not leaked.
	KindNotFound
	// KindConflict is a state conflict, e.g. a uniqueness violation.
	KindConflict
)

// Error is a classified, client-safe application error. Message is safe to
// return to callers; cause carries internal detail that must never be exposed.
type Error struct {
	Kind    Kind
	Code    string
	Message string
	cause   error
}

// Error implements the error interface. The cause, when present, is included so
// server-side logs retain the full chain; callers that surface errors to
// clients use Message instead.
func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.cause)
	}
	return e.Message
}

// Unwrap exposes the underlying cause for errors.Is / errors.As.
func (e *Error) Unwrap() error { return e.cause }

// Wrap attaches an underlying cause and returns the same *Error for chaining.
// The cause is kept out of the client-facing Message.
func (e *Error) Wrap(cause error) *Error {
	e.cause = cause
	return e
}

// New constructs an *Error with an explicit Kind.
func New(kind Kind, code, message string) *Error {
	return &Error{Kind: kind, Code: code, Message: message}
}

// NewInvalid builds a KindInvalid error (malformed/invalid request).
func NewInvalid(code, message string) *Error { return New(KindInvalid, code, message) }

// NewUnauthorized builds a KindUnauthorized error.
func NewUnauthorized(code, message string) *Error { return New(KindUnauthorized, code, message) }

// NewForbidden builds a KindForbidden error.
func NewForbidden(code, message string) *Error { return New(KindForbidden, code, message) }

// NewNotFound builds a KindNotFound error.
func NewNotFound(code, message string) *Error { return New(KindNotFound, code, message) }

// NewConflict builds a KindConflict error.
func NewConflict(code, message string) *Error { return New(KindConflict, code, message) }

// Internal wraps an arbitrary cause as an internal failure with a generic,
// non-leaking client message.
func Internal(cause error) *Error {
	return &Error{Kind: KindInternal, Code: "INTERNAL", Message: "internal server error", cause: cause}
}

// As returns the first *Error in err's chain, if any.
func As(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

// KindOf returns the Kind of err, or KindInternal when err is not an *Error.
func KindOf(err error) Kind {
	if e, ok := As(err); ok {
		return e.Kind
	}
	return KindInternal
}
