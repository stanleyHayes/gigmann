// Package localembedder is a deterministic, dependency-free ports.Embedder used
// when no cloud embedding provider is configured. It feature-hashes a bag of
// lowercased word tokens into a fixed-dimension unit vector, so texts that share
// words land closer in cosine space (lexical, not semantic, similarity) — enough
// to demonstrate grounded NL retrieval offline.
package localembedder

import (
	"context"
	"encoding/binary"
	"hash/fnv"
	"math"
	"strings"
	"unicode"

	"github.com/xcreativs/gigmann/internal/ports"
)

// Dim is the embedding dimension; it must match the vector(Dim) DB column.
const Dim = 1024

// Embedder is the deterministic local ports.Embedder.
type Embedder struct{}

// New creates a local embedder.
func New() *Embedder { return &Embedder{} }

var _ ports.Embedder = (*Embedder)(nil)

// Dimensions returns the fixed embedding dimension.
func (e *Embedder) Dimensions() int { return Dim }

// Embed returns one unit vector per input text. The kind is ignored (symmetric).
func (e *Embedder) Embed(_ context.Context, texts []string, _ ports.EmbedKind) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i, t := range texts {
		out[i] = embed(t)
	}
	return out, nil
}

func embed(text string) []float32 {
	v := make([]float32, Dim)
	for _, tok := range tokenize(text) {
		idx := hashWith(tok, 0) % Dim
		sign := float32(1)
		if hashWith(tok, 0x9e3779b9)&1 == 1 {
			sign = -1
		}
		v[idx] += sign
	}
	normalize(v)
	return v
}

func tokenize(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func hashWith(tok string, seed uint32) uint32 {
	h := fnv.New32a()
	if seed != 0 {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], seed) // same low-to-high bytes; avoids gosec G115
		_, _ = h.Write(b[:])
	}
	_, _ = h.Write([]byte(tok))
	return h.Sum32()
}

func normalize(v []float32) {
	var sum float64
	for _, x := range v {
		sum += float64(x) * float64(x)
	}
	if sum == 0 {
		return
	}
	inv := float32(1 / math.Sqrt(sum))
	for i := range v {
		v[i] *= inv
	}
}
