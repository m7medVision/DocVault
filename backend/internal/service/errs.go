package service

import "errors"

var (
	ErrDocumentRepositoryNotConfigured = errors.New("document repository not configured")
	ErrFolderRepositoryNotConfigured   = errors.New("folder repository not configured")
	ErrTagRepositoryNotConfigured      = errors.New("tag repository not configured")
	ErrFolderNotFound                  = errors.New("folder not found")
	ErrTagNotFound                     = errors.New("tag not found")
	ErrDocumentNotFound                = errors.New("document not found")
)
