package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/ports"
)

const defaultSearchLimit = 5

// FacilityMatch is a facility resolved from a natural-language query, with a
// similarity score in [0,1] (1 = identical).
type FacilityMatch struct {
	FacilityID string
	Name       string
	Score      float64
}

// FacilitySearchService resolves a natural-language phrase to facilities using
// vector similarity. It embeds the query and ranks stored facility embeddings.
type FacilitySearchService struct {
	embedder ports.Embedder
	repo     ports.FacilityEmbeddingRepository
	names    map[string]string
}

// NewFacilitySearchService wires the service with an embedder, an embedding
// store, and the facilities (for resolving display names).
func NewFacilitySearchService(
	embedder ports.Embedder, repo ports.FacilityEmbeddingRepository, facilities []facility.Facility,
) *FacilitySearchService {
	names := make(map[string]string, len(facilities))
	for _, f := range facilities {
		names[f.ID] = f.Name
	}
	return &FacilitySearchService{embedder: embedder, repo: repo, names: names}
}

// Resolve returns up to `limit` facilities most similar to the query, best first.
func (s *FacilitySearchService) Resolve(ctx context.Context, query string, limit int) ([]FacilityMatch, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}
	if limit <= 0 || limit > 20 {
		limit = defaultSearchLimit
	}
	vecs, err := s.embedder.Embed(ctx, []string{query}, ports.EmbedQuery)
	if err != nil {
		return nil, fmt.Errorf("app: embed query: %w", err)
	}
	if len(vecs) == 0 {
		return nil, nil
	}
	matches, err := s.repo.Search(ctx, vecs[0], limit)
	if err != nil {
		return nil, fmt.Errorf("app: search facilities: %w", err)
	}
	out := make([]FacilityMatch, 0, len(matches))
	for _, m := range matches {
		out = append(out, FacilityMatch{
			FacilityID: m.FacilityID,
			Name:       s.names[m.FacilityID],
			Score:      1 - m.Distance,
		})
	}
	return out, nil
}
