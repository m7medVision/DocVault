-- name: SearchDocumentChunks :many
WITH RECURSIVE folder_ancestors AS (
  SELECT f.id AS origin_folder_id, f.id AS ancestor_id, f.parent_id, f.is_restricted, ARRAY[f.id] AS path
  FROM folders f WHERE f.org_id = sqlc.arg(org_id)
  UNION ALL
  SELECT fa.origin_folder_id, pf.id, pf.parent_id, pf.is_restricted, fa.path || pf.id
  FROM folders pf JOIN folder_ancestors fa ON pf.id = fa.parent_id
  WHERE NOT pf.id = ANY(fa.path) AND array_length(fa.path, 1) < 100
),
filtered_chunks AS (
  SELECT
    c.document_id,
    d.title,
    d.doc_type,
    c.id AS chunk_id,
    c.chunk_text,
    COALESCE(p.page_number, 0)::integer AS page_number,
    COALESCE(d.language, '') AS language,
    FALSE AS is_translation,
    c.embedding <=> sqlc.arg(query_vector)::text::vector AS distance,
    GREATEST(0.0, 1 - (c.embedding <=> sqlc.arg(query_vector)::text::vector)) AS raw_semantic_score,
    POSITION(LOWER(sqlc.arg(query_text)) IN LOWER(c.chunk_text)) > 0 AS chunk_contains_query,
    POSITION(LOWER(sqlc.arg(query_text)) IN LOWER(d.title)) > 0 AS title_contains_query
  FROM extracted_text_chunks c
  JOIN documents d ON c.document_id = d.id
  LEFT JOIN document_pages p ON c.page_id = p.id
  WHERE c.embedding IS NOT NULL
    AND d.tenant_id = sqlc.arg(tenant_id)
    AND d.org_id = sqlc.arg(org_id)
    AND (sqlc.narg(doc_type)::text IS NULL OR d.doc_type::text = sqlc.narg(doc_type)::text)
    AND (sqlc.narg(language)::text IS NULL OR d.language = sqlc.narg(language)::text)
    AND (sqlc.narg(status)::text IS NULL OR d.status::text = sqlc.narg(status)::text)
    AND (sqlc.narg(folder_id)::uuid IS NULL OR d.folder_id = sqlc.narg(folder_id)::uuid)
    AND (sqlc.narg(document_id)::uuid IS NULL OR c.document_id = sqlc.narg(document_id)::uuid)
    AND (sqlc.narg(start_date)::timestamptz IS NULL OR d.created_at >= sqlc.narg(start_date)::timestamptz)
    AND (sqlc.narg(end_date)::timestamptz IS NULL OR d.created_at <= sqlc.narg(end_date)::timestamptz)
    AND (
      cardinality(sqlc.arg(tags)::text[]) = 0
      OR NOT EXISTS (
        SELECT 1
        FROM unnest(sqlc.arg(tags)::text[]) AS required_tag(name)
        WHERE NOT EXISTS (
          SELECT 1
          FROM document_tags dt
          JOIN tags t ON t.id = dt.tag_id
          WHERE dt.document_id = d.id
            AND t.tenant_id = d.tenant_id
            AND t.name = required_tag.name
        )
      )
    )
    AND (
      sqlc.arg(is_admin)::boolean
      OR d.owner_id = sqlc.arg(user_id)::uuid
      OR (NOT d.is_restricted AND NOT EXISTS (
           SELECT 1 FROM folder_ancestors fa WHERE fa.origin_folder_id=d.folder_id AND fa.is_restricted))
      OR EXISTS (SELECT 1 FROM acl_grants g
           WHERE g.resource_type='document' AND g.resource_id=d.id AND g.permission='read'
             AND g.org_id = sqlc.arg(org_id)
             AND ((g.principal_type='user' AND g.principal_id=sqlc.arg(user_id)::uuid)
               OR (g.principal_type='group' AND g.principal_id = ANY(sqlc.arg(group_ids)::uuid[]))))
      OR EXISTS (SELECT 1 FROM folder_ancestors fa
           JOIN acl_grants g ON g.resource_type='folder' AND g.resource_id=fa.ancestor_id AND g.permission='read'
             AND g.org_id = sqlc.arg(org_id)
           WHERE fa.origin_folder_id=d.folder_id
             AND ((g.principal_type='user' AND g.principal_id=sqlc.arg(user_id)::uuid)
               OR (g.principal_type='group' AND g.principal_id = ANY(sqlc.arg(group_ids)::uuid[])))))
),
scored_chunks AS (
  SELECT
    document_id,
    title,
    doc_type,
    chunk_id,
    chunk_text,
    page_number,
    language,
    is_translation,
    distance,
    GREATEST(
      LEAST(1.0, GREATEST(0.0, (raw_semantic_score - 0.4) / 0.6)),
      CASE
        WHEN LOWER(BTRIM(title)) = LOWER(BTRIM(sqlc.arg(query_text))) THEN 1.0
        WHEN LOWER(BTRIM(chunk_text)) = LOWER(BTRIM(sqlc.arg(query_text))) THEN 0.99
        WHEN chunk_contains_query THEN 0.97
        WHEN title_contains_query THEN 0.95
        ELSE 0.0
      END
    )::double precision AS score
  FROM filtered_chunks
),
vector_matches AS (
  SELECT *
  FROM scored_chunks
  ORDER BY distance ASC
  LIMIT sqlc.arg(limit_count)
),
lexical_matches AS (
  SELECT *
  FROM scored_chunks
  WHERE score >= 0.95
  ORDER BY score DESC, distance ASC
  LIMIT sqlc.arg(limit_count)
),
candidate_matches AS (
  SELECT * FROM vector_matches
  UNION ALL
  SELECT * FROM lexical_matches
),
deduped_matches AS (
  SELECT DISTINCT ON (chunk_id)
    document_id,
    title,
    doc_type,
    chunk_id,
    chunk_text,
    page_number,
    language,
    is_translation,
    score,
    distance
  FROM candidate_matches
  ORDER BY chunk_id, score DESC, distance ASC
)
SELECT
  document_id,
  title,
  doc_type,
  chunk_id,
  chunk_text,
  page_number,
  language,
  is_translation,
  score
FROM deduped_matches
ORDER BY score DESC
LIMIT sqlc.arg(limit_count);
