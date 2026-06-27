package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/xcreativs/gigmann/internal/core/facility"
	"github.com/xcreativs/gigmann/internal/ports"
)

// FacilityContent is the text embedded for a facility: the fields a user is
// likely to name in a natural-language query (name, region, town, type, manager).
func FacilityContent(f facility.Facility) string {
	return strings.Join([]string{f.Name, string(f.Region), f.Town, f.Type, f.ManagerName}, " ")
}

// SeedFacilityEmbeddings embeds each facility's content and stores it, but only
// when the store is empty (idempotent first-run). It reports whether it embedded.
func SeedFacilityEmbeddings(
	ctx context.Context, embedder ports.Embedder, repo ports.FacilityEmbeddingRepository, facilities []facility.Facility,
) (bool, error) {
	n, err := repo.Count(ctx)
	if err != nil {
		return false, err
	}
	if n > 0 || len(facilities) == 0 {
		return false, nil
	}
	contents := make([]string, len(facilities))
	for i, f := range facilities {
		contents[i] = FacilityContent(f)
	}
	vecs, err := embedder.Embed(ctx, contents, ports.EmbedDocument)
	if err != nil {
		return false, fmt.Errorf("app: embed facilities: %w", err)
	}
	if len(vecs) != len(facilities) {
		return false, fmt.Errorf("app: embedder returned %d vectors for %d facilities", len(vecs), len(facilities))
	}
	for i, f := range facilities {
		if err := repo.Upsert(ctx, f.ID, contents[i], vecs[i]); err != nil {
			return false, err
		}
	}
	return true, nil
}
