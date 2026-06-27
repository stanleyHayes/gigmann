-- 000004_facility_embeddings.up.sql — pgvector NL retrieval (GEC-13). The vector
-- extension is created in 000001. Embeddings are derived data: each facility's
-- text (name/region/town/type/manager) is embedded so natural-language queries
-- can fuzzy-resolve to a facility. Dimension 1024 matches the configured embedder
-- (Voyage voyage-3.5-lite @ output_dimension 1024, or the local fallback).
CREATE TABLE facility_embeddings (
    facility_id text PRIMARY KEY REFERENCES facilities(id) ON DELETE CASCADE,
    content     text NOT NULL,
    embedding   vector(1024) NOT NULL
);

-- HNSW approximate-nearest-neighbour index for cosine distance (<=>).
CREATE INDEX idx_facility_embeddings_hnsw
    ON facility_embeddings USING hnsw (embedding vector_cosine_ops);
