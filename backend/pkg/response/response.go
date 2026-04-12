// Package response provides standardized HTTP response helpers.
package response

import (
	"encoding/json"
	"net/http"
)

// APIResponse is the standard response wrapper.
type APIResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error *APIError   `json:"error,omitempty"`
}

// APIError represents an error response.
type APIError struct {
	Message   string `json:"message"`
	Code      string `json:"code"`
	RequestID string `json:"request_id,omitempty"`
}

// Success writes a successful JSON response.
func Success(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{Data: data})
}

// Error writes an error JSON response.
func Error(w http.ResponseWriter, status int, message, code, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Error: &APIError{
			Message:   message,
			Code:      code,
			RequestID: requestID,
		},
	})
}

// BadRequest writes a 400 error response.
func BadRequest(w http.ResponseWriter, message, requestID string) {
	Error(w, http.StatusBadRequest, message, "BAD_REQUEST", requestID)
}

// Unauthorized writes a 401 error response.
func Unauthorized(w http.ResponseWriter, message, requestID string) {
	Error(w, http.StatusUnauthorized, message, "UNAUTHORIZED", requestID)
}

// Forbidden writes a 403 error response.
func Forbidden(w http.ResponseWriter, message, requestID string) {
	Error(w, http.StatusForbidden, message, "FORBIDDEN", requestID)
}

// NotFound writes a 404 error response.
func NotFound(w http.ResponseWriter, message, requestID string) {
	Error(w, http.StatusNotFound, message, "NOT_FOUND", requestID)
}

// InternalError writes a 500 error response.
func InternalError(w http.ResponseWriter, message, requestID string) {
	Error(w, http.StatusInternalServerError, message, "INTERNAL_ERROR", requestID)
}
