package localembedder_test

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xcreativs/gigmann/internal/adapters/outbound/localembedder"
	"github.com/xcreativs/gigmann/internal/ports"
)

func cosine(a, b []float32) float64 {
	var dot float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
	}
	return dot
}

func TestDeterministicAndUnitNorm(t *testing.T) {
	e := localembedder.New()
	assert.Equal(t, localembedder.Dim, e.Dimensions())

	v1, err := e.Embed(context.Background(), []string{"Kasoa Polyclinic"}, ports.EmbedDocument)
	require.NoError(t, err)
	v2, err := e.Embed(context.Background(), []string{"Kasoa Polyclinic"}, ports.EmbedQuery)
	require.NoError(t, err)
	require.Len(t, v1, 1)
	assert.Len(t, v1[0], localembedder.Dim)
	assert.Equal(t, v1[0], v2[0], "deterministic regardless of kind")
	assert.InDelta(t, 1.0, math.Sqrt(cosine(v1[0], v1[0])), 1e-5, "unit norm")
}

func TestLexicalSimilarityRanksSharedTokensHigher(t *testing.T) {
	e := localembedder.New()
	// Mirrors the real write path: facility content is name+region+town+type+manager.
	vecs, err := e.Embed(context.Background(), []string{
		"Kasoa Polyclinic Central Kasoa High-volume OPD Ama Owusu", // 0: the facility
		"how is Kasoa doing today",                                 // 1: a query naming Kasoa
		"Tamale North Clinic Northern Tamale Fuseini Abdulai",      // 2: an unrelated facility
	}, ports.EmbedDocument)
	require.NoError(t, err)

	simRelated := cosine(vecs[1], vecs[0])
	simUnrelated := cosine(vecs[1], vecs[2])
	assert.Greater(t, simRelated, simUnrelated, "query should be closer to the facility it names")
}

func TestEmptyTextIsZeroVector(t *testing.T) {
	e := localembedder.New()
	v, err := e.Embed(context.Background(), []string{"!!! ???"}, ports.EmbedDocument)
	require.NoError(t, err)
	assert.Equal(t, math.Float64bits(0), math.Float64bits(cosine(v[0], v[0])), "no tokens -> zero vector")
}
