// Package validator provides request validation helpers.
package validator

import (
	"strings"
)

// ValidationError represents a field-level validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var messages []string
	for _, e := range ve {
		messages = append(messages, e.Field+": "+e.Message)
	}
	return strings.Join(messages, "; ")
}

// Validator holds validation state.
type Validator struct {
	errors ValidationErrors
}

// New creates a new Validator.
func New() *Validator {
	return &Validator{}
}

// AddError adds a validation error.
func (v *Validator) AddError(field, message string) {
	v.errors = append(v.errors, ValidationError{Field: field, Message: message})
}

// HasErrors returns true if any validation errors exist.
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors returns all validation errors.
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

// ValidateRequired checks that a field is not empty.
func (v *Validator) ValidateRequired(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, "is required")
	}
}

// ValidateMaxLength checks that a field does not exceed a maximum length.
func (v *Validator) ValidateMaxLength(field, value string, max int) {
	if len(value) > max {
		v.AddError(field, "must be at most "+string(rune(max))+" characters")
	}
}

// ValidateEnum checks that a field value is one of the allowed values.
func (v *Validator) ValidateEnum(field, value string, allowed []string) {
	for _, a := range allowed {
		if value == a {
			return
		}
	}
	v.AddError(field, "must be one of: "+strings.Join(allowed, ", "))
}

// ValidateUUID checks that a field looks like a UUID.
func (v *Validator) ValidateUUID(field, value string) {
	if len(value) != 36 {
		v.AddError(field, "must be a valid UUID")
	}
}
