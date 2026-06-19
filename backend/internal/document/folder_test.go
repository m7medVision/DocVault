package document

import "testing"

func TestMaxFolderDepth(t *testing.T) {
	if MaxFolderDepth != 12 {
		t.Errorf("MaxFolderDepth = %d, want 12", MaxFolderDepth)
	}
}

func TestIsSelfParent(t *testing.T) {
	if !IsSelfParent("f1", "f1") {
		t.Error("a folder targeting itself must be a self-parent")
	}
	if IsSelfParent("f1", "f2") {
		t.Error("distinct folders must not be a self-parent")
	}
	// An empty folder id must never be treated as a self-parent.
	if IsSelfParent("", "") {
		t.Error("empty folder id must not be a self-parent")
	}
}
