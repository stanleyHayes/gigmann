package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/xcreativs/gigmann/internal/ports"
)

// FacilityEmbeddingRepo is a pgvector implementation of
// ports.FacilityEmbeddingRepository. Vectors are passed as a text literal cast to
// ::vector, so no pgvector-go dependency is needed; the HNSW index serves the
// cosine-distance (<=>) ORDER BY.
type FacilityEmbeddingRepo struct {
	pool *pgxpool.Pool
}

var _ ports.FacilityEmbeddingRepository = (*FacilityEmbeddingRepo)(nil)

// NewFacilityEmbeddingRepo builds the repo over a pgx pool.
func NewFacilityEmbeddingRepo(pool *pgxpool.Pool) *FacilityEmbeddingRepo {
	return &FacilityEmbeddingRepo{pool: pool}
}

// Upsert stores (or replaces) a facility's content and embedding.
func (r *FacilityEmbeddingRepo) Upsert(ctx context.Context, facilityID, content string, embedding []float32) error {
	if _, err := r.pool.Exec(ctx,
		`INSERT INTO facility_embeddings (facility_id, content, embedding)
		 VALUES ($1, $2, $3::vector)
		 ON CONFLICT (facility_id) DO UPDATE SET content = EXCLUDED.content, embedding = EXCLUDED.embedding`,
		facilityID, content, vectorString(embedding)); err != nil {
		return fmt.Errorf("postgres: upsert facility embedding %q: %w", facilityID, err)
	}
	return nil
}

// Count returns the number of stored facility embeddings.
func (r *FacilityEmbeddingRepo) Count(ctx context.Context) (int, error) {
	var n int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM facility_embeddings`).Scan(&n); err != nil {
		return 0, fmt.Errorf("postgres: count facility embeddings: %w", err)
	}
	return n, nil
}

// Search returns the `limit` facilities whose embeddings are nearest (cosine) to
// the query embedding, closest first.
func (r *FacilityEmbeddingRepo) Search(ctx context.Context, embedding []float32, limit int) ([]ports.FacilityMatch, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT facility_id, content, embedding <=> $1::vector AS distance
		 FROM facility_embeddings
		 ORDER BY embedding <=> $1::vector
		 LIMIT $2`,
		vectorString(embedding), limit)
	if err != nil {
		return nil, fmt.Errorf("postgres: search facility embeddings: %w", err)
	}
	defer rows.Close()
	var out []ports.FacilityMatch
	for rows.Next() {
		var m ports.FacilityMatch
		if err := rows.Scan(&m.FacilityID, &m.Content, &m.Distance); err != nil {
			return nil, fmt.Errorf("postgres: scan facility match: %w", err)
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: iterate facility matches: %w", err)
	}
	return out, nil
}

// vectorString formats a float slice as a pgvector text literal: [a,b,c].
func vectorString(vec []float32) string {
	var b strings.Builder
	b.WriteByte('[')
	for i, x := range vec {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(x), 'f', -1, 32))
	}
	b.WriteByte(']')
	return b.String()
}
