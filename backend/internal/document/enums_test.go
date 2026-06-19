package document

import (
	"errors"
	"testing"
)

func TestParseDocType(t *testing.T) {
	valid := []string{"invoice", "contract", "identity", "warranty", "receipt", "other"}
	for _, s := range valid {
		got, err := ParseDocType(s)
		if err != nil {
			t.Errorf("ParseDocType(%q) unexpected error: %v", s, err)
		}
		if string(got) != s {
			t.Errorf("ParseDocType(%q) = %q", s, got)
		}
	}

	for _, s := range []string{"", "INVOICE", "passport", "bill"} {
		_, err := ParseDocType(s)
		if err == nil {
			t.Errorf("ParseDocType(%q) expected error, got nil", s)
		}
		if !errors.Is(err, ErrInvalidDocType) {
			t.Errorf("ParseDocType(%q) error not ErrInvalidDocType: %v", s, err)
		}
	}

	// Error string must keep the legacy "invalid document type: <value>" shape.
	_, err := ParseDocType("passport")
	if err.Error() != "invalid document type: passport" {
		t.Errorf("error string = %q, want legacy format", err.Error())
	}
}

// TestProcessingStageProgress pins the exact (message, percent) pairs the
// previous stageProgress map produced, so the usecase swap is behaviour-preserving.
func TestProcessingStageProgress(t *testing.T) {
	cases := []struct {
		stage   ProcessingStage
		message string
		percent int
	}{
		{StageUploaded, "Upload received", 5},
		{StageOCRQueued, "Queued for OCR...", 15},
		{StageOCRRunning, "Running OCR...", 30},
		{StageOCRComplete, "OCR complete", 55},
		{StageProcessingQueued, "Queued for processing...", 65},
		{StageProcessingRunning, "Extracting metadata...", 75},
		{StageIndexing, "Indexing document...", 85},
		{StageSuggesting, "Generating suggestions...", 90},
		{StageCompleted, "Processing complete", 100},
		{StageOCRFailed, "OCR failed", 0},
		{StageProcessingFailed, "Processing failed", 0},
	}
	for _, tc := range cases {
		msg, pct, known := tc.stage.Progress()
		if !known {
			t.Errorf("Progress(%q) known=false, want true", tc.stage)
		}
		if msg != tc.message || pct != tc.percent {
			t.Errorf("Progress(%q) = (%q,%d), want (%q,%d)", tc.stage, msg, pct, tc.message, tc.percent)
		}
	}

	if _, _, known := ProcessingStage("bogus").Progress(); known {
		t.Error("Progress(bogus) known=true, want false")
	}
}

func TestProcessingStageIsFailed(t *testing.T) {
	if !StageOCRFailed.IsFailed() || !StageProcessingFailed.IsFailed() {
		t.Error("failed stages must report IsFailed() == true")
	}
	for _, s := range []ProcessingStage{StageUploaded, StageIndexing, StageCompleted} {
		if s.IsFailed() {
			t.Errorf("%q must not be IsFailed()", s)
		}
	}
}
