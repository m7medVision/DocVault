-- +goose Up

-- HNSW index for approximate nearest-neighbour cosine similarity search
-- m=16: number of bi-directional links per node (good balance for 1024-dim vectors)
-- ef_construction=64: search width during index build (higher = better recall, slower build)
CREATE INDEX IF NOT EXISTS idx_chunks_embedding_hnsw
    ON extracted_text_chunks
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);

-- GIN index over the generated chunk_tsv column for keyword / phrase matching
-- (to_tsquery @@ chunk_tsv).
CREATE INDEX IF NOT EXISTS idx_chunks_fts
    ON extracted_text_chunks
    USING gin (chunk_tsv);

-- Trigram GIN indexes backing word_similarity() fuzzy containment for proper
-- nouns, IDs, dates and typos (e.g. "teepee" vs "tepee").
CREATE INDEX IF NOT EXISTS idx_chunks_trgm
    ON extracted_text_chunks
    USING gin (chunk_text gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_documents_title_trgm
    ON documents
    USING gin (title gin_trgm_ops);

-- +goose Down
DROP INDEX IF EXISTS idx_documents_title_trgm;
DROP INDEX IF EXISTS idx_chunks_trgm;
DROP INDEX IF EXISTS idx_chunks_fts;
DROP INDEX IF EXISTS idx_chunks_embedding_hnsw;
