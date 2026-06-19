package repository

import "errors"

// Sentinel errors returned by the data-access layer and matched by the usecase
// layer with errors.Is. They live here, in the repository contracts package,
// rather than in the per-context postgres adapters so both the adapters (which
// return them) and the usecase layer (which matches them) can reference them
// without the consumer depending on an adapter package.
var (
	// ErrNotificationNotFound is returned when no notification matches the
	// tenant/user/id.
	ErrNotificationNotFound = errors.New("notification not found")

	// ErrReminderNotFound is returned when no reminder rule matches the
	// tenant/id.
	ErrReminderNotFound = errors.New("reminder not found")

	// ErrDocumentNotFound is returned when no document matches the tenant/org/id.
	// Callers detect it via errors.Is rather than matching error-string substrings.
	ErrDocumentNotFound = errors.New("document not found")

	// ErrDocumentTitleExists is returned when a write collides with the per-folder
	// unique document-title constraint (idx_documents_unique_title).
	ErrDocumentTitleExists = errors.New("a document with this title already exists in the folder")
)
