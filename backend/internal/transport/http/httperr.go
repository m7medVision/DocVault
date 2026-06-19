package handler

import (
	"log/slog"
	"net/http"

	"github.com/docvault/backend/internal/platform/apperr"
)

// errorEnvelope is the single error response shape: a client-safe message plus
// a stable machine-readable code.
type errorEnvelope struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// statusForKind is the one place a domain error Kind becomes an HTTP status.
func statusForKind(k apperr.Kind) int {
	switch k {
	case apperr.KindInvalid:
		return http.StatusBadRequest
	case apperr.KindUnauthorized:
		return http.StatusUnauthorized
	case apperr.KindForbidden:
		return http.StatusForbidden
	case apperr.KindNotFound:
		return http.StatusNotFound
	case apperr.KindConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// writeErr translates a domain error into an HTTP response. It is the single
// seam handlers use to report errors: an *apperr.Error maps to its Kind's
// status and exposes only its client-safe Message and Code; any other error is
// treated as internal — its detail is logged server-side and never leaked.
func writeErr(w http.ResponseWriter, r *http.Request, err error) {
	ae, ok := apperr.As(err)
	if !ok {
		ae = apperr.Internal(err)
	}

	status := statusForKind(ae.Kind)
	if status >= http.StatusInternalServerError {
		slog.ErrorContext(r.Context(), "request failed", "code", ae.Code, "error", err)
	}

	writeJSON(w, status, errorEnvelope{Error: ae.Message, Code: ae.Code})
}
