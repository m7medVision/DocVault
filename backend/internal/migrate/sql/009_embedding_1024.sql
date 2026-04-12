-- +goose Up

-- Recreate the embedding column and HNSW index for mistral-embed-2312 (1024 dimensions).
-- Must drop the index first, then alter the column type, then rebuild the index.

DROP INDEX IF EXISTS idx_chunks_embedding_hnsw;

ALTER TABLE extracted_text_chunks
    ALTER COLUMN embedding TYPE vector(1024);

-- Rebuild HNSW index tuned for 1024-dim vectors
-- m=16: bi-directional links per node (good balance)
-- ef_construction=64: search width during build (higher = better recall)
CREATE INDEX idx_chunks_embedding_hnsw
    ON extracted_text_chunks
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);

-- +goose Down
DROP INDEX IF EXISTS idx_chunks_embedding_hnsw;

-- The retired 768-dimension contract is no longer restored on rollback.
-- Keep the database aligned with the current 1024-dimension embedding model.
ALTER TABLE extracted_text_chunks
    ALTER COLUMN embedding TYPE vector(1024);

CREATE INDEX idx_chunks_embedding_hnsw
    ON extracted_text_chunks
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);
