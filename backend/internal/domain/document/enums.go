package document

import (
	"errors"
	"fmt"
)

// MaxFileSize is the largest accepted upload, in bytes (50MB).
const MaxFileSize = 50 * 1024 * 1024

// DocType is the validated kind of an uploaded document.
type DocType string

const (
	DocTypeInvoice  DocType = "invoice"
	DocTypeContract DocType = "contract"
	DocTypeIdentity DocType = "identity"
	DocTypeWarranty DocType = "warranty"
	DocTypeReceipt  DocType = "receipt"
	DocTypeOther    DocType = "other"
)

// ErrInvalidDocType is returned by ParseDocType for an unrecognised value.
var ErrInvalidDocType = errors.New("invalid document type")

// Valid reports whether d is a recognised document type.
func (d DocType) Valid() bool {
	switch d {
	case DocTypeInvoice, DocTypeContract, DocTypeIdentity, DocTypeWarranty, DocTypeReceipt, DocTypeOther:
		return true
	default:
		return false
	}
}

func (d DocType) String() string { return string(d) }

// ParseDocType validates a raw doc-type string, returning ErrInvalidDocType
// (wrapped) when it is not one of the recognised values.
func ParseDocType(s string) (DocType, error) {
	d := DocType(s)
	if !d.Valid() {
		return "", fmt.Errorf("%w: %s", ErrInvalidDocType, s)
	}
	return d, nil
}

// Progress maps a processing stage to its human-readable message and a
// completion percentage. known is false for an unrecognised stage. Failed
// stages report a 0 percentage; callers distinguish them with IsFailed.
func (s ProcessingStage) Progress() (message string, percent int, known bool) {
	switch s {
	case StageUploaded:
		return "Upload received", 5, true
	case StageOCRQueued:
		return "Queued for OCR...", 15, true
	case StageOCRRunning:
		return "Running OCR...", 30, true
	case StageOCRComplete:
		return "OCR complete", 55, true
	case StageProcessingQueued:
		return "Queued for processing...", 65, true
	case StageProcessingRunning:
		return "Extracting metadata...", 75, true
	case StageIndexing:
		return "Indexing document...", 85, true
	case StageSuggesting:
		return "Generating suggestions...", 90, true
	case StageCompleted:
		return "Processing complete", 100, true
	case StageOCRFailed:
		return "OCR failed", 0, true
	case StageProcessingFailed:
		return "Processing failed", 0, true
	default:
		return "", 0, false
	}
}

// IsFailed reports whether the stage is a terminal failure.
func (s ProcessingStage) IsFailed() bool {
	return s == StageOCRFailed || s == StageProcessingFailed
}
