package memory

import (
	"context"
	"math"
	"sort"
	"sync"

	"github.com/xcreativs/gigmann/internal/ports"
)

type embeddingRecord struct {
	facilityID string
	content    string
	vec        []float32
}

// FacilityEmbeddingRepo is an in-memory ports.FacilityEmbeddingRepository that
// brute-force ranks by cosine distance — fine for the 12-facility demo network.
type FacilityEmbeddingRepo struct {
	mu    sync.RWMutex
	items []embeddingRecord
}

// NewFacilityEmbeddingRepo creates an empty in-memory embedding store.
func NewFacilityEmbeddingRepo() *FacilityEmbeddingRepo { return &FacilityEmbeddingRepo{} }

var _ ports.FacilityEmbeddingRepository = (*FacilityEmbeddingRepo)(nil)

// Upsert stores or replaces a facility's content and embedding.
func (r *FacilityEmbeddingRepo) Upsert(_ context.Context, facilityID, content string, embedding []float32) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := make([]float32, len(embedding))
	copy(cp, embedding)
	for i := range r.items {
		if r.items[i].facilityID == facilityID {
			r.items[i] = embeddingRecord{facilityID, content, cp}
			return nil
		}
	}
	r.items = append(r.items, embeddingRecord{facilityID, content, cp})
	return nil
}

// Count returns the number of stored embeddings.
func (r *FacilityEmbeddingRepo) Count(_ context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items), nil
}

// Search ranks all stored facilities by cosine distance to the query, nearest first.
func (r *FacilityEmbeddingRepo) Search(_ context.Context, embedding []float32, limit int) ([]ports.FacilityMatch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	matches := make([]ports.FacilityMatch, 0, len(r.items))
	for _, it := range r.items {
		matches = append(matches, ports.FacilityMatch{
			FacilityID: it.facilityID,
			Content:    it.content,
			Distance:   cosineDistance(embedding, it.vec),
		})
	}
	sort.SliceStable(matches, func(i, j int) bool { return matches[i].Distance < matches[j].Distance })
	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}
	return matches, nil
}

func cosineDistance(a, b []float32) float64 {
	if len(a) != len(b) {
		return 2 // maximally distant on mismatch
	}
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	if na == 0 || nb == 0 {
		return 1
	}
	return 1 - dot/(math.Sqrt(na)*math.Sqrt(nb))
}
