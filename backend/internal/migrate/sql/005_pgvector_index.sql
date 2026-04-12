-- +goose Up

-- HNSW index for approximate nearest-neighbour cosine similarity search
-- m=16: number of bi-directional links per node (good balance for 1024-dim vectors)
-- ef_construction=64: search width during index build (higher = better recall, slower build)
CREATE INDEX IF NOT EXISTS idx_chunks_embedding_hnsw
    ON extracted_text_chunks
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);

-- +goose Down
DROP INDEX IF EXISTS idx_chunks_embedding_hnsw;
