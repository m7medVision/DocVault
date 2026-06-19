package document

// MaxFolderDepth caps how deeply folders may nest. The root level is depth 1; a
// folder N levels below root has depth N+1. A reparent is rejected when the
// moved subtree's deepest resulting leaf would exceed this cap. The cap is
// enforced authoritatively inside the advisory-locked repository reparent; this
// constant is the single source of truth both layers share.
const MaxFolderDepth = 12

// IsSelfParent reports whether a reparent would make a folder its own parent —
// the simplest cycle to reject. The locked, cycle-checked move rejects it too,
// but checking here yields a precise error before the lock is taken.
func IsSelfParent(folderID, targetParentID string) bool {
	return folderID != "" && targetParentID == folderID
}
