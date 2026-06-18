-- name: CreateDocument :exec
INSERT INTO documents (id, tenant_id, org_id, folder_id, owner_id, title, doc_type, status, language, processing_stage, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW());

-- name: GetDocumentByID :one
SELECT id, tenant_id, org_id, folder_id, owner_id, title, doc_type, status, language,
       processing_stage, processing_error, created_at
FROM documents
WHERE id = $1 AND tenant_id = $2 AND org_id = $3;

-- name: ListDocuments :many
SELECT d.id, d.tenant_id, d.org_id, d.folder_id, d.owner_id, d.title, d.doc_type, d.status, d.language, d.created_at
FROM documents d
WHERE d.tenant_id = sqlc.arg(tenant_id_arg)
  AND d.org_id = sqlc.arg(org_id_arg)
  AND (sqlc.narg(doc_type)::text IS NULL OR d.doc_type::text = sqlc.narg(doc_type)::text)
  AND (sqlc.narg(folder_id)::uuid IS NULL OR d.folder_id = sqlc.narg(folder_id)::uuid)
  AND (sqlc.narg(status)::text IS NULL OR d.status::text = sqlc.narg(status)::text)
  AND (sqlc.narg(language)::text IS NULL OR d.language = sqlc.narg(language)::text)
  AND (sqlc.narg(cursor_id)::uuid IS NULL OR d.created_at < (SELECT created_at FROM documents WHERE id = sqlc.narg(cursor_id)::uuid))
ORDER BY d.created_at DESC
LIMIT sqlc.arg(limit_count);

-- name: UpdateDocument :exec
UPDATE documents
SET title = $1, doc_type = $2, status = $3, language = $4, folder_id = $5
WHERE id = $6 AND tenant_id = $7 AND org_id = $8;

-- name: DeleteDocument :execrows
DELETE FROM documents
WHERE id = $1 AND tenant_id = $2 AND org_id = $3;

-- name: CreateDocumentVersion :exec
INSERT INTO document_versions (id, document_id, version_number, storage_key, mime_type, file_size, uploaded_by, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW());

-- name: GetDocumentVersions :many
SELECT dv.id, dv.document_id, dv.version_number, dv.storage_key, dv.mime_type, dv.file_size, dv.uploaded_by, dv.created_at
FROM document_versions dv
JOIN documents d ON dv.document_id = d.id
WHERE dv.document_id = $1 AND d.tenant_id = $2
ORDER BY dv.version_number DESC;

-- name: CreateDocumentPage :exec
INSERT INTO document_pages (id, document_id, version_id, page_number, ocr_text, translated_text, confidence, ocr_model, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW());

-- name: GetDocumentPages :many
SELECT p.id, p.document_id, p.version_id, p.page_number, p.ocr_text, p.translated_text, p.confidence, p.ocr_model, p.created_at
FROM document_pages p
JOIN documents d ON d.id = p.document_id
WHERE p.document_id = $1 AND d.tenant_id = $2
ORDER BY page_number ASC;

-- name: SetDocumentMetadata :exec
INSERT INTO document_metadata (id, document_id, key, extracted_value, corrected_value, corrected_by, corrected_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
ON CONFLICT (document_id, key) DO UPDATE SET
  extracted_value = EXCLUDED.extracted_value,
  corrected_value = EXCLUDED.corrected_value,
  corrected_by = EXCLUDED.corrected_by,
  corrected_at = EXCLUDED.corrected_at;

-- name: GetDocumentMetadata :many
SELECT m.id, m.document_id, m.key, m.extracted_value, m.corrected_value, m.corrected_by, m.corrected_at, m.created_at
FROM document_metadata m
JOIN documents d ON d.id = m.document_id
WHERE m.document_id = $1 AND d.tenant_id = $2
ORDER BY key ASC;

-- name: UpdateDocumentMetadataField :execrows
UPDATE document_metadata m
SET corrected_value = $1, corrected_by = $2, corrected_at = NOW()
FROM documents d
WHERE m.document_id = d.id AND m.document_id = $3 AND m.key = $4 AND d.tenant_id = $5;

-- name: UpdateDocumentProcessingFields :exec
UPDATE documents
SET processing_stage = COALESCE($1, processing_stage),
    processing_error = COALESCE($2, processing_error)
WHERE id = $3 AND tenant_id = $4;

-- name: GetDocumentStats :one
WITH doc_stats AS (
  SELECT
    COUNT(*) as total_documents,
    COUNT(*) FILTER (WHERE status = 'pending' OR status = 'processing') as pending_documents,
    COUNT(*) FILTER (WHERE status = 'processed' AND created_at >= NOW() - INTERVAL '7 days') as completed_this_week
  FROM documents d
  WHERE d.tenant_id = sqlc.arg(tenant_id_arg)
    AND d.org_id = sqlc.arg(org_id_arg)
),
storage_stats AS (
  SELECT COALESCE(SUM(v.file_size), 0)::bigint as storage_used_bytes
  FROM documents d
  JOIN document_versions v ON d.id = v.document_id
  WHERE d.tenant_id = sqlc.arg(tenant_id_arg)
    AND d.org_id = sqlc.arg(org_id_arg)
)
SELECT d.total_documents, d.pending_documents, d.completed_this_week, s.storage_used_bytes
FROM doc_stats d, storage_stats s;
