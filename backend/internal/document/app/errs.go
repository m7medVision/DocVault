package app

import "errors"

var (
	ErrDocumentRepositoryNotConfigured = errors.New("document repository not configured")
	ErrFolderRepositoryNotConfigured   = errors.New("folder repository not configured")
	ErrTagRepositoryNotConfigured      = errors.New("tag repository not configured")
	ErrFolderNotFound                  = errors.New("folder not found")
	ErrTagNotFound                     = errors.New("tag not found")
	ErrDocumentNotFound                = errors.New("document not found")

	// ErrFolderSelfParent is returned when a folder is reparented onto itself.
	ErrFolderSelfParent = errors.New("a folder cannot be its own parent")
	// ErrFolderCycle is returned when a reparent would create a cycle, i.e. the
	// target parent is the folder itself or one of its descendants.
	ErrFolderCycle = errors.New("cannot move a folder into itself or one of its descendants")
	// ErrFolderDepthExceeded is returned when a reparent would push the folder
	// past the maximum nesting depth.
	ErrFolderDepthExceeded = errors.New("resulting folder depth exceeds the maximum allowed")
	// ErrTargetParentNotFound is returned when the requested target parent does
	// not exist in the caller's tenant/org.
	ErrTargetParentNotFound = errors.New("target parent folder not found")
)
